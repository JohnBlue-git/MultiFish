# Job Service

## Overview

The **Job Service** provides a sophisticated scheduling system for automating BMC operations across multiple machines. It enables time-based execution of recurring or one-time tasks with worker pool management, detailed execution logging, and comprehensive validation.

## Table of Contents

- [Architecture](#architecture)
- [Core Concepts](#core-concepts)
- [Job Structure](#job-structure)
- [Schedule Types](#schedule-types)
- [Actions](#actions)
- [Payload Validation](#payload-validation)
- [Worker Pools](#worker-pools)
- [Execution Flow](#execution-flow)
- [API Endpoints](#api-endpoints)
- [Usage Examples](#usage-examples)
- [Execution Logs](#execution-logs)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Architecture

```
┌────────────────────────────────────────────────────────┐
│              Job Service API Layer                     │
│   GET/POST/PATCH/DELETE /MultiFish/v1/JobService      │
└────────────────────┬───────────────────────────────────┘
                     │
         ┌───────────▼─────────────┐
         │     JobService          │
         │  - Job Registry         │
         │  - Scheduler Loop       │
         │  - Worker Pool Mgmt     │
         └───┬──────────────┬──────┘
             │              │
    ┌────────▼────┐    ┌────▼──────────┐
    │  Validator  │    │  Executor     │
    │  - Schedule │    │  - Actions    │
    │  - Payload  │    │  - Results    │
    │  - Machines │    │  - Logging    │
    └─────────────┘    └───┬───────────┘
                           │
               ┌───────────▼────────────┐
               │   Worker Pool          │
               │  - Semaphore (99)      │
               │  - Concurrent Jobs     │
               │  - Resource Control    │
               └───┬────────────────────┘
                   │
        ┌──────────┴──────────┐
        │                     │
   ┌────▼─────┐        ┌─────▼──────┐
   │ Machine1 │        │  Machine2  │
   │ Actions  │   ...  │  Actions   │
   └──────────┘        └────────────┘
                           │
                    ┌──────▼──────┐
                    │  JSON Logs  │
                    │ (per exec)  │
                    └─────────────┘
```

## Core Concepts

### Job Service Components

1. **Job Registry**: In-memory storage for all jobs
2. **Scheduler Loop**: Background goroutine checking due jobs
3. **Validator**: Pre-execution validation (schedule, payload, machines)
4. **Executor**: Executes actions on target machines
5. **Worker Pool**: Controls concurrent job execution

### Key Features

- ✅ **Time-based Scheduling**: Once or continuous (daily/weekly/monthly)
- ✅ **Multi-machine Support**: Execute on multiple BMCs simultaneously
- ✅ **Worker Pool**: Configurable concurrency control (1-10000 workers)
- ✅ **Comprehensive Validation**: Schedule, payload, and machine validation
- ✅ **Detailed Logging**: JSON execution logs per job per machine
- ✅ **Automatic Rescheduling**: Continuous jobs reschedule after execution
- ✅ **Immediate Trigger**: Override schedule and run now
- ✅ **Thread-safe**: Concurrent job management with mutex protection

## Job Structure

### Job Model

```go
type Job struct {
    ID             string        // Unique identifier (auto-generated)
    Name           string        // Human-readable description
    Machines       []string      // Target machine IDs
    Action         ActionType    // Operation to perform
    Payload        Payload       // Action-specific data
    Schedule       Schedule      // Timing information
    Status         JobStatus     // Current state
    CreatedTime    time.Time     // Creation timestamp
    LastRunTime    *time.Time    // Last execution time
    NextRunTime    *time.Time    // Next scheduled execution
    ExecutionCount int           // Number of executions
}
```

### Job Lifecycle

```
Created → Pending → Scheduled
                        ↓
                    Running → Completed
                        ↓
                    Running → Failed → Retried (continuous)
                        ↓
                    Cancelled
```

**Status Transitions:**

| Status | Description | Next States |
|--------|-------------|-------------|
| `Pending` | Job created, not yet scheduled | `Scheduled` |
| `Scheduled` | Waiting for next run time | `Running` |
| `Running` | Currently executing | `Completed`, `Failed` |
| `Completed` | Successfully executed | `Scheduled` (continuous) or terminal (once) |
| `Failed` | Execution failed | `Scheduled` (continuous) or terminal (once) |
| `Cancelled` | User cancelled | Terminal state |

## Schedule Types

### Once Schedule

Execute exactly one time at specified time.

**Structure:**
```json
{
  "Type": "Once",
  "Time": "14:30:00",
  "Period": null
}
```

**Characteristics:**
- Single execution
- Status → `Completed` after run
- Not rescheduled
- NextRunTime cleared after execution

**Use Cases:**
- One-time configuration changes
- Immediate or near-future operations
- Testing new configurations
- Emergency responses

**Example:**
```json
{
  "Name": "Emergency Fan Speed Increase",
  "Machines": ["server-1"],
  "Action": "PatchFanController",
  "Payload": [...],
  "Schedule": {
    "Type": "Once",
    "Time": "15:00:00"
  }
}
```

### Continuous Schedule

Execute repeatedly based on period configuration.

**Structure:**
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

**Period Configuration:**

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `StartDay` | string | Start date (YYYY-MM-DD) | `"2026-02-10"` |
| `EndDay` | string | End date (YYYY-MM-DD) | `"2026-12-31"` |
| `DaysOfWeek` | []string | Specific weekdays | `["Monday", "Friday"]` |
| `DaysOfMonth` | string | Specific days (1-31) | `"1,15,30"` |

**Schedule Patterns:**

#### Daily Schedule

Run every day at specified time.

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

**Use Cases:**
- Daily backup operations
- Nightly performance mode
- Regular maintenance tasks

#### Weekly Schedule

Run on specific weekdays.

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

**Common Patterns:**
- **Weekdays**: `["Monday", "Tuesday", "Wednesday", "Thursday", "Friday"]`
- **Weekend**: `["Saturday", "Sunday"]`
- **MWF**: `["Monday", "Wednesday", "Friday"]`
- **Tue/Thu**: `["Tuesday", "Thursday"]`

#### Monthly Schedule

Run on specific days of the month.

```json
{
  "Type": "Continuous",
  "Time": "00:00:00",
  "Period": {
    "StartDay": "2026-02-01",
    "EndDay": "2026-12-31",
    "DaysOfWeek": [],
    "DaysOfMonth": "1,15"
  }
}
```

**Examples:**
- **Monthly (1st)**: `"DaysOfMonth": "1"`
- **Bi-weekly (1st & 15th)**: `"DaysOfMonth": "1,15"`
- **Quarter End**: `"DaysOfMonth": "31"`
- **Multiple Days**: `"DaysOfMonth": "1,10,20,30"`

## Actions

### Supported Actions

| Action | Description | Target | Payload Type |
|--------|-------------|--------|--------------|
| `PatchProfile` | Update thermal profile | Manager OEM | Profile selection |
| `PatchManager` | Update manager properties | Manager | Manager attributes |
| `PatchFanController` | Configure fan controller | Manager OEM | Fan controller settings |
| `PatchFanZone` | Configure fan zone | Manager OEM | Fan zone settings |
| `PatchPidController` | Configure PID controller | Manager OEM | PID parameters |

### PatchProfile

Update thermal management profile.

**Payload Structure:**
```json
{
  "ManagerID": "bmc",
  "Payload": {
    "Profile": "Performance"
  }
}
```

**Valid Profiles:**
- `Performance` - Maximum performance, higher power
- `Balanced` - Optimal performance/efficiency balance
- `PowerSaver` - Minimize power consumption
- `Custom` - User-defined settings

**Example Job:**
```json
{
  "Name": "Nightly PowerSaver Mode",
  "Machines": ["server-1", "server-2"],
  "Action": "PatchProfile",
  "Payload": [
    {
      "ManagerID": "bmc",
      "Payload": {"Profile": "PowerSaver"}
    }
  ],
  "Schedule": {
    "Type": "Continuous",
    "Time": "22:00:00",
    "Period": {
      "StartDay": "2026-02-10",
      "EndDay": "2026-12-31",
      "DaysOfWeek": []
    }
  }
}
```

### PatchManager

Update manager properties.

**Payload Structure:**
```json
{
  "ManagerID": "bmc",
  "Payload": {
    "ServiceIdentification": "Production BMC v2.1"
  }
}
```

**Allowed Fields:**
- `ServiceIdentification` - Manager description
- Other manager-specific properties (varies by BMC)

### PatchFanController

Configure fan controller settings.

**Payload Structure:**
```json
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

**Parameters:**
- `Multiplier` - Fan speed multiplier (0.0-2.0)
- `StepDown` - Temperature decrease step (°C)
- `StepUp` - Temperature increase step (°C)

### PatchFanZone

Configure fan zone settings.

**Payload Structure:**
```json
{
  "ManagerID": "bmc",
  "FanZoneID": "zone_1",
  "Payload": {
    "MinThermalOutput": 30.0,
    "FailSafePercent": 100.0
  }
}
```

### PatchPidController

Configure PID controller parameters.

**Payload Structure:**
```json
{
  "ManagerID": "bmc",
  "PidControllerID": "pid_cpu",
  "Payload": {
    "SetPoint": 70.0,
    "ProportionalCoefficient": 0.5,
    "IntegralCoefficient": 0.1,
    "DerivativeCoefficient": 0.01
  }
}
```

## Payload Validation

### Validation Process

The job service performs comprehensive validation before job creation:

```
┌─────────────────┐
│ Create Request  │
└────────┬────────┘
         │
    ┌────▼─────────────┐
    │ Schedule Valid?  │  ← Time format, period logic
    └────┬─────────────┘
         │ Yes
    ┌────▼─────────────┐
    │ Action Valid?    │  ← Known action type
    └────┬─────────────┘
         │ Yes
    ┌────▼─────────────┐
    │ Payload Valid?   │  ← Type-specific validation
    └────┬─────────────┘
         │ Yes
    ┌────▼─────────────┐
    │ Machines Exist?  │  ← Platform registry check
    └────┬─────────────┘
         │ All Valid
    ┌────▼─────────────┐
    │  Job Created     │
    └──────────────────┘
```

### Validation Response

**Success:**
```json
{
  "@odata.type": "#Job.v1_0_0.Job",
  "@odata.id": "/MultiFish/v1/JobService/Jobs/Job-1707489234567890",
  "Id": "Job-1707489234567890",
  "Name": "Daily Profile Update",
  "Status": "Pending",
  ...
}
```

**Failure:**
```json
{
  "error": {
    "code": "JobValidationFailed",
    "message": "Job validation failed",
    "@Message.ExtendedInfo": [
      {
        "MessageId": "Base.1.0.JobValidationFailed",
        "Message": "Validation failed: schedule errors, machine validation failed",
        "Severity": "Critical",
        "ValidationDetails": {
          "ScheduleValid": false,
          "ScheduleErrors": ["Invalid time format"],
          "ActionValid": true,
          "ActionErrors": [],
          "PayloadValid": true,
          "PayloadErrors": [],
          "MachineResults": [
            {
              "MachineId": "server-1",
              "Valid": false,
              "Message": "Machine not found",
              "Errors": ["machine server-1 not found in platform"]
            }
          ]
        }
      }
    ]
  }
}
```

### Common Validation Errors

**Schedule Errors:**
- Invalid time format (must be HH:MM:SS)
- Missing Period for Continuous type
- Invalid DaysOfWeek values
- Invalid date format (must be YYYY-MM-DD)
- EndDay before StartDay

**Action Errors:**
- Unknown action type
- Action not supported for machine type

**Payload Errors:**
- Missing required fields
- Invalid field values
- Type mismatch
- Unknown fields

**Machine Errors:**
- Machine not found in platform
- Machine not connected
- Insufficient permissions

## Worker Pools

### Configuration

**Default:** 99 concurrent workers

**Configurable via:**
1. **Environment Variable:**
   ```bash
   WORKER_POOL_SIZE=150 ./multifish
   ```

2. **Configuration File:**
   ```yaml
   worker_pool_size: 150
   ```

3. **Runtime API:**
   ```bash
   curl -X PATCH http://localhost:8080/MultiFish/v1/JobService \
     -H "Content-Type: application/json" \
     -d '{"ServiceCapabilities": {"WorkerPoolSize": 150}}'
   ```

### Worker Pool Sizing

**Guidelines:**

| Scenario | Recommended Size | Reasoning |
|----------|-----------------|-----------|
| Small deployment (1-10 machines) | 10-25 | Avoid resource waste |
| Medium deployment (10-50 machines) | 50-100 | Balance throughput/resources |
| Large deployment (50-200 machines) | 100-500 | Maximize parallelism |
| Enterprise (200+ machines) | 500-1000 | Handle peak loads |

**Considerations:**
- CPU cores available
- Network bandwidth
- BMC response times
- Memory constraints
- Concurrent job frequency

### Worker Pool Metrics

```bash
curl http://localhost:8080/MultiFish/v1/JobService
```

**Response:**
```json
{
  "ServiceCapabilities": {
    "WorkerPoolSize": 99,
    "ActiveWorkers": 15,
    "AvailableWorkers": 84,
    "TotalJobs": 42,
    "RunningJobs": 3
  }
}
```

**Metrics:**
- `WorkerPoolSize` - Maximum concurrent jobs
- `ActiveWorkers` - Currently executing
- `AvailableWorkers` - Free workers
- `TotalJobs` - All jobs in registry
- `RunningJobs` - Jobs in Running status

### Worker Pool Behavior

**Semaphore-based Control:**
```go
// Worker pool uses semaphore
type WorkerPool struct {
    sem chan struct{}  // Buffered channel (size = pool size)
}

// Acquire worker
<-wp.sem

// Release worker
wp.sem <- struct{}{}
```

**Blocking:**
- If all workers busy, new jobs wait
- Jobs queue until worker available
- No job is dropped (queued indefinitely)

## Execution Flow

### Job Execution Lifecycle

```
1. Scheduler Check (every 1 second)
   ↓
2. Find Due Jobs (NextRunTime <= Now)
   ↓
3. Acquire Worker from Pool
   ↓
4. Update Status → Running
   ↓
5. Execute Action on Each Machine
   │  - Connect to machine
   │  - Perform action
   │  - Log result
   ↓
6. Aggregate Results
   ↓
7. Update Job Status → Completed/Failed
   ↓
8. Write Execution Log (JSON)
   ↓
9. Reschedule (if Continuous)
   ↓
10. Release Worker
```

### Scheduler Loop

**Implementation:**
```go
func (js *JobService) Start() {
    ticker := time.NewTicker(1 * time.Second)
    for range ticker.C {
        jobs := js.findDueJobs()
        for _, job := range jobs {
            go js.executeJob(job)  // Async execution
        }
    }
}
```

**Characteristics:**
- Checks every 1 second
- Spawns goroutine per due job
- Non-blocking (doesn't wait for completion)
- Concurrent execution of multiple jobs

### Per-Machine Execution

```go
func executeOnMachine(machine, action, payload) {
    start := time.Now()
    
    // Execute action
    err := action.Execute(machine, payload)
    
    duration := time.Since(start)
    
    // Log result
    logResult(machine.ID, action, err, duration)
}
```

**Logged Information:**
- Job ID and machine ID
- Action type
- Execution timestamp
- Duration
- Success/failure status
- Error details (if failed)
- Payload applied

## API Endpoints

### GET /MultiFish/v1/JobService

Get JobService capabilities and metrics.

**Request:**
```bash
curl http://localhost:8080/MultiFish/v1/JobService
```

**Response:**
```json
{
  "@odata.type": "#JobService.v1_0_0.JobService",
  "@odata.id": "/MultiFish/v1/JobService",
  "Id": "JobService",
  "Name": "Job Service",
  "Jobs": {
    "@odata.id": "/MultiFish/v1/JobService/Jobs"
  },
  "ServiceCapabilities": {
    "WorkerPoolSize": 99,
    "ActiveWorkers": 5,
    "AvailableWorkers": 94,
    "TotalJobs": 12,
    "RunningJobs": 2
  }
}
```

### PATCH /MultiFish/v1/JobService

Update JobService configuration (worker pool size).

**Request:**
```bash
curl -X PATCH http://localhost:8080/MultiFish/v1/JobService \
  -H "Content-Type: application/json" \
  -d '{
    "ServiceCapabilities": {
      "WorkerPoolSize": 150
    }
  }'
```

**Constraints:**
- Min: 1
- Max: 10000
- Changes take effect immediately
- Running jobs not affected

### GET /MultiFish/v1/JobService/Jobs

List all jobs.

**Request:**
```bash
curl http://localhost:8080/MultiFish/v1/JobService/Jobs
```

**Response:**
```json
{
  "@odata.type": "#JobCollection.JobCollection",
  "@odata.id": "/MultiFish/v1/JobService/Jobs",
  "Name": "Job Collection",
  "Members": [
    {"@odata.id": "/MultiFish/v1/JobService/Jobs/Job-1707489234567890"},
    {"@odata.id": "/MultiFish/v1/JobService/Jobs/Job-1707489245678901"}
  ],
  "Members@odata.count": 2
}
```

### POST /MultiFish/v1/JobService/Jobs

Create a new job.

**Request:**
```bash
curl -X POST http://localhost:8080/MultiFish/v1/JobService/Jobs \
  -H "Content-Type: application/json" \
  -d @payloads/patch_profile.json
```

**Response (201 Created):**
```json
{
  "@odata.type": "#Job.v1_0_0.Job",
  "@odata.id": "/MultiFish/v1/JobService/Jobs/Job-1707489234567890",
  "Id": "Job-1707489234567890",
  "Name": "Daily PowerSaver Mode",
  "Machines": ["server-1"],
  "Action": "PatchProfile",
  "Payload": [...],
  "Schedule": {...},
  "Status": "Pending",
  "CreatedTime": "2026-02-10T14:03:54Z",
  "NextRunTime": "2026-02-10T22:00:00Z",
  "ExecutionCount": 0
}
```

### GET /MultiFish/v1/JobService/Jobs/{jobId}

Get job details.

**Request:**
```bash
curl http://localhost:8080/MultiFish/v1/JobService/Jobs/Job-1707489234567890
```

**Response:**
```json
{
  "@odata.type": "#Job.v1_0_0.Job",
  "@odata.id": "/MultiFish/v1/JobService/Jobs/Job-1707489234567890",
  "Id": "Job-1707489234567890",
  "Name": "Daily PowerSaver Mode",
  "Machines": ["server-1"],
  "Action": "PatchProfile",
  "Payload": [...],
  "Schedule": {...},
  "Status": "Completed",
  "CreatedTime": "2026-02-10T14:03:54Z",
  "LastRunTime": "2026-02-10T22:00:00Z",
  "NextRunTime": "2026-02-11T22:00:00Z",
  "ExecutionCount": 1
}
```

### DELETE /MultiFish/v1/JobService/Jobs/{jobId}

Delete a job.

**Request:**
```bash
curl -X DELETE http://localhost:8080/MultiFish/v1/JobService/Jobs/Job-1707489234567890
```

**Response:**
```json
{
  "message": "Job Job-1707489234567890 deleted successfully"
}
```

### POST /MultiFish/v1/JobService/Jobs/{jobId}/Actions/Cancel

Cancel a running or scheduled job.

**Request:**
```bash
curl -X POST http://localhost:8080/MultiFish/v1/JobService/Jobs/Job-1707489234567890/Actions/Cancel
```

**Response:**
```json
{
  "message": "Job Job-1707489234567890 cancelled successfully"
}
```

**Effects:**
- Status → `Cancelled`
- NextRunTime cleared
- Not rescheduled
- Running execution completes (not interrupted)

## Usage Examples

### Example 1: Daily Profile Switch

Switch to PowerSaver mode every night, Performance every morning.

**Night Mode (PowerSaver at 10 PM):**
```bash
curl -X POST http://localhost:8080/MultiFish/v1/JobService/Jobs \
  -H "Content-Type: application/json" \
  -d '{
    "Name": "Nightly PowerSaver Mode",
    "Machines": ["server-1", "server-2", "server-3"],
    "Action": "PatchProfile",
    "Payload": [
      {
        "ManagerID": "bmc",
        "Payload": {"Profile": "PowerSaver"}
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

**Day Mode (Performance at 8 AM):**
```bash
curl -X POST http://localhost:8080/MultiFish/v1/JobService/Jobs \
  -H "Content-Type: application/json" \
  -d '{
    "Name": "Morning Performance Mode",
    "Machines": ["server-1", "server-2", "server-3"],
    "Action": "PatchProfile",
    "Payload": [
      {
        "ManagerID": "bmc",
        "Payload": {"Profile": "Performance"}
      }
    ],
    "Schedule": {
      "Type": "Continuous",
      "Time": "08:00:00",
      "Period": {
        "StartDay": "2026-02-10",
        "EndDay": "2026-12-31",
        "DaysOfWeek": ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday"],
        "DaysOfMonth": null
      }
    }
  }'
```

### Example 2: Weekday Workload Pattern

```bash
curl -X POST http://localhost:8080/MultiFish/v1/JobService/Jobs \
  -H "Content-Type: application/json" \
  -d @payloads/continuous_weekdays.json
```

**File: payloads/continuous_weekdays.json**
```json
{
  "Name": "Weekday Performance Mode",
  "Machines": ["production-cluster-1", "production-cluster-2"],
  "Action": "PatchProfile",
  "Payload": [
    {
      "ManagerID": "bmc",
      "Payload": {"Profile": "Performance"}
    }
  ],
  "Schedule": {
    "Type": "Continuous",
    "Time": "08:00:00",
    "Period": {
      "StartDay": "2026-02-10",
      "EndDay": "2026-12-31",
      "DaysOfWeek": ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday"],
      "DaysOfMonth": null
    }
  }
}
```

### Example 3: Monthly Maintenance

```bash
curl -X POST http://localhost:8080/MultiFish/v1/JobService/Jobs \
  -H "Content-Type: application/json" \
  -d @payloads/continuous_monthly.json
```

**File: payloads/continuous_monthly.json**
```json
{
  "Name": "Monthly Maintenance Window",
  "Machines": ["server-1", "server-2"],
  "Action": "PatchManager",
  "Payload": [
    {
      "ManagerID": "bmc",
      "Payload": {
        "ServiceIdentification": "Production BMC - Maintenance"
      }
    }
  ],
  "Schedule": {
    "Type": "Continuous",
    "Time": "02:00:00",
    "Period": {
      "StartDay": "2026-02-01",
      "EndDay": "2026-12-31",
      "DaysOfWeek": [],
      "DaysOfMonth": "1"
    }
  }
}
```

### Example 4: Multiple Managers Per Machine

```bash
curl -X POST http://localhost:8080/MultiFish/v1/JobService/Jobs \
  -H "Content-Type: application/json" \
  -d '{
    "Name": "Update Multiple BMC Managers",
    "Machines": ["dual-bmc-server"],
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
    ],
    "Schedule": {
      "Type": "Once",
      "Time": "15:00:00",
      "Period": null
    }
  }'
```

### Example 5: Fan Controller Automation

```bash
curl -X POST http://localhost:8080/MultiFish/v1/JobService/Jobs \
  -H "Content-Type: application/json" \
  -d '{
    "Name": "Adjust Fan Controllers for Night",
    "Machines": ["server-1"],
    "Action": "PatchFanController",
    "Payload": [
      {
        "ManagerID": "bmc",
        "FanControllerID": "cpu_fan_controller",
        "Payload": {
          "Multiplier": 0.8,
          "StepDown": 3,
          "StepUp": 7
        }
      }
    ],
    "Schedule": {
      "Type": "Continuous",
      "Time": "21:00:00",
      "Period": {
        "StartDay": "2026-02-10",
        "EndDay": "2026-12-31",
        "DaysOfWeek": [],
        "DaysOfMonth": null
      }
    }
  }'
```

## Execution Logs

### Log Location

**Default:** `./logs/`

**Configurable:**
```bash
LOGS_DIR=/var/log/multifish/jobs ./multifish
```

### Log File Naming

Format: `job{jobId}_machine{machineId}_{action}_{timestamp}_{status}.json`

**Examples:**
```
job1_machine1_PatchProfile_20260210_140350_success.json
job2_machine3_PatchManager_20260210_153022_failed.json
job3_machine1_PatchFanController_20260210_160000_success.json
```

### Log File Structure

```json
{
  "job_id": "Job-1707489234567890",
  "machine_id": "server-1",
  "action": "PatchProfile",
  "timestamp": "2026-02-10T22:00:00Z",
  "status": "Success",
  "duration": "1.234s",
  "payload": {
    "ManagerID": "bmc",
    "Payload": {
      "Profile": "PowerSaver"
    }
  },
  "result": {
    "message": "Profile updated successfully",
    "details": "..."
  },
  "error": null
}
```

**Failed Execution:**
```json
{
  "job_id": "Job-1707489234567890",
  "machine_id": "server-2",
  "action": "PatchProfile",
  "timestamp": "2026-02-10T22:00:05Z",
  "status": "Failed",
  "duration": "2.345s",
  "payload": {...},
  "result": null,
  "error": {
    "code": "InternalError",
    "message": "Failed to update profile: connection timeout",
    "details": "BMC not responding within timeout period"
  }
}
```

### Log Retention

**Recommendations:**
- **Development:** Keep 7-30 days
- **Production:** Keep 90-365 days
- **Compliance:** As required by policy

**Rotation Strategy:**
```bash
# Find logs older than 90 days
find logs -name "*.json" -mtime +90

# Delete old logs
find logs -name "*.json" -mtime +90 -delete

# Archive to compressed storage
tar -czf logs-archive-$(date +%Y%m).tar.gz logs/*.json
```

## Best Practices

### 1. Job Naming

**Good Names:**
```
"Nightly PowerSaver Mode - Production Cluster"
"Weekday Performance Boost - Database Servers"
"Monthly Maintenance - BMC Identification Update"
```

**Avoid:**
```
"Job 1"
"Test"
"asdf"
```

**Guidelines:**
- Describe what and when
- Include target group
- Be specific and clear
- Use consistent naming conventions

### 2. Schedule Planning

**Time Zone Considerations:**
- Jobs use server local time
- Document time zone in job name/description
- Consider daylight saving time changes

**Avoid Peak Times:**
```json
{
  "Name": "Avoid: Profile change during business hours",
  "Time": "14:00:00"  // Bad: middle of workday
}

{
  "Name": "Better: Off-hours maintenance",
  "Time": "02:00:00"  // Good: low-usage period
}
```

**Maintenance Windows:**
```json
{
  "Name": "Scheduled Maintenance Window",
  "Schedule": {
    "Type": "Continuous",
    "Time": "02:00:00",
    "Period": {
      "DaysOfWeek": ["Sunday"]  // Weekly maintenance
    }
  }
}
```

### 3. Worker Pool Sizing

**Start Conservative:**
```yaml
worker_pool_size: 50  # Start here
```

**Monitor and Adjust:**
```bash
# Check metrics regularly
curl http://localhost:8080/MultiFish/v1/JobService | jq '.ServiceCapabilities'

# If ActiveWorkers often == WorkerPoolSize, increase
# If ActiveWorkers always << WorkerPoolSize, decrease
```

**Performance vs Resources:**
- More workers = higher concurrency, more memory
- Fewer workers = lower resource usage, longer queue times

### 4. Error Handling

**Monitor Logs:**
```bash
# Check recent failures
ls -ltr logs/*failed.json | tail -5

# Parse error patterns
grep -h "error" logs/*.json | jq -r '.error.message' | sort | uniq -c
```

**Respond to Failures:**
- Continuous jobs auto-retry on next schedule
- Once jobs require manual intervention
- Review logs to identify systematic issues

### 5. Testing Jobs

**Test Before Production:**
```bash
# Create test job with Once schedule
curl -X POST .../JobService/Jobs \
  -d '{
    "Name": "TEST - Profile Update",
    "Machines": ["test-machine"],
    "Action": "PatchProfile",
    "Schedule": {
      "Type": "Once",
      "Time": "14:05:00"  // 5 minutes from now
    }
  }'

# Monitor execution
watch -n 1 'curl -s .../JobService/Jobs/{jobId} | jq ".Status,.LastRunTime"'

# Check logs
ls -ltr logs/ | tail -1
cat logs/job*_TEST_*.json | jq
```

### 6. Machine Grouping

**Logical Grouping:**
```json
{
  "Name": "Production Cluster - Profile Update",
  "Machines": [
    "prod-web-01",
    "prod-web-02",
    "prod-web-03"
  ]
}

{
  "Name": "Database Cluster - Performance Mode",
  "Machines": [
    "prod-db-primary",
    "prod-db-replica-01",
    "prod-db-replica-02"
  ]
}
```

**Benefits:**
- Easier management
- Clear responsibility
- Simplified troubleshooting
- Better audit trails

### 7. Payload Validation

**Always validate before production:**
1. Test payload on single machine
2. Verify expected behavior
3. Check logs for errors
4. Roll out to larger group

**Use payload examples:**
```bash
# Start with examples
cp payloads/patch_profile.json my_job.json

# Customize
vim my_job.json

# Validate JSON syntax
cat my_job.json | jq '.'

# Create job
curl -X POST .../JobService/Jobs -d @my_job.json
```

## Troubleshooting

### Job Not Executing

**Check 1: Job Status**
```bash
curl http://localhost:8080/MultiFish/v1/JobService/Jobs/{jobId}
```

**Look for:**
- `Status`: Should be `Pending` or `Scheduled`
- `NextRunTime`: Should be in the future
- `NextRunTime`: Null means job won't run

**Check 2: Worker Pool**
```bash
curl http://localhost:8080/MultiFish/v1/JobService | jq '.ServiceCapabilities'
```

**If `AvailableWorkers == 0`:**
- All workers busy
- Jobs queued
- Consider increasing worker pool size

**Check 3: Schedule Validation**
```bash
# Get job details and examine schedule
curl http://localhost:8080/MultiFish/v1/JobService/Jobs/{jobId} | jq '.Schedule'
```

**Common issues:**
- Time in the past
- Invalid time format
- Wrong DaysOfWeek
- EndDay in the past

### Validation Errors

**Problem:** Job creation returns validation errors

**Solution:**
1. **Read error details carefully:**
   ```bash
   curl -X POST ... | jq '.error."@Message.ExtendedInfo"[0].ValidationDetails'
   ```

2. **Check each validation category:**
   - ScheduleValid
   - ActionValid
   - PayloadValid
   - MachineResults

3. **Fix identified issues:**
   ```json
   "ScheduleErrors": ["Invalid time format"]
   → Fix: Change "Time": "2pm" to "Time": "14:00:00"
   ```

### Jobs Queuing Up

**Symptoms:**
- Many jobs in `Pending` status
- `AvailableWorkers` always 0
- Long delays between jobs

**Solutions:**

1. **Increase worker pool:**
   ```bash
   curl -X PATCH .../JobService \
     -d '{"ServiceCapabilities": {"WorkerPoolSize": 200}}'
   ```

2. **Reduce job frequency:**
   ```json
   {
     "Time": "02:00:00",  // Spread jobs across time
     "Period": {"DaysOfWeek": ["Sunday"]}  // Less frequent
   }
   ```

3. **Stagger execution times:**
   ```json
   Job 1: "Time": "02:00:00"
   Job 2: "Time": "02:15:00"
   Job 3: "Time": "02:30:00"
   ```

### Machine Not Found

**Error:**
```json
{
  "MachineResults": [
    {
      "MachineId": "server-xyz",
      "Valid": false,
      "Message": "Machine not found"
    }
  ]
}
```

**Solutions:**
1. **List available machines:**
   ```bash
   curl http://localhost:8080/MultiFish/v1/Platform
   ```

2. **Check machine ID spelling:**
   - Case-sensitive
   - No extra spaces
   - Exact match required

3. **Register machine if missing:**
   ```bash
   curl -X POST .../Platform -d '{...}'
   ```

### Action Execution Failures

**Check execution logs:**
```bash
# Find recent failures
ls -ltr logs/*failed.json | tail -5

# Examine error
cat logs/job1_machine1_PatchProfile_*_failed.json | jq
```

**Common errors:**

| Error | Cause | Solution |
|-------|-------|----------|
| Connection timeout | BMC unreachable | Check network, BMC status |
| Authentication failed | Wrong credentials | Update machine config |
| Invalid payload | Wrong field values | Validate against examples |
| Resource not found | Wrong manager/controller ID | Check available resources |

### Performance Issues

**Symptoms:**
- Slow job execution
- High CPU usage
- Memory growth

**Diagnostics:**
```bash
# Check active jobs
curl .../JobService | jq '.ServiceCapabilities.RunningJobs'

# Monitor system resources
top
htop

# Check log file size
du -sh logs/
```

**Solutions:**
1. **Reduce worker pool if CPU high**
2. **Clean old logs if disk full**
3. **Increase timeout for slow BMCs**
4. **Distribute jobs across time**

## Integration with Platform Management

The Job Service is tightly integrated with Platform Management:

### Machine Lookup

```go
// Job executor uses platform manager
machine, err := platformMgr.GetMachine(machineID)
if err != nil {
    return fmt.Errorf("machine not found: %w", err)
}
```

**Flow:**
1. Job lists target machines
2. Executor requests each machine from platform
3. Platform returns machine connection
4. Executor performs action
5. Result logged

### Service Type Awareness

```go
// Different execution based on service type
switch machine.Config.Type {
case "Base":
    // Use BaseService for action
case "Extend":
    // Use ExtendService for action
}
```

**Impact:**
- Actions validated against machine capabilities
- Extend-only actions fail on Base machines
- Payload validation considers service type

### Validation Integration

```go
// Platform validator checks machines exist
for _, machineID := range job.Machines {
    machine, err := platform.GetMachine(machineID)
    if err != nil {
        validationErrors = append(validationErrors, err)
    }
}
```

See [PLATFORM.md](PLATFORM.md) for detailed platform management documentation.
