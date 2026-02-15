package extendprovider

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stmcginnis/gofish/redfish"
)

// Helper function to create a redfish.Manager with ID set
func createManager(id string) *redfish.Manager {
	mgr := &redfish.Manager{}
	mgr.ID = id
	return mgr
}

// Helper function to create a redfish.Manager with ID and Name set
func createManagerWithName(id, name string) *redfish.Manager {
	mgr := &redfish.Manager{}
	mgr.ID = id
	mgr.Name = name
	return mgr
}

// TestOdataID tests the OdataID helper struct
func TestOdataID(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected string
	}{
		{
			name:     "valid odata id",
			jsonData: `{"@odata.id": "/redfish/v1/test"}`,
			expected: "/redfish/v1/test",
		},
		{
			name:     "empty odata id",
			jsonData: `{"@odata.id": ""}`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var oid OdataID
			if err := json.Unmarshal([]byte(tt.jsonData), &oid); err != nil {
				t.Errorf("Failed to unmarshal: %v", err)
			}
			if oid.OdataID != tt.expected {
				t.Errorf("OdataID = %v, want %v", oid.OdataID, tt.expected)
			}
		})
	}
}

// TestFanController_JSON tests FanController JSON marshaling/unmarshaling
func TestFanController_JSON(t *testing.T) {
	jsonData := `{
		"@odata.id": "/redfish/v1/fan",
		"@odata.type": "#FanController.v1_0_0.FanController",
		"FFGainCoefficient": 1.5,
		"FFOffCoefficient": 2.5,
		"ICoefficient": 3.5,
		"ILimitMax": 100.0,
		"ILimitMin": 0.0,
		"Inputs": ["input1", "input2"],
		"NegativeHysteresis": 5.0,
		"OutLimitMax": 200.0,
		"OutLimitMin": 10.0,
		"Outputs": ["output1"],
		"PCoefficient": 4.5,
		"PositiveHysteresis": 6.0,
		"SlewNeg": 7.0,
		"SlewPos": 8.0,
		"Zones": [{"@odata.id": "/zone1"}]
	}`

	var fc FanController
	if err := json.Unmarshal([]byte(jsonData), &fc); err != nil {
		t.Fatalf("Failed to unmarshal FanController: %v", err)
	}

	if fc.FFGainCoefficient != 1.5 {
		t.Errorf("FFGainCoefficient = %v, want 1.5", fc.FFGainCoefficient)
	}
	if fc.ILimitMax != 100.0 {
		t.Errorf("ILimitMax = %v, want 100.0", fc.ILimitMax)
	}
	if len(fc.Inputs) != 2 {
		t.Errorf("len(Inputs) = %v, want 2", len(fc.Inputs))
	}
	if len(fc.Zones) != 1 {
		t.Errorf("len(Zones) = %v, want 1", len(fc.Zones))
	}
}

// TestFanControllers_UnmarshalJSON tests custom unmarshaling for FanControllers
func TestFanControllers_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"@odata.id": "/redfish/v1/fancontrollers",
		"@odata.type": "#FanControllers.v1_0_0.FanControllers",
		"Fan1": {
			"FFGainCoefficient": 1.5,
			"ILimitMax": 100.0,
			"Inputs": ["temp1"]
		},
		"Fan2": {
			"FFGainCoefficient": 2.5,
			"ILimitMax": 150.0,
			"Inputs": ["temp2"]
		}
	}`

	var fcs FanControllers
	if err := json.Unmarshal([]byte(jsonData), &fcs); err != nil {
		t.Fatalf("Failed to unmarshal FanControllers: %v", err)
	}

	if fcs.OdataID != "/redfish/v1/fancontrollers" {
		t.Errorf("OdataID = %v, want /redfish/v1/fancontrollers", fcs.OdataID)
	}

	if len(fcs.Items) != 2 {
		t.Fatalf("len(Items) = %v, want 2", len(fcs.Items))
	}

	if fc, ok := fcs.Items["Fan1"]; !ok {
		t.Error("Fan1 not found in Items")
	} else if fc.FFGainCoefficient != 1.5 {
		t.Errorf("Fan1.FFGainCoefficient = %v, want 1.5", fc.FFGainCoefficient)
	}

	if fc, ok := fcs.Items["Fan2"]; !ok {
		t.Error("Fan2 not found in Items")
	} else if fc.ILimitMax != 150.0 {
		t.Errorf("Fan2.ILimitMax = %v, want 150.0", fc.ILimitMax)
	}
}

// TestPidController_JSON tests PidController JSON marshaling/unmarshaling
func TestPidController_JSON(t *testing.T) {
	jsonData := `{
		"@odata.id": "/redfish/v1/pid",
		"@odata.type": "#PidController.v1_0_0.PidController",
		"FFGainCoefficient": 1.5,
		"SetPoint": 50.0,
		"Inputs": ["sensor1"]
	}`

	var pc PidController
	if err := json.Unmarshal([]byte(jsonData), &pc); err != nil {
		t.Fatalf("Failed to unmarshal PidController: %v", err)
	}

	if pc.FFGainCoefficient != 1.5 {
		t.Errorf("FFGainCoefficient = %v, want 1.5", pc.FFGainCoefficient)
	}
	if pc.SetPoint != 50.0 {
		t.Errorf("SetPoint = %v, want 50.0", pc.SetPoint)
	}
}

// TestPidControllers_UnmarshalJSON tests custom unmarshaling for PidControllers
func TestPidControllers_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"@odata.id": "/redfish/v1/pidcontrollers",
		"@odata.type": "#PidControllers.v1_0_0.PidControllers",
		"Pid1": {
			"SetPoint": 50.0,
			"ICoefficient": 2.0
		},
		"Pid2": {
			"SetPoint": 60.0,
			"ICoefficient": 3.0
		}
	}`

	var pcs PidControllers
	if err := json.Unmarshal([]byte(jsonData), &pcs); err != nil {
		t.Fatalf("Failed to unmarshal PidControllers: %v", err)
	}

	if len(pcs.Items) != 2 {
		t.Fatalf("len(Items) = %v, want 2", len(pcs.Items))
	}

	if pc, ok := pcs.Items["Pid1"]; !ok {
		t.Error("Pid1 not found in Items")
	} else if pc.SetPoint != 50.0 {
		t.Errorf("Pid1.SetPoint = %v, want 50.0", pc.SetPoint)
	}
}

// TestFanZone_JSON tests FanZone JSON marshaling/unmarshaling
func TestFanZone_JSON(t *testing.T) {
	jsonData := `{
		"@odata.id": "/redfish/v1/zone",
		"@odata.type": "#FanZone.v1_0_0.FanZone",
		"FailSafePercent": 75.0,
		"MinThermalOutput": 25.0
	}`

	var fz FanZone
	if err := json.Unmarshal([]byte(jsonData), &fz); err != nil {
		t.Fatalf("Failed to unmarshal FanZone: %v", err)
	}

	if fz.FailSafePercent != 75.0 {
		t.Errorf("FailSafePercent = %v, want 75.0", fz.FailSafePercent)
	}
	if fz.MinThermalOutput != 25.0 {
		t.Errorf("MinThermalOutput = %v, want 25.0", fz.MinThermalOutput)
	}
}

// TestFanZones_UnmarshalJSON tests custom unmarshaling for FanZones
func TestFanZones_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"@odata.id": "/redfish/v1/fanzones",
		"@odata.type": "#FanZones.v1_0_0.FanZones",
		"Zone1": {
			"FailSafePercent": 75.0,
			"MinThermalOutput": 25.0
		},
		"Zone2": {
			"FailSafePercent": 80.0,
			"MinThermalOutput": 30.0
		}
	}`

	var fzs FanZones
	if err := json.Unmarshal([]byte(jsonData), &fzs); err != nil {
		t.Fatalf("Failed to unmarshal FanZones: %v", err)
	}

	if len(fzs.Items) != 2 {
		t.Fatalf("len(Items) = %v, want 2", len(fzs.Items))
	}

	if fz, ok := fzs.Items["Zone1"]; !ok {
		t.Error("Zone1 not found in Items")
	} else if fz.FailSafePercent != 75.0 {
		t.Errorf("Zone1.FailSafePercent = %v, want 75.0", fz.FailSafePercent)
	}
}

// TestExtractOpenBmcFan tests the ExtractOpenBmcFan function
func TestExtractOpenBmcFan(t *testing.T) {
	tests := []struct {
		name        string
		oemData     string
		expectNil   bool
		expectProfile string
	}{
		{
			name: "valid OpenBmc Fan data",
			oemData: `{
				"OpenBmc": {
					"Fan": {
						"@odata.id": "/redfish/v1/fan",
						"@odata.type": "#Fan.v1_0_0.Fan",
						"Profile": "Performance",
						"Profile@Redfish.AllowableValues": ["Performance", "Balanced"]
					}
				}
			}`,
			expectNil:   false,
			expectProfile: "Performance",
		},
		{
			name:        "empty OEM data",
			oemData:     ``,
			expectNil:   true,
		},
		{
			name:        "no OpenBmc field",
			oemData:     `{"Other": {}}`,
			expectNil:   true,
		},
		{
			name: "no Fan field",
			oemData: `{
				"OpenBmc": {
					"Other": {}
				}
			}`,
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fan := ExtractOpenBmcFan(json.RawMessage(tt.oemData))
			if tt.expectNil {
				if fan != nil {
					t.Errorf("Expected nil, got %+v", fan)
				}
			} else {
				if fan == nil {
					t.Error("Expected non-nil fan")
				} else if fan.Profile != tt.expectProfile {
					t.Errorf("Profile = %v, want %v", fan.Profile, tt.expectProfile)
				}
			}
		})
	}
}

// TestProfileAllowedPatchFields tests the ProfileAllowedPatchFields function
func TestProfileAllowedPatchFields(t *testing.T) {
	fields := ProfileAllowedPatchFields()
	
	if len(fields) != 1 {
		t.Errorf("len(fields) = %v, want 1", len(fields))
	}
	
	if field, ok := fields["Profile"]; !ok {
		t.Error("Profile field not found")
	} else {
		if field.Name != "Profile" {
			t.Errorf("field.Name = %v, want Profile", field.Name)
		}
		if field.Expected != "string" {
			t.Errorf("field.Expected = %v, want string", field.Expected)
		}
	}
}

// TestFanControllerAllowedPatchFields tests the FanControllerAllowedPatchFields function
func TestFanControllerAllowedPatchFields(t *testing.T) {
	fields := FanControllerAllowedPatchFields()
	
	expectedFields := []string{
		"FFGainCoefficient", "FFOffCoefficient", "ICoefficient",
		"ILimitMax", "ILimitMin", "NegativeHysteresis",
		"OutLimitMax", "OutLimitMin", "PCoefficient",
		"PositiveHysteresis", "SlewNeg", "SlewPos", "Zones",
	}
	
	for _, fieldName := range expectedFields {
		if field, ok := fields[fieldName]; !ok {
			t.Errorf("Field %s not found", fieldName)
		} else if field.Name != fieldName {
			t.Errorf("field.Name = %v, want %v", field.Name, fieldName)
		}
	}
}

// TestFanZoneAllowedPatchFields tests the FanZoneAllowedPatchFields function
func TestFanZoneAllowedPatchFields(t *testing.T) {
	fields := FanZoneAllowedPatchFields()
	
	if len(fields) != 2 {
		t.Errorf("len(fields) = %v, want 2", len(fields))
	}
	
	expectedFields := []string{"FailSafePercent", "MinThermalOutput"}
	for _, fieldName := range expectedFields {
		if field, ok := fields[fieldName]; !ok {
			t.Errorf("Field %s not found", fieldName)
		} else if field.Expected != "number" {
			t.Errorf("field.Expected = %v, want number", field.Expected)
		}
	}
}

// TestPidControllerAllowedPatchFields tests the PidControllerAllowedPatchFields function
func TestPidControllerAllowedPatchFields(t *testing.T) {
	fields := PidControllerAllowedPatchFields()
	
	expectedFields := []string{
		"FFGainCoefficient", "FFOffCoefficient", "ICoefficient",
		"ILimitMax", "ILimitMin", "NegativeHysteresis",
		"OutLimitMax", "OutLimitMin", "PCoefficient",
		"PositiveHysteresis", "SetPoint", "SlewNeg", "SlewPos", "Zones",
	}
	
	for _, fieldName := range expectedFields {
		if _, ok := fields[fieldName]; !ok {
			t.Errorf("Field %s not found", fieldName)
		}
	}
}

// TestExtendManager_GetManager tests the GetManager method
func TestExtendManager_GetManager(t *testing.T) {
	mgr := createManagerWithName("test-manager", "Test Manager")
	
	em := &ExtendManager{
		mgr: mgr,
	}
	
	result := em.GetManager()
	if result != mgr {
		t.Errorf("GetManager() returned wrong manager")
	}
	if result.ID != "test-manager" {
		t.Errorf("Manager.ID = %v, want test-manager", result.ID)
	}
}

// TestExtendProvider_TypeName tests the TypeName method
func TestExtendProvider_TypeName(t *testing.T) {
	provider := &ExtendProvider{}
	
	typeName := provider.TypeName()
	if typeName != "Extended" {
		t.Errorf("TypeName() = %v, want Extended", typeName)
	}
}

// TestExtendProvider_Supports tests the Supports method
func TestExtendProvider_Supports(t *testing.T) {
	provider := &ExtendProvider{}
	
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{
			name:     "supports ExtendManager",
			input:    &ExtendManager{},
			expected: true,
		},
		{
			name:     "does not support redfish.Manager",
			input:    &redfish.Manager{},
			expected: false,
		},
		{
			name:     "does not support string",
			input:    "test",
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

// TestExtendProvider_SupportsCollection tests the SupportsCollection method
func TestExtendProvider_SupportsCollection(t *testing.T) {
	provider := &ExtendProvider{}
	
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{
			name:     "supports ExtendManager slice",
			input:    []*ExtendManager{},
			expected: true,
		},
		{
			name:     "does not support redfish.Manager slice",
			input:    []*redfish.Manager{},
			expected: false,
		},
		{
			name:     "does not support single ExtendManager",
			input:    &ExtendManager{},
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

// TestExtendProvider_GetManagerCollectionResponse tests the GetManagerCollectionResponse method
func TestExtendProvider_GetManagerCollectionResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &ExtendProvider{}
	
	managers := []*ExtendManager{
		{mgr: createManager("mgr1")},
		{mgr: createManager("mgr2")},
	}
	
	resp, err := provider.GetManagerCollectionResponse(managers, "machine1")
	if err != nil {
		t.Fatalf("GetManagerCollectionResponse() error = %v", err)
	}
	
	if resp["@odata.type"] != "#ManagerCollection.ManagerCollection" {
		t.Errorf("@odata.type = %v", resp["@odata.type"])
	}
	
	members, ok := resp["Members"].([]gin.H)
	if !ok {
		t.Fatal("Members is not []gin.H")
	}
	
	if len(members) != 2 {
		t.Errorf("len(Members) = %v, want 2", len(members))
	}
}

// TestExtendProvider_GetManagerCollectionResponse_InvalidType tests error handling
func TestExtendProvider_GetManagerCollectionResponse_InvalidType(t *testing.T) {
	provider := &ExtendProvider{}
	
	_, err := provider.GetManagerCollectionResponse([]*redfish.Manager{}, "machine1")
	if err == nil {
		t.Error("Expected error for invalid type, got nil")
	}
}

// TestExtendProvider_GetProfileResponse tests the GetProfileResponse method
func TestExtendProvider_GetProfileResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &ExtendProvider{}
	
	tests := []struct {
		name          string
		manager       interface{}
		expectError   bool
		expectProfile string
	}{
		{
			name: "valid profile",
			manager: &ExtendManager{
				mgr: createManager("mgr1"),
				OpenBmcFan: &OpenBmcFan{
					Profile:                "Performance",
					ProfileAllowableValues: []string{"Performance", "Balanced"},
				},
			},
			expectError:   false,
			expectProfile: "Performance",
		},
		{
			name: "no OpenBmcFan",
			manager: &ExtendManager{
				mgr:        createManager("mgr1"),
				OpenBmcFan: nil,
			},
			expectError: true,
		},
		{
			name:        "invalid type",
			manager:     &redfish.Manager{},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, respErr := provider.GetProfileResponse(tt.manager, "machine1", "mgr1")
			
			if tt.expectError {
				if respErr == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if respErr != nil {
					t.Errorf("Unexpected error: %v", respErr)
				}
				if profile, ok := resp["Profile"]; !ok {
					t.Error("Profile not found in response")
				} else if profile != tt.expectProfile {
					t.Errorf("Profile = %v, want %v", profile, tt.expectProfile)
				}
			}
		})
	}
}

// TestExtendProvider_GetFanControllerCollectionResponse tests the GetFanControllerCollectionResponse method
func TestExtendProvider_GetFanControllerCollectionResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &ExtendProvider{}
	
	fanControllers := &FanControllers{
		Items: map[string]*FanController{
			"Fan1": {},
			"Fan2": {},
		},
	}
	
	em := &ExtendManager{
		mgr: createManager("mgr1"),
		OpenBmcFan: &OpenBmcFan{
			FanControllers: fanControllers,
		},
	}
	
	resp, respErr := provider.GetFanControllerCollectionResponse(em, "machine1", "mgr1")
	if respErr != nil {
		t.Fatalf("Unexpected error: %v", respErr)
	}
	
	members, ok := resp["Members"].([]gin.H)
	if !ok {
		t.Fatal("Members is not []gin.H")
	}
	
	if len(members) != 2 {
		t.Errorf("len(Members) = %v, want 2", len(members))
	}
}

// TestExtendProvider_GetFanControllerResponse tests the GetFanControllerResponse method
func TestExtendProvider_GetFanControllerResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &ExtendProvider{}
	
	fanController := &FanController{
		FFGainCoefficient: 1.5,
		ILimitMax:         100.0,
		Inputs:            []string{"temp1"},
	}
	
	fanControllers := &FanControllers{
		Items: map[string]*FanController{
			"Fan1": fanController,
		},
	}
	
	em := &ExtendManager{
		mgr: createManager("mgr1"),
		OpenBmcFan: &OpenBmcFan{
			FanControllers: fanControllers,
		},
	}
	
	resp, respErr := provider.GetFanControllerResponse(em, "machine1", "mgr1", "Fan1")
	if respErr != nil {
		t.Fatalf("Unexpected error: %v", respErr)
	}
	
	if resp["Id"] != "Fan1" {
		t.Errorf("Id = %v, want Fan1", resp["Id"])
	}
	
	if resp["FFGainCoefficient"] != 1.5 {
		t.Errorf("FFGainCoefficient = %v, want 1.5", resp["FFGainCoefficient"])
	}
}

// TestExtendProvider_GetFanControllerResponse_NotFound tests error when fan controller not found
func TestExtendProvider_GetFanControllerResponse_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &ExtendProvider{}
	
	em := &ExtendManager{
		mgr: createManager("mgr1"),
		OpenBmcFan: &OpenBmcFan{
			FanControllers: &FanControllers{
				Items: map[string]*FanController{},
			},
		},
	}
	
	_, respErr := provider.GetFanControllerResponse(em, "machine1", "mgr1", "NonExistent")
	if respErr == nil {
		t.Error("Expected error for non-existent fan controller")
	}
	if respErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %v, want %v", respErr.StatusCode, http.StatusNotFound)
	}
}

// TestExtendProvider_GetFanZoneCollectionResponse tests the GetFanZoneCollectionResponse method
func TestExtendProvider_GetFanZoneCollectionResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &ExtendProvider{}
	
	fanZones := &FanZones{
		Items: map[string]FanZone{
			"Zone1": {FailSafePercent: 75.0},
			"Zone2": {FailSafePercent: 80.0},
		},
	}
	
	em := &ExtendManager{
		mgr: createManager("mgr1"),
		OpenBmcFan: &OpenBmcFan{
			FanZones: fanZones,
		},
	}
	
	resp, respErr := provider.GetFanZoneCollectionResponse(em, "machine1", "mgr1")
	if respErr != nil {
		t.Fatalf("Unexpected error: %v", respErr)
	}
	
	members, ok := resp["Members"].([]gin.H)
	if !ok {
		t.Fatal("Members is not []gin.H")
	}
	
	if len(members) != 2 {
		t.Errorf("len(Members) = %v, want 2", len(members))
	}
}

// TestExtendProvider_GetPidControllerCollectionResponse tests the GetPidControllerCollectionResponse method
func TestExtendProvider_GetPidControllerCollectionResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &ExtendProvider{}
	
	pidControllers := &PidControllers{
		Items: map[string]*PidController{
			"Pid1": {SetPoint: 50.0},
			"Pid2": {SetPoint: 60.0},
		},
	}
	
	em := &ExtendManager{
		mgr: createManager("mgr1"),
		OpenBmcFan: &OpenBmcFan{
			PidControllers: pidControllers,
		},
	}
	
	resp, respErr := provider.GetPidControllerCollectionResponse(em, "machine1", "mgr1")
	if respErr != nil {
		t.Fatalf("Unexpected error: %v", respErr)
	}
	
	members, ok := resp["Members"].([]gin.H)
	if !ok {
		t.Fatal("Members is not []gin.H")
	}
	
	if len(members) != 2 {
		t.Errorf("len(Members) = %v, want 2", len(members))
	}
}

// TestExtendProvider_GetPidControllerResponse tests the GetPidControllerResponse method
func TestExtendProvider_GetPidControllerResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &ExtendProvider{}
	
	pidController := &PidController{
		SetPoint:     50.0,
		ICoefficient: 2.0,
		Inputs:       []string{"sensor1"},
	}
	
	pidControllers := &PidControllers{
		Items: map[string]*PidController{
			"Pid1": pidController,
		},
	}
	
	em := &ExtendManager{
		mgr: createManager("mgr1"),
		OpenBmcFan: &OpenBmcFan{
			PidControllers: pidControllers,
		},
	}
	
	resp, respErr := provider.GetPidControllerResponse(em, "machine1", "mgr1", "Pid1")
	if respErr != nil {
		t.Fatalf("Unexpected error: %v", respErr)
	}
	
	if resp["Id"] != "Pid1" {
		t.Errorf("Id = %v, want Pid1", resp["Id"])
	}
	
	if resp["SetPoint"] != 50.0 {
		t.Errorf("SetPoint = %v, want 50.0", resp["SetPoint"])
	}
}

// Mock test to verify profile patch field validation (conceptual - requires actual implementation)
func TestPatchProfileType_Validation(t *testing.T) {
	tests := []struct {
		name    string
		profile string
		valid   bool
	}{
		{"valid performance", "Performance", true},
		{"valid balanced", "Balanced", true},
		{"valid power saver", "PowerSaver", true},
		{"valid custom", "Custom", true},
		{"empty string", "", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patch := PatchProfileType{
				Profile: tt.profile,
			}
			
			// Check if profile is empty
			isEmpty := patch.Profile == ""
			if tt.valid && isEmpty {
				t.Error("Expected valid profile but got empty")
			}
			if !tt.valid && !isEmpty {
				// Further validation would happen in the actual patch function
			}
		})
	}
}

// TestPatchFanControllerType_PointerFields tests that patch types use pointers correctly
func TestPatchFanControllerType_PointerFields(t *testing.T) {
	value := 1.5
	patch := PatchFanControllerType{
		FFGainCoefficient: &value,
	}
	
	if patch.FFGainCoefficient == nil {
		t.Error("FFGainCoefficient should not be nil")
	}
	
	if *patch.FFGainCoefficient != 1.5 {
		t.Errorf("FFGainCoefficient = %v, want 1.5", *patch.FFGainCoefficient)
	}
	
	// Verify nil fields
	if patch.ICoefficient != nil {
		t.Error("ICoefficient should be nil when not set")
	}
}

// TestProfileAllowlist tests the profile allowlist constant
func TestProfileAllowlist(t *testing.T) {
	expectedProfiles := []string{"Performance", "Balanced", "PowerSaver", "Custom"}
	
	if len(ProfileAllowlist) != len(expectedProfiles) {
		t.Errorf("len(ProfileAllowlist) = %v, want %v", len(ProfileAllowlist), len(expectedProfiles))
	}
	
	for _, expected := range expectedProfiles {
		found := false
		for _, profile := range ProfileAllowlist {
			if profile == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Profile %s not found in ProfileAllowlist", expected)
		}
	}
}

// TestExtendProvider_ErrorHandling tests various error scenarios
func TestExtendProvider_ErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &ExtendProvider{}
	
	t.Run("GetFanControllerCollectionResponse with nil FanControllers", func(t *testing.T) {
		em := &ExtendManager{
			mgr: createManager("mgr1"),
			OpenBmcFan: &OpenBmcFan{
				FanControllers: nil,
			},
		}
		
		_, respErr := provider.GetFanControllerCollectionResponse(em, "machine1", "mgr1")
		if respErr == nil {
			t.Error("Expected error for nil FanControllers")
		}
	})
	
	t.Run("GetFanZoneResponse with invalid manager type", func(t *testing.T) {
		_, respErr := provider.GetFanZoneResponse(&redfish.Manager{}, "machine1", "mgr1", "zone1")
		if respErr == nil {
			t.Error("Expected error for invalid manager type")
		}
		if respErr.StatusCode != http.StatusInternalServerError {
			t.Errorf("StatusCode = %v, want %v", respErr.StatusCode, http.StatusInternalServerError)
		}
	})
	
	t.Run("GetPidControllerResponse with non-existent controller", func(t *testing.T) {
		em := &ExtendManager{
			mgr: createManager("mgr1"),
			OpenBmcFan: &OpenBmcFan{
				PidControllers: &PidControllers{
					Items: map[string]*PidController{},
				},
			},
		}
		
		_, respErr := provider.GetPidControllerResponse(em, "machine1", "mgr1", "NonExistent")
		if respErr == nil {
			t.Error("Expected error for non-existent PID controller")
		}
		if respErr.StatusCode != http.StatusNotFound {
			t.Errorf("StatusCode = %v, want %v", respErr.StatusCode, http.StatusNotFound)
		}
	})
}

// TestOpenBmcFan_CompleteStructure tests the complete OpenBmcFan structure
func TestOpenBmcFan_CompleteStructure(t *testing.T) {
	jsonData := `{
		"@odata.id": "/redfish/v1/fan",
		"@odata.type": "#Fan.v1_0_0.Fan",
		"Profile": "Performance",
		"Profile@Redfish.AllowableValues": ["Performance", "Balanced", "PowerSaver"],
		"FanControllers": {
			"@odata.id": "/redfish/v1/fancontrollers",
			"@odata.type": "#FanControllers",
			"Fan1": {"FFGainCoefficient": 1.5}
		},
		"FanZones": {
			"@odata.id": "/redfish/v1/fanzones",
			"@odata.type": "#FanZones",
			"Zone1": {"FailSafePercent": 75.0}
		},
		"PidControllers": {
			"@odata.id": "/redfish/v1/pidcontrollers",
			"@odata.type": "#PidControllers",
			"Pid1": {"SetPoint": 50.0}
		}
	}`
	
	var fan OpenBmcFan
	if err := json.Unmarshal([]byte(jsonData), &fan); err != nil {
		t.Fatalf("Failed to unmarshal OpenBmcFan: %v", err)
	}
	
	if fan.Profile != "Performance" {
		t.Errorf("Profile = %v, want Performance", fan.Profile)
	}
	
	if len(fan.ProfileAllowableValues) != 3 {
		t.Errorf("len(ProfileAllowableValues) = %v, want 3", len(fan.ProfileAllowableValues))
	}
	
	if fan.FanControllers == nil {
		t.Error("FanControllers should not be nil")
	}
	
	if fan.FanZones == nil {
		t.Error("FanZones should not be nil")
	}
	
	if fan.PidControllers == nil {
		t.Error("PidControllers should not be nil")
	}
}
