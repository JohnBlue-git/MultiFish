package redfish

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stmcginnis/gofish/redfish"
	
	"multifish/utility"
)

// ========== Types ==========

// PatchManagerType defines the fields that can be patched on a manager
type PatchManagerType struct {
	ServiceIdentification string `json:"ServiceIdentification,omitempty"`
}

// ManagerAllowedPatchFields returns the allowed fields for patching a manager
func ManagerAllowedPatchFields() utility.FieldSpecMap {
	return utility.FieldSpecMap{
		"ServiceIdentification": {Name: "ServiceIdentification", Expected: "string"},
	}
}

// ========== Helper Functions ==========

// RefreshManager re-fetches the manager from the server
func RefreshManager(manager *redfish.Manager) error {
	log := utility.GetLogger()

	client := manager.GetClient()
	if client == nil {
		log.Error().Msg("cannot refresh: no client available")
		return fmt.Errorf("cannot refresh manager: Redfish client is not initialized or connection was closed. Reconnect to the BMC endpoint")
	}

	// Re-fetch the manager from the server using its ODataID
	newMgr, err := redfish.GetManager(client, manager.ODataID)
	if err != nil {
		log.Error().Msgf("failed to refresh manager: %v", err)
		return fmt.Errorf("failed to refresh manager from endpoint '%s': %w. Check network connectivity and BMC availability", manager.ODataID, err)
	}

	// Update the manager object
	manager = newMgr

	return nil
}

// UpdateAndRefreshManager updates the manager and refreshes the object
func UpdateAndRefreshManager(manager *redfish.Manager) error {
	log := utility.GetLogger()

	// Apply ETag matching setting from machine configuration
	manager.DisableEtagMatch(true)

	// Update the manager
	if err := manager.Update(); err != nil {
		log.Error().Msgf("Manager update failed: %v", err)
		return fmt.Errorf("failed to update manager at endpoint '%s': %w. Verify payload is valid and BMC supports the requested changes", manager.ODataID, err)
	}

	// Refresh again to get the new ETag and updated data after successful update
	// if err := RefreshManager(manager); err != nil {
	// 	log.Printf("Warning: Failed to refresh manager after update: %v", err)
	// 	// Don't return error, as the update was successful
	// }
	
	return nil
}

// ========== Implementations ==========

// PatchManagerData patches the ServiceIdentification field of the manager
func PatchManagerData(manager *redfish.Manager, patch *PatchManagerType) *utility.ResponseError {
	log := utility.GetLogger()

	// Validate patch
	if patch == nil {
		log.Error().Msg("patch cannot be nil")
		return &utility.ResponseError{
			StatusCode: http.StatusBadRequest,
			Error:      fmt.Errorf("patch cannot be nil"),
			Message:    "InvalidRequest",
		}
	}

	// Update local value
	manager.ServiceIdentification = patch.ServiceIdentification

	// Use the Update and Refresh method to apply the change
	if err := UpdateAndRefreshManager(manager); err != nil {
		log.Error().Msgf("failed to patch ServiceIdentification: %v", err)
		return &utility.ResponseError{
			StatusCode: http.StatusInternalServerError,
			Error:      fmt.Errorf("failed to patch ServiceIdentification: %v", err),
			Message:    "InternalError",
		}
	}

	return nil
}

// ========== Provider Implementation ==========

// RedfishProvider handles base redfish.Manager types
type RedfishProvider struct{}

// TypeName returns the name of this provider
func (p *RedfishProvider) TypeName() string {
	return "Redfish"
}

// Supports checks if this provider can handle the given manager instance
func (p *RedfishProvider) Supports(v interface{}) bool {
	_, ok := v.(*redfish.Manager)
	return ok
}

// SupportsCollection checks if this provider can handle the given manager collection
func (p *RedfishProvider) SupportsCollection(v interface{}) bool {
	_, ok := v.([]*redfish.Manager)
	return ok
}

// GetManagerCollectionResponse builds the response for a collection of managers
func (p *RedfishProvider) GetManagerCollectionResponse(managers interface{}, machineID string) (gin.H, error) {
	log := utility.GetLogger()

	mgrs, ok := managers.([]*redfish.Manager)
	if !ok {
		log.Error().Msgf("invalid manager collection type for RedfishProvider: %T", managers)
		return nil, fmt.Errorf("invalid manager type for RedfishProvider: %T", managers)
	}

	var members []gin.H
	for _, mgr := range mgrs {
		member := gin.H{
			"@odata.id": fmt.Sprintf("/MultiFish/v1/Platform/%s/Managers/%s", machineID, mgr.ID),
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
func (p *RedfishProvider) GetManagerResponse(manager interface{}, machineID, managerID string) (gin.H, *utility.ResponseError) {
	log := utility.GetLogger()

	mgr, ok := manager.(*redfish.Manager)
	if !ok {
		log.Error().Msgf("invalid manager type for RedfishProvider: %T", manager)
		return nil, &utility.ResponseError{
			StatusCode: http.StatusInternalServerError,
			Error:      fmt.Errorf("invalid manager type for RedfishProvider: %T", manager),
			Message:    "InternalError",
		}
	}
	
	return gin.H{
		"@odata.type":           mgr.ODataType,
		"@odata.id":             fmt.Sprintf("/MultiFish/v1/Platform/%s/Managers/%s", machineID, managerID),
		"Id":                    mgr.ID,
		"Name":                  mgr.Name,
		"ManagerType":           mgr.ManagerType,
		"FirmwareVersion":       mgr.FirmwareVersion,
		"UUID":                  mgr.UUID,
		"Status":                mgr.Status,
		"Description":           mgr.Description,
		"Model":                 mgr.Model,
		"DateTime":              mgr.DateTime,
		"PowerState":            mgr.PowerState,
		"GraphicalConsole":      mgr.GraphicalConsole,
		"SerialConsole":         mgr.SerialConsole,
		"CommandShell":          mgr.CommandShell,
		"NetworkProtocol":       mgr.NetworkProtocol,
		"ServiceIdentification": mgr.ServiceIdentification,
	}, nil
}

// PatchManager patches the manager with the given updates
func (p *RedfishProvider) PatchManager(manager interface{}, updates interface{}) *utility.ResponseError {
	log := utility.GetLogger()

	mgr, ok := manager.(*redfish.Manager)
	if !ok {
		log.Error().Msgf("invalid manager type for RedfishProvider: %T", manager)
		return &utility.ResponseError{
			StatusCode: http.StatusInternalServerError,
			Error:      fmt.Errorf("invalid manager type for RedfishProvider: %T", manager),
			Message:    "InternalError",
		}
	}
	
	patch, ok := updates.(*PatchManagerType)
	if !ok {
		log.Error().Msgf("invalid patch type for RedfishProvider: %T", updates)
		return &utility.ResponseError{
			StatusCode: http.StatusBadRequest,
			Error:      fmt.Errorf("invalid patch type: %T", updates),
			Message:    "InvalidRequest",
		}
	}
	
	err := PatchManagerData(mgr, patch)
	if err != nil {
		log.Error().Msgf("failed to patch manager: %v", err)
		return &utility.ResponseError{
			StatusCode: err.StatusCode,
			Error:      err.Error,
			Message:    err.Message,
		}
	}
	
	return nil
}

// GetProfileResponse - Redfish base managers don't support profiles
func (p *RedfishProvider) GetProfileResponse(manager interface{}, machineID, managerID string) (gin.H, *utility.ResponseError) {
	return nil, &utility.ResponseError{
		StatusCode: http.StatusNotFound,
		Error:      fmt.Errorf("Profile not supported for base Redfish managers"),
		Message:    "ResourceNotFound",
	}
}

// PatchProfile - Redfish base managers don't support profiles
func (p *RedfishProvider) PatchProfile(manager interface{}, updates interface{}) *utility.ResponseError {
	return &utility.ResponseError{
		StatusCode: http.StatusNotFound,
		Error:      fmt.Errorf("Profile not supported for base Redfish managers"),
		Message:    "ResourceNotFound",
	}
}

// GetFanControllerCollectionResponse - Redfish base managers don't support fan controllers
func (p *RedfishProvider) GetFanControllerCollectionResponse(manager interface{}, machineID, managerID string) (gin.H, *utility.ResponseError) {
	return nil, &utility.ResponseError{
		StatusCode: http.StatusNotFound,
		Error:      fmt.Errorf("FanControllers not supported for base Redfish managers"),
		Message:    "ResourceNotFound",
	}
}

// GetFanControllerResponse - Redfish base managers don't support fan controllers
func (p *RedfishProvider) GetFanControllerResponse(manager interface{}, machineID, managerID, fanID string) (gin.H, *utility.ResponseError) {
	return nil, &utility.ResponseError{
		StatusCode: http.StatusNotFound,
		Error:      fmt.Errorf("FanControllers not supported for base Redfish managers"),
		Message:    "ResourceNotFound",
	}
}

// PatchFanController - Redfish base managers don't support fan controllers
func (p *RedfishProvider) PatchFanController(manager interface{}, fanID string, updates interface{}) *utility.ResponseError {
	return &utility.ResponseError{
		StatusCode: http.StatusInternalServerError,
		Error:      fmt.Errorf("FanControllers not supported for base Redfish managers"),
		Message:    "InternalError",
	}
}

// GetFanZoneCollectionResponse - Redfish base managers don't support fan zones
func (p *RedfishProvider) GetFanZoneCollectionResponse(manager interface{}, machineID, managerID string) (gin.H, *utility.ResponseError) {
	return nil, &utility.ResponseError{
		StatusCode: http.StatusNotFound,
		Error:      fmt.Errorf("FanZones not supported for base Redfish managers"),
		Message:    "ResourceNotFound",
	}
}

// GetFanZoneResponse - Redfish base managers don't support fan zones
func (p *RedfishProvider) GetFanZoneResponse(manager interface{}, machineID, managerID, zoneID string) (gin.H, *utility.ResponseError) {
	return nil, &utility.ResponseError{
		StatusCode: http.StatusNotFound,
		Error:      fmt.Errorf("FanZones not supported for base Redfish managers"),
		Message:    "ResourceNotFound",
	}
}

// PatchFanZone - Redfish base managers don't support fan zones
func (p *RedfishProvider) PatchFanZone(manager interface{}, zoneID string, updates interface{}) *utility.ResponseError {
	return &utility.ResponseError{
		StatusCode: http.StatusInternalServerError,
		Error:      fmt.Errorf("FanZones not supported for base Redfish managers"),
		Message:    "InternalError",
	}
}

// GetPidControllerCollectionResponse - Redfish base managers don't support PID controllers
func (p *RedfishProvider) GetPidControllerCollectionResponse(manager interface{}, machineID, managerID string) (gin.H, *utility.ResponseError) {
	return nil, &utility.ResponseError{
		StatusCode: http.StatusNotFound,
		Error:      fmt.Errorf("PidControllers not supported for base Redfish managers"),
		Message:    "ResourceNotFound",
	}
}

// GetPidControllerResponse - Redfish base managers don't support PID controllers
func (p *RedfishProvider) GetPidControllerResponse(manager interface{}, machineID, managerID, pidID string) (gin.H, *utility.ResponseError) {
	return nil, &utility.ResponseError{
		StatusCode: http.StatusNotFound,
		Error:      fmt.Errorf("PidControllers not supported for base Redfish managers"),
		Message:    "ResourceNotFound",
	}
}

// PatchPidController - Redfish base managers don't support PID controllers
func (p *RedfishProvider) PatchPidController(manager interface{}, pidID string, updates interface{}) *utility.ResponseError {
	return &utility.ResponseError{
		StatusCode: http.StatusInternalServerError,
		Error:      fmt.Errorf("PidControllers not supported for base Redfish managers"),
		Message:    "InternalError",
	}
}
