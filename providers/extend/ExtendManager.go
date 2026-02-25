package extendprovider

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stmcginnis/gofish/redfish"

	"multifish/utility"
	redfishprovider "multifish/providers/redfish"
)

// ========== Constants ==========

// ProfileAllowlist contains the valid profile values
var ProfileAllowlist = []string{"Performance", "Balanced", "PowerSaver", "Custom"}

// ========== Type Definitions ==========

// OdataID is a helper struct for @odata.id fields.
type OdataID struct {
	OdataID string `json:"@odata.id"`
}

// FanController represents a single fan controller.
type FanController struct {
	OdataID             string     `json:"@odata.id",omitempty`
	OdataType           string     `json:"@odata.type,omitempty"`
	FFGainCoefficient   float64    `json:"FFGainCoefficient"`
	FFOffCoefficient    float64    `json:"FFOffCoefficient"`
	ICoefficient        float64    `json:"ICoefficient"`
	ILimitMax           float64    `json:"ILimitMax"`
	ILimitMin           float64    `json:"ILimitMin"`
	Inputs              []string   `json:"Inputs"`
	NegativeHysteresis  float64    `json:"NegativeHysteresis"`
	OutLimitMax         float64    `json:"OutLimitMax"`
	OutLimitMin         float64    `json:"OutLimitMin"`
	Outputs             []string   `json:"Outputs"`
	PCoefficient        float64    `json:"PCoefficient"`
	PositiveHysteresis  float64    `json:"PositiveHysteresis"`
	SlewNeg             float64    `json:"SlewNeg"`
	SlewPos             float64    `json:"SlewPos"`
	Zones               []OdataID  `json:"Zones"`
}

// FanControllers is a map of fan controller name to FanController.
type FanControllers struct {
	OdataID   string                    `json:"@odata.id"`
	OdataType string                    `json:"@odata.type"`
	Items     map[string]*FanController `json:"-"`
}

// UnmarshalJSON custom unmarshal for FanControllers to extract dynamic keys.
func (fc *FanControllers) UnmarshalJSON(data []byte) error {
	type Alias FanControllers
	aux := &struct {
		*Alias
	}{Alias: (*Alias)(fc)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	// Extract dynamic keys
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	fc.Items = make(map[string]*FanController)
	for k, v := range raw {
		if k == "@odata.id" || k == "@odata.type" {
			continue
		}
		var ctrl FanController
		if err := json.Unmarshal(v, &ctrl); err == nil {
			fc.Items[k] = &ctrl
		}
	}
	return nil
}

// PidController represents a single PID controller.
type PidController struct {
	OdataID            string    `json:"@odata.id",omitempty`
	OdataType          string    `json:"@odata.type",omitempty"`
	FFGainCoefficient  float64   `json:"FFGainCoefficient"`
	FFOffCoefficient   float64   `json:"FFOffCoefficient"`
	ICoefficient       float64   `json:"ICoefficient"`
	ILimitMax          float64   `json:"ILimitMax"`
	ILimitMin          float64   `json:"ILimitMin"`
	Inputs             []string  `json:"Inputs"`
	NegativeHysteresis float64   `json:"NegativeHysteresis"`
	OutLimitMax        float64   `json:"OutLimitMax"`
	OutLimitMin        float64   `json:"OutLimitMin"`
	PCoefficient       float64   `json:"PCoefficient"`
	PositiveHysteresis float64   `json:"PositiveHysteresis"`
	SetPoint           float64   `json:"SetPoint"`
	SlewNeg            float64   `json:"SlewNeg"`
	SlewPos            float64   `json:"SlewPos"`
	Zones              []OdataID `json:"Zones"`
}

// PidControllers is a map of PID controller name to PidController.
type PidControllers struct {
	OdataID   string                      `json:"@odata.id"`
	OdataType string                      `json:"@odata.type"`
	Items     map[string]*PidController   `json:"-"`
}

// UnmarshalJSON custom unmarshal for PidControllers to extract dynamic keys.
func (pc *PidControllers) UnmarshalJSON(data []byte) error {
	type Alias PidControllers
	aux := &struct {
		*Alias
	}{Alias: (*Alias)(pc)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	// Extract dynamic keys
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	pc.Items = make(map[string]*PidController)
	for k, v := range raw {
		if k == "@odata.id" || k == "@odata.type" {
			continue
		}
		var ctrl PidController
		if err := json.Unmarshal(v, &ctrl); err == nil {
			pc.Items[k] = &ctrl
		}
	}
	return nil
}

// FanZone represents a single fan zone.
type FanZone struct {
	OdataID          string   `json:"@odata.id,omitempty"`
	OdataType        string   `json:"@odata.type,omitempty"`
	FailSafePercent  float64  `json:"FailSafePercent"`
	MinThermalOutput float64  `json:"MinThermalOutput"`
}

// FanZones represents the FanZones group.
type FanZones struct {
	OdataID   string              `json:"@odata.id,omitempty"`
	OdataType string              `json:"@odata.type,omitempty"`
	Items     map[string]FanZone  `json:"-"`
}

// UnmarshalJSON custom unmarshal for FanZones to extract dynamic keys.
func (fz *FanZones) UnmarshalJSON(data []byte) error {
	type Alias FanZones
	aux := &struct {
		*Alias
	}{Alias: (*Alias)(fz)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	// Extract dynamic keys
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	fz.Items = make(map[string]FanZone)
	for k, v := range raw {
		if k == "@odata.id" || k == "@odata.type" {
			continue
		}
		var zone FanZone
		if err := json.Unmarshal(v, &zone); err == nil {
			fz.Items[k] = zone
		}
	}
	return nil
}

// OpenBmcFan is the top-level OEM OpenBmc Fan structure.
type OpenBmcFan struct {
	OdataID                      string          `json:"@odata.id"`
	OdataType                    string          `json:"@odata.type"`
	FanControllers               *FanControllers `json:"FanControllers"`
	FanZones                     *FanZones       `json:"FanZones"`
	PidControllers               *PidControllers `json:"PidControllers"`
	Profile                      string          `json:"Profile"`
	ProfileAllowableValues       []string        `json:"Profile@Redfish.AllowableValues"`
}

// ExtendManager extends the redfish.Manager with OEM fields.
type ExtendManager struct {
	// Embed the original Manager
	mgr *redfish.Manager

	// Oem.OpenBmc.Fan
	OpenBmcFan *OpenBmcFan
}

// PatchProfileType represents the PATCH request body for Profile
type PatchProfileType struct {
	Profile string `json:"Profile,omitempty"`
}

// ProfileAllowedPatchFields returns the allowed fields for Profile PATCH operations.
func ProfileAllowedPatchFields() utility.FieldSpecMap {
	return utility.FieldSpecMap{
		"Profile": {Name: "Profile", Expected: "string"},
	}
}

// PatchFanControllerType represents the PATCH request body for FanController
type PatchFanControllerType struct {
	FFGainCoefficient   *float64    `json:"FFGainCoefficient,omitempty"`
	FFOffCoefficient    *float64    `json:"FFOffCoefficient,omitempty"`
	ICoefficient        *float64    `json:"ICoefficient,omitempty"`
	ILimitMax           *float64    `json:"ILimitMax,omitempty"`
	ILimitMin           *float64    `json:"ILimitMin,omitempty"`
	Inputs              []string    `json:"Inputs,omitempty"`
	NegativeHysteresis  *float64    `json:"NegativeHysteresis,omitempty"`
	OutLimitMax         *float64    `json:"OutLimitMax,omitempty"`
	OutLimitMin         *float64    `json:"OutLimitMin,omitempty"`
	Outputs             []string    `json:"Outputs,omitempty"`
	PCoefficient        *float64    `json:"PCoefficient,omitempty"`
	PositiveHysteresis  *float64    `json:"PositiveHysteresis,omitempty"`
	SlewNeg             *float64    `json:"SlewNeg,omitempty"`
	SlewPos             *float64    `json:"SlewPos,omitempty"`
	Zones               []OdataID   `json:"Zones,omitempty"`
}

// FanControllerAllowedPatchFields returns the allowed fields for FanController PATCH operations.
func FanControllerAllowedPatchFields() utility.FieldSpecMap {
	return utility.FieldSpecMap{
		"FFGainCoefficient":  {Name: "FFGainCoefficient", Expected: "number"},
		"FFOffCoefficient":   {Name: "FFOffCoefficient", Expected: "number"},
		"ICoefficient":       {Name: "ICoefficient", Expected: "number"},
		"ILimitMax":          {Name: "ILimitMax", Expected: "number"},
		"ILimitMin":          {Name: "ILimitMin", Expected: "number"},
		"NegativeHysteresis": {Name: "NegativeHysteresis", Expected: "number"},
		"OutLimitMax":        {Name: "OutLimitMax", Expected: "number"},
		"OutLimitMin":        {Name: "OutLimitMin", Expected: "number"},
		"PCoefficient":       {Name: "PCoefficient", Expected: "number"},
		"PositiveHysteresis": {Name: "PositiveHysteresis", Expected: "number"},
		"SlewNeg":            {Name: "SlewNeg", Expected: "number"},
		"SlewPos":            {Name: "SlewPos", Expected: "number"},
		"Zones":              {Name: "Zones", Expected: "array"},
	}
}

// PatchFanZoneType represents the PATCH request body for FanZone
type PatchFanZoneType struct {
	FailSafePercent  *float64 `json:"FailSafePercent,omitempty"`
	MinThermalOutput *float64 `json:"MinThermalOutput,omitempty"`
}

// FanZoneAllowedPatchFields returns the allowed fields for FanZone PATCH operations.
func FanZoneAllowedPatchFields() utility.FieldSpecMap {
	return utility.FieldSpecMap{
		"FailSafePercent":  {Name: "FailSafePercent", Expected: "number"},
		"MinThermalOutput": {Name: "MinThermalOutput", Expected: "number"},
	}
}

// PatchPidControllerType represents the PATCH request body for PidController
type PatchPidControllerType struct {
	FFGainCoefficient   *float64    `json:"FFGainCoefficient,omitempty"`
	FFOffCoefficient    *float64    `json:"FFOffCoefficient,omitempty"`
	ICoefficient        *float64    `json:"ICoefficient,omitempty"`
	ILimitMax           *float64    `json:"ILimitMax,omitempty"`
	ILimitMin           *float64    `json:"ILimitMin,omitempty"`
	Inputs              []string    `json:"Inputs,omitempty"`
	NegativeHysteresis  *float64    `json:"NegativeHysteresis,omitempty"`
	OutLimitMax         *float64    `json:"OutLimitMax,omitempty"`
	OutLimitMin         *float64    `json:"OutLimitMin,omitempty"`
	PCoefficient        *float64    `json:"PCoefficient,omitempty"`
	PositiveHysteresis  *float64    `json:"PositiveHysteresis,omitempty"`
	SetPoint            *float64    `json:"SetPoint,omitempty"`
	SlewNeg             *float64    `json:"SlewNeg,omitempty"`
	SlewPos             *float64    `json:"SlewPos,omitempty"`
	Zones               []OdataID   `json:"Zones,omitempty"`
}

// PidControllerAllowedPatchFields returns the allowed fields for PidController PATCH operations.
func PidControllerAllowedPatchFields() utility.FieldSpecMap {
	return utility.FieldSpecMap{
		"FFGainCoefficient":  {Name: "FFGainCoefficient", Expected: "number"},
		"FFOffCoefficient":   {Name: "FFOffCoefficient", Expected: "number"},
		"ICoefficient":       {Name: "ICoefficient", Expected: "number"},
		"ILimitMax":          {Name: "ILimitMax", Expected: "number"},
		"ILimitMin":          {Name: "ILimitMin", Expected: "number"},
		"NegativeHysteresis": {Name: "NegativeHysteresis", Expected: "number"},
		"OutLimitMax":        {Name: "OutLimitMax", Expected: "number"},
		"OutLimitMin":        {Name: "OutLimitMin", Expected: "number"},
		"PCoefficient":       {Name: "PCoefficient", Expected: "number"},
		"PositiveHysteresis": {Name: "PositiveHysteresis", Expected: "number"},
		"SetPoint":           {Name: "SetPoint", Expected: "number"},
		"SlewNeg":            {Name: "SlewNeg", Expected: "number"},
		"SlewPos":            {Name: "SlewPos", Expected: "number"},
		"Zones":              {Name: "Zones", Expected: "array"},
	}
}

// ========== Helper Functions ==========

// ExtractOpenBmcFan extracts the OpenBmcFan from the given OEM raw message.
func ExtractOpenBmcFan(oem json.RawMessage) *OpenBmcFan {
	log := utility.GetLogger()

	if len(oem) == 0 {
		log.Debug().Msg("OEM data is empty")
		return nil
	}

	var oemMap map[string]json.RawMessage
	if err := json.Unmarshal(oem, &oemMap); err != nil {
		log.Debug().Msgf("Failed to unmarshal OEM data: %v", err)
		return nil
	}

	openBmcRaw, ok := oemMap["OpenBmc"]
	if !ok {
		log.Debug().Msg("OpenBmc data is missing")
		return nil
	}

	var openBmcMap map[string]json.RawMessage
	if err := json.Unmarshal(openBmcRaw, &openBmcMap); err != nil {
		log.Debug().Msgf("Failed to unmarshal OpenBmc data: %v", err)
		return nil
	}

	fanRaw, ok := openBmcMap["Fan"]
	if !ok {
		log.Debug().Msg("Fan data is missing in OpenBmc OEM")
		return nil
	}

	var fan OpenBmcFan
	if err := json.Unmarshal(fanRaw, &fan); err != nil {
		log.Debug().Msgf("Failed to unmarshal Fan data: %v", err)
		return nil
	}

	return &fan
}

// Managers retrieves all managers and fills in the extended struct.
func (es *ExtendService) Managers() ([]*ExtendManager, error) {
	log := utility.GetLogger()

	// Get the base managers
	managers, err := es.Service.Managers()
	if err != nil {
		log.Error().Msgf("Failed to get managers: %v", err)
		return nil, err
	}

	// Create extended managers
	var extendedManagers []*ExtendManager
	for _, manager := range managers {
		ext := &ExtendManager{
			mgr:    manager,
		}
		
		log := utility.GetLogger()
		if ext.OpenBmcFan = ExtractOpenBmcFan(manager.Oem); ext.OpenBmcFan == nil {
			log.Debug().Str("managerID", manager.ID).Msg("Oem.OpenBmc.Fan not found for extended manager")
		}
		extendedManagers = append(extendedManagers, ext)
	}

	return extendedManagers, nil
}

// GetManager returns the underlying redfish.Manager.
func (em *ExtendManager) GetManager() *redfish.Manager {
	return em.mgr
}

// refreshExtendManager re-fetches the manager from the server and updates all fields including OEM data
func refreshExtendManager(em *ExtendManager) error {
	log := utility.GetLogger()

	// Re-fetch base manager from the server using its ODataID
	if err := redfishprovider.RefreshManager(em.GetManager()); err != nil {
		log.Error().Msgf("Failed to refresh manager: %v", err)
		return fmt.Errorf("failed to refresh extended manager '%s': %w", em.GetManager().ID, err)
	}
	
	// Re-extract OEM data
	em.OpenBmcFan = ExtractOpenBmcFan(em.GetManager().Oem)
	if em.OpenBmcFan == nil {
		log.Debug().Str("managerID", em.GetManager().ID).Msg("Oem.OpenBmc.Fan not found after refresh")
	}

	return nil
}

// updateAndRefreshExtendManager updates the manager with the given fields and refreshes the object
func updateAndRefreshExtendManager(em *ExtendManager, payload *map[string]interface{}) error {
	log := utility.GetLogger()

	// Apply ETag matching setting from machine configuration
	em.GetManager().DisableEtagMatch(true)

	// Patch payload
	if err := em.GetManager().Patch(em.GetManager().ODataID, payload); err != nil {
		log.Error().Msgf("Failed to patch manager: %v", err)
		return err
	}

	// Note: We could refresh here, but it's optional
	// if err := refreshExtendManager(em); err != nil {
	// 	log.Printf("Warning: Failed to refresh manager after update: %v", err)
	// }
	
	return nil
}

// ========= Implementations ==========

// patchExtendProfile is the internal implementation for patching profile
func patchExtendProfile(em *ExtendManager, patch PatchProfileType) *utility.ResponseError {
	log := utility.GetLogger()

	// Validate
	if em.OpenBmcFan == nil {
		log.Debug().Msg("OpenBmcFan is not available")
		return &utility.ResponseError{
			StatusCode: http.StatusNotFound,
			Error:      fmt.Errorf("OpenBmcFan is not available"),
			Message:    "ResourceNotFound",
		}
	}
	
	if patch.Profile == "" {
		log.Debug().Msg("Profile cannot be empty")
		return &utility.ResponseError{
			StatusCode: http.StatusBadRequest,
			Error:      fmt.Errorf("Profile cannot be empty"),
			Message:    "InvalidRequest",
		}
	}
	
	// Check if profile is in allowable values
	allowed := false
	for _, v := range em.OpenBmcFan.ProfileAllowableValues {
		if v == patch.Profile {
			allowed = true
			break
		}
	}
	if !allowed {
		log.Debug().Msgf("profile value '%s' is not allowed", patch.Profile)
		return &utility.ResponseError{
			StatusCode: http.StatusBadRequest,
			Error:      fmt.Errorf("profile value '%s' is not allowed", patch.Profile),
			Message:    "InvalidRequest",
		}
	}
	
	// Patch the value via Redfish PATCH
	patchPayload := map[string]interface{}{
		"Oem": map[string]interface{}{
			"OpenBmc": map[string]interface{}{
				"Fan": map[string]interface{}{
					"Profile": patch.Profile,
				},
			},
		},
	}
	
	// Patch the value
	if err := updateAndRefreshExtendManager(em, &patchPayload); err != nil {
		log.Error().Msgf("Failed to patch profile: %v", err)
		return &utility.ResponseError{
			StatusCode: http.StatusInternalServerError,
			Error:      fmt.Errorf("failed to patch profile: %v", err),
			Message:    "InternalError",
		}
	}
	
	return nil
}

// patchExtendFanController is the internal implementation for patching fan controller
func patchExtendFanController(em *ExtendManager, fanID string, fcPatch *PatchFanControllerType) *utility.ResponseError {
	log := utility.GetLogger()

	// Check if FanControllers are available
	if em.OpenBmcFan == nil || em.OpenBmcFan.FanControllers == nil {
		log.Debug().Msg("FanControllers not available")
		return &utility.ResponseError{
			StatusCode: http.StatusNotFound,
			Error:      fmt.Errorf("FanControllers not available"),
			Message:    "ResourceNotFound",
		}
	}
	
	// Check if the specified FanController exists
	_, ok := em.OpenBmcFan.FanControllers.Items[fanID]
	if !ok {
		log.Debug().Msgf("FanController '%s' not found", fanID)
		return &utility.ResponseError{
			StatusCode: http.StatusNotFound,
			Error:      fmt.Errorf("FanController '%s' not found", fanID),
			Message:    "ResourceNotFound",
		}
	}
	
	// Validate fcPatch
	if fcPatch == nil {
		log.Debug().Msg("fcPatch cannot be nil")
		return &utility.ResponseError{
			StatusCode: http.StatusBadRequest,
			Error:      fmt.Errorf("fcPatch cannot be nil"),
			Message:    "InvalidRequest",
		}
	}
	
	// Prepare patch body
	patch := map[string]interface{}{
		"Oem": map[string]interface{}{
			"OpenBmc": map[string]interface{}{
				"Fan": map[string]interface{}{
					"FanControllers": map[string]interface{}{
						fanID: fcPatch,
					},
				},
			},
		},
	}
	
	// Use the update and refresh method to apply the change
	if err := updateAndRefreshExtendManager(em, &patch); err != nil {
		log.Error().Msgf("Failed to patch FanController: %v", err)
		return &utility.ResponseError{
			StatusCode: http.StatusInternalServerError,
			Error:      fmt.Errorf("failed to patch FanController: %v", err),
			Message:    "InternalError",
		}
	}
	
	return nil
}

// patchExtendFanZone is the internal implementation for patching fan zone
func patchExtendFanZone(em *ExtendManager, zoneID string, fzPatch *PatchFanZoneType) *utility.ResponseError {
	log := utility.GetLogger()

	// Check if FanZones are available
	if em.OpenBmcFan == nil || em.OpenBmcFan.FanZones == nil {
		log.Debug().Msg("FanZones not available")
		return &utility.ResponseError{
			StatusCode: http.StatusNotFound,
			Error:      fmt.Errorf("FanZones not available"),
			Message:    "ResourceNotFound",
		}
	}
	
	// Check if the specified FanZone exists
	_, ok := em.OpenBmcFan.FanZones.Items[zoneID]
	if !ok {
		log.Debug().Msgf("FanZone '%s' not found", zoneID)
		return &utility.ResponseError{
			StatusCode: http.StatusNotFound,
			Error:      fmt.Errorf("FanZone '%s' not found", zoneID),
			Message:    "ResourceNotFound",
		}
	}
	
	// Validate fzPatch
	if fzPatch == nil {
		log.Debug().Msg("fzPatch cannot be nil")
		return &utility.ResponseError{
			StatusCode: http.StatusBadRequest,
			Error:      fmt.Errorf("fzPatch cannot be nil"),
			Message:    "InvalidRequest",
		}
	}
	
	// Prepare patch body
	patch := map[string]interface{}{
		"Oem": map[string]interface{}{
			"OpenBmc": map[string]interface{}{
				"Fan": map[string]interface{}{
					"FanZones": map[string]interface{}{
						zoneID: fzPatch,
					},
				},
			},
		},
	}
	
	// Use the update and refresh method to apply the change
	if err := updateAndRefreshExtendManager(em, &patch); err != nil {
		log.Error().Msgf("Failed to patch FanZone: %v", err)
		return &utility.ResponseError{
			StatusCode: http.StatusInternalServerError,
			Error:      fmt.Errorf("failed to patch FanZone: %v", err),
			Message:    "InternalError",
		}
	}
	
	return nil
}

// patchExtendPidController is the internal implementation for patching PID controller
func patchExtendPidController(em *ExtendManager, pidID string, pcPatch *PatchPidControllerType) *utility.ResponseError {
	log := utility.GetLogger()

	// Check if PidControllers are available
	if em.OpenBmcFan == nil || em.OpenBmcFan.PidControllers == nil {
		log.Debug().Msg("PidControllers not available")
		return &utility.ResponseError{
			StatusCode: http.StatusNotFound,
			Error:      fmt.Errorf("PidControllers not available"),
			Message:    "ResourceNotFound",
		}
	}
	
	// Check if the specified PidController exists
	_, ok := em.OpenBmcFan.PidControllers.Items[pidID]
	if !ok {
		log.Debug().Msgf("PidController '%s' not found", pidID)
		return &utility.ResponseError{
			StatusCode: http.StatusNotFound,
			Error:      fmt.Errorf("PidController '%s' not found", pidID),
			Message:    "ResourceNotFound",
		}
	}
	
	// Validate pcPatch
	if pcPatch == nil {
		log.Debug().Msg("pcPatch cannot be nil")
		return &utility.ResponseError{
			StatusCode: http.StatusBadRequest,
			Error:      fmt.Errorf("pcPatch cannot be nil"),
			Message:    "InvalidRequest",
		}
	}
	
	// Prepare patch body
	patch := map[string]interface{}{
		"Oem": map[string]interface{}{
			"OpenBmc": map[string]interface{}{
				"Fan": map[string]interface{}{
					"PidControllers": map[string]interface{}{
						pidID: pcPatch,
					},
				},
			},
		},
	}
	
	// Use the update and refresh method to apply the change
	if err := updateAndRefreshExtendManager(em, &patch); err != nil {
		log.Error().Msgf("Failed to patch PidController: %v", err)
		return &utility.ResponseError{
			StatusCode: http.StatusInternalServerError,
			Error:      fmt.Errorf("failed to patch PidController: %v", err),
			Message:    "InternalError",
		}
	}
	
	return nil
}

// ========== Provider Implementation ==========

// ExtendProvider handles ExtendManager types
type ExtendProvider struct{}

// TypeName returns the name of this provider
func (p *ExtendProvider) TypeName() string {
	return "Extended"
}

// Supports checks if this provider can handle the given manager instance
func (p *ExtendProvider) Supports(v interface{}) bool {
	_, ok := v.(*ExtendManager)
	return ok
}

// SupportsCollection checks if this provider can handle the given manager collection
func (p *ExtendProvider) SupportsCollection(v interface{}) bool {
	_, ok := v.([]*ExtendManager)
	return ok
}

// GetManagerCollectionResponse builds the response for a collection of managers
func (p *ExtendProvider) GetManagerCollectionResponse(managers interface{}, machineID string) (gin.H, error) {
	log := utility.GetLogger()

	mgrs, ok := managers.([]*ExtendManager)
	if !ok {
		log.Error().Msgf("invalid manager type for ExtendProvider: %T", managers)
		return nil, fmt.Errorf("invalid manager type for ExtendProvider: %T", managers)
	}
	
	var members []gin.H
	for _, em := range mgrs {
		member := gin.H{
			"@odata.id": fmt.Sprintf("/MultiFish/v1/Platform/%s/Managers/%s", machineID, em.GetManager().ID),
		}
		members = append(members, member)
	}
	
	return gin.H{
		"@odata.type":         "#ManagerCollection.ManagerCollection",
		"@odata.id":           fmt.Sprintf("/MultiFish/v1/Platform/%s/Managers", machineID),
		"Name":                "Manager Collection",
		"Members":             members,
		"Members@odata.count": len(members),
	}, nil
}

// GetManagerResponse builds the response for a single manager
func (p *ExtendProvider) GetManagerResponse(manager interface{}, machineID, managerID string) (gin.H, *utility.ResponseError) {
	log := utility.GetLogger()

	em, ok := manager.(*ExtendManager)
	if !ok {
		log.Error().Msgf("invalid manager type for ExtendProvider: %T", manager)
		return nil, &utility.ResponseError{
			StatusCode: http.StatusInternalServerError,
			Error:      fmt.Errorf("invalid manager type for ExtendProvider: %T", manager),
			Message:    "InternalError",
		}
	}
	
	return p.GetManagerResponse(em.GetManager(), machineID, managerID)
}

// PatchManager patches the manager with the given updates
func (p *ExtendProvider) PatchManager(manager interface{}, updates interface{}) *utility.ResponseError {
	log := utility.GetLogger()

	em, ok := manager.(*ExtendManager)
	if !ok {
		log.Error().Msgf("invalid manager type for ExtendProvider: %T", manager)
		return &utility.ResponseError{
			StatusCode: http.StatusInternalServerError,
			Error:      fmt.Errorf("invalid manager type for ExtendProvider: %T", manager),
			Message:    "InternalError",
		}
	}
	
	patch, ok := updates.(*redfishprovider.PatchManagerType)
	if !ok {
		log.Error().Msgf("invalid patch type for ExtendProvider: %T", updates)
		return &utility.ResponseError{
			StatusCode: http.StatusBadRequest,
			Error:      fmt.Errorf("invalid patch type: %T", updates),
			Message:    "InvalidRequest",
		}
	}
	
	if patch == nil {
		log.Error().Msg("patch cannot be nil")
		return &utility.ResponseError{
			StatusCode: http.StatusBadRequest,
			Error:      fmt.Errorf("patch cannot be nil"),
			Message:    "InvalidRequest",
		}
	}
	
	mgr := em.GetManager()
	err := redfishprovider.PatchManagerData(mgr, patch)
	if err != nil {
		log.Error().Msgf("Failed to patch manager: %v", err)
		return &utility.ResponseError{
			StatusCode: err.StatusCode,
			Error:      err.Error,
			Message:    err.Message,
		}
	}
	
	return nil
}

// GetProfileResponse builds the response for the manager's profile
func (p *ExtendProvider) GetProfileResponse(manager interface{}, machineID, managerID string) (gin.H, *utility.ResponseError) {
	log := utility.GetLogger()

	em, ok := manager.(*ExtendManager)
	if !ok {
		log.Error().Msgf("invalid manager type for ExtendProvider: %T", manager)
		return nil, &utility.ResponseError{
			StatusCode: http.StatusInternalServerError,
			Error:      fmt.Errorf("invalid manager type for ExtendProvider: %T", manager),
			Message:    "InternalError",
		}
	}
	
	if em.OpenBmcFan == nil {
		log.Debug().Msg("OpenBmcFan is not available, cannot get profile")
		return nil, &utility.ResponseError{
			StatusCode: http.StatusNotFound,
			Error:      fmt.Errorf("Profile not available"),
			Message:    "ResourceNotFound",
		}
	}
	
	return gin.H{
		"@odata.type":                     "#OemProfile.v1_0_0.Profile",
		"@odata.id":                       fmt.Sprintf("/MultiFish/v1/Platform/%s/Managers/%s/Oem/OpenBmc/Fan/Profile", machineID, managerID),
		"Profile":                         em.OpenBmcFan.Profile,
		"Profile@Redfish.AllowableValues": em.OpenBmcFan.ProfileAllowableValues,
	}, nil
}

// PatchProfile patches the manager's profile
func (p *ExtendProvider) PatchProfile(manager interface{}, updates interface{}) *utility.ResponseError {
	log := utility.GetLogger()

	em, ok := manager.(*ExtendManager)
	if !ok {
		log.Error().Msgf("invalid manager type for ExtendProvider: %T", manager)
		return &utility.ResponseError{
			StatusCode: http.StatusInternalServerError,
			Error:      fmt.Errorf("invalid manager type for ExtendProvider: %T", manager),
			Message:    "InternalError",
		}
	}
	
	patch, ok := updates.(PatchProfileType)
	if !ok {
		log.Error().Msgf("invalid patch type for ExtendProvider: %T", updates)
		return &utility.ResponseError{
			StatusCode: http.StatusBadRequest,
			Error:      fmt.Errorf("invalid patch type: %T", updates),
			Message:    "InvalidRequest",
		}
	}
	
	return patchExtendProfile(em, patch)
}

// GetFanControllerCollectionResponse builds the response for the FanControllers collection
func (p *ExtendProvider) GetFanControllerCollectionResponse(manager interface{}, machineID, managerID string) (gin.H, *utility.ResponseError) {
	log := utility.GetLogger()

	em, ok := manager.(*ExtendManager)
	if !ok {
		log.Error().Msgf("invalid manager type for ExtendProvider: %T", manager)
		return nil, &utility.ResponseError{
			StatusCode: http.StatusInternalServerError,
			Error:      fmt.Errorf("invalid manager type for ExtendProvider: %T", manager),
			Message:    "InternalError",
		}
	}
	
	// Check if FanControllers are available
	if em.OpenBmcFan == nil || em.OpenBmcFan.FanControllers == nil {
		log.Debug().Msg("FanControllers not available, cannot get collection response")
		return nil, &utility.ResponseError{
			StatusCode: http.StatusNotFound,
			Error:      fmt.Errorf("FanControllers not available"),
			Message:    "ResourceNotFound",
		}
	}
	
	// Populate members
	members := make([]gin.H, 0, len(em.OpenBmcFan.FanControllers.Items))
	for fanID := range em.OpenBmcFan.FanControllers.Items {
		members = append(members, gin.H{
			"@odata.id": fmt.Sprintf("/MultiFish/v1/Platform/%s/Managers/%s/Oem/OpenBmc/Fan/FanControllers/%s", machineID, managerID, fanID),
		})
	}
	
	// Build response
	response := gin.H{
		"@odata.type":         em.OpenBmcFan.FanControllers.OdataType,
		"@odata.id":           fmt.Sprintf("/MultiFish/v1/Platform/%s/Managers/%s/Oem/OpenBmc/Fan/FanControllers", machineID, managerID),
		"Id":                  "FanControllers",
		"Members":             members,
		"Members@odata.count": len(members),
	}
	return response, nil
}

// GetFanControllerResponse builds the response for a specific FanController
func (p *ExtendProvider) GetFanControllerResponse(manager interface{}, machineID, managerID, fanID string) (gin.H, *utility.ResponseError) {
	log := utility.GetLogger()

	em, ok := manager.(*ExtendManager)
	if !ok {
		log.Error().Msgf("invalid manager type for ExtendProvider: %T", manager)
		return nil, &utility.ResponseError{
			StatusCode: http.StatusInternalServerError,
			Error:      fmt.Errorf("invalid manager type for ExtendProvider: %T", manager),
			Message:    "InternalError",
		}
	}
	
	// Check if FanControllers are available
	if em.OpenBmcFan == nil || em.OpenBmcFan.FanControllers == nil {
		log.Debug().Msg("FanControllers not available, cannot get resource response")
		return nil, &utility.ResponseError{
			StatusCode: http.StatusNotFound,
			Error:      fmt.Errorf("FanControllers not available"),
			Message:    "ResourceNotFound",
		}
	}
	
	// Check if the specified FanController exists
	fc, ok := em.OpenBmcFan.FanControllers.Items[fanID]
	if !ok {
		log.Debug().Msgf("FanController '%s' not found", fanID)
		return nil, &utility.ResponseError{
			StatusCode: http.StatusNotFound,
			Error:      fmt.Errorf("FanController '%s' not found", fanID),
			Message:    "ResourceNotFound",
		}
	}
	
	// Build response
	response := gin.H{
		"@odata.type":         fc.OdataType,
		"@odata.id":           fmt.Sprintf("/MultiFish/v1/Platform/%s/Managers/%s/Oem/OpenBmc/Fan/FanControllers/%s", machineID, managerID, fanID),
		"Id":                  fanID,
		"FFGainCoefficient":   fc.FFGainCoefficient,
		"FFOffCoefficient":    fc.FFOffCoefficient,
		"ICoefficient":        fc.ICoefficient,
		"ILimitMax":           fc.ILimitMax,
		"ILimitMin":           fc.ILimitMin,
		"Inputs":              fc.Inputs,
		"NegativeHysteresis":  fc.NegativeHysteresis,
		"OutLimitMax":         fc.OutLimitMax,
		"OutLimitMin":         fc.OutLimitMin,
		"Outputs":             fc.Outputs,
		"PCoefficient":        fc.PCoefficient,
		"PositiveHysteresis":  fc.PositiveHysteresis,
		"SlewNeg":             fc.SlewNeg,
		"SlewPos":             fc.SlewPos,
		"Zones":               fc.Zones,
	}
	return response, nil
}

// PatchFanController patches the specified FanController
func (p *ExtendProvider) PatchFanController(manager interface{}, fanID string, updates interface{}) *utility.ResponseError {
	log := utility.GetLogger()

	em, ok := manager.(*ExtendManager)
	if !ok {
		log.Error().Msgf("invalid manager type for ExtendProvider: %T", manager)
		return &utility.ResponseError{
			StatusCode: http.StatusInternalServerError,
			Error:      fmt.Errorf("invalid manager type for ExtendProvider: %T", manager),
			Message:    "InternalError",
		}
	}
	
	fcPatch, ok := updates.(*PatchFanControllerType)
	if !ok {
		log.Error().Msgf("invalid patch type for ExtendProvider: %T", updates)
		return &utility.ResponseError{
			StatusCode: http.StatusBadRequest,
			Error:      fmt.Errorf("invalid patch type: %T", updates),
			Message:    "InvalidRequest",
		}
	}
	
	return patchExtendFanController(em, fanID, fcPatch)
}

// GetFanZoneCollectionResponse builds the response for the FanZones collection
func (p *ExtendProvider) GetFanZoneCollectionResponse(manager interface{}, machineID, managerID string) (gin.H, *utility.ResponseError) {
	log := utility.GetLogger()

	em, ok := manager.(*ExtendManager)
	if !ok {
		log.Error().Msgf("invalid manager type for ExtendProvider: %T", manager)
		return nil, &utility.ResponseError{
			StatusCode: http.StatusInternalServerError,
			Error:      fmt.Errorf("invalid manager type for ExtendProvider: %T", manager),
			Message:    "InternalError",
		}
	}
	
	// Check if FanZones are available
	if em.OpenBmcFan == nil || em.OpenBmcFan.FanZones == nil {
		log.Debug().Msg("FanZones not available, cannot get collection response")
		return nil, &utility.ResponseError{
			StatusCode: http.StatusNotFound,
			Error:      fmt.Errorf("FanZones not available"),
			Message:    "ResourceNotFound",
		}
	}
	
	// Populate members
	members := make([]gin.H, 0, len(em.OpenBmcFan.FanZones.Items))
	for zoneID := range em.OpenBmcFan.FanZones.Items {
		members = append(members, gin.H{
			"@odata.id": fmt.Sprintf("/MultiFish/v1/Platform/%s/Managers/%s/Oem/OpenBmc/Fan/FanZones/%s", machineID, managerID, zoneID),
		})
	}
	
	// Build response
	response := gin.H{
		"@odata.type":         em.OpenBmcFan.FanZones.OdataType,
		"@odata.id":           fmt.Sprintf("/MultiFish/v1/Platform/%s/Managers/%s/Oem/OpenBmc/Fan/FanZones", machineID, managerID),
		"Id":                  "FanZones",
		"Members":             members,
		"Members@odata.count": len(members),
	}
	return response, nil
}

// GetFanZoneResponse builds the response for a specific FanZone
func (p *ExtendProvider) GetFanZoneResponse(manager interface{}, machineID, managerID, zoneID string) (gin.H, *utility.ResponseError) {
	log := utility.GetLogger()

	em, ok := manager.(*ExtendManager)
	if !ok {
		log.Error().Msgf("invalid manager type for ExtendProvider: %T", manager)
		return nil, &utility.ResponseError{
			StatusCode: http.StatusInternalServerError,
			Error:      fmt.Errorf("invalid manager type for ExtendProvider: %T", manager),
			Message:    "InternalError",
		}
	}
	
	// Check if FanZones are available
	if em.OpenBmcFan == nil || em.OpenBmcFan.FanZones == nil {
		log.Debug().Msg("FanZones not available, cannot get resource response")
		return nil, &utility.ResponseError{
			StatusCode: http.StatusNotFound,
			Error:      fmt.Errorf("FanZones not available"),
			Message:    "ResourceNotFound",
		}
	}
	
	// Check if the specified FanZone exists
	fz, ok := em.OpenBmcFan.FanZones.Items[zoneID]
	if !ok {
		log.Debug().Msgf("FanZone '%s' not found", zoneID)
		return nil, &utility.ResponseError{
			StatusCode: http.StatusNotFound,
			Error:      fmt.Errorf("FanZone '%s' not found", zoneID),
			Message:    "ResourceNotFound",
		}
	}
	
	// Build response
	response := gin.H{
		"@odata.type":        fz.OdataType,
		"@odata.id":          fmt.Sprintf("/MultiFish/v1/Platform/%s/Managers/%s/Oem/OpenBmc/Fan/FanZones/%s", machineID, managerID, zoneID),
		"Id":                 zoneID,
		"FailSafePercent":    fz.FailSafePercent,
		"MinThermalOutput":   fz.MinThermalOutput,
	}
	return response, nil
}

// PatchFanZone patches the specified FanZone
func (p *ExtendProvider) PatchFanZone(manager interface{}, zoneID string, updates interface{}) *utility.ResponseError {
	log := utility.GetLogger()

	em, ok := manager.(*ExtendManager)
	if !ok {
		log.Error().Msgf("invalid manager type for ExtendProvider: %T", manager)
		return &utility.ResponseError{
			StatusCode: http.StatusInternalServerError,
			Error:      fmt.Errorf("invalid manager type for ExtendProvider: %T", manager),
			Message:    "InternalError",
		}
	}
	
	fzPatch, ok := updates.(*PatchFanZoneType)
	if !ok {
		log.Error().Msgf("invalid patch type for ExtendProvider: %T", updates)
		return &utility.ResponseError{
			StatusCode: http.StatusBadRequest,
			Error:      fmt.Errorf("invalid patch type: %T", updates),
			Message:    "InvalidRequest",
		}
	}
	
	return patchExtendFanZone(em, zoneID, fzPatch)
}

// GetPidControllerCollectionResponse builds the response for the PidControllers collection
func (p *ExtendProvider) GetPidControllerCollectionResponse(manager interface{}, machineID, managerID string) (gin.H, *utility.ResponseError) {
	log := utility.GetLogger()

	em, ok := manager.(*ExtendManager)
	if !ok {
		log.Error().Msgf("invalid manager type for ExtendProvider: %T", manager)
		return nil, &utility.ResponseError{
			StatusCode: http.StatusInternalServerError,
			Error:      fmt.Errorf("invalid manager type for ExtendProvider: %T", manager),
			Message:    "InternalError",
		}
	}
	
	// Check if PidControllers are available
	if em.OpenBmcFan == nil || em.OpenBmcFan.PidControllers == nil {
		log.Debug().Msg("PidControllers not available, cannot get collection response")
		return nil, &utility.ResponseError{
			StatusCode: http.StatusNotFound,
			Error:      fmt.Errorf("PidControllers not available"),
			Message:    "ResourceNotFound",
		}
	}
	
	// Populate members
	members := make([]gin.H, 0, len(em.OpenBmcFan.PidControllers.Items))
	for pidID := range em.OpenBmcFan.PidControllers.Items {
		members = append(members, gin.H{
			"@odata.id": fmt.Sprintf("/MultiFish/v1/Platform/%s/Managers/%s/Oem/OpenBmc/Fan/PidControllers/%s", machineID, managerID, pidID),
		})
	}
	
	// Build response
	response := gin.H{
		"@odata.type":         em.OpenBmcFan.PidControllers.OdataType,
		"@odata.id":           fmt.Sprintf("/MultiFish/v1/Platform/%s/Managers/%s/Oem/OpenBmc/Fan/PidControllers", machineID, managerID),
		"Id":                  "PidControllers",
		"Members":             members,
		"Members@odata.count": len(members),
	}
	return response, nil
}

// GetPidControllerResponse builds the response for a specific PidController
func (p *ExtendProvider) GetPidControllerResponse(manager interface{}, machineID, managerID, pidID string) (gin.H, *utility.ResponseError) {
	log := utility.GetLogger()

	em, ok := manager.(*ExtendManager)
	if !ok {
		log.Error().Msgf("invalid manager type for ExtendProvider: %T", manager)
		return nil, &utility.ResponseError{
			StatusCode: http.StatusInternalServerError,
			Error:      fmt.Errorf("invalid manager type for ExtendProvider: %T", manager),
			Message:    "InternalError",
		}
	}
	
	// Check if PidControllers are available
	if em.OpenBmcFan == nil || em.OpenBmcFan.PidControllers == nil {
		log.Debug().Msg("PidControllers not available, cannot get collection response")
		return nil, &utility.ResponseError{
			StatusCode: http.StatusNotFound,
			Error:      fmt.Errorf("PidControllers not available"),
			Message:    "ResourceNotFound",
		}
	}
	
	// Check if the specified PidController exists
	pc, ok := em.OpenBmcFan.PidControllers.Items[pidID]
	if !ok {
		log.Debug().Msgf("PidController '%s' not found", pidID)
		return nil, &utility.ResponseError{
			StatusCode: http.StatusNotFound,
			Error:      fmt.Errorf("PidController '%s' not found", pidID),
			Message:    "ResourceNotFound",
		}
	}
	
	// Build response
	response := gin.H{
		"@odata.type":         pc.OdataType,
		"@odata.id":           fmt.Sprintf("/MultiFish/v1/Platform/%s/Managers/%s/Oem/OpenBmc/Fan/PidControllers/%s", machineID, managerID, pidID),
		"Id":                  pidID,
		"FFGainCoefficient":   pc.FFGainCoefficient,
		"FFOffCoefficient":    pc.FFOffCoefficient,
		"ICoefficient":        pc.ICoefficient,
		"ILimitMax":           pc.ILimitMax,
		"ILimitMin":           pc.ILimitMin,
		"Inputs":              pc.Inputs,
		"NegativeHysteresis":  pc.NegativeHysteresis,
		"OutLimitMax":         pc.OutLimitMax,
		"OutLimitMin":         pc.OutLimitMin,
		"PCoefficient":        pc.PCoefficient,
		"PositiveHysteresis":  pc.PositiveHysteresis,
		"SetPoint":            pc.SetPoint,
		"SlewNeg":             pc.SlewNeg,
		"SlewPos":             pc.SlewPos,
		"Zones":               pc.Zones,
	}
	return response, nil
}

// PatchPidController patches the specified PidController
func (p *ExtendProvider) PatchPidController(manager interface{}, pidID string, updates interface{}) *utility.ResponseError {
	log := utility.GetLogger()

	em, ok := manager.(*ExtendManager)
	if !ok {
		log.Error().Msgf("invalid manager type for ExtendProvider: %T", manager)
		return &utility.ResponseError{
			StatusCode: http.StatusInternalServerError,
			Error:      fmt.Errorf("invalid manager type for ExtendProvider: %T", manager),
			Message:    "InternalError",
		}
	}
	
	pcPatch, ok := updates.(*PatchPidControllerType)
	if !ok {
		log.Error().Msgf("invalid patch type for ExtendProvider: %T", updates)
		return &utility.ResponseError{
			StatusCode: http.StatusBadRequest,
			Error:      fmt.Errorf("invalid patch type: %T", updates),
			Message:    "InvalidRequest",
		}
	}
	
	return patchExtendPidController(em, pidID, pcPatch)
}
