package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"multifish/config"
	"multifish/scheduler"
	extendprovider "multifish/providers/extend"
)

func setupJobServiceTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// Initialize platform manager (required for job service)
	PlatformMgr = &PlatformManager{
		machines: make(map[string]*MachineConnection),
	}

	// Add test machines
	testMachine1 := MachineConfig{
		ID:                "machine-1",
		Name:              "Test Machine 1",
		Type:              "Extend",
		Endpoint:          "https://127.0.0.1:8443",
		Username:          "admin",
		Password:          "password",
		Insecure:          true,
		HTTPClientTimeout: 30,
	}

	testMachine2 := MachineConfig{
	ID:                "machine-2",
	Name:              "Test Machine 2",
	Type:              "Extend",
	Endpoint:          "https://127.0.0.1:8444",
	Username:          "admin",
	Password:          "password",
	Insecure:          true,
	HTTPClientTimeout: 30,
}

// Note: We can't actually connect without real BMC endpoints,
// but we can test the validation logic
_ = testMachine1
_ = testMachine2

// Create test configuration
cfg := config.DefaultConfig()

// Setup job service routes
JobServiceRoutes(router, cfg)

return router
}

func TestGetJobServiceRoot(t *testing.T) {
	router := setupJobServiceTestRouter()

	req, _ := http.NewRequest("GET", "/MultiFish/v1/JobService", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "#JobService.v1_0_0.JobService", response["@odata.type"])
	assert.Equal(t, "/MultiFish/v1/JobService", response["@odata.id"])
	assert.Equal(t, "JobService", response["Id"])

	serviceCapabilities := response["ServiceCapabilities"].(map[string]interface{})
	assert.Equal(t, float64(99), serviceCapabilities["WorkerPoolSize"])
	assert.Equal(t, float64(0), serviceCapabilities["ActiveWorkers"])
	assert.Equal(t, float64(99), serviceCapabilities["AvailableWorkers"])
	assert.Equal(t, float64(0), serviceCapabilities["TotalJobs"])
	assert.Equal(t, float64(0), serviceCapabilities["RunningJobs"])
}

func TestPatchJobServiceRoot(t *testing.T) {
	router := setupJobServiceTestRouter()

	// Test updating WorkerPoolSize
	patchBody := map[string]interface{}{
		"ServiceCapabilities": map[string]interface{}{
			"WorkerPoolSize": 150,
		},
	}
	body, _ := json.Marshal(patchBody)

	req, _ := http.NewRequest("PATCH", "/MultiFish/v1/JobService", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	serviceCapabilities := response["ServiceCapabilities"].(map[string]interface{})
	assert.Equal(t, float64(150), serviceCapabilities["WorkerPoolSize"])
	assert.Equal(t, float64(150), serviceCapabilities["AvailableWorkers"])
}

func TestPatchJobServiceRoot_InvalidSize(t *testing.T) {
	router := setupJobServiceTestRouter()

	// Test with invalid WorkerPoolSize (0)
	patchBody := map[string]interface{}{
		"ServiceCapabilities": map[string]interface{}{
			"WorkerPoolSize": 0,
		},
	}
	body, _ := json.Marshal(patchBody)

	req, _ := http.NewRequest("PATCH", "/MultiFish/v1/JobService", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestPatchJobServiceRoot_LargeSize(t *testing.T) {
	router := setupJobServiceTestRouter()

	// Test with a very large WorkerPoolSize (no limit)
	patchBody := map[string]interface{}{
		"ServiceCapabilities": map[string]interface{}{
			"WorkerPoolSize": 10000,
		},
	}
	body, _ := json.Marshal(patchBody)

	req, _ := http.NewRequest("PATCH", "/MultiFish/v1/JobService", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	serviceCapabilities := response["ServiceCapabilities"].(map[string]interface{})
	assert.Equal(t, float64(10000), serviceCapabilities["WorkerPoolSize"])
}

func TestGetJobsCollectionEmpty(t *testing.T) {
	router := setupJobServiceTestRouter()

	req, _ := http.NewRequest("GET", "/MultiFish/v1/JobService/Jobs", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "#JobCollection.JobCollection", response["@odata.type"])
	assert.Equal(t, "/MultiFish/v1/JobService/Jobs", response["@odata.id"])

	members := response["Members"].([]interface{})
	assert.Equal(t, 0, len(members))
	assert.Equal(t, float64(0), response["Members@odata.count"])
}

func TestCreateJobValidation(t *testing.T) {
	router := setupJobServiceTestRouter()

	tests := []struct {
		name           string
		request        scheduler.JobCreateRequest
		expectError    bool
		errorContains  string
		checkSchedule  bool
		checkAction    bool
		checkPayload   bool
		checkMachines  bool
	}{
		{
			name: "Valid One-Time Job",
			request: scheduler.JobCreateRequest{
				Name:     "Test One-Time Job",
				Machines: []string{"machine-1"},
				Action:   scheduler.ActionPatchProfile,
				Payload:  []scheduler.ExecutePatchProfilePayload{{ManagerID: "bmc", Payload: extendprovider.PatchProfileType{Profile: "Performance"}}},
				Schedule: scheduler.Schedule{
					Type:   scheduler.ScheduleTypeOnce,
					Time:   "08:00:00",
					Period: nil,
				},
			},
			expectError: true, // Will fail because machines aren't actually connected
			checkMachines: true,
		},
		{
			name: "Invalid Action",
			request: scheduler.JobCreateRequest{
				Name:     "Invalid Action Job",
				Machines: []string{"machine-1"},
				Action:   "InvalidAction",
				Payload:  []scheduler.ExecutePatchProfilePayload{{ManagerID: "bmc", Payload: extendprovider.PatchProfileType{Profile: "Performance"}}},
				Schedule: scheduler.Schedule{
					Type:   scheduler.ScheduleTypeOnce,
					Time:   "08:00:00",
					Period: nil,
				},
			},
			expectError:   true,
			checkAction:   true,
		},
		{
			name: "Invalid Payload - Empty Profile",
			request: scheduler.JobCreateRequest{
				Name:     "Invalid Payload Job",
				Machines: []string{"machine-1"},
				Action:   scheduler.ActionPatchProfile,
				Payload:  []scheduler.ExecutePatchProfilePayload{{ManagerID: "bmc", Payload: extendprovider.PatchProfileType{Profile: ""}}},
				Schedule: scheduler.Schedule{
					Type:   scheduler.ScheduleTypeOnce,
					Time:   "08:00:00",
					Period: nil,
				},
			},
			expectError:   true,
			checkPayload:  true,
		},
		{
			name: "Invalid scheduler.Schedule - Wrong Time Format",
			request: scheduler.JobCreateRequest{
				Name:     "Invalid scheduler.Schedule Job",
				Machines: []string{"machine-1"},
				Action:   scheduler.ActionPatchProfile,
				Payload:  []scheduler.ExecutePatchProfilePayload{{ManagerID: "bmc", Payload: extendprovider.PatchProfileType{Profile: "Performance"}}},
				Schedule: scheduler.Schedule{
					Type:   scheduler.ScheduleTypeOnce,
					Time:   "25:00:00", // Invalid hour
					Period: nil,
				},
			},
			expectError:   true,
			checkSchedule: true,
		},
		{
			name: "Invalid scheduler.Schedule - Continuous without scheduler.Period",
			request: scheduler.JobCreateRequest{
				Name:     "Invalid Continuous Job",
				Machines: []string{"machine-1"},
				Action:   scheduler.ActionPatchProfile,
				Payload:  []scheduler.ExecutePatchProfilePayload{{ManagerID: "bmc", Payload: extendprovider.PatchProfileType{Profile: "Performance"}}},
				Schedule: scheduler.Schedule{
					Type:   scheduler.ScheduleTypeContinuous,
					Time:   "08:00:00",
					Period: nil, // Should not be nil for Continuous
				},
			},
			expectError:   true,
			checkSchedule: true,
		},
		{
			name: "Invalid scheduler.Schedule - Once with scheduler.Period",
			request: scheduler.JobCreateRequest{
				Name:     "Invalid Once Job",
				Machines: []string{"machine-1"},
				Action:   scheduler.ActionPatchProfile,
				Payload:  []scheduler.ExecutePatchProfilePayload{{ManagerID: "bmc", Payload: extendprovider.PatchProfileType{Profile: "Performance"}}},
				Schedule: scheduler.Schedule{
					Type: scheduler.ScheduleTypeOnce,
					Time: "08:00:00",
					Period: &scheduler.Period{
						DaysOfWeek: []scheduler.DayOfWeek{scheduler.Monday},
					},
				},
			},
			expectError:   true,
			checkSchedule: true,
		},
		{
			name: "Empty Machines",
			request: scheduler.JobCreateRequest{
				Name:     "No Machines Job",
				Machines: []string{},
				Action:   scheduler.ActionPatchProfile,
				Payload:  []scheduler.ExecutePatchProfilePayload{{ManagerID: "bmc", Payload: extendprovider.PatchProfileType{Profile: "Performance"}}},
				Schedule: scheduler.Schedule{
					Type:   scheduler.ScheduleTypeOnce,
					Time:   "08:00:00",
					Period: nil,
				},
			},
			expectError:   true,
			errorContains: "validation failed",
		},
		{
			name: "Duplicate Machines",
			request: scheduler.JobCreateRequest{
				Name:     "Duplicate Machines Job",
				Machines: []string{"machine-1", "machine-2", "machine-1"},
				Action:   scheduler.ActionPatchProfile,
				Payload:  []scheduler.ExecutePatchProfilePayload{{ManagerID: "bmc", Payload: extendprovider.PatchProfileType{Profile: "Performance"}}},
				Schedule: scheduler.Schedule{
					Type:   scheduler.ScheduleTypeOnce,
					Time:   "08:00:00",
					Period: nil,
				},
			},
			expectError:   true,
			errorContains: "duplicate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest("POST", "/MultiFish/v1/JobService/Jobs", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if tt.expectError {
				assert.Equal(t, http.StatusBadRequest, w.Code)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				errorObj := response["error"].(map[string]interface{})
				assert.Equal(t, "JobValidationFailed", errorObj["code"])

				if tt.errorContains != "" {
					message := errorObj["message"].(string)
					
					// Check extended info message for more specific error details
					if len(errorObj["@Message.ExtendedInfo"].([]interface{})) > 0 {
						extendedInfo := errorObj["@Message.ExtendedInfo"].([]interface{})[0].(map[string]interface{})
						validationMessage := extendedInfo["Message"].(string)
						// The specific validation message contains the details
						if !strings.Contains(message, tt.errorContains) {
							assert.Contains(t, validationMessage, tt.errorContains)
						} else {
							assert.Contains(t, message, tt.errorContains)
						}
					} else {
						assert.Contains(t, message, tt.errorContains)
					}
				}

				// Check specific validation details
				if len(errorObj["@Message.ExtendedInfo"].([]interface{})) > 0 {
					extendedInfo := errorObj["@Message.ExtendedInfo"].([]interface{})[0].(map[string]interface{})
					validationDetails := extendedInfo["ValidationDetails"].(map[string]interface{})

					if tt.checkSchedule {
						assert.False(t, validationDetails["ScheduleValid"].(bool))
					}
					if tt.checkAction {
						assert.False(t, validationDetails["ActionValid"].(bool))
					}
					if tt.checkPayload {
						assert.False(t, validationDetails["PayloadValid"].(bool))
					}
				}
			}
		})
	}
}

func TestCreateJobWithContinuousSchedule(t *testing.T) {
	router := setupJobServiceTestRouter()

	startDay := "2026-02-01"
	endDay := "2026-03-01"
	daysOfMonth := "1"

	request := scheduler.JobCreateRequest{
		Name:     "Continuous Job Test",
		Machines: []string{"machine-1"},
		Action:   scheduler.ActionPatchProfile,
		Payload:  []scheduler.ExecutePatchProfilePayload{{ManagerID: "bmc", Payload: extendprovider.PatchProfileType{Profile: "Performance"}}},
		Schedule: scheduler.Schedule{
			Type: scheduler.ScheduleTypeContinuous,
			Time: "08:00:00",
			Period: &scheduler.Period{
				StartDay:    &startDay,
				EndDay:      &endDay,
				DaysOfWeek:  []scheduler.DayOfWeek{scheduler.Monday, scheduler.Sunday},
				DaysOfMonth: &daysOfMonth,
			},
		},
	}

	jsonData, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/MultiFish/v1/JobService/Jobs", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Will fail because machines aren't actually connected, but we can verify the request structure
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// The error should be about machine validation, not schedule validation
	errorObj := response["error"].(map[string]interface{})
	assert.Equal(t, "JobValidationFailed", errorObj["code"])
}

func TestGetNonExistentJob(t *testing.T) {
	router := setupJobServiceTestRouter()

	req, _ := http.NewRequest("GET", "/MultiFish/v1/JobService/Jobs/non-existent-job", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	errorObj := response["error"].(map[string]interface{})
	// RedfishError adds "Base.1.0." prefix
	assert.Contains(t, errorObj["code"].(string), "ResourceNotFound")
}

func TestDeleteNonExistentJob(t *testing.T) {
	router := setupJobServiceTestRouter()

	req, _ := http.NewRequest("DELETE", "/MultiFish/v1/JobService/Jobs/non-existent-job", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCancelNonExistentJob(t *testing.T) {
	router := setupJobServiceTestRouter()

	req, _ := http.NewRequest("POST", "/MultiFish/v1/JobService/Jobs/non-existent-job/Actions/Cancel", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestJobValidation(t *testing.T) {
	tests := []struct {
		name        string
		request     scheduler.JobCreateRequest
		shouldFail  bool
		failureType string // "schedule", "action", "payload"
	}{
		{
			name: "Valid Continuous scheduler.Schedule with DaysOfWeek",
			request: scheduler.JobCreateRequest{
				Machines: []string{"machine-1"},
				Action:   scheduler.ActionPatchProfile,
				Payload:  []scheduler.ExecutePatchProfilePayload{{ManagerID: "bmc", Payload: extendprovider.PatchProfileType{Profile: "Performance"}}},
				Schedule: scheduler.Schedule{
					Type: scheduler.ScheduleTypeContinuous,
					Time: "08:00:00",
					Period: &scheduler.Period{
						DaysOfWeek: []scheduler.DayOfWeek{scheduler.Monday},
					},
				},
			},
			shouldFail: false,
		},
		{
			name: "Valid Continuous scheduler.Schedule with DaysOfMonth",
			request: scheduler.JobCreateRequest{
				Machines: []string{"machine-1"},
				Action:   scheduler.ActionPatchProfile,
				Payload:  []scheduler.ExecutePatchProfilePayload{{ManagerID: "bmc", Payload: extendprovider.PatchProfileType{Profile: "Balanced"}}},
				Schedule: scheduler.Schedule{
					Type: scheduler.ScheduleTypeContinuous,
					Time: "08:00:00",
					Period: &scheduler.Period{
						DaysOfMonth: stringPtr("1"),
					},
				},
			},
			shouldFail: false,
		},
		{
			name: "Invalid - Continuous without DaysOfWeek or DaysOfMonth",
			request: scheduler.JobCreateRequest{
				Machines: []string{"machine-1"},
				Action:   scheduler.ActionPatchProfile,
				Payload:  []scheduler.ExecutePatchProfilePayload{{ManagerID: "bmc", Payload: extendprovider.PatchProfileType{Profile: "Performance"}}},
				Schedule: scheduler.Schedule{
					Type: scheduler.ScheduleTypeContinuous,
					Time: "08:00:00",
					Period: &scheduler.Period{
						DaysOfWeek:  []scheduler.DayOfWeek{},
						DaysOfMonth: stringPtr(""),
					},
				},
			},
			shouldFail:  true,
			failureType: "schedule",
		},
		{
			name: "Invalid Profile",
			request: scheduler.JobCreateRequest{
				Machines: []string{"machine-1"},
				Action:   scheduler.ActionPatchProfile,
				Payload:  []scheduler.ExecutePatchProfilePayload{{ManagerID: "bmc", Payload: extendprovider.PatchProfileType{Profile: "InvalidProfile"}}},
				Schedule: scheduler.Schedule{
					Type:   scheduler.ScheduleTypeOnce,
					Time:   "08:00:00",
					Period: nil,
				},
			},
			shouldFail:  true,
			failureType: "payload",
		},
		{
			name: "Invalid - Duplicate Machines",
			request: scheduler.JobCreateRequest{
				Machines: []string{"machine-1", "machine-2", "machine-1", "machine-3", "machine-2"},
				Action:   scheduler.ActionPatchProfile,
				Payload:  []scheduler.ExecutePatchProfilePayload{{ManagerID: "bmc", Payload: extendprovider.PatchProfileType{Profile: "Performance"}}},
				Schedule: scheduler.Schedule{
					Type:   scheduler.ScheduleTypeOnce,
					Time:   "08:00:00",
					Period: nil,
				},
			},
			shouldFail:  true,
			failureType: "machines",
		},
		{
			name: "Valid - No Duplicate Machines",
			request: scheduler.JobCreateRequest{
				Machines: []string{"machine-1", "machine-2", "machine-3"},
				Action:   scheduler.ActionPatchProfile,
				Payload:  []scheduler.ExecutePatchProfilePayload{{ManagerID: "bmc", Payload: extendprovider.PatchProfileType{Profile: "Performance"}}},
				Schedule: scheduler.Schedule{
					Type:   scheduler.ScheduleTypeOnce,
					Time:   "08:00:00",
					Period: nil,
				},
			},
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validationResp := tt.request.Validate()

			if tt.shouldFail {
				assert.False(t, validationResp.Valid)
				switch tt.failureType {
				case "schedule":
					assert.False(t, validationResp.ScheduleValid)
				case "action":
					assert.False(t, validationResp.ActionValid)
				case "payload":
					assert.False(t, validationResp.PayloadValid)
				case "machines":
					// For duplicate machines, the error is in the general message
					assert.Contains(t, validationResp.Message, "duplicate machine IDs")
				}
			} else {
				assert.True(t, validationResp.Valid)
			}
		})
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
