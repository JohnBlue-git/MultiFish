package providers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"multifish/utility"
)

// ManagerProvider defines the interface that all manager type providers must implement
// This allows adding new manager types (Dell, HP, Cisco, etc.) without modifying existing code
type ManagerProvider interface {
	// TypeName returns a human-readable name for this provider (e.g., "Redfish", "Extended", "Dell")
	TypeName() string
	
	// Supports checks if this provider can handle the given manager instance
	Supports(v interface{}) bool
	
	// SupportsCollection checks if this provider can handle the given manager collection
	SupportsCollection(v interface{}) bool
	
	// ========== Manager Collection Operations ==========
	
	// GetManagerCollectionResponse builds the response for a collection of managers
	GetManagerCollectionResponse(managers interface{}, machineID string) (gin.H, error)
	
	// ========== Individual Manager Operations ==========
	
	// GetManagerResponse builds the response for a single manager
	GetManagerResponse(manager interface{}, machineID, managerID string) (gin.H, *utility.ResponseError)
	
	// PatchManager patches the manager with the given updates
	PatchManager(manager interface{}, updates interface{}) *utility.ResponseError
	
	// ========== Profile Operations ==========
	
	// GetProfileResponse builds the response for the manager's profile
	// Returns nil, nil if profiles are not supported by this provider
	GetProfileResponse(manager interface{}, machineID, managerID string) (gin.H, *utility.ResponseError)
	
	// PatchProfile patches the manager's profile
	// Returns error if profiles are not supported by this provider
	PatchProfile(manager interface{}, updates interface{}) *utility.ResponseError
	
	// ========== FanController Operations ==========
	
	// GetFanControllerCollectionResponse builds the response for the FanControllers collection
	// Returns nil, error if fan controllers are not supported by this provider
	GetFanControllerCollectionResponse(manager interface{}, machineID, managerID string) (gin.H, *utility.ResponseError)
	
	// GetFanControllerResponse builds the response for a specific FanController
	// Returns nil, error if fan controllers are not supported by this provider
	GetFanControllerResponse(manager interface{}, machineID, managerID, fanID string) (gin.H, *utility.ResponseError)
	
	// PatchFanController patches the specified FanController
	// Returns error if fan controllers are not supported by this provider
	PatchFanController(manager interface{}, fanID string, updates interface{}) *utility.ResponseError
	
	// ========== FanZone Operations ==========
	
	// GetFanZoneCollectionResponse builds the response for the FanZones collection
	// Returns nil, error if fan zones are not supported by this provider
	GetFanZoneCollectionResponse(manager interface{}, machineID, managerID string) (gin.H, *utility.ResponseError)
	
	// GetFanZoneResponse builds the response for a specific FanZone
	// Returns nil, error if fan zones are not supported by this provider
	GetFanZoneResponse(manager interface{}, machineID, managerID, zoneID string) (gin.H, *utility.ResponseError)
	
	// PatchFanZone patches the specified FanZone
	// Returns error if fan zones are not supported by this provider
	PatchFanZone(manager interface{}, zoneID string, updates interface{}) *utility.ResponseError
	
	// ========== PidController Operations ==========
	
	// GetPidControllerCollectionResponse builds the response for the PidControllers collection
	// Returns nil, error if PID controllers are not supported by this provider
	GetPidControllerCollectionResponse(manager interface{}, machineID, managerID string) (gin.H, *utility.ResponseError)
	
	// GetPidControllerResponse builds the response for a specific PidController
	// Returns nil, error if PID controllers are not supported by this provider
	GetPidControllerResponse(manager interface{}, machineID, managerID, pidID string) (gin.H, *utility.ResponseError)
	
	// PatchPidController patches the specified PidController
	// Returns error if PID controllers are not supported by this provider
	PatchPidController(manager interface{}, pidID string, updates interface{}) *utility.ResponseError
}

// ========== Manager Registry ==========

// ManagerRegistry is a specialized registry for ManagerProvider instances
// It wraps the generic registry and adds delegation methods for manager-specific operations
type ManagerRegistry struct {
	*GenericRegistry[ManagerProvider]
}

// NewManagerRegistry creates a new manager provider registry
func NewManagerRegistry() *ManagerRegistry {
	return &ManagerRegistry{
		GenericRegistry: NewGenericRegistry[ManagerProvider](),
	}
}

// ========== Delegation Methods ==========

// GetManagerCollectionResponse delegates to the appropriate provider
func (r *ManagerRegistry) GetManagerCollectionResponse(managers interface{}, machineID string) (gin.H, error) {
	provider, err := r.FindCollectionProvider(managers)
	if err != nil {
		utility.GetLogger().Error().Msgf("failed to find collection provider: %v", err)
		return nil, err
	}
	return provider.GetManagerCollectionResponse(managers, machineID)
}

// GetManagerResponse delegates to the appropriate provider
func (r *ManagerRegistry) GetManagerResponse(manager interface{}, machineID, managerID string) (gin.H, *utility.ResponseError) {
	provider, err := r.FindProvider(manager)
	if err != nil {
		utility.GetLogger().Error().Msgf("failed to find provider: %v", err)
		return nil, &utility.ResponseError{
			StatusCode: http.StatusNotFound,
			Error:      err,
			Message:    "UnsupportedManager",
		}
	}
	return provider.GetManagerResponse(manager, machineID, managerID)
}

// PatchManager delegates to the appropriate provider
func (r *ManagerRegistry) PatchManager(manager interface{}, updates interface{}) *utility.ResponseError {
	provider, err := r.FindProvider(manager)
	if err != nil {
		utility.GetLogger().Error().Msgf("failed to find provider: %v", err)
		return &utility.ResponseError{
			StatusCode: http.StatusNotFound,
			Error:      err,
			Message:    "UnsupportedManager",
		}
	}
	return provider.PatchManager(manager, updates)
}

// GetProfileResponse delegates to the appropriate provider
func (r *ManagerRegistry) GetProfileResponse(manager interface{}, machineID, managerID string) (gin.H, *utility.ResponseError) {
	provider, err := r.FindProvider(manager)
	if err != nil {
		utility.GetLogger().Error().Msgf("failed to find provider: %v", err)
		return nil, &utility.ResponseError{
			StatusCode: http.StatusNotFound,
			Error:      err,
			Message:    "UnsupportedManager",
		}
	}
	return provider.GetProfileResponse(manager, machineID, managerID)
}

// PatchProfile delegates to the appropriate provider
func (r *ManagerRegistry) PatchProfile(manager interface{}, updates interface{}) *utility.ResponseError {
	provider, err := r.FindProvider(manager)
	if err != nil {
		utility.GetLogger().Error().Msgf("failed to find provider: %v", err)
		return &utility.ResponseError{
			StatusCode: http.StatusNotFound,
			Error:      err,
			Message:    "UnsupportedManager",
		}
	}
	return provider.PatchProfile(manager, updates)
}

// GetFanControllerCollectionResponse delegates to the appropriate provider
func (r *ManagerRegistry) GetFanControllerCollectionResponse(manager interface{}, machineID, managerID string) (gin.H, *utility.ResponseError) {
	provider, err := r.FindProvider(manager)
	if err != nil {
		utility.GetLogger().Error().Msgf("failed to find provider: %v", err)
		return nil, &utility.ResponseError{
			StatusCode: http.StatusNotFound,
			Error:      err,
			Message:    "UnsupportedManager",
		}
	}
	return provider.GetFanControllerCollectionResponse(manager, machineID, managerID)
}

// GetFanControllerResponse delegates to the appropriate provider
func (r *ManagerRegistry) GetFanControllerResponse(manager interface{}, machineID, managerID, fanID string) (gin.H, *utility.ResponseError) {
	provider, err := r.FindProvider(manager)
	if err != nil {
		utility.GetLogger().Error().Msgf("failed to find provider: %v", err)
		return nil, &utility.ResponseError{
			StatusCode: http.StatusNotFound,
			Error:      err,
			Message:    "UnsupportedManager",
		}
	}
	return provider.GetFanControllerResponse(manager, machineID, managerID, fanID)
}

// PatchFanController delegates to the appropriate provider
func (r *ManagerRegistry) PatchFanController(manager interface{}, fanID string, updates interface{}) *utility.ResponseError {
	provider, err := r.FindProvider(manager)
	if err != nil {
		utility.GetLogger().Error().Msgf("failed to find provider: %v", err)
		return &utility.ResponseError{
			StatusCode: http.StatusInternalServerError,
			Error:      err,
			Message:    "UnsupportedManager",
		}
	}
	return provider.PatchFanController(manager, fanID, updates)
}

// GetFanZoneCollectionResponse delegates to the appropriate provider
func (r *ManagerRegistry) GetFanZoneCollectionResponse(manager interface{}, machineID, managerID string) (gin.H, *utility.ResponseError) {
	provider, err := r.FindProvider(manager)
	if err != nil {
		utility.GetLogger().Error().Msgf("failed to find provider: %v", err)
		return nil, &utility.ResponseError{
			StatusCode: http.StatusNotFound,
			Error:      err,
			Message:    "UnsupportedManager",
		}
	}
	return provider.GetFanZoneCollectionResponse(manager, machineID, managerID)
}

// GetFanZoneResponse delegates to the appropriate provider
func (r *ManagerRegistry) GetFanZoneResponse(manager interface{}, machineID, managerID, zoneID string) (gin.H, *utility.ResponseError) {
	provider, err := r.FindProvider(manager)
	if err != nil {
		utility.GetLogger().Error().Msgf("failed to find provider: %v", err)
		return nil, &utility.ResponseError{
			StatusCode: http.StatusNotFound,
			Error:      err,
			Message:    "UnsupportedManager",
		}
	}
	return provider.GetFanZoneResponse(manager, machineID, managerID, zoneID)
}

// PatchFanZone delegates to the appropriate provider
func (r *ManagerRegistry) PatchFanZone(manager interface{}, zoneID string, updates interface{}) *utility.ResponseError {
	provider, err := r.FindProvider(manager)
	if err != nil {
		utility.GetLogger().Error().Msgf("failed to find provider: %v", err)
		return &utility.ResponseError{
			StatusCode: http.StatusInternalServerError,
			Error:      err,
			Message:    "UnsupportedManager",
		}
	}
	return provider.PatchFanZone(manager, zoneID, updates)
}

// GetPidControllerCollectionResponse delegates to the appropriate provider
func (r *ManagerRegistry) GetPidControllerCollectionResponse(manager interface{}, machineID, managerID string) (gin.H, *utility.ResponseError) {
	provider, err := r.FindProvider(manager)
	if err != nil {
		utility.GetLogger().Error().Msgf("failed to find provider: %v", err)
		return nil, &utility.ResponseError{
			StatusCode: http.StatusNotFound,
			Error:      err,
			Message:    "UnsupportedManager",
		}
	}
	return provider.GetPidControllerCollectionResponse(manager, machineID, managerID)
}

// GetPidControllerResponse delegates to the appropriate provider
func (r *ManagerRegistry) GetPidControllerResponse(manager interface{}, machineID, managerID, pidID string) (gin.H, *utility.ResponseError) {
	provider, err := r.FindProvider(manager)
	if err != nil {
		utility.GetLogger().Error().Msgf("failed to find provider: %v", err)
		return nil, &utility.ResponseError{
			StatusCode: http.StatusNotFound,
			Error:      err,
			Message:    "UnsupportedManager",
		}
	}
	return provider.GetPidControllerResponse(manager, machineID, managerID, pidID)
}

// PatchPidController delegates to the appropriate provider
func (r *ManagerRegistry) PatchPidController(manager interface{}, pidID string, updates interface{}) *utility.ResponseError {
	provider, err := r.FindProvider(manager)
	if err != nil {
		utility.GetLogger().Error().Msgf("failed to find provider: %v", err)
		return &utility.ResponseError{
			StatusCode: http.StatusInternalServerError,
			Error:      err,
			Message:    "UnsupportedManager",
		}
	}
	return provider.PatchPidController(manager, pidID, updates)
}
