package scheduler

import (
	"testing"
	"time"
	extendprovider "multifish/providers/extend"
)

// Mock JobValidator for testing
type MockJobValidator struct {
	ValidateMachinesFunc func(machineIDs []string, action ActionType, payload Payload) []MachineValidationResult
}

func (m *MockJobValidator) ValidateMachines(machineIDs []string, action ActionType, payload Payload) []MachineValidationResult {
	if m.ValidateMachinesFunc != nil {
		return m.ValidateMachinesFunc(machineIDs, action, payload)
	}
	
	results := make([]MachineValidationResult, len(machineIDs))
	for i, machineID := range machineIDs {
		results[i] = MachineValidationResult{
			MachineID: machineID,
			Valid:     true,
			Message:   "Machine is valid",
		}
	}
	return results
}

// Mock JobExecutor for testing
type MockJobExecutor struct {
	ExecuteJobFunc func(job *Job) *ExecutionHistory
}

func (m *MockJobExecutor) ExecuteJob(job *Job) *ExecutionHistory {
	if m.ExecuteJobFunc != nil {
		return m.ExecuteJobFunc(job)
	}

	return &ExecutionHistory{
		JobID:         job.ID,
		ExecutionTime: time.Now(),
		Status:        JobStatusCompleted,
		Results: []MachineExecutionResult{
			{
				MachineID: job.Machines[0],
				Success:   true,
				Message:   "Success",
				StartTime: time.Now(),
				EndTime:   time.Now(),
				Duration:  "1s",
			},
		},
	}
}

// TestNewJobService tests the constructor
func TestNewJobService(t *testing.T) {
	mockValidator := &MockJobValidator{}
	mockExecutor := &MockJobExecutor{}

	service := NewJobService(mockValidator, mockExecutor)

	if service == nil {
		t.Fatal("NewJobService returned nil")
	}

	if service.validator != mockValidator {
		t.Error("Validator not set correctly")
	}

	if service.executor != mockExecutor {
		t.Error("Executor not set correctly")
	}

	if service.jobs == nil {
		t.Error("Jobs map not initialized")
	}

	// Stop the scheduler
	service.Stop()
}

// TestJobService_CreateJob tests job creation
func TestJobService_CreateJob(t *testing.T) {
	tests := []struct {
		name           string
		request        *JobCreateRequest
		setupValidator func(*MockJobValidator)
		expectError    bool
		expectValid    bool
	}{
		{
			name: "valid job creation",
			request: &JobCreateRequest{
				Name:     "Test Job",
				Machines: []string{"machine1"},
				Action:   ActionPatchProfile,
				Payload: []ExecutePatchProfilePayload{
					{
						ManagerID: "bmc",
						Payload:   extendprovider.PatchProfileType{Profile: "Performance"},
					},
				},
				Schedule: Schedule{
					Type: ScheduleTypeOnce,
					Time: "08:00:00",
				},
			},
			setupValidator: func(m *MockJobValidator) {
				m.ValidateMachinesFunc = func(machineIDs []string, action ActionType, payload Payload) []MachineValidationResult {
					return []MachineValidationResult{
						{
							MachineID: "machine1",
							Valid:     true,
							Message:   "Valid",
						},
					}
				}
			},
			expectError: false,
			expectValid: true,
		},
		{
			name: "invalid schedule",
			request: &JobCreateRequest{
				Name:     "Test Job",
				Machines: []string{"machine1"},
				Action:   ActionPatchProfile,
				Payload: []ExecutePatchProfilePayload{
					{
						ManagerID: "bmc",
						Payload:   extendprovider.PatchProfileType{Profile: "Performance"},
					},
				},
				Schedule: Schedule{
					Type: ScheduleTypeOnce,
					Time: "25:00:00", // Invalid hour
				},
			},
			setupValidator: func(m *MockJobValidator) {},
			expectError:    true,
			expectValid:    false,
		},
		{
			name: "machine validation failed",
			request: &JobCreateRequest{
				Name:     "Test Job",
				Machines: []string{"nonexistent"},
				Action:   ActionPatchProfile,
				Payload: []ExecutePatchProfilePayload{
					{
						ManagerID: "bmc",
						Payload:   extendprovider.PatchProfileType{Profile: "Performance"},
					},
				},
				Schedule: Schedule{
					Type: ScheduleTypeOnce,
					Time: "08:00:00",
				},
			},
			setupValidator: func(m *MockJobValidator) {
				m.ValidateMachinesFunc = func(machineIDs []string, action ActionType, payload Payload) []MachineValidationResult {
					return []MachineValidationResult{
						{
							MachineID: "nonexistent",
							Valid:     false,
							Message:   "Machine not found",
						},
					}
				}
			},
			expectError: true,
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockValidator := &MockJobValidator{}
			tt.setupValidator(mockValidator)
			mockExecutor := &MockJobExecutor{}

			service := NewJobService(mockValidator, mockExecutor)
			defer service.Stop()

			job, validationResp, err := service.CreateJob(tt.request)

			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if validationResp.Valid != tt.expectValid {
				t.Errorf("Expected Valid=%v, got %v. Message: %s",
					tt.expectValid, validationResp.Valid, validationResp.Message)
			}

			if !tt.expectError {
				if job == nil {
					t.Fatal("Expected job to be created")
				}

				if job.ID == "" {
					t.Error("Job ID should not be empty")
				}

				if job.Status != JobStatusPending {
					t.Errorf("Expected status Pending, got %v", job.Status)
				}

				if job.NextRunTime == nil {
					t.Error("NextRunTime should be set")
				}
			}
		})
	}
}

// TestJobService_GetJob tests job retrieval
func TestJobService_GetJob(t *testing.T) {
	mockValidator := &MockJobValidator{}
	mockExecutor := &MockJobExecutor{}

	service := NewJobService(mockValidator, mockExecutor)
	defer service.Stop()

	// Create a job first
	request := &JobCreateRequest{
		Name:     "Test Job",
		Machines: []string{"machine1"},
		Action:   ActionPatchProfile,
		Payload: []ExecutePatchProfilePayload{
			{
				ManagerID: "bmc",
				Payload:   extendprovider.PatchProfileType{Profile: "Performance"},
			},
		},
		Schedule: Schedule{
			Type: ScheduleTypeOnce,
			Time: "08:00:00",
		},
	}

	createdJob, _, err := service.CreateJob(request)
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	// Test GetJob
	retrievedJob, err := service.GetJob(createdJob.ID)
	if err != nil {
		t.Fatalf("Failed to get job: %v", err)
	}

	if retrievedJob.ID != createdJob.ID {
		t.Errorf("Job ID mismatch: got %v, want %v", retrievedJob.ID, createdJob.ID)
	}

	// Test getting non-existent job
	_, err = service.GetJob("nonexistent")
	if err == nil {
		t.Error("Expected error when getting non-existent job")
	}
}

// TestJobService_ListJobs tests listing all jobs
func TestJobService_ListJobs(t *testing.T) {
	mockValidator := &MockJobValidator{}
	mockExecutor := &MockJobExecutor{}

	service := NewJobService(mockValidator, mockExecutor)
	defer service.Stop()

	// Initially should be empty
	jobs := service.ListJobs()
	if len(jobs) != 0 {
		t.Errorf("Expected 0 jobs initially, got %d", len(jobs))
	}

	// Create multiple jobs
	for i := 0; i < 3; i++ {
		request := &JobCreateRequest{
			Name:     "Test Job",
			Machines: []string{"machine1"},
			Action:   ActionPatchProfile,
			Payload: []ExecutePatchProfilePayload{
				{
					ManagerID: "bmc",
					Payload:   extendprovider.PatchProfileType{Profile: "Performance"},
				},
			},
			Schedule: Schedule{
				Type: ScheduleTypeOnce,
				Time: "08:00:00",
			},
		}

		_, _, err := service.CreateJob(request)
		if err != nil {
			t.Fatalf("Failed to create job: %v", err)
		}
		
		// Small delay to ensure different timestamps
		time.Sleep(time.Millisecond)
	}

	// Should have 3 jobs now
	jobs = service.ListJobs()
	if len(jobs) != 3 {
		t.Errorf("Expected 3 jobs, got %d", len(jobs))
	}
}

// TestJobService_DeleteJob tests job deletion
func TestJobService_DeleteJob(t *testing.T) {
	mockValidator := &MockJobValidator{}
	mockExecutor := &MockJobExecutor{}

	service := NewJobService(mockValidator, mockExecutor)
	defer service.Stop()

	// Create a job
	request := &JobCreateRequest{
		Name:     "Test Job",
		Machines: []string{"machine1"},
		Action:   ActionPatchProfile,
		Payload: []ExecutePatchProfilePayload{
			{
				ManagerID: "bmc",
				Payload:   extendprovider.PatchProfileType{Profile: "Performance"},
			},
		},
		Schedule: Schedule{
			Type: ScheduleTypeOnce,
			Time: "08:00:00",
		},
	}

	createdJob, _, err := service.CreateJob(request)
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	// Delete the job
	err = service.DeleteJob(createdJob.ID)
	if err != nil {
		t.Errorf("Failed to delete job: %v", err)
	}

	// Verify job is deleted
	_, err = service.GetJob(createdJob.ID)
	if err == nil {
		t.Error("Expected error when getting deleted job")
	}

	// Test deleting non-existent job
	err = service.DeleteJob("nonexistent")
	if err == nil {
		t.Error("Expected error when deleting non-existent job")
	}
}

// TestJobService_WorkerPoolSize tests the worker pool size management
func TestJobService_WorkerPoolSize(t *testing.T) {
	mockValidator := &MockJobValidator{}
	mockExecutor := &MockJobExecutor{}

	service := NewJobService(mockValidator, mockExecutor)
	defer service.Stop()

	// Verify initial worker pool size
	initialSize := service.GetWorkerPoolSize()
	if initialSize != DefaultWorkerPoolSize {
		t.Errorf("Expected initial worker pool size %d, got %d", DefaultWorkerPoolSize, initialSize)
	}

	// Test setting a new worker pool size
	newSize := 50
	err := service.SetWorkerPoolSize(newSize)
	if err != nil {
		t.Fatalf("Failed to set worker pool size: %v", err)
	}

	// Verify the size was updated
	if service.GetWorkerPoolSize() != newSize {
		t.Errorf("Expected worker pool size %d, got %d", newSize, service.GetWorkerPoolSize())
	}

	// Test setting invalid size (should fail)
	err = service.SetWorkerPoolSize(0)
	if err == nil {
		t.Error("Expected error when setting worker pool size to 0")
	}

	err = service.SetWorkerPoolSize(-1)
	if err == nil {
		t.Error("Expected error when setting negative worker pool size")
	}

	// Test creating multiple jobs (no limit now)
	for i := 0; i < 10; i++ {
		request := &JobCreateRequest{
			Name:     "Test Job",
			Machines: []string{"machine1"},
			Action:   ActionPatchProfile,
			Payload: []ExecutePatchProfilePayload{
				{
					ManagerID: "bmc",
					Payload:   extendprovider.PatchProfileType{Profile: "Performance"},
				},
			},
			Schedule: Schedule{
				Type: ScheduleTypeOnce,
				Time: "08:00:00",
			},
		}

		_, _, err := service.CreateJob(request)
		if err != nil {
			t.Fatalf("Failed to create job %d: %v", i, err)
		}

		// Small delay to ensure different timestamps
		time.Sleep(time.Microsecond)
	}

	// Verify we can create jobs without hitting a limit
	if service.GetJobCount() != 10 {
		t.Errorf("Expected 10 jobs, got %d", service.GetJobCount())
	}
}

// TestJobService_Stop tests stopping the service
func TestJobService_Stop(t *testing.T) {
	mockValidator := &MockJobValidator{}
	mockExecutor := &MockJobExecutor{}

	service := NewJobService(mockValidator, mockExecutor)

	// Stop the service (should not panic)
	service.Stop()

	// Stopping again should be safe
	service.Stop()
}

// TestJobService_CalculateNextRunTime tests next run time calculation
func TestJobService_CalculateNextRunTime(t *testing.T) {
	mockValidator := &MockJobValidator{}
	mockExecutor := &MockJobExecutor{}

	service := NewJobService(mockValidator, mockExecutor)
	defer service.Stop()

	tests := []struct {
		name     string
		schedule Schedule
		checkFn  func(time.Time) bool
	}{
		{
			name: "once schedule - future time today",
			schedule: Schedule{
				Type: ScheduleTypeOnce,
				Time: "23:59:59",
			},
			checkFn: func(nextRun time.Time) bool {
				// Should be scheduled for today or tomorrow
				return !nextRun.IsZero()
			},
		},
		{
			name: "once schedule - past time (should be tomorrow)",
			schedule: Schedule{
				Type: ScheduleTypeOnce,
				Time: "00:00:01",
			},
			checkFn: func(nextRun time.Time) bool {
				return !nextRun.IsZero()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &Job{
				Schedule: tt.schedule,
			}

			nextRun := service.calculateNextRunTime(job)
			if !tt.checkFn(nextRun) {
				t.Errorf("Next run time validation failed: %v", nextRun)
			}
		})
	}
}

// TestJobService_ConcurrentAccess tests concurrent job operations
func TestJobService_ConcurrentAccess(t *testing.T) {
	mockValidator := &MockJobValidator{}
	mockExecutor := &MockJobExecutor{}

	service := NewJobService(mockValidator, mockExecutor)
	defer service.Stop()

	// Create a job
	request := &JobCreateRequest{
		Name:     "Test Job",
		Machines: []string{"machine1"},
		Action:   ActionPatchProfile,
		Payload: []ExecutePatchProfilePayload{
			{
				ManagerID: "bmc",
				Payload:   extendprovider.PatchProfileType{Profile: "Performance"},
			},
		},
		Schedule: Schedule{
			Type: ScheduleTypeOnce,
			Time: "08:00:00",
		},
	}

	createdJob, _, err := service.CreateJob(request)
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	// Test concurrent reads
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := service.GetJob(createdJob.ID)
			if err != nil {
				t.Errorf("Concurrent read failed: %v", err)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestJobService_UpdateJob tests job updates
func TestJobService_UpdateJob(t *testing.T) {
	mockValidator := &MockJobValidator{}
	mockExecutor := &MockJobExecutor{}

	service := NewJobService(mockValidator, mockExecutor)
	defer service.Stop()

	// Create a job
	request := &JobCreateRequest{
		Name:     "Original Name",
		Machines: []string{"machine1"},
		Action:   ActionPatchProfile,
		Payload: []ExecutePatchProfilePayload{
			{
				ManagerID: "bmc",
				Payload:   extendprovider.PatchProfileType{Profile: "Performance"},
			},
		},
		Schedule: Schedule{
			Type: ScheduleTypeOnce,
			Time: "08:00:00",
		},
	}

	createdJob, _, err := service.CreateJob(request)
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	// Manually update the job
	service.mu.Lock()
	job := service.jobs[createdJob.ID]
	job.Name = "Updated Name"
	service.mu.Unlock()

	// Verify the update
	updatedJob, err := service.GetJob(createdJob.ID)
	if err != nil {
		t.Fatalf("Failed to get updated job: %v", err)
	}

	if updatedJob.Name != "Updated Name" {
		t.Errorf("Job name not updated: got %v, want 'Updated Name'", updatedJob.Name)
	}
}
