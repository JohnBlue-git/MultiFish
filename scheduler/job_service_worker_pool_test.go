package scheduler

import (
	"fmt"
	"sync"
	"testing"
	"time"

	extendprovider "multifish/providers/extend"

	"github.com/stretchr/testify/assert"
)

// TestWorkerPool_ConcurrencyControl tests that the worker pool properly limits concurrent executions
func TestWorkerPool_ConcurrencyControl(t *testing.T) {
	// Create a mock executor that tracks concurrent executions
	executor := &ConcurrencyTrackingExecutor{
		maxConcurrent: 0,
		currentCount:  0,
	}

	mockValidator := &MockJobValidator{}
	service := NewJobService(mockValidator, executor)
	defer service.Stop()

	// Create multiple jobs with past execution times so they run immediately
	numJobs := 20
	now := time.Now()
	pastTime := now.Add(-2 * time.Hour)

	for i := 0; i < numJobs; i++ {
		// Create job with past run time
		jobID := time.Now().UnixNano() + int64(i)
		job := &Job{
			ID:             fmt.Sprintf("Job-%d", jobID),
			Name:           "Concurrent Test Job",
			Machines:       []string{"machine1"},
			Action:         ActionPatchProfile,
			Status:         JobStatusPending,
			CreatedTime:    now,
			NextRunTime:    &pastTime, // Set to past
			ExecutionCount: 0,
		}

		service.mu.Lock()
		service.jobs[job.ID] = job
		service.mu.Unlock()
	}

	// Manually trigger job execution check
	service.checkAndExecuteJobs()

	// Wait a bit for jobs to start executing
	time.Sleep(100 * time.Millisecond)

	// Check that we never exceeded the worker pool size
	executor.mu.Lock()
	maxConcurrent := executor.maxConcurrent
	executor.mu.Unlock()

	poolSize := service.GetWorkerPoolSize()
	assert.Greater(t, maxConcurrent, 0, "Should have executed some jobs concurrently")
	assert.LessOrEqual(t, maxConcurrent, poolSize,
		"Maximum concurrent executions (%d) should not exceed WorkerPoolSize (%d)",
		maxConcurrent, poolSize)

	t.Logf("Max concurrent executions: %d (limit: %d)", maxConcurrent, poolSize)

	// Wait for all jobs to complete
	time.Sleep(300 * time.Millisecond)
}

// TestWorkerPool_Metrics tests the worker pool metrics methods
func TestWorkerPool_Metrics(t *testing.T) {
	mockValidator := &MockJobValidator{}
	mockExecutor := &MockJobExecutor{}

	service := NewJobService(mockValidator, mockExecutor)
	defer service.Stop()

	// Initially, all workers should be available
	assert.Equal(t, DefaultWorkerPoolSize, service.GetWorkerPoolSize())
	assert.Equal(t, 0, service.GetActiveWorkers())
	assert.Equal(t, DefaultWorkerPoolSize, service.GetAvailableWorkers())
	assert.Equal(t, 0, service.GetRunningJobsCount())

	// Create a job
	request := &JobCreateRequest{
		Name:     "Metrics Test Job",
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
	assert.NoError(t, err)

	assert.Equal(t, 1, service.GetJobCount())
}

// ConcurrencyTrackingExecutor tracks maximum concurrent executions
type ConcurrencyTrackingExecutor struct {
	mu            sync.Mutex
	maxConcurrent int
	currentCount  int
}

func (e *ConcurrencyTrackingExecutor) ExecuteJob(job *Job) *ExecutionHistory {
	e.mu.Lock()
	e.currentCount++
	if e.currentCount > e.maxConcurrent {
		e.maxConcurrent = e.currentCount
	}
	e.mu.Unlock()

	// Simulate work
	time.Sleep(50 * time.Millisecond)

	e.mu.Lock()
	e.currentCount--
	e.mu.Unlock()

	return &ExecutionHistory{
		JobID:         job.ID,
		ExecutionTime: time.Now(),
		Status:        JobStatusCompleted,
		Results: []MachineExecutionResult{
			{
				MachineID: "machine1",
				Success:   true,
				Message:   "Success",
			},
		},
	}
}
