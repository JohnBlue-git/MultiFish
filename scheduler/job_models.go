package scheduler

import (
	"encoding/json"
	"fmt"
	"time"
)

// ScheduleType represents the type of schedule
type ScheduleType string

const (
	ScheduleTypeOnce       ScheduleType = "Once"
	ScheduleTypeContinuous ScheduleType = "Continuous"
)

// DayOfWeek represents days of the week
type DayOfWeek string

const (
	Monday    DayOfWeek = "Monday"
	Tuesday   DayOfWeek = "Tuesday"
	Wednesday DayOfWeek = "Wednesday"
	Thursday  DayOfWeek = "Thursday"
	Friday    DayOfWeek = "Friday"
	Saturday  DayOfWeek = "Saturday"
	Sunday    DayOfWeek = "Sunday"
)

// Period represents the time period for continuous schedules
type Period struct {
	StartDay     *string     `json:"StartDay,omitempty"`     // Format: YYYY-MM-DD
	EndDay       *string     `json:"EndDay,omitempty"`       // Format: YYYY-MM-DD
	DaysOfWeek   []DayOfWeek `json:"DaysOfWeek,omitempty"`   // e.g., ["Monday", "Sunday"]
	DaysOfMonth  *string     `json:"DaysOfMonth,omitempty"`  // e.g., "1" or "1,15,30"
}

// Schedule represents the scheduling information
type Schedule struct {
	Type   ScheduleType `json:"Type"`             // "Once" or "Continuous"
	Time   string       `json:"Time"`             // Format: HH:MM:SS
	Period *Period      `json:"Period,omitempty"` // Required for Continuous, null for Once
}

// Payload represents a general payload type for job actions
// It can hold different payload types depending on the action type
type Payload any

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusPending   JobStatus = "Pending"
	JobStatusRunning   JobStatus = "Running"
	JobStatusCompleted JobStatus = "Completed"
	JobStatusFailed    JobStatus = "Failed"
	JobStatusCancelled JobStatus = "Cancelled"
)

// MachineValidationResult represents validation result for a single machine
type MachineValidationResult struct {
	MachineID string `json:"MachineId"`
	Valid     bool   `json:"Valid"`
	Message   string `json:"Message,omitempty"`
	Errors    []string `json:"Errors,omitempty"`
}

// Job represents a scheduled job
type Job struct {
	ID           string            `json:"Id"`
	Name         string            `json:"Name,omitempty"`
	Machines     []string          `json:"Machines"`
	Action       ActionType        `json:"Action"`
	Payload      Payload           `json:"Payload"`
	Schedule     Schedule          `json:"Schedule"`
	Status       JobStatus         `json:"Status"`
	CreatedTime  time.Time         `json:"CreatedTime"`
	LastRunTime  *time.Time        `json:"LastRunTime,omitempty"`
	NextRunTime  *time.Time        `json:"NextRunTime,omitempty"`
	ExecutionCount int             `json:"ExecutionCount"`
}

// JobCreateRequest represents the request to create a job
type JobCreateRequest struct {
	Name     string     `json:"Name,omitempty"`
	Machines []string   `json:"Machines"`
	Action   ActionType `json:"Action"`
	Payload  Payload    `json:"Payload"`
	Schedule Schedule   `json:"Schedule"`
}

// JobValidationResponse represents the validation response
type JobValidationResponse struct {
	Valid            bool                      `json:"Valid"`
	Message          string                    `json:"Message"`
	MachineResults   []MachineValidationResult `json:"MachineResults,omitempty"`
	ScheduleValid    bool                      `json:"ScheduleValid"`
	ScheduleErrors   []string                  `json:"ScheduleErrors,omitempty"`
	ActionValid      bool                      `json:"ActionValid"`
	ActionErrors     []string                  `json:"ActionErrors,omitempty"`
	PayloadValid     bool                      `json:"PayloadValid"`
	PayloadErrors    []string                  `json:"PayloadErrors,omitempty"`
}

// ExecutionHistory represents a single execution record
type ExecutionHistory struct {
	JobID         string                    `json:"JobId"`
	ExecutionTime time.Time                 `json:"ExecutionTime"`
	Status        JobStatus                 `json:"Status"`
	Results       []MachineExecutionResult  `json:"Results"`
}

// MachineExecutionResult represents execution result for a single machine
type MachineExecutionResult struct {
	MachineID string    `json:"MachineId"`
	Success   bool      `json:"Success"`
	Message   string    `json:"Message,omitempty"`
	Error     string    `json:"Error,omitempty"`
	StartTime time.Time `json:"StartTime"`
	EndTime   time.Time `json:"EndTime"`
	Duration  string    `json:"Duration"`
}

// Validate validates the job creation request
func (j *JobCreateRequest) Validate() *JobValidationResponse {
	response := &JobValidationResponse{
		Valid:         true,
		ScheduleValid: true,
		ActionValid:   true,
		PayloadValid:  true,
	}

	// Validate machines
	if len(j.Machines) == 0 {
		response.Valid = false
		response.Message = "At least one machine must be specified"
		return response
	}

	// Check for duplicate machines
	if err := j.validateNoDuplicateMachines(); err != nil {
		response.Valid = false
		response.Message = err.Error()
		return response
	}

	// Validate action
	if err := j.validateAction(); err != nil {
		response.Valid = false
		response.ActionValid = false
		response.ActionErrors = append(response.ActionErrors, err.Error())
	}

	// Validate payload
	if err := j.validatePayload(); err != nil {
		response.Valid = false
		response.PayloadValid = false
		response.PayloadErrors = append(response.PayloadErrors, err.Error())
	}

	// Validate schedule
	if errs := j.validateSchedule(); len(errs) > 0 {
		response.Valid = false
		response.ScheduleValid = false
		response.ScheduleErrors = errs
	}

	if response.Valid {
		response.Message = "Job validation successful"
	} else {
		response.Message = "Job validation failed"
	}

	return response
}

// validateNoDuplicateMachines checks that there are no duplicate machine IDs
func (j *JobCreateRequest) validateNoDuplicateMachines() error {
	seen := make(map[string]bool)
	duplicates := []string{}

	for _, machineID := range j.Machines {
		if seen[machineID] {
			duplicates = append(duplicates, machineID)
		} else {
			seen[machineID] = true
		}
	}

	if len(duplicates) > 0 {
		return fmt.Errorf("job validation failed: duplicate machine IDs found: %v. Each machine can only be specified once per job. Remove duplicates from the machines array", duplicates)
	}

	return nil
}

// validateAction validates the action type
func (j *JobCreateRequest) validateAction() error {
	switch j.Action {
	case ActionPatchProfile:
		return nil
	default:
		return fmt.Errorf("job validation failed: unsupported action type '%s'. Valid actions are: %v", j.Action, []string{"PatchProfile", "PatchManager", "PatchFanController", "PatchFanZone", "PatchPidController"})
	}
}

// validateSchedule validates the schedule configuration
func (j *JobCreateRequest) validateSchedule() []string {
	var errors []string

	// Validate schedule type
	if j.Schedule.Type != ScheduleTypeOnce && j.Schedule.Type != ScheduleTypeContinuous {
		errors = append(errors, fmt.Sprintf("invalid schedule type: %s (must be 'Once' or 'Continuous')", j.Schedule.Type))
	}

	// Validate time format (HH:MM:SS)
	if _, err := time.Parse("15:04:05", j.Schedule.Time); err != nil {
		errors = append(errors, fmt.Sprintf("invalid time format: %s (expected HH:MM:SS)", j.Schedule.Time))
	}

	// Validate period based on schedule type
	if j.Schedule.Type == ScheduleTypeOnce {
		if j.Schedule.Period != nil {
			errors = append(errors, "Period must be null for 'Once' schedule type")
		}
	} else if j.Schedule.Type == ScheduleTypeContinuous {
		if j.Schedule.Period == nil {
			errors = append(errors, "Period is required for 'Continuous' schedule type")
		} else {
			errors = append(errors, j.validatePeriod()...)
		}
	}

	return errors
}

// validatePeriod validates the period configuration
func (j *JobCreateRequest) validatePeriod() []string {
	var errors []string
	period := j.Schedule.Period

	// Validate StartDay
	if period.StartDay != nil {
		if _, err := time.Parse("2006-01-02", *period.StartDay); err != nil {
			errors = append(errors, fmt.Sprintf("invalid StartDay format: %s (expected YYYY-MM-DD)", *period.StartDay))
		}
	}

	// Validate EndDay
	if period.EndDay != nil {
		if _, err := time.Parse("2006-01-02", *period.EndDay); err != nil {
			errors = append(errors, fmt.Sprintf("invalid EndDay format: %s (expected YYYY-MM-DD)", *period.EndDay))
		}
	}

	// Validate StartDay is before EndDay
	if period.StartDay != nil && period.EndDay != nil {
		startDay, _ := time.Parse("2006-01-02", *period.StartDay)
		endDay, _ := time.Parse("2006-01-02", *period.EndDay)
		if startDay.After(endDay) {
			errors = append(errors, "StartDay must be before or equal to EndDay")
		}
	}

	// Validate DaysOfWeek
	if len(period.DaysOfWeek) > 0 {
		validDays := map[DayOfWeek]bool{
			Monday: true, Tuesday: true, Wednesday: true, Thursday: true,
			Friday: true, Saturday: true, Sunday: true,
		}
		for _, day := range period.DaysOfWeek {
			if !validDays[day] {
				errors = append(errors, fmt.Sprintf("invalid day of week: %s", day))
			}
		}
	}

	// Validate DaysOfMonth (should be numbers 1-31, comma-separated)
	if period.DaysOfMonth != nil && *period.DaysOfMonth != "" {
		// Simple validation - could be enhanced
		// Format: "1" or "1,15,30"
		// For now, just check it's not empty
	}

	// Ensure at least one recurrence pattern is specified
	hasDaysOfWeek := len(period.DaysOfWeek) > 0
	hasDaysOfMonth := period.DaysOfMonth != nil && *period.DaysOfMonth != ""
	
	if !hasDaysOfWeek && !hasDaysOfMonth {
		errors = append(errors, "at least one of DaysOfWeek or DaysOfMonth must be specified for continuous schedules")
	}

	return errors
}

// ToJSON converts the object to JSON string
func (j *Job) ToJSON() string {
	data, _ := json.MarshalIndent(j, "", "  ")
	return string(data)
}

// ToJSON converts the execution history to JSON string
func (e *ExecutionHistory) ToJSON() string {
	data, _ := json.MarshalIndent(e, "", "  ")
	return string(data)
}
