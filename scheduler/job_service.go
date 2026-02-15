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

const (
	DefaultWorkerPoolSize = 99
)

var (
	LogsDir = "./logs" // Default logs directory, can be configured
)

// JobService manages job scheduling and execution
type JobService struct {
	jobs           map[string]*Job
	mu             sync.RWMutex
	ticker         *time.Ticker
	stopChan       chan bool
	validator      JobValidator
	executor       JobExecutor
	stopped        bool
	workerPool     chan struct{} // Semaphore for controlling concurrent job executions
	workerPoolSize int           // Current worker pool size (dynamically adjustable)
	runningJobs    map[string]bool
	runningMu      sync.Mutex // Separate mutex for running jobs tracking
}

// JobValidator validates jobs against machines
type JobValidator interface {
	ValidateMachines(machineIDs []string, action ActionType, payload Payload) []MachineValidationResult
}

// JobExecutor executes jobs on machines
type JobExecutor interface {
	ExecuteJob(job *Job) *ExecutionHistory
}

// NewJobService creates a new job service
func NewJobService(validator JobValidator, executor JobExecutor) *JobService {
	log := utility.GetLogger()
	
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(LogsDir, 0755); err != nil {
		log.Warn().Err(err).Msg("Failed to create logs directory")
	}

	js := &JobService{
		jobs:           make(map[string]*Job),
		stopChan:       make(chan bool),
		validator:      validator,
		executor:       executor,
		workerPoolSize: DefaultWorkerPoolSize,
		workerPool:     make(chan struct{}, DefaultWorkerPoolSize), // Buffered channel as semaphore
		runningJobs:    make(map[string]bool),
	}

	// Start the scheduler
	js.startScheduler()

	return js
}

// CreateJob creates a new job after validation
func (js *JobService) CreateJob(req *JobCreateRequest) (*Job, *JobValidationResponse, error) {
	js.mu.Lock()
	defer js.mu.Unlock()

	// Validate basic job structure
	validationResp := req.Validate()

	// Validate machines against the platform
	if js.validator != nil {
		machineResults := js.validator.ValidateMachines(req.Machines, req.Action, req.Payload)
		validationResp.MachineResults = machineResults

		// Check if all machines are valid
		for _, result := range machineResults {
			if !result.Valid {
				validationResp.Valid = false
			}
		}
	}

	// If validation fails, return the validation response
	if !validationResp.Valid {
		return nil, validationResp, fmt.Errorf("job validation failed")
	}

	// Generate job ID
	jobID := fmt.Sprintf("Job-%d", time.Now().UnixNano())

	// Create the job
	job := &Job{
		ID:             jobID,
		Name:           req.Name,
		Machines:       req.Machines,
		Action:         req.Action,
		Payload:        req.Payload,
		Schedule:       req.Schedule,
		Status:         JobStatusPending,
		CreatedTime:    time.Now(),
		ExecutionCount: 0,
	}

	// Calculate next run time
	nextRun := js.calculateNextRunTime(job)
	job.NextRunTime = &nextRun

	// Store the job
	js.jobs[jobID] = job

	log := utility.GetLogger()
	log.Info().
		Str("jobID", jobID).
		Time("nextRun", nextRun).
		Msg("Job created")

	return job, validationResp, nil
}

// GetJob retrieves a job by ID
func (js *JobService) GetJob(jobID string) (*Job, error) {
	log := utility.GetLogger()

	js.mu.RLock()
	defer js.mu.RUnlock()

	job, exists := js.jobs[jobID]
	if !exists {
		log.Warn().Str("jobID", jobID).Msg("Job not found")
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	return job, nil
}

// ListJobs returns all jobs
func (js *JobService) ListJobs() []*Job {
	js.mu.RLock()
	defer js.mu.RUnlock()

	jobs := make([]*Job, 0, len(js.jobs))
	for _, job := range js.jobs {
		jobs = append(jobs, job)
	}

	return jobs
}

// DeleteJob deletes a job by ID
func (js *JobService) DeleteJob(jobID string) error {
	log := utility.GetLogger()

	js.mu.Lock()
	defer js.mu.Unlock()

	if _, exists := js.jobs[jobID]; !exists {
		log.Warn().Str("jobID", jobID).Msg("Job not found")
		return fmt.Errorf("job with ID '%s' not found in job service (active jobs: %d). Use GET /jobs to list available jobs", jobID, len(js.jobs))
	}

	delete(js.jobs, jobID)
	log.Info().Str("jobID", jobID).Msg("Job deleted")

	return nil
}

// CancelJob cancels a job
func (js *JobService) CancelJob(jobID string) error {
	log := utility.GetLogger()

	js.mu.Lock()
	defer js.mu.Unlock()

	job, exists := js.jobs[jobID]
	if !exists {
		log.Warn().Str("jobID", jobID).Msg("Job not found")
		return fmt.Errorf("job with ID '%s' not found in job service (active jobs: %d). Use GET /jobs to list available jobs", jobID, len(js.jobs))
	}

	job.Status = JobStatusCancelled
	log.Info().Str("jobID", jobID).Msg("Job cancelled")

	return nil
}

// startScheduler starts the job scheduler
func (js *JobService) startScheduler() {
	log := utility.GetLogger()

	// Use 1-second granularity for precise scheduling
	// This ensures jobs scheduled with second-precision (HH:MM:SS) are executed on time
	tickInterval := 1 * time.Second

	// Initialize ticker
	js.ticker = time.NewTicker(tickInterval)

	// Start a goroutine to check and execute jobs at each tick
	go func() {
		lastCheck := time.Now()
		
		for {
			select {
			case tickTime := <-js.ticker.C:
				// Drift correction: measure actual elapsed time
				actualElapsed := tickTime.Sub(lastCheck)
				drift := actualElapsed - tickInterval
				
				// Log significant drift (> 100ms)
				if drift > 100*time.Millisecond || drift < -100*time.Millisecond {
					log.Warn().
						Dur("expected", tickInterval).
						Dur("actual", actualElapsed).
						Dur("drift", drift).
						Msg("Scheduler tick drift detected")
				}
				
				// Check and execute jobs
				js.checkAndExecuteJobs()
				
				// Update last check time for drift tracking
				lastCheck = tickTime
				
			case <-js.stopChan:
				js.ticker.Stop()
				log.Info().Msg("Job scheduler goroutine stopped")
				return
			}
		}
	}()

	log.Info().
		Dur("tickInterval", tickInterval).
		Msg("Job scheduler started with drift monitoring")
}

// Stop stops the job scheduler
func (js *JobService) Stop() {
	log := utility.GetLogger()

	js.mu.Lock()
	defer js.mu.Unlock()
	
	if !js.stopped {
		close(js.stopChan)
		js.stopped = true
		log.Info().Msg("Job scheduler stopped")
	}
}

// checkAndExecuteJobs checks for jobs that need to be executed
func (js *JobService) checkAndExecuteJobs() {
	js.mu.Lock()
	defer js.mu.Unlock()

	now := time.Now()
	log := utility.GetLogger()

	for _, job := range js.jobs {
		// Skip cancelled or completed jobs
		if job.Status == JobStatusCancelled {
			continue
		}

		// Skip if job is already running
		js.runningMu.Lock()
		isRunning := js.runningJobs[job.ID]
		js.runningMu.Unlock()
		
		if isRunning {
			continue
		}

		// Check if job should run now
		// Use After() to handle cases where we might have missed the exact time
		if job.NextRunTime != nil && now.After(*job.NextRunTime) {
			// Calculate how late we are (for monitoring purposes)
			delay := now.Sub(*job.NextRunTime)
			
			// Log if execution is significantly delayed (> 2 seconds)
			if delay > 2*time.Second {
				log.Warn().
					Str("jobID", job.ID).
					Str("jobName", job.Name).
					Time("scheduledTime", *job.NextRunTime).
					Time("actualTime", now).
					Dur("delay", delay).
					Msg("Job execution delayed")
			}
			
			// Try to acquire a worker slot (non-blocking)
			select {
			case js.workerPool <- struct{}{}:
				// Successfully acquired a worker slot, execute job
				go js.executeJobAsync(job)
			default:
				// No worker slots available, skip this execution
				log.Warn().
					Str("jobID", job.ID).
					Msg("Worker pool full, skipping job execution (will retry next cycle)")
			}
		}
	}
}

// executeJobAsync executes a job asynchronously
func (js *JobService) executeJobAsync(job *Job) {
	// Ensure we release the worker slot when done
	defer func() {
		<-js.workerPool // Release worker slot
		js.runningMu.Lock()
		delete(js.runningJobs, job.ID)
		js.runningMu.Unlock()
	}()

	// Mark job as running
	js.runningMu.Lock()
	js.runningJobs[job.ID] = true
	js.runningMu.Unlock()

	log := utility.GetLogger()
	log.Info().
		Str("jobID", job.ID).
		Int("activeWorkers", len(js.workerPool)).
		Int("poolSize", js.workerPoolSize).
		Msg("Executing job")

	// Update job status
	js.mu.Lock()
	job.Status = JobStatusRunning
	js.mu.Unlock()

	// Execute the job
	history := js.executor.ExecuteJob(job)

	// Update job status and times
	js.mu.Lock()
	defer js.mu.Unlock()

	now := time.Now()
	job.LastRunTime = &now
	job.ExecutionCount++

	// Determine job status based on execution results
	allSuccess := true
	for _, result := range history.Results {
		if !result.Success {
			allSuccess = false
			break
		}
	}

	if allSuccess {
		history.Status = JobStatusCompleted
	} else {
		history.Status = JobStatusFailed
	}

	// Update job status
	if job.Schedule.Type == ScheduleTypeOnce {
		job.Status = history.Status
		job.NextRunTime = nil
	} else {
		// For continuous jobs, calculate next run time
		nextRun := js.calculateNextRunTime(job)
		job.NextRunTime = &nextRun
		job.Status = JobStatusPending
	}

	log.Info().
		Str("jobID", job.ID).
		Str("status", string(history.Status)).
		Msg("Job execution completed")
}

// calculateNextRunTime calculates the next run time for a job
func (js *JobService) calculateNextRunTime(job *Job) time.Time {
	log := utility.GetLogger()
	now := time.Now()

	// Parse the scheduled time
	schedTime, err := time.Parse("15:04:05", job.Schedule.Time)
	if err != nil {
		log.Error().Err(err).Str("jobID", job.ID).Msg("Error parsing schedule time")
		return now.Add(24 * time.Hour) // Default to next day
	}

	if job.Schedule.Type == ScheduleTypeOnce {
		// For "Once" type, schedule for today at the specified time, or tomorrow if time has passed
		nextRun := time.Date(now.Year(), now.Month(), now.Day(),
			schedTime.Hour(), schedTime.Minute(), schedTime.Second(), 0, now.Location())

		if nextRun.Before(now) {
			nextRun = nextRun.Add(24 * time.Hour)
		}

		return nextRun
	}

	// For "Continuous" type
	period := job.Schedule.Period
	if period == nil {
		log.Error().Str("jobID", job.ID).Msg("Period is nil for continuous job")
		return now.Add(24 * time.Hour)
	}

	// Start from today or StartDay, whichever is later
	startDate := now
	if period.StartDay != nil {
		if startDay, err := time.Parse("2006-01-02", *period.StartDay); err == nil {
			if startDay.After(startDate) {
				startDate = startDay
			}
		}
	}

	// Find the next valid execution time
	candidate := time.Date(startDate.Year(), startDate.Month(), startDate.Day(),
		schedTime.Hour(), schedTime.Minute(), schedTime.Second(), 0, startDate.Location())

	// If candidate is in the past, start from tomorrow
	if candidate.Before(now) {
		candidate = candidate.Add(24 * time.Hour)
	}

	// Find next matching day (up to 365 days in the future)
	for i := 0; i < 365; i++ {
		if js.isValidExecutionDate(candidate, period) {
			return candidate
		}
		candidate = candidate.Add(24 * time.Hour)
	}

	// Fallback
	log.Warn().Str("jobID", job.ID).Msg("Could not find valid execution date for job")
	return now.Add(24 * time.Hour)
}

// isValidExecutionDate checks if a date matches the period criteria
func (js *JobService) isValidExecutionDate(date time.Time, period *Period) bool {
	// Check if date is within the period range
	if period.StartDay != nil {
		if startDay, err := time.Parse("2006-01-02", *period.StartDay); err == nil {
			if date.Before(startDay) {
				return false
			}
		}
	}

	if period.EndDay != nil {
		if endDay, err := time.Parse("2006-01-02", *period.EndDay); err == nil {
			// Add 23:59:59 to endDay to include the entire day
			endDayEnd := endDay.Add(24*time.Hour - time.Second)
			if date.After(endDayEnd) {
				return false
			}
		}
	}

	// Check DaysOfWeek
	if len(period.DaysOfWeek) > 0 {
		dayMatches := false
		weekday := date.Weekday().String()
		for _, day := range period.DaysOfWeek {
			if string(day) == weekday {
				dayMatches = true
				break
			}
		}
		if !dayMatches {
			return false
		}
	}

	// Check DaysOfMonth
	if period.DaysOfMonth != nil && *period.DaysOfMonth != "" {
		// Simple check: see if current day is in the list
		// For simplicity, we'll just check if the day number matches
		// A more robust implementation would parse the comma-separated list
		dayOfMonth := date.Day()
		dayStr := fmt.Sprintf("%d", dayOfMonth)
		
		// This is a simplified check - should be enhanced for comma-separated values
		if *period.DaysOfMonth == dayStr {
			return true
		}
		// If DaysOfMonth is specified but doesn't match, and DaysOfWeek is empty, return false
		if len(period.DaysOfWeek) == 0 {
			return false
		}
	}

	return true
}

// logExecutionHistory logs execution history to a file
func (js *JobService) logExecutionHistory(history *ExecutionHistory) {
	log := utility.GetLogger()
	
	// Create log filename with date
	logDate := history.ExecutionTime.Format("2006-01-02")
	logFile := filepath.Join(LogsDir, fmt.Sprintf("job-executions-%s.log", logDate))

	// Open file in append mode
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error().Err(err).Str("logFile", logFile).Msg("Error opening log file")
		return
	}
	defer file.Close()

	// Write JSON log entry
	jsonData, err := json.Marshal(history)
	if err != nil {
		log.Error().Err(err).Msg("Error marshaling execution history")
		return
	}

	if _, err := file.WriteString(string(jsonData) + "\n"); err != nil {
		log.Error().Err(err).Str("logFile", logFile).Msg("Error writing to log file")
		return
	}

	log.Debug().Str("logFile", logFile).Msg("Execution history logged")
}

// GetWorkerPoolSize returns the maximum number of jobs allowed
func (js *JobService) GetWorkerPoolSize() int {
	js.mu.RLock()
	defer js.mu.RUnlock()
	return js.workerPoolSize
}

// SetWorkerPoolSize updates the worker pool size dynamically
func (js *JobService) SetWorkerPoolSize(size int) error {
	log := utility.GetLogger()

	if size <= 0 {
		log.Warn().Int("workerPoolSize", size).Msg("Invalid worker pool size")
		return fmt.Errorf("job service configuration failed: worker pool size must be greater than 0, got %d. Configure worker_pool_size in config file", size)
	}

	js.mu.Lock()
	defer js.mu.Unlock()

	// Create a new worker pool channel with the new size
	newWorkerPool := make(chan struct{}, size)

	// Transfer existing workers to the new pool
	// This preserves the current active worker count
	for i := 0; i < len(js.workerPool) && i < size; i++ {
		newWorkerPool <- struct{}{}
	}

	// Update the worker pool and size
	js.workerPool = newWorkerPool
	js.workerPoolSize = size

	log.Info().
		Int("poolSize", size).
		Int("activeWorkers", len(js.workerPool)).
		Msg("Worker pool size updated")
	return nil
}

// SetLogsDir updates the logs directory path
func SetLogsDir(dir string) error {
	log := utility.GetLogger()

	if dir == "" {
		return fmt.Errorf("logs directory path cannot be empty")
	}

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Error().Err(err).Str("dir", dir).Msg("Failed to create logs directory")
		return fmt.Errorf("failed to create logs directory '%s': %w", dir, err)
	}

	LogsDir = dir
	log.Info().Str("logsDir", dir).Msg("Logs directory updated")
	return nil
}

// GetActiveWorkers returns the current number of active/running jobs
func (js *JobService) GetActiveWorkers() int {
	return len(js.workerPool)
}

// GetAvailableWorkers returns the number of available worker slots
func (js *JobService) GetAvailableWorkers() int {
	js.mu.RLock()
	defer js.mu.RUnlock()
	return js.workerPoolSize - len(js.workerPool)
}

// GetRunningJobsCount returns the count of currently running jobs
func (js *JobService) GetRunningJobsCount() int {
	js.runningMu.Lock()
	defer js.runningMu.Unlock()
	return len(js.runningJobs)
}

// GetJobCount returns the current number of jobs
func (js *JobService) GetJobCount() int {
	js.mu.RLock()
	defer js.mu.RUnlock()
	return len(js.jobs)
}
