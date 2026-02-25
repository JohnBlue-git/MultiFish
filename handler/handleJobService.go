package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"multifish/config"
	"multifish/scheduler"
	"multifish/utility"
)

// ========== Job Service ==========

// MachineActionExecutorAdapter adapts main package functions for scheduler
type MachineActionExecutorAdapter struct{}

// formatJobResponse formats a job for the API response
func formatJobResponse(job *scheduler.Job) gin.H {
	response := gin.H{
		"@odata.type":    "#Job.v1_0_0.Job",
		"@odata.id":      fmt.Sprintf("/MultiFish/v1/JobService/Jobs/%s", job.ID),
		"Id":             job.ID,
		"Name":           job.Name,
		"Machines":       job.Machines,
		"Action":         job.Action,
		"Payload":        job.Payload,
		"Schedule":       formatSchedule(job.Schedule),
		"Status":         job.Status,
		"CreatedTime":    job.CreatedTime.Format("2006-01-02T15:04:05Z07:00"),
		"ExecutionCount": job.ExecutionCount,
	}

	if job.LastRunTime != nil {
		response["LastRunTime"] = job.LastRunTime.Format("2006-01-02T15:04:05Z07:00")
	}

	if job.NextRunTime != nil {
		response["NextRunTime"] = job.NextRunTime.Format("2006-01-02T15:04:05Z07:00")
	}

	return response
}

// formatSchedule formats a schedule for the API response
func formatSchedule(schedule scheduler.Schedule) gin.H {
	result := gin.H{
		"Type": schedule.Type,
		"Time": schedule.Time,
	}

	if schedule.Period != nil {
		period := gin.H{}

		if schedule.Period.StartDay != nil {
			period["StartDay"] = *schedule.Period.StartDay
		}

		if schedule.Period.EndDay != nil {
			period["EndDay"] = *schedule.Period.EndDay
		}

		if len(schedule.Period.DaysOfWeek) > 0 {
			period["DaysOfWeek"] = schedule.Period.DaysOfWeek
		} else {
			period["DaysOfWeek"] = []string{}
		}

		if schedule.Period.DaysOfMonth != nil {
			period["DaysOfMonth"] = *schedule.Period.DaysOfMonth
		}

		result["Period"] = period
	}

	return result
}

// GET /MultiFish/v1/JobService - Get JobService root
func getJobServiceRoot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"@odata.type": "#JobService.v1_0_0.JobService",
		"@odata.id":   "/MultiFish/v1/JobService",
		"Id":          "JobService",
		"Name":        "Job Service",
		"Jobs": gin.H{
			"@odata.id": "/MultiFish/v1/JobService/Jobs",
		},
		"ServiceCapabilities": gin.H{
			"WorkerPoolSize":    JobService.GetWorkerPoolSize(),
			"ActiveWorkers":     JobService.GetActiveWorkers(),
			"AvailableWorkers":  JobService.GetAvailableWorkers(),
			"TotalJobs":         JobService.GetJobCount(),
			"RunningJobs":       JobService.GetRunningJobsCount(),
		},
	})
}

// PATCH /MultiFish/v1/JobService - Update JobService configuration
func patchJobServiceRoot(c *gin.Context) {
	var req struct {
		ServiceCapabilities *struct {
			WorkerPoolSize *int `json:"WorkerPoolSize"`
		} `json:"ServiceCapabilities"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utility.RedfishError(c, http.StatusBadRequest,
			fmt.Sprintf("Invalid request body: %v", err),
			"InvalidJSON")
		return
	}

	// Check if WorkerPoolSize was provided
	if req.ServiceCapabilities != nil && req.ServiceCapabilities.WorkerPoolSize != nil {
		newSize := *req.ServiceCapabilities.WorkerPoolSize

		if err := JobService.SetWorkerPoolSize(newSize); err != nil {
			utility.RedfishError(c, http.StatusBadRequest,
				fmt.Sprintf("Invalid WorkerPoolSize: %v", err),
				"PropertyValueNotInList")
			return
		}
	}

	// Return updated JobService root
	getJobServiceRoot(c)
}

// GET /MultiFish/v1/JobService/Jobs - Get jobs collection
func getJobsCollection(c *gin.Context) {
	jobs := JobService.ListJobs()

	members := make([]gin.H, len(jobs))
	for i, job := range jobs {
		members[i] = gin.H{
			"@odata.id": fmt.Sprintf("/MultiFish/v1/JobService/Jobs/%s", job.ID),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"@odata.type":        "#JobCollection.JobCollection",
		"@odata.id":          "/MultiFish/v1/JobService/Jobs",
		"Name":               "Job Collection",
		"Members":            members,
		"Members@odata.count": len(members),
	})
}

// POST /MultiFish/v1/JobService/Jobs - Create a new job
func createJob(c *gin.Context) {
	var req scheduler.JobCreateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utility.RedfishError(c, http.StatusBadRequest,
			fmt.Sprintf("Invalid request body: %v", err),
			"InvalidJSON")
		return
	}

	// Create the job
	job, validationResp, err := JobService.CreateJob(&req)

	// If validation failed, return detailed error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "JobValidationFailed",
				"message": err.Error(),
				"@Message.ExtendedInfo": []gin.H{
					{
						"MessageId": "Base.1.0.JobValidationFailed",
						"Message":   validationResp.Message,
						"Severity":  "Critical",
						"ValidationDetails": gin.H{
							"ScheduleValid":  validationResp.ScheduleValid,
							"ScheduleErrors": validationResp.ScheduleErrors,
							"ActionValid":    validationResp.ActionValid,
							"ActionErrors":   validationResp.ActionErrors,
							"PayloadValid":   validationResp.PayloadValid,
							"PayloadErrors":  validationResp.PayloadErrors,
							"MachineResults": validationResp.MachineResults,
						},
					},
				},
			},
		})
		return
	}

	// Return created job
	c.JSON(http.StatusCreated, formatJobResponse(job))
}

// GET /MultiFish/v1/JobService/Jobs/:jobId - Get a specific job
func getJob(c *gin.Context) {
	jobID := c.Param("jobId")

	job, err := JobService.GetJob(jobID)
	if err != nil {
		utility.RedfishError(c, http.StatusNotFound,
			fmt.Sprintf("Job not found: %s", jobID),
			"ResourceNotFound")
		return
	}

	c.JSON(http.StatusOK, formatJobResponse(job))
}

// DELETE /MultiFish/v1/JobService/Jobs/:jobId - Delete a job
func deleteJob(c *gin.Context) {
	jobID := c.Param("jobId")

	if err := JobService.DeleteJob(jobID); err != nil {
		utility.RedfishError(c, http.StatusNotFound,
			fmt.Sprintf("Job not found: %s", jobID),
			"ResourceNotFound")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Job %s deleted successfully", jobID),
	})
}

// POST /MultiFish/v1/JobService/Jobs/:jobId/Actions/Cancel - Cancel a job
func cancelJob(c *gin.Context) {
	jobID := c.Param("jobId")

	if err := JobService.CancelJob(jobID); err != nil {
		utility.RedfishError(c, http.StatusNotFound,
			fmt.Sprintf("Job not found: %s", jobID),
			"ResourceNotFound")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Job %s cancelled successfully", jobID),
	})
}

// ========== Job Service Initialization ==========

// JobService is the global job service instance
var JobService *scheduler.JobService

// InitJobService initializes the job service
func InitJobService(cfg *config.Config) {
	log := utility.GetLogger()
	
	// Set logs directory from configuration
	if err := scheduler.SetLogsDir(cfg.LogsDir); err != nil {
		log.Warn().Err(err).Msg("Failed to set logs directory, using default")
	}
	
	// Create adapter for platform manager
	platformAdapter := NewPlatformManagerAdapter(PlatformMgr)
	
	// Create validator and executor
	validator := scheduler.NewPlatformValidator(platformAdapter)
	machineExecutorAdapter := &MachineActionExecutorAdapter{}
	actionExecutor := scheduler.NewDefaultActionExecutor(machineExecutorAdapter)
	executor := scheduler.NewPlatformExecutor(platformAdapter, actionExecutor)

	// Create job service with configured worker pool size
	JobService = scheduler.NewJobService(validator, executor)
	
	// Set worker pool size from configuration
	if err := JobService.SetWorkerPoolSize(cfg.WorkerPoolSize); err != nil {
		log.Warn().Err(err).Msg("Failed to set worker pool size")
	}
}

// ========== Job Service Routes ==========

func JobServiceRoutes(router *gin.Engine, cfg *config.Config) {
	// Initialize job service with configuration
	InitJobService(cfg)

	// JobService root
	router.GET("/MultiFish/v1/JobService", getJobServiceRoot)
	router.PATCH("/MultiFish/v1/JobService", patchJobServiceRoot)

	// Jobs collection
	router.GET("/MultiFish/v1/JobService/Jobs", getJobsCollection)
	router.POST("/MultiFish/v1/JobService/Jobs", createJob)

	// Individual job
	router.GET("/MultiFish/v1/JobService/Jobs/:jobId", getJob)
	router.DELETE("/MultiFish/v1/JobService/Jobs/:jobId", deleteJob)
	router.POST("/MultiFish/v1/JobService/Jobs/:jobId/Actions/Cancel", cancelJob)
}
