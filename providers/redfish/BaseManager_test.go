package redfish

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stmcginnis/gofish/redfish"
	
	"multifish/utility"
)

// Helper function to create a redfish.Manager with ID set
func createManager(id string) *redfish.Manager {
	mgr := &redfish.Manager{}
	mgr.ID = id
	return mgr
}

// Helper function to create a full redfish.Manager with all fields
func createFullManager(id, name, fwVersion, uuid, svcID string) *redfish.Manager {
	mgr := &redfish.Manager{}
	mgr.ID = id
	mgr.Name = name
	mgr.FirmwareVersion = fwVersion
	mgr.UUID = uuid
	mgr.ServiceIdentification = svcID
	return mgr
}

// TestPatchManagerType tests the PatchManagerType structure
func TestPatchManagerType(t *testing.T) {
	tests := []struct {
		name                  string
		serviceIdentification string
	}{
		{"empty service ID", ""},
		{"valid service ID", "test-service-123"},
		{"service ID with spaces", "Service 123"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patch := PatchManagerType{
				ServiceIdentification: tt.serviceIdentification,
			}
			
			if patch.ServiceIdentification != tt.serviceIdentification {
				t.Errorf("ServiceIdentification = %v, want %v", 
					patch.ServiceIdentification, tt.serviceIdentification)
			}
		})
	}
}

// TestManagerAllowedPatchFields tests the ManagerAllowedPatchFields function
func TestManagerAllowedPatchFields(t *testing.T) {
	fields := ManagerAllowedPatchFields()
	
	if len(fields) != 1 {
		t.Errorf("len(fields) = %v, want 1", len(fields))
	}
	
	field, ok := fields["ServiceIdentification"]
	if !ok {
		t.Fatal("ServiceIdentification field not found")
	}
	
	if field.Name != "ServiceIdentification" {
		t.Errorf("field.Name = %v, want ServiceIdentification", field.Name)
	}
	
	if field.Expected != "string" {
		t.Errorf("field.Expected = %v, want string", field.Expected)
	}
}

// TestPatchManagerData tests the PatchManagerData function
func TestPatchManagerData(t *testing.T) {
	tests := []struct {
		name           string
		patch          *PatchManagerType
		expectError    bool
		expectedStatus int
	}{
		{
			name:        "nil patch",
			patch:       nil,
			expectError: true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "valid patch",
			patch: &PatchManagerType{
				ServiceIdentification: "new-service-id",
			},
			expectError: true, // Will fail without mock client
			expectedStatus: http.StatusInternalServerError,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a manager (without client, so Update will fail)
			mgr := &redfish.Manager{}
			mgr.ID = "test-manager"
			mgr.ServiceIdentification = "old-id"
			
			respErr := PatchManagerData(mgr, tt.patch)
			
			if tt.expectError {
				if respErr == nil {
					t.Error("Expected error, got nil")
				} else if respErr.StatusCode != tt.expectedStatus {
					t.Errorf("StatusCode = %v, want %v", respErr.StatusCode, tt.expectedStatus)
				}
			}
		})
	}
}

// TestRedfishProvider_TypeName tests the TypeName method
func TestRedfishProvider_TypeName(t *testing.T) {
	provider := &RedfishProvider{}
	
	typeName := provider.TypeName()
	if typeName != "Redfish" {
		t.Errorf("TypeName() = %v, want Redfish", typeName)
	}
}

// TestRedfishProvider_Supports tests the Supports method
func TestRedfishProvider_Supports(t *testing.T) {
	provider := &RedfishProvider{}
	
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{
			name:     "supports redfish.Manager",
			input:    &redfish.Manager{},
			expected: true,
		},
		{
			name:     "does not support string",
			input:    "test",
			expected: false,
		},
		{
			name:     "does not support int",
			input:    123,
			expected: false,
		},
		{
			name:     "does not support nil",
			input:    nil,
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.Supports(tt.input)
			if result != tt.expected {
				t.Errorf("Supports() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestRedfishProvider_SupportsCollection tests the SupportsCollection method
func TestRedfishProvider_SupportsCollection(t *testing.T) {
	provider := &RedfishProvider{}
	
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{
			name:     "supports redfish.Manager slice",
			input:    []*redfish.Manager{},
			expected: true,
		},
		{
			name:     "does not support single manager",
			input:    &redfish.Manager{},
			expected: false,
		},
		{
			name:     "does not support string slice",
			input:    []string{},
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.SupportsCollection(tt.input)
			if result != tt.expected {
				t.Errorf("SupportsCollection() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestRedfishProvider_GetManagerCollectionResponse tests the GetManagerCollectionResponse method
func TestRedfishProvider_GetManagerCollectionResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &RedfishProvider{}
	
	tests := []struct {
		name         string
		managers     interface{}
		machineID    string
		expectError  bool
		expectCount  int
	}{
		{
			name: "valid collection",
			managers: []*redfish.Manager{
				createManager("mgr1"),
				createManager("mgr2"),
			},
			machineID:   "machine1",
			expectError: false,
			expectCount: 2,
		},
		{
			name:        "empty collection",
			managers:    []*redfish.Manager{},
			machineID:   "machine1",
			expectError: false,
			expectCount: 0,
		},
		{
			name:        "invalid type",
			managers:    "invalid",
			machineID:   "machine1",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := provider.GetManagerCollectionResponse(tt.managers, tt.machineID)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				
				if resp["@odata.type"] != "#ManagerCollection.ManagerCollection" {
					t.Errorf("@odata.type = %v", resp["@odata.type"])
				}
				
				members, ok := resp["Members"].([]gin.H)
				if !ok {
					t.Fatal("Members is not []gin.H")
				}
				
				if len(members) != tt.expectCount {
					t.Errorf("len(Members) = %v, want %v", len(members), tt.expectCount)
				}
				
				odataID := fmt.Sprintf("/MultiFish/v1/Platform/%s/Managers", tt.machineID)
				if resp["@odata.id"] != odataID {
					t.Errorf("@odata.id = %v, want %v", resp["@odata.id"], odataID)
				}
			}
		})
	}
}

// TestRedfishProvider_GetManagerResponse tests the GetManagerResponse method
func TestRedfishProvider_GetManagerResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &RedfishProvider{}
	
	tests := []struct {
		name        string
		manager     interface{}
		machineID   string
		managerID   string
		expectError bool
	}{
		{
			name: "valid manager",
			manager: createFullManager("mgr1", "Test Manager", "1.0.0", "uuid-123", "service-1"),
			machineID:   "machine1",
			managerID:   "mgr1",
			expectError: false,
		},
		{
			name:        "invalid type",
			manager:     "invalid",
			machineID:   "machine1",
			managerID:   "mgr1",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, respErr := provider.GetManagerResponse(tt.manager, tt.machineID, tt.managerID)
			
			if tt.expectError {
				if respErr == nil {
					t.Error("Expected error, got nil")
				}
				if respErr.StatusCode != http.StatusInternalServerError {
					t.Errorf("StatusCode = %v, want %v", 
						respErr.StatusCode, http.StatusInternalServerError)
				}
			} else {
				if respErr != nil {
					t.Fatalf("Unexpected error: %v", respErr)
				}
				
				mgr := tt.manager.(*redfish.Manager)
				
				if resp["Id"] != mgr.ID {
					t.Errorf("Id = %v, want %v", resp["Id"], mgr.ID)
				}
				
				if resp["Name"] != mgr.Name {
					t.Errorf("Name = %v, want %v", resp["Name"], mgr.Name)
				}
				
				if resp["ServiceIdentification"] != mgr.ServiceIdentification {
					t.Errorf("ServiceIdentification = %v, want %v", 
						resp["ServiceIdentification"], mgr.ServiceIdentification)
				}
				
				expectedOdataID := fmt.Sprintf("/MultiFish/v1/Platform/%s/Managers/%s", 
					tt.machineID, tt.managerID)
				if resp["@odata.id"] != expectedOdataID {
					t.Errorf("@odata.id = %v, want %v", resp["@odata.id"], expectedOdataID)
				}
			}
		})
	}
}

// TestRedfishProvider_PatchManager tests the PatchManager method
func TestRedfishProvider_PatchManager(t *testing.T) {
	provider := &RedfishProvider{}
	
	tests := []struct {
		name           string
		manager        interface{}
		updates        interface{}
		expectError    bool
		expectedStatus int
	}{
		{
			name:           "invalid manager type",
			manager:        "invalid",
			updates:        &PatchManagerType{},
			expectError:    true,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "invalid patch type",
			manager:        &redfish.Manager{},
			updates:        "invalid",
			expectError:    true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "nil patch",
			manager:        &redfish.Manager{},
			updates:        (*PatchManagerType)(nil),
			expectError:    true,
			expectedStatus: http.StatusBadRequest,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			respErr := provider.PatchManager(tt.manager, tt.updates)
			
			if tt.expectError {
				if respErr == nil {
					t.Error("Expected error, got nil")
				}
				if respErr.StatusCode != tt.expectedStatus {
					t.Errorf("StatusCode = %v, want %v", 
						respErr.StatusCode, tt.expectedStatus)
				}
			}
		})
	}
}

// TestRedfishProvider_GetProfileResponse tests that base Redfish doesn't support profiles
func TestRedfishProvider_GetProfileResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &RedfishProvider{}
	
	mgr := createManager("mgr1")
	resp, respErr := provider.GetProfileResponse(mgr, "machine1", "mgr1")
	
	if respErr == nil {
		t.Error("Expected error for unsupported profile")
	}
	
	if respErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %v, want %v", respErr.StatusCode, http.StatusNotFound)
	}
	
	if resp != nil {
		t.Error("Response should be nil for unsupported feature")
	}
}

// TestRedfishProvider_PatchProfile tests that base Redfish doesn't support profile patching
func TestRedfishProvider_PatchProfile(t *testing.T) {
	provider := &RedfishProvider{}
	
	mgr := createManager("mgr1")
	respErr := provider.PatchProfile(mgr, map[string]interface{}{"Profile": "Performance"})
	
	if respErr == nil {
		t.Error("Expected error for unsupported profile patch")
	}
	
	if respErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %v, want %v", respErr.StatusCode, http.StatusNotFound)
	}
}

// TestRedfishProvider_GetFanControllerCollectionResponse tests unsupported fan controllers
func TestRedfishProvider_GetFanControllerCollectionResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &RedfishProvider{}
	
	mgr := createManager("mgr1")
	resp, respErr := provider.GetFanControllerCollectionResponse(mgr, "machine1", "mgr1")
	
	if respErr == nil {
		t.Error("Expected error for unsupported fan controllers")
	}
	
	if respErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %v, want %v", respErr.StatusCode, http.StatusNotFound)
	}
	
	if resp != nil {
		t.Error("Response should be nil for unsupported feature")
	}
}

// TestRedfishProvider_GetFanControllerResponse tests unsupported fan controller retrieval
func TestRedfishProvider_GetFanControllerResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &RedfishProvider{}
	
	mgr := createManager("mgr1")
	resp, respErr := provider.GetFanControllerResponse(mgr, "machine1", "mgr1", "fan1")
	
	if respErr == nil {
		t.Error("Expected error for unsupported fan controller")
	}
	
	if respErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %v, want %v", respErr.StatusCode, http.StatusNotFound)
	}
	
	if resp != nil {
		t.Error("Response should be nil for unsupported feature")
	}
}

// TestRedfishProvider_PatchFanController tests unsupported fan controller patching
func TestRedfishProvider_PatchFanController(t *testing.T) {
	provider := &RedfishProvider{}
	
	mgr := createManager("mgr1")
	respErr := provider.PatchFanController(mgr, "fan1", map[string]interface{}{})
	
	if respErr == nil {
		t.Error("Expected error for unsupported fan controller patch")
	}
	
	if respErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %v, want %v", respErr.StatusCode, http.StatusInternalServerError)
	}
}

// TestRedfishProvider_GetFanZoneCollectionResponse tests unsupported fan zones
func TestRedfishProvider_GetFanZoneCollectionResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &RedfishProvider{}
	
	mgr := createManager("mgr1")
	resp, respErr := provider.GetFanZoneCollectionResponse(mgr, "machine1", "mgr1")
	
	if respErr == nil {
		t.Error("Expected error for unsupported fan zones")
	}
	
	if respErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %v, want %v", respErr.StatusCode, http.StatusNotFound)
	}
	
	if resp != nil {
		t.Error("Response should be nil for unsupported feature")
	}
}

// TestRedfishProvider_GetFanZoneResponse tests unsupported fan zone retrieval
func TestRedfishProvider_GetFanZoneResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &RedfishProvider{}
	
	mgr := createManager("mgr1")
	resp, respErr := provider.GetFanZoneResponse(mgr, "machine1", "mgr1", "zone1")
	
	if respErr == nil {
		t.Error("Expected error for unsupported fan zone")
	}
	
	if respErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %v, want %v", respErr.StatusCode, http.StatusNotFound)
	}
	
	if resp != nil {
		t.Error("Response should be nil for unsupported feature")
	}
}

// TestRedfishProvider_PatchFanZone tests unsupported fan zone patching
func TestRedfishProvider_PatchFanZone(t *testing.T) {
	provider := &RedfishProvider{}
	
	mgr := createManager("mgr1")
	respErr := provider.PatchFanZone(mgr, "zone1", map[string]interface{}{})
	
	if respErr == nil {
		t.Error("Expected error for unsupported fan zone patch")
	}
	
	if respErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %v, want %v", respErr.StatusCode, http.StatusInternalServerError)
	}
}

// TestRedfishProvider_GetPidControllerCollectionResponse tests unsupported PID controllers
func TestRedfishProvider_GetPidControllerCollectionResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &RedfishProvider{}
	
	mgr := createManager("mgr1")
	resp, respErr := provider.GetPidControllerCollectionResponse(mgr, "machine1", "mgr1")
	
	if respErr == nil {
		t.Error("Expected error for unsupported PID controllers")
	}
	
	if respErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %v, want %v", respErr.StatusCode, http.StatusNotFound)
	}
	
	if resp != nil {
		t.Error("Response should be nil for unsupported feature")
	}
}

// TestRedfishProvider_GetPidControllerResponse tests unsupported PID controller retrieval
func TestRedfishProvider_GetPidControllerResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &RedfishProvider{}
	
	mgr := createManager("mgr1")
	resp, respErr := provider.GetPidControllerResponse(mgr, "machine1", "mgr1", "pid1")
	
	if respErr == nil {
		t.Error("Expected error for unsupported PID controller")
	}
	
	if respErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %v, want %v", respErr.StatusCode, http.StatusNotFound)
	}
	
	if resp != nil {
		t.Error("Response should be nil for unsupported feature")
	}
}

// TestRedfishProvider_PatchPidController tests unsupported PID controller patching
func TestRedfishProvider_PatchPidController(t *testing.T) {
	provider := &RedfishProvider{}
	
	mgr := createManager("mgr1")
	respErr := provider.PatchPidController(mgr, "pid1", map[string]interface{}{})
	
	if respErr == nil {
		t.Error("Expected error for unsupported PID controller patch")
	}
	
	if respErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %v, want %v", respErr.StatusCode, http.StatusInternalServerError)
	}
}

// TestRedfishProvider_AllUnsupportedFeatures tests that all extended features return appropriate errors
func TestRedfishProvider_AllUnsupportedFeatures(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &RedfishProvider{}
	mgr := createManager("mgr1")
	
	unsupportedTests := []struct {
		name   string
		testFn func() *utility.ResponseError
	}{
		{
			name:   "PatchProfile",
			testFn: func() *utility.ResponseError { 
				return provider.PatchProfile(mgr, nil) 
			},
		},
		{
			name:   "PatchFanController",
			testFn: func() *utility.ResponseError { 
				return provider.PatchFanController(mgr, "fan1", nil) 
			},
		},
		{
			name:   "PatchFanZone",
			testFn: func() *utility.ResponseError { 
				return provider.PatchFanZone(mgr, "zone1", nil) 
			},
		},
		{
			name:   "PatchPidController",
			testFn: func() *utility.ResponseError { 
				return provider.PatchPidController(mgr, "pid1", nil) 
			},
		},
	}
	
	for _, tt := range unsupportedTests {
		t.Run(tt.name, func(t *testing.T) {
			respErr := tt.testFn()
			if respErr == nil {
				t.Errorf("%s should return error for unsupported feature", tt.name)
			}
		})
	}
}

// TestRedfishProvider_ManagerResponseFields tests all fields in manager response
func TestRedfishProvider_ManagerResponseFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &RedfishProvider{}
	
	mgr := &redfish.Manager{}
	mgr.ID = "test-mgr"
	mgr.Name = "Test Manager"
	mgr.FirmwareVersion = "2.0.1"
	mgr.UUID = "uuid-456"
	mgr.Model = "Model X"
	mgr.DateTime = "2026-02-04T00:00:00Z"
	mgr.ServiceIdentification = "svc-123"
	
	resp, respErr := provider.GetManagerResponse(mgr, "machine1", "test-mgr")
	if respErr != nil {
		t.Fatalf("Unexpected error: %v", respErr)
	}
	
	expectedFields := map[string]interface{}{
		"Id":                    "test-mgr",
		"Name":                  "Test Manager",
		"FirmwareVersion":       "2.0.1",
		"UUID":                  "uuid-456",
		"Model":                 "Model X",
		"DateTime":              "2026-02-04T00:00:00Z",
		"ServiceIdentification": "svc-123",
	}
	
	for field, expectedValue := range expectedFields {
		if resp[field] != expectedValue {
			t.Errorf("%s = %v, want %v", field, resp[field], expectedValue)
		}
	}
}

// TestUpdateAndRefreshManager tests manager update behavior
func TestUpdateAndRefreshManager(t *testing.T) {
	// Create a manager (without client, so Update will fail)
	mgr := &redfish.Manager{}
	mgr.ID = "test-manager"
	mgr.ServiceIdentification = "old-id"
	
	// This should fail because there's no client
	err := UpdateAndRefreshManager(mgr)
	if err == nil {
		t.Error("Expected error when updating manager without client")
	}
}

// TestPatchManagerData_ValidatesNilPatch tests nil patch validation
func TestPatchManagerData_ValidatesNilPatch(t *testing.T) {
	mgr := createManager("test")
	
	respErr := PatchManagerData(mgr, nil)
	if respErr == nil {
		t.Error("Expected error for nil patch")
	}
	
	if respErr.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %v, want %v", respErr.StatusCode, http.StatusBadRequest)
	}
	
	if respErr.Message != "InvalidRequest" {
		t.Errorf("Message = %v, want InvalidRequest", respErr.Message)
	}
}
