package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stmcginnis/gofish"

	"multifish/utility"
	extendprovider "multifish/providers/extend"
)

// ========== Utility Functions ==========

// PlatformManagerAdapter adapts *PlatformManager to the scheduler.JobPlatformManager interface
type PlatformManagerAdapter struct {
	mgr *PlatformManager
}

// NewPlatformManagerAdapter creates a new adapter
func NewPlatformManagerAdapter(mgr *PlatformManager) *PlatformManagerAdapter {
	return &PlatformManagerAdapter{mgr: mgr}
}

// GetMachine retrieves a machine and returns it as interface{}
func (pma *PlatformManagerAdapter) GetMachine(machineID string) (interface{}, error) {
	return pma.mgr.GetMachineInterface(machineID)
}

// getService retrieves the appropriate service interface based on the machine's service type
func getService(machine *MachineConnection) (interface{}, *utility.ResponseError) {
	log := utility.GetLogger()

	switch {
	case machine.Config.Type == string(ServiceTypeBase) && machine.BaseService != nil:
		return machine.BaseService, nil
	case machine.Config.Type == string(ServiceTypeExtend) && machine.ExtendService != nil:
		return machine.ExtendService, nil
	default:
		log.Error().Msgf("no service available for machine %s with type %s", machine.Config.ID, machine.Config.Type)
		return nil, &utility.ResponseError{
			StatusCode: http.StatusInternalServerError,
			Error:      fmt.Errorf("no service available for machine %s", machine.Config.ID),
			Message:    "ServiceUnavailable",
		}
	}
}

// ========== Data Structures ==========

// MachineConfig represents a single BMC machine configuration
type MachineConfig struct {
	ID                    string   `json:"Id"`
	Name                  string   `json:"Name,omitempty"`
	Type                  string   `json:"Type,omitempty"`
	TypeAllowableValues   []string `json:"Type@Redfish.AllowableValues,omitempty"`
	Endpoint              string   `json:"Endpoint"`
	Username              string   `json:"Username"`
	Password              string   `json:"Password"`
	Insecure              bool     `json:"Insecure"`
	HTTPClientTimeout     int      `json:"HTTPClientTimeout,omitempty"` // default: 30
	DisableEtagMatch      bool     `json:"DisableEtagMatch,omitempty"`  // default: false
}

// PatchMachineConfig represents allowed fields for PATCH operations
type PatchMachineConfig struct {
	Endpoint          *string `json:"Endpoint,omitempty"`
	Username          *string `json:"Username,omitempty"`
	Password          *string `json:"Password,omitempty"`
	HTTPClientTimeout *int    `json:"HTTPClientTimeout,omitempty"`
	DisableEtagMatch  *bool   `json:"DisableEtagMatch,omitempty"`
	Type              *string `json:"Type,omitempty"`
}

// ServiceType represents the type of Redfish service to use
type ServiceType string

const (
	ServiceTypeBase   ServiceType = "Base"
	ServiceTypeExtend ServiceType = "Extend"
)

// MachineConnection represents an active connection to a BMC
type MachineConnection struct {
	Config         MachineConfig
	Client         *gofish.APIClient
	BaseService    *gofish.Service
	ExtendService  *extendprovider.ExtendService
}

// PlatformManager manages multiple BMC connections
type PlatformManager struct {
	machines map[string]*MachineConnection
	mu       sync.RWMutex
}

// Global platform manager
var platformMgr = &PlatformManager{
	machines: make(map[string]*MachineConnection),
}

// ========== PlatformManager Methods ==========

// AddMachine adds a new machine to the platform
func (pm *PlatformManager) AddMachine(config MachineConfig) error {
	log := utility.GetLogger()

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Set default values
	if config.HTTPClientTimeout == 0 {
		config.HTTPClientTimeout = 30
	}
	
	// Set default Type and AllowableValues if not provided
	if config.Type == "" {
		config.Type = string(ServiceTypeExtend) // Default to Extend for backward compatibility
	}
	if len(config.TypeAllowableValues) == 0 {
		config.TypeAllowableValues = []string{string(ServiceTypeBase), string(ServiceTypeExtend)}
	}
	
	// Validate Type
	validType := false
	for _, allowable := range config.TypeAllowableValues {
		if config.Type == allowable {
			validType = true
			break
		}
	}
	if !validType {
		log.Error().Msgf("invalid Type '%s', must be one of: %v", config.Type, config.TypeAllowableValues)
		return fmt.Errorf("machine configuration validation failed: invalid Type '%s' for endpoint '%s', must be one of: %v (check your config file)", config.Type, config.Endpoint, config.TypeAllowableValues)
	}

	// Create custom HTTP client with timeout
	transport := &http.Transport{
		TLSHandshakeTimeout: 10 * time.Second,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     50,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false, // default is false, but we keep it explicit for clarity
	}
	
	// Skip TLS verification if Insecure is true
	if config.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	
	httpClient := &http.Client{
		Timeout:   time.Duration(config.HTTPClientTimeout) * time.Second,
		Transport: transport,
	}

	clientConfig := gofish.ClientConfig{
		Endpoint:          config.Endpoint,
		Username:          config.Username,
		Password:          config.Password,
		Insecure:          config.Insecure,
		HTTPClient:        httpClient,
	}

	client, err := gofish.Connect(clientConfig)
	if err != nil {
		log.Error().Msgf("failed to connect to %s: %v", config.Endpoint, err)
		return fmt.Errorf("failed to establish Redfish connection to endpoint '%s' (user: %s, timeout: %ds, insecure: %v): %w", config.Endpoint, config.Username, config.HTTPClientTimeout, config.Insecure, err)
	}

	// Create connection based on Type
	connection := &MachineConnection{
		Config:        config,
		Client:        client,
	}
	
	if config.Type == string(ServiceTypeExtend) {
		connection.ExtendService = extendprovider.NewExtendService(client)
		log.Info().
			Str("machineID", config.ID).
			Str("endpoint", config.Endpoint).
			Str("type", "ExtendService").
			Msg("Added machine")
	} else {
		connection.BaseService = client.Service
		log.Info().
			Str("machineID", config.ID).
			Str("endpoint", config.Endpoint).
			Str("type", "BaseService").
			Msg("Added machine")
	}
	
	pm.machines[config.ID] = connection

	return nil
}

// GetMachine retrieves a machine connection by ID
func (pm *PlatformManager) GetMachine(id string) (*MachineConnection, error) {
	log := utility.GetLogger()

	pm.mu.RLock()
	defer pm.mu.RUnlock()

	machine, ok := pm.machines[id]
	if !ok {
		log.Warn().Msgf("machine %s not found", id)
		return nil, fmt.Errorf("machine %s not found", id)
	}
	return machine, nil
}

// GetMachineInterface retrieves a machine connection as interface{} (implements providers.PlatformManager)
func (pm *PlatformManager) GetMachineInterface(id string) (interface{}, error) {
	return pm.GetMachine(id)
}

// ListMachines returns all machine configurations with passwords masked
func (pm *PlatformManager) ListMachines() []MachineConfig {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	configs := make([]MachineConfig, 0, len(pm.machines))
	for _, machine := range pm.machines {
		// Don't expose password in listings - use security utility
		config := machine.Config
		config.Password = utility.MaskPassword(config.Password)
		configs = append(configs, config)
	}
	return configs
}

// RemoveMachine removes a machine and closes its connection
func (pm *PlatformManager) RemoveMachine(id string) error {
	log := utility.GetLogger()

	pm.mu.Lock()
	defer pm.mu.Unlock()

	machine, ok := pm.machines[id]
	if !ok {
		log.Warn().Msgf("machine %s not found for removal", id)
		return fmt.Errorf("machine with ID '%s' not found in registry (available machines: %d). Verify the machine ID or check /machines endpoint", id, len(pm.machines))
	}

	// Logout to close the session
	machine.Client.Logout()
	
	// Close idle connections to prevent connection leak
	if machine.Client.HTTPClient != nil && machine.Client.HTTPClient.Transport != nil {
		if transport, ok := machine.Client.HTTPClient.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}
	
	delete(pm.machines, id)
	log.Info().Str("machineID", id).Msg("Removed machine")
	return nil
}

// CleanupAll closes all machine connections
func (pm *PlatformManager) CleanupAll() {
	log := utility.GetLogger()

	pm.mu.Lock()
	defer pm.mu.Unlock()

	for id, machine := range pm.machines {
		// Logout to close the session
		machine.Client.Logout()
		
		// Close idle connections to prevent connection leak
		if machine.Client.HTTPClient != nil && machine.Client.HTTPClient.Transport != nil {
			if transport, ok := machine.Client.HTTPClient.Transport.(*http.Transport); ok {
				transport.CloseIdleConnections()
			}
		}
		log.Info().Str("machineID", id).Msg("Cleaned up machine")
	}
	
	// Clear the map
	pm.machines = make(map[string]*MachineConnection)
}

// ========== API Handlers ==========

// GET /MultiFish/v1/Platform - List all machines
func getPlatform(c *gin.Context) {
	machines := platformMgr.ListMachines()
	
	members := make([]map[string]string, len(machines))
	for i, machine := range machines {
		members[i] = map[string]string{
			"@odata.id": fmt.Sprintf("/MultiFish/v1/Platform/%s", machine.ID),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"@odata.type":  "#MachineCollection.MachineCollection",
		"@odata.id":    "/MultiFish/v1/Platform",
		"Name":         "Platform Machine Collection",
		"Members":      members,
		"Members@odata.count": len(members),
	})
}

// POST /MultiFish/v1/Platform - Add a new machine
func addMachine(c *gin.Context) {
	var config MachineConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		utility.RedfishError(c, http.StatusBadRequest, "Invalid request body", "InvalidJSON")
		return
	}

	if config.ID == "" {
		utility.RedfishError(c, http.StatusBadRequest, "Machine ID is required", "PropertyMissing")
		return
	}

	if err := platformMgr.AddMachine(config); err != nil {
		utility.RedfishError(c, http.StatusInternalServerError, err.Error(), "InternalError")
		return
	}

	c.Header("Location", fmt.Sprintf("/MultiFish/v1/Platform/%s", config.ID))
	c.JSON(http.StatusCreated, gin.H{
		"@odata.id": fmt.Sprintf("/MultiFish/v1/Platform/%s", config.ID),
		"Id":        config.ID,
		"Name":      config.Name,
	})
}

// GET /MultiFish/v1/Platform/:machineId - Get machine details
func getMachine(c *gin.Context) {
	machineID := c.Param("machineId")
	
	machine, err := platformMgr.GetMachine(machineID)
	if err != nil {
		utility.RedfishError(c, http.StatusNotFound, err.Error(), "ResourceNotFound")
		return
	}

	// Get managers based on service type
	var managerLinks []map[string]string
	
	if machine.Config.Type == string(ServiceTypeBase) && machine.BaseService != nil {
		managers, err := machine.BaseService.Managers()
		if err != nil {
			utility.RedfishError(c, http.StatusInternalServerError, "Failed to get managers", "InternalError")
			return
		}
		managerLinks = make([]map[string]string, len(managers))
		for i, mgr := range managers {
			managerLinks[i] = map[string]string{
				"@odata.id": fmt.Sprintf("/MultiFish/v1/Platform/%s/Managers/%s", machineID, mgr.ID),
			}
		}
	} else if machine.Config.Type == string(ServiceTypeExtend) && machine.ExtendService != nil {
		managers, err := machine.ExtendService.Managers()
		if err != nil {
			utility.RedfishError(c, http.StatusInternalServerError, "Failed to get managers", "InternalError")
			return
		}
		managerLinks = make([]map[string]string, len(managers))
		for i, mgr := range managers {
			managerLinks[i] = map[string]string{
				"@odata.id": fmt.Sprintf("/MultiFish/v1/Platform/%s/Managers/%s", machineID, mgr.GetManager().ID),
			}
		}
	}

	response := gin.H{
		"@odata.type": "#Machine.v1_0_0.Machine",
		"@odata.id":   fmt.Sprintf("/MultiFish/v1/Platform/%s", machineID),
		"Id":          machine.Config.ID,
		"Name":        machine.Config.Name,
		"Type":        machine.Config.Type,
		"Type@Redfish.AllowableValues": machine.Config.TypeAllowableValues,
		"Description": "BMC Machine Resource",
		"Connection": gin.H{
			"Endpoint": machine.Config.Endpoint,
			"Username": machine.Config.Username,
			"Password": utility.MaskPassword(machine.Config.Password), // Never expose password
			"Insecure": machine.Config.Insecure,
			"HTTPClientTimeout":  machine.Config.HTTPClientTimeout,
			"DisableEtagMatch": machine.Config.DisableEtagMatch,
		},
		"Managers": gin.H{
			"@odata.id":           fmt.Sprintf("/MultiFish/v1/Platform/%s/Managers", machineID),
			"Members":             managerLinks,
			"Members@odata.count": len(managerLinks),
		},
	}

	c.JSON(http.StatusOK, response)
}

// PATCH /MultiFish/v1/Platform/:machineId - Update machine configuration
func updateMachine(c *gin.Context) {
	machineID := c.Param("machineId")
	
	// Define allowed fields with their expected types for PATCH operations
	allowedPatchFields := utility.FieldSpecMap{
		"Endpoint":          {Name: "Endpoint", Expected: "string"},
		"Username":          {Name: "Username", Expected: "string"},
		"Password":          {Name: "Password", Expected: "string"},
		"HTTPClientTimeout": {Name: "HTTPClientTimeout", Expected: "number"},
		"DisableEtagMatch":  {Name: "DisableEtagMatch", Expected: "bool"},
		"Type":              {Name: "Type", Expected: "string"},
	}

	// Validate and bind the patch data to struct
	var updates PatchMachineConfig
	if !utility.CheckAndBindPatchPayload(c, allowedPatchFields, &updates) {
		return
	}

	machine, err := platformMgr.GetMachine(machineID)
	if err != nil {
		utility.RedfishError(c, http.StatusNotFound, err.Error(), "ResourceNotFound")
		return
	}

	// Update configuration
	platformMgr.mu.Lock()
	needsReconnect := false
	
	if updates.Endpoint != nil && *updates.Endpoint != machine.Config.Endpoint {
		machine.Config.Endpoint = *updates.Endpoint
		needsReconnect = true
	}
	if updates.Username != nil && *updates.Username != machine.Config.Username {
		machine.Config.Username = *updates.Username
		needsReconnect = true
	}
	if updates.Password != nil {
		machine.Config.Password = *updates.Password
		needsReconnect = true
	}
	if updates.HTTPClientTimeout != nil && *updates.HTTPClientTimeout > 0 {
		machine.Config.HTTPClientTimeout = *updates.HTTPClientTimeout
	}
	// Update DisableEtagMatch setting (can be toggled without reconnection)
	if updates.DisableEtagMatch != nil && *updates.DisableEtagMatch != machine.Config.DisableEtagMatch {
		log := utility.GetLogger()
		machine.Config.DisableEtagMatch = *updates.DisableEtagMatch
		log.Info().
			Str("machineID", machineID).
			Bool("disableEtagMatch", *updates.DisableEtagMatch).
			Msg("Updated DisableEtagMatch setting")
	}
	if updates.Type != nil && *updates.Type != machine.Config.Type {
		// Validate new Type
		validType := false
		for _, allowable := range machine.Config.TypeAllowableValues {
			if *updates.Type == allowable {
				validType = true
				break
			}
		}
		if !validType {
			platformMgr.mu.Unlock()
			utility.RedfishError(c, http.StatusBadRequest, 
				fmt.Sprintf("Invalid Type '%s', must be one of: %v", *updates.Type, machine.Config.TypeAllowableValues),
				"InvalidValue")
			return
		}
		machine.Config.Type = *updates.Type
		needsReconnect = true
	}
	platformMgr.mu.Unlock()

	// If critical settings changed, suggest reconnection
	message := "Configuration updated successfully"
	if needsReconnect {
		message = "Configuration updated successfully. Reconnection may be required for some changes to take effect."
	}

	c.JSON(http.StatusOK, gin.H{
		"@odata.id": fmt.Sprintf("/MultiFish/v1/Platform/%s", machineID),
		"Id":        machineID,
		"Message":   message,
	})
}

// DELETE /MultiFish/v1/Platform/:machineId - Remove a machine
func deleteMachine(c *gin.Context) {
	machineID := c.Param("machineId")
	
	if err := platformMgr.RemoveMachine(machineID); err != nil {
		utility.RedfishError(c, http.StatusNotFound, err.Error(), "ResourceNotFound")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Message": fmt.Sprintf("Machine %s removed successfully", machineID),
	})
}

// ========== Route Setup ==========

// platformRoutes sets up the platform-related routes
func platformRoutes(router *gin.Engine) {
	router.GET("/MultiFish/v1/Platform", getPlatform)
	router.POST("/MultiFish/v1/Platform", addMachine)
	router.GET("/MultiFish/v1/Platform/:machineId", getMachine)
	router.PATCH("/MultiFish/v1/Platform/:machineId", updateMachine)
	router.DELETE("/MultiFish/v1/Platform/:machineId", deleteMachine)
}

