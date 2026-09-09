# Scheduler Module

## Overview

The `scheduler/` package implements a sophisticated job scheduling system that enables automated, time-based execution of BMC operations across multiple machines. It supports both one-time and recurring schedules with flexible timing options.

## Why Job Scheduler?

The scheduler solves several operational challenges:
- **Automation**: Execute repetitive tasks without manual intervention
- **Coordination**: Apply changes to multiple machines simultaneously
- **Timing Control**: Schedule operations during maintenance windows
- **Resource Management**: Control concurrency with worker pools
- **Audit Trail**: Track job execution history and results

## Architecture

```
scheduler/
├── job_models.go              # Core data structures and validation
├── job_models_test.go         # Model tests
├── job_action.go              # Action execution logic
├── job_executor.go            # Job execution engine
├── job_executor_test.go       # Executor tests
├── job_service.go             # Job service and management
├── job_service_test.go        # Service tests
├── job_service_worker_pool_test.go  # Worker pool tests
├── payload_models.go          # Payload structures and validation
├── README.md                  # This file
└── logs/                      # Job execution logs
    └── job*.json              # Individual job execution results
```

## Core Concepts

### 1. Job Structure (`job_models.go`)

The fundamental unit of work in the scheduler.

```go
type Job struct {
    ID             string
    Name           string
    Machines       []string
    Action         ActionType
    Payload        Payload
    Schedule       Schedule
    Status         JobStatus
    CreatedTime    time.Time
    LastRunTime    *time.Time
    NextRunTime    *time.Time
    ExecutionCount int
}
```

**Fields Explained:**
- **ID**: Unique identifier (generated automatically)
- **Name**: Human-readable description
- **Machines**: List of target machine IDs
- **Action**: Operation to perform (e.g., PatchProfile)
- **Payload**: Action-specific configuration data
- **Schedule**: When and how often to run
- **Status**: Current state (Pending, Running, Completed, Failed)
- **CreatedTime**: When job was created
- **LastRunTime**: Last execution timestamp
- **NextRunTime**: Scheduled next execution
- **ExecutionCount**: Number of times executed

**Job Lifecycle:**
```
Created → Pending → Scheduled → Running → Completed
                                    ↓
                                  Failed
                                    ↓
                                 Retried (for continuous jobs)
```

### 2. Schedule Types (`job_models.go`)

#### Once Schedule

Executes exactly one time at specified time.

```json
{
  "Type": "Once",
  "Time": "14:30:00",
  "Period": null
}
```

**Use Cases:**
- One-time configuration changes
- Immediate or near-future operations
- Testing new configurations

**Behavior:**
- Runs once at specified time
- Status changes to Completed after execution
- Not rescheduled

#### Continuous Schedule

Executes repeatedly based on specified period.

```json
{
  "Type": "Continuous",
  "Time": "22:00:00",
  "Period": {
    "StartDay": "2026-02-10",
    "EndDay": "2026-12-31",
    "DaysOfWeek": ["Monday", "Wednesday", "Friday"],
    "DaysOfMonth": null
  }
}
```

**Period Options:**
- **Daily**: Empty DaysOfWeek and no DaysOfMonth
- **Weekly**: Specify DaysOfWeek (e.g., ["Monday", "Friday"])
- **Monthly**: Specify DaysOfMonth (e.g., "1,15,30")
- **Date Range**: Use StartDay and EndDay

**Examples:**

**Daily at 2 AM:**
```json
{
  "Type": "Continuous",
  "Time": "02:00:00",
  "Period": {
    "StartDay": "2026-02-10",
    "EndDay": "2026-12-31",
    "DaysOfWeek": [],
    "DaysOfMonth": null
  }
}
```

**Weekdays at 8 AM:**
```json
{
  "Type": "Continuous",
  "Time": "08:00:00",
  "Period": {
    "StartDay": "2026-02-10",
    "EndDay": "2026-12-31",
    "DaysOfWeek": ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday"],
    "DaysOfMonth": null
  }
}
```

**1st and 15th of each month:**
```json
{
  "Type": "Continuous",
  "Time": "00:00:00",
  "Period": {
    "StartDay": "2026-02-10",
    "EndDay": "2026-12-31",
    "DaysOfWeek": [],
    "DaysOfMonth": "1,15"
  }
}
```

### 3. Action Types (`job_action.go`)

Supported operations that can be scheduled.

```go
const (
    ActionPatchProfile       ActionType = "PatchProfile"
    ActionPatchManager       ActionType = "PatchManager"
    ActionPatchFanController ActionType = "PatchFanController"
    ActionPatchFanZone       ActionType = "PatchFanZone"
    ActionPatchPidController ActionType = "PatchPidController"
)
```

**Action Execution Flow:**
```
Job Triggered
     ↓
For each machine:
     ↓
  Get machine connection
     ↓
  Execute action with payload
     ↓
  Log result
     ↓
Aggregate results
     ↓
Update job status
```

### 4. Payload Models (`payload_models.go`)

Action-specific configuration structures.

#### PatchProfile Payload

```go
type ExecutePatchProfilePayload struct {
    ManagerID string
    Payload   extendprovider.PatchProfileType
}

// Example
{
  "ManagerID": "bmc",
  "Payload": {
    "Profile": "Performance"
  }
}
```

**Validation:**
- ManagerID cannot be empty
- Profile must be one of: Performance, Balanced, PowerSaver, Custom
- No duplicate ManagerIDs

#### PatchManager Payload

```go
type ExecutePatchManagerPayload struct {
    ManagerID string
    Payload   redfish.PatchManagerType
}

// Example
{
  "ManagerID": "bmc",
  "Payload": {
    "ServiceIdentification": "Production BMC"
  }
}
```

#### PatchFanController Payload

```go
type ExecutePatchFanControllerPayload struct {
    ManagerID       string
    FanControllerID string
    Payload         extendprovider.PatchFanControllerType
}

// Example
{
  "ManagerID": "bmc",
  "FanControllerID": "cpu_fan_controller",
  "Payload": {
    "Multiplier": 1.2,
    "StepDown": 2,
    "StepUp": 5
  }
}
```

#### PatchFanZone Payload

```go
type ExecutePatchFanZonePayload struct {
    ManagerID string
    FanZoneID string
    Payload   extendprovider.PatchFanZoneType
}

// Example
{
  "ManagerID": "bmc",
  "FanZoneID": "cpu_zone",
  "Payload": {
    "FailSafePercent": 100.0,
    "MinThermalOutput": 30.0
  }
}
```

#### PatchPidController Payload

```go
type ExecutePatchPidControllerPayload struct {
    ManagerID       string
    PidControllerID string
    Payload         extendprovider.PatchPidControllerType
}

// Example
{
  "ManagerID": "bmc",
  "PidControllerID": "cpu_temp_controller",
  "Payload": {
    "PCoefficient": 0.8,
    "ICoefficient": 0.1,
    "SetPoint": 75.0
  }
}
```

**Multiple Managers:**

Jobs can target multiple managers in a single payload:

```json
{
  "Name": "Update Multiple Profiles",
  "Machines": ["machine-1", "machine-2"],
  "Action": "PatchProfile",
  "Payload": [
    {
      "ManagerID": "bmc",
      "Payload": {"Profile": "Performance"}
    },
    {
      "ManagerID": "bmc2",
      "Payload": {"Profile": "Balanced"}
    }
  ]
}
```

### 5. Job Executor (`job_executor.go`)

Executes jobs and manages their lifecycle.

**Key Components:**

#### JobExecutor Interface

```go
type JobExecutor interface {
    ExecuteJob(job *Job) error
    ValidateJob(job *Job) error
}
```

#### PlatformExecutor

```go
type PlatformExecutor struct {
    platformMgr    JobPlatformManager
    actionExecutor ActionExecutor
}
```

**Note:** The `JobPlatformManager` interface is defined in the scheduler package and provides platform-specific machine operations.
**Execution Flow:**
```
1. Validate job structure
2. For each machine:
   a. Get platform connection
   b. Validate platform supports action
   c. Execute action with payload
   d. Collect results
3. Log execution details
4. Update job status
5. Calculate next run time (if continuous)
```

**Error Handling:**
- Individual machine failures don't stop job
- Results logged per machine
- Job marked failed if all machines fail
- Continuous jobs reschedule even after failures

### 6. Job Service (`job_service.go`)

Manages job lifecycle and scheduling.

**Key Features:**

#### Worker Pool

Concurrent job execution with controlled parallelism.

```go
type JobService struct {
    jobs          map[string]*Job
    workerPool    chan struct{}
    activeWorkers int32
    // ...
}
```

**Configuration:**
```json
{
  "WorkerPoolSize": 10
}
```

**Why Worker Pool?**
- Prevents resource exhaustion
- Controls system load
- Ensures fair job scheduling
- Limits concurrent BMC connections

**Worker Pool Behavior:**
- Default size: 5 workers
- Configurable via PATCH /JobService
- Blocks when pool exhausted
- Releases worker after job completion

#### Job Ticker

Checks for due jobs every minute.

```go
func (js *JobService) Start() {
    ticker := time.NewTicker(1 * time.Minute)
    go func() {
        for range ticker.C {
            js.checkAndExecuteJobs()
        }
    }()
}
```

**Schedule Check Logic:**
```
Every minute:
  For each job:
    If job.NextRunTime <= now:
      If worker available:
        Execute job
      Else:
        Wait for worker
```

#### Job Management Operations

**Create Job:**
```go
job, err := jobService.CreateJob(request)
// Returns: Job with generated ID and calculated NextRunTime
```

**Get Job:**
```go
job, err := jobService.GetJob(jobID)
// Returns: Current job state
```

**Update Job:**
```go
err := jobService.UpdateJob(jobID, updates)
// Allows: Name, Schedule changes
// Recalculates NextRunTime
```

**Delete Job:**
```go
err := jobService.DeleteJob(jobID)
// Cancels future executions
// Does not cancel running job
```

**Trigger Job:**
```go
err := jobService.TriggerJob(jobID)
// Executes immediately regardless of schedule
// Does not modify NextRunTime
```

### 7. Job Logging

Execution results are logged to `logs/`.

**Log Format:**
```
{jobID}_{machineID}_{action}_{timestamp}_{status}.json
```

**Example:**
```
job1_machine1_PatchProfile_20260209_154349_success.json
```

**Log Contents:**
```json
{
  "JobID": "job1",
  "MachineID": "machine-1",
  "Action": "PatchProfile",
  "Status": "success",
  "Timestamp": "2026-02-09T15:43:49Z",
  "ExecutionTime": "1.234s",
  "Result": {
    "ManagerID": "bmc",
    "OldProfile": "Balanced",
    "NewProfile": "Performance"
  },
  "Error": null
}
```

**Log Retention:**
- Logs persist indefinitely
- Manual cleanup required
- Consider log rotation for production

## Usage Examples

### Create One-Time Job

```bash
curl -X POST http://localhost:8080/MultiFish/v1/JobService/Jobs \
  -H "Content-Type: application/json" \
  -d '{
    "Name": "Update Profile to Performance",
    "Machines": ["machine-1", "machine-2"],
    "Action": "PatchProfile",
    "Payload": [
      {
        "ManagerID": "bmc",
        "Payload": {
          "Profile": "Performance"
        }
      }
    ],
    "Schedule": {
      "Type": "Once",
      "Time": "14:00:00",
      "Period": null
    }
  }'
```

### Create Daily Recurring Job

```bash
curl -X POST http://localhost:8080/MultiFish/v1/JobService/Jobs \
  -H "Content-Type: application/json" \
  -d '{
    "Name": "Nightly PowerSaver Mode",
    "Machines": ["machine-1"],
    "Action": "PatchProfile",
    "Payload": [
      {
        "ManagerID": "bmc",
        "Payload": {
          "Profile": "PowerSaver"
        }
      }
    ],
    "Schedule": {
      "Type": "Continuous",
      "Time": "22:00:00",
      "Period": {
        "StartDay": "2026-02-10",
        "EndDay": "2026-12-31",
        "DaysOfWeek": [],
        "DaysOfMonth": null
      }
    }
  }'
```

### Update Job Schedule

```bash
curl -X PATCH http://localhost:8080/MultiFish/v1/JobService/Jobs/job_123 \
  -H "Content-Type: application/json" \
  -d '{
    "Schedule": {
      "Type": "Once",
      "Time": "16:00:00",
      "Period": null
    }
  }'
```

### Trigger Job Immediately

```bash
curl -X POST http://localhost:8080/MultiFish/v1/JobService/Jobs/job_123/Actions/Trigger
```

### Get Job Status

```bash
curl http://localhost:8080/MultiFish/v1/JobService/Jobs/job_123
```

**Response:**
```json
{
  "@odata.type": "#Job.v1_0_0.Job",
  "@odata.id": "/MultiFish/v1/JobService/Jobs/job_123",
  "Id": "job_123",
  "Name": "Update Profile",
  "Status": "Completed",
  "ExecutionCount": 5,
  "LastRunTime": "2026-02-09T15:43:49Z",
  "NextRunTime": "2026-02-10T15:43:00Z"
}
```

## Advanced Topics

### Validation Pipeline

Every job goes through multi-stage validation:

```
1. Structure Validation
   - Required fields present
   - Valid data types
   
2. Payload Validation
   - Action-specific rules
   - Field whitelists
   - Value ranges
   
3. Platform Validation
   - Machines exist
   - Platforms support action
   - Payload compatible
   
4. Schedule Validation
   - Valid time format
   - Period constraints
   - Date range logic
```

### Next Run Time Calculation

**Once Schedule:**
```go
if scheduleTime.After(now) {
    nextRun = scheduleTime today
} else {
    nextRun = scheduleTime tomorrow
}
```

**Continuous Schedule:**
```go
1. Start from current time
2. Check if today matches criteria (DaysOfWeek/DaysOfMonth)
3. If yes and time not passed: schedule for today
4. Otherwise: find next matching day
5. Apply time component
6. Ensure within StartDay/EndDay range
```

### Worker Pool Management

**Acquiring Worker:**
```go
select {
case js.workerPool <- struct{}{}:
    // Worker acquired
    defer func() { <-js.workerPool }()
    executeJob()
case <-time.After(timeout):
    // Worker unavailable
    return error("worker pool exhausted")
}
```

**Metrics:**
```go
WorkerPoolSize:   cap(js.workerPool)
ActiveWorkers:    js.activeWorkers
AvailableWorkers: cap(js.workerPool) - js.activeWorkers
```

### Error Handling Strategies

**Transient Errors:**
- Network timeouts
- Temporary BMC unavailability
- Continuous jobs retry on next schedule

**Permanent Errors:**
- Invalid payload
- Unsupported action
- Job marked failed immediately

**Partial Failures:**
- Some machines succeed, others fail
- Job continues executing
- Results logged per machine
- Overall status reflects worst outcome

## Testing

### Unit Tests

```go
func TestJobValidation(t *testing.T) {
    tests := []struct {
        name    string
        job     *Job
        wantErr bool
    }{
        {
            name: "valid job",
            job: &Job{
                Name:     "Test",
                Machines: []string{"m1"},
                Action:   ActionPatchProfile,
                Payload:  validPayload,
                Schedule: validSchedule,
            },
            wantErr: false,
        },
        {
            name: "missing name",
            job: &Job{
                Machines: []string{"m1"},
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.job.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("unexpected error: %v", err)
            }
        })
    }
}
```

### Integration Tests

```go
func TestJobExecution(t *testing.T) {
    // Setup test environment
    executor := setupTestExecutor()
    
    job := &Job{
        // ... test job configuration
    }
    
    err := executor.ExecuteJob(job)
    assert.NoError(t, err)
    assert.Equal(t, JobStatusCompleted, job.Status)
    assert.NotNil(t, job.LastRunTime)
}
```
