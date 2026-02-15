package scheduler

import (
	"testing"
	"time"
	extendprovider "multifish/providers/extend"
)

func TestJobCreateRequestValidate(t *testing.T) {
	tests := []struct {
		name          string
		request       JobCreateRequest
		expectedValid bool
	}{
		{
			name: "Valid Once Schedule",
			request: JobCreateRequest{
				Name:     "Test Job",
				Machines: []string{"machine-1"},
				Action:   ActionPatchProfile,
				Payload: []ExecutePatchProfilePayload{
					{
						ManagerID: "bmc",
						Payload:   extendprovider.PatchProfileType{Profile: "Performance"},
					},
				},
				Schedule: Schedule{
					Type:   ScheduleTypeOnce,
					Time:   "08:00:00",
					Period: nil,
				},
			},
			expectedValid: true,
		},
		{
			name: "Invalid - No Machines",
			request: JobCreateRequest{
				Name:   "Test Job",
				Action: ActionPatchProfile,
				Payload: []ExecutePatchProfilePayload{
					{
						ManagerID: "bmc",
						Payload:   extendprovider.PatchProfileType{Profile: "Performance"},
					},
				},
				Schedule: Schedule{
					Type:   ScheduleTypeOnce,
					Time:   "08:00:00",
					Period: nil,
				},
			},
			expectedValid: false,
		},
		{
			name: "Invalid - Bad Time Format",
			request: JobCreateRequest{
				Machines: []string{"machine-1"},
				Action:   ActionPatchProfile,
				Payload: []ExecutePatchProfilePayload{
					{
						ManagerID: "bmc",
						Payload:   extendprovider.PatchProfileType{Profile: "Performance"},
					},
				},
				Schedule: Schedule{
					Type:   ScheduleTypeOnce,
					Time:   "25:00:00", // Invalid hour
					Period: nil,
				},
			},
			expectedValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.request.Validate()
			if result.Valid != tt.expectedValid {
				t.Errorf("Expected Valid=%v, got Valid=%v. Message: %s", 
					tt.expectedValid, result.Valid, result.Message)
			}
		})
	}
}

func TestJobToJSON(t *testing.T) {
	now := time.Now()
	job := &Job{
		ID:             "job-1",
		Name:           "Test Job",
		Machines:       []string{"machine-1"},
		Action:         ActionPatchProfile,
		Status:         JobStatusPending,
		CreatedTime:    now,
		ExecutionCount: 0,
	}

	json := job.ToJSON()
	if json == "" {
		t.Error("Expected non-empty JSON string")
	}
	if json == "{}" {
		t.Error("Expected JSON with content")
	}
}

func TestValidateProfilePayloads(t *testing.T) {
	tests := []struct {
		name        string
		payload     Payload
		expectError bool
	}{
		{
			name: "Valid Payload",
			payload: []ExecutePatchProfilePayload{
				{
					ManagerID: "bmc1",
					Payload:   extendprovider.PatchProfileType{Profile: "Performance"},
				},
			},
			expectError: false,
		},
		{
			name: "Invalid - Empty Profile",
			payload: []ExecutePatchProfilePayload{
				{
					ManagerID: "bmc1",
					Payload:   extendprovider.PatchProfileType{Profile: ""},
				},
			},
			expectError: true,
		},
		{
			name: "Invalid - Invalid Profile Value",
			payload: []ExecutePatchProfilePayload{
				{
					ManagerID: "bmc1",
					Payload:   extendprovider.PatchProfileType{Profile: "InvalidProfile"},
				},
			},
			expectError: true,
		},
		{
			name: "Invalid - Duplicate ManagerID",
			payload: []ExecutePatchProfilePayload{
				{
					ManagerID: "bmc1",
					Payload:   extendprovider.PatchProfileType{Profile: "Performance"},
				},
				{
					ManagerID: "bmc1",
					Payload:   extendprovider.PatchProfileType{Profile: "Balanced"},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProfilePayloads(tt.payload)
			if (err != nil) != tt.expectError {
				t.Errorf("Expected error=%v, got error=%v (%v)", 
					tt.expectError, err != nil, err)
			}
		})
	}
}
