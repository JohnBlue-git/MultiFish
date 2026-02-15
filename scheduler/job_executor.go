package scheduler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"multifish/utility"
)

// JobPlatformManager interface for platform operations used by job executor
type JobPlatformManager interface {
	GetMachine(machineID string) (interface{}, error)
}

// ExecutionErrorLog represents the error log structure for JSON output
type ExecutionErrorLog struct {
	JobID       string    `json:"job_id"`
	MachineID   string    `json:"machine_id"`
	Action      string    `json:"action"`
	Timestamp   string    `json:"timestamp"`
	ErrorType   string    `json:"error_type"`
	ErrorDetail string    `json:"error_detail"`
	Payload     Payload   `json:"payload,omitempty"`
}

// ExecutionSuccessLog represents the success log structure for JSON output
type ExecutionSuccessLog struct {
	JobID       string    `json:"job_id"`
	MachineID   string    `json:"machine_id"`
	Action      string    `json:"action"`
	Timestamp   string    `json:"timestamp"`
	Status      string    `json:"status"`
	Message     string    `json:"message"`
	Duration    string    `json:"duration"`
	Payload     Payload   `json:"payload,omitempty"`
}

// writeErrorToJSON writes the execution error to a JSON file
func writeErrorToJSON(jobID string, machineID string, action ActionType, execErr error, payload Payload) error {
	log := utility.GetLogger()

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(LogsDir, 0755); err != nil {
		log.Error().Msgf("failed to create logs directory: %v", err)
		return fmt.Errorf("failed to create logs directory '%s': %w. Check filesystem permissions and available disk space", LogsDir, err)
	}

	// Generate timestamp
	timestamp := time.Now()
	timestampStr := timestamp.Format("20060102_150405")
	
	// Create filename: jobID_machine_action_timestamp.json
	filename := fmt.Sprintf("%s_%s_%s_%s_error.json", jobID, machineID, action, timestampStr)
	filepath := filepath.Join(LogsDir, filename)

	// Determine error type
	errorType := "ExecutionError"
	if execErr != nil {
		log.Error().Msgf("failed to execute action %s on machine %s: %v", action, machineID, execErr)
		errorType = fmt.Sprintf("%T", execErr)
	}

	// Create error log structure
	errorLog := ExecutionErrorLog{
		JobID:       jobID,
		MachineID:   machineID,
		Action:      string(action),
		Timestamp:   timestamp.Format(time.RFC3339),
		ErrorType:   errorType,
		ErrorDetail: execErr.Error(),
		Payload:     payload,
	}

	// Marshal to JSON with indentation
	jsonData, err := json.MarshalIndent(errorLog, "", "  ")
	if err != nil {
		log.Error().Msgf("failed to marshal error log to JSON: %v", err)
		return fmt.Errorf("failed to serialize execution error log (jobID: %s, machine: %s, action: %s) to JSON: %w", jobID, machineID, action, err)
	}

	// Write to file
	if err := os.WriteFile(filepath, jsonData, 0644); err != nil {
		log.Error().Msgf("failed to write error log to file: %v", err)
		return fmt.Errorf("failed to write execution error log to file '%s': %w. Check disk space and permissions", filepath, err)
	}

	return nil
}

// writeSuccessToJSON writes the execution success to a JSON file
func writeSuccessToJSON(jobID string, machineID string, action ActionType, duration string, payload Payload) error {
	log := utility.GetLogger()

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(LogsDir, 0755); err != nil {
		log.Error().Msgf("failed to create logs directory: %v", err)
		return fmt.Errorf("failed to create logs directory '%s': %w. Check filesystem permissions and available disk space", LogsDir, err)
	}

	// Generate timestamp
	timestamp := time.Now()
	timestampStr := timestamp.Format("20060102_150405")
	
	// Create filename: jobID_machine_action_timestamp_success.json
	filename := fmt.Sprintf("%s_%s_%s_%s_success.json", jobID, machineID, action, timestampStr)
	filepath := filepath.Join(LogsDir, filename)

	// Create success log structure
	successLog := ExecutionSuccessLog{
		JobID:     jobID,
		MachineID: machineID,
		Action:    string(action),
		Timestamp: timestamp.Format(time.RFC3339),
		Status:    "Success",
		Message:   fmt.Sprintf("Successfully executed %s", action),
		Duration:  duration,
		Payload:   payload,
	}

	// Marshal to JSON with indentation
	jsonData, err := json.MarshalIndent(successLog, "", "  ")
	if err != nil {
		log.Error().Msgf("failed to marshal success log to JSON: %v", err)
		return fmt.Errorf("failed to serialize execution success log (jobID: %s, machine: %s, action: %s) to JSON: %w", jobID, machineID, action, err)
	}

	// Write to file
	if err := os.WriteFile(filepath, jsonData, 0644); err != nil {
		log.Error().Msgf("failed to write success log to file: %v", err)
		return fmt.Errorf("failed to write execution success log to file '%s': %w. Check disk space and permissions", filepath, err)
	}

	return nil
}

// PlatformValidator validates jobs against platform machines
type PlatformValidator struct {
	platformMgr JobPlatformManager
}

// NewPlatformValidator creates a new platform validator
func NewPlatformValidator(platformMgr JobPlatformManager) *PlatformValidator {
	return &PlatformValidator{
		platformMgr: platformMgr,
	}
}

// ValidateMachines validates that all machines exist and support the action
func (pv *PlatformValidator) ValidateMachines(machineIDs []string, action ActionType, payload Payload) []MachineValidationResult {
	results := make([]MachineValidationResult, len(machineIDs))

	for i, machineID := range machineIDs {
		result := MachineValidationResult{
			MachineID: machineID,
			Valid:     true,
			Errors:    []string{},
		}

		// Check if machine exists
		machine, err := pv.platformMgr.GetMachine(machineID)
		if err != nil {
			result.Valid = false
			result.Message = "Machine not found"
			result.Errors = append(result.Errors, err.Error())
			results[i] = result
			continue
		}

		// Validate action support for this machine
		if err := pv.validateActionSupport(machine, action); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, err.Error())
		}

		// Validate payload for this machine
		if err := pv.validatePayloadSupport(machine, action, payload); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, err.Error())
		}

		if result.Valid {
			result.Message = "Machine is valid for the requested action"
		} else {
			result.Message = "Machine validation failed"
		}

		results[i] = result
	}

	return results
}

// PlatformExecutor executes jobs on platform machines
type PlatformExecutor struct {
	platformMgr JobPlatformManager
	actionExecutor ActionExecutor
}

// NewPlatformExecutor creates a new platform executor
func NewPlatformExecutor(platformMgr JobPlatformManager, actionExecutor ActionExecutor) *PlatformExecutor {
	return &PlatformExecutor{
		platformMgr:    platformMgr,
		actionExecutor: actionExecutor,
	}
}

// ExecuteJob executes a job on all specified machines
func (pe *PlatformExecutor) ExecuteJob(job *Job) *ExecutionHistory {
	history := &ExecutionHistory{
		JobID:         job.ID,
		ExecutionTime: time.Now(),
		Status:        JobStatusRunning,
		Results:       make([]MachineExecutionResult, len(job.Machines)),
	}

	// Execute on all machines in parallel
	var wg sync.WaitGroup
	for i, machineID := range job.Machines {
		wg.Add(1)
		go func(idx int, mID string) {
			defer wg.Done()
			history.Results[idx] = pe.executeMachine(job.ID, mID, job.Action, job.Payload)
		}(i, machineID)
	}

	wg.Wait()

	return history
}

// executeMachine executes the action on a single machine
func (pe *PlatformExecutor) executeMachine(jobID string, machineID string, action ActionType, payload Payload) MachineExecutionResult {
	log := utility.GetLogger()

	result := MachineExecutionResult{
		MachineID: machineID,
		StartTime: time.Now(),
	}

	// Get machine
	machine, err := pe.platformMgr.GetMachine(machineID)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Failed to get machine: %v", err)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime).String()
		
		log.Error().
			Err(err).
			Str("jobID", jobID).
			Str("machineID", machineID).
			Str("action", string(action)).
			Msg("Failed to get machine")
		
		// Write error to JSON file
		if logErr := writeErrorToJSON(jobID, machineID, action, err, payload); logErr != nil {
			log.Warn().
				Err(logErr).
				Str("machineID", machineID).
				Msg("Failed to write error log")
		}
		
		return result
	}

	// Execute action
	execErr := ExecuteAction(pe.actionExecutor, action, machine, payload)

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).String()

	if execErr != nil {
		result.Success = false
		result.Error = execErr.Error()
		result.Message = fmt.Sprintf("Failed to execute %s", action)
		
		log.Error().
			Err(execErr).
			Str("jobID", jobID).
			Str("machineID", machineID).
			Str("action", string(action)).
			Msg("Failed to execute action")
		
		// Write error to JSON file
		if logErr := writeErrorToJSON(jobID, machineID, action, execErr, payload); logErr != nil {
			log.Warn().
				Err(logErr).
				Str("machineID", machineID).
				Msg("Failed to write error log")
		}
	} else {
		result.Success = true
		result.Message = fmt.Sprintf("Successfully executed %s", action)
		
		log.Info().
			Str("jobID", jobID).
			Str("machineID", machineID).
			Str("action", string(action)).
			Str("duration", result.Duration).
			Msg("Successfully executed action")
		
		// Write success to JSON file
		if logErr := writeSuccessToJSON(jobID, machineID, action, result.Duration, payload); logErr != nil {
			log.Warn().
				Err(logErr).
				Str("machineID", machineID).
				Msg("Failed to write success log")
		}
	}

	return result
}
