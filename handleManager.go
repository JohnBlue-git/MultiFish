package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"

	"multifish/utility"
	"multifish/providers"
	redfishprovider "multifish/providers/redfish"
	extendprovider "multifish/providers/extend"
)

// Global manager provider registry
var managerProviders *providers.ManagerRegistry

// init registers all manager providers
func init() {
	managerProviders = providers.NewManagerRegistry()
	
	// Register providers in order of preference
	// Extended managers should be checked before base redfish managers
	managerProviders.Register(&extendprovider.ExtendProvider{})
	managerProviders.Register(&redfishprovider.RedfishProvider{})
	
	// Future: Add more providers here
	// managerProviders.Register(&dellprovider.DellProvider{})
	// managerProviders.Register(&hpprovider.HPProvider{})
}

// ========== Helper type for manager callback ==========

type ManagerResponseCallback func(targetManager interface{})

func getManagerByService(machine *MachineConnection, managerID string) (interface{}, *utility.ResponseError) {
	log := utility.GetLogger()

	// Get the service for the machine
	service, err := getService(machine)
	if err != nil {
		log.Error().Msgf("failed to get service for machine %s: %v", machine.Config.ID, err)
		return nil, err
	}

	// Get managers from the service
	var managers interface{}
	var managersErr error
	switch svc := service.(type) {
	case *gofish.Service:
		managers, managersErr = svc.Managers()
	case *extendprovider.ExtendService:
		managers, managersErr = svc.Managers()
	default:
		log.Error().Msgf("failed to get managers from machine %s: %v", machine.Config.ID, err)
		return nil, &utility.ResponseError{
			StatusCode: http.StatusInternalServerError,
			Error:      fmt.Errorf("unsupported service type: %T", svc),
			Message:    "InternalError",
		}
	}
	if managersErr != nil {
		log.Error().Msgf("failed to get managers from machine %s: %v", machine.Config.ID, managersErr)
		return nil, &utility.ResponseError{
			StatusCode: http.StatusInternalServerError,
			Error:      fmt.Errorf("failed to get managers: %v", managersErr),
			Message:    "InternalError",
		}
	}

	// Find the target manager by ID
	switch mgrs := managers.(type) {
	case []*redfish.Manager:
		for _, manager := range mgrs {
			if manager.ID == managerID {
				return manager, nil
			}
		}
	case []*extendprovider.ExtendManager:
		for _, manager := range mgrs {
			if manager.GetManager().ID == managerID {
				return manager, nil
			}
		}
	default:
		log.Error().Msgf("unsupported manager collection type: %T", mgrs)
		return nil, &utility.ResponseError{
			StatusCode: http.StatusInternalServerError,
			Error:      fmt.Errorf("unsupported manager type: %T", mgrs),
			Message:    "InternalError",
		}
	}
	log.Warn().Msgf("manager %s not found for machine %s", managerID, machine.Config.ID)
	return nil, &utility.ResponseError{
		StatusCode: http.StatusNotFound,
		Error:      fmt.Errorf("manager %s not found", managerID),
		Message:    "ResourceNotFound",
	}
}

func getManagersCallback(c *gin.Context, machine *MachineConnection, managerID string, callback ManagerResponseCallback) {
	log := utility.GetLogger()

	targetManager, err := getManagerByService(machine, managerID)
	if err != nil {
		log.Error().Msgf("failed to get manager %s for machine %s: %v", managerID, machine.Config.ID, err)
		utility.RedfishError(c, err.StatusCode, err.Error.Error(), err.Message)
		return
	}
	callback(targetManager)
}

// ========= MachineActionExecutor implementation ==========

func (m *MachineActionExecutorAdapter) GetManagerByService(machine interface{}, managerID string) (interface{}, error) {
	machineConn, ok := machine.(*MachineConnection)
	if !ok {
		return nil, fmt.Errorf("invalid machine type: expected *MachineConnection, got %T", machine)
	}
	
	manager, respErr := getManagerByService(machineConn, managerID)
	if respErr != nil {
		return nil, respErr.Error
	}
	return manager, nil
}

func (m *MachineActionExecutorAdapter) PatchManager(manager interface{}, patch interface{}) error {
	respErr := PatchManager(manager, patch)
	if respErr != nil {
		return respErr.Error
	}
	return nil
}

func (m *MachineActionExecutorAdapter) PatchProfile(manager interface{}, patch extendprovider.PatchProfileType) error {
	respErr := PatchProfile(manager, patch)
	if respErr != nil {
		return respErr.Error
	}
	return nil
}

func (m *MachineActionExecutorAdapter) PatchFanController(manager interface{}, fanControllerID string, patch *extendprovider.PatchFanControllerType) error {
	respErr := PatchFanController(manager, fanControllerID, patch)
	if respErr != nil {
		return respErr.Error
	}
	return nil
}

func (m *MachineActionExecutorAdapter) PatchFanZone(manager interface{}, fanZoneID string, patch *extendprovider.PatchFanZoneType) error {
	respErr := PatchFanZone(manager, fanZoneID, patch)
	if respErr != nil {
		return respErr.Error
	}
	return nil
}

func (m *MachineActionExecutorAdapter) PatchPidController(manager interface{}, pidControllerID string, patch *extendprovider.PatchPidControllerType) error {
	respErr := PatchPidController(manager, pidControllerID, patch)
	if respErr != nil {
		return respErr.Error
	}
	return nil
}

// ========== /MultiFish/v1/Platform/:machineId/Managers ==========

// GET /MultiFish/v1/Platform/:machineId/Managers - Get managers collection 
func GetManagerCollectionResponse(v interface{}, machineID string) (gin.H, error) {
	return managerProviders.GetManagerCollectionResponse(v, machineID)
}
func getManagers(c *gin.Context) {
	machineID := c.Param("machineId")
	machine, err := platformMgr.GetMachine(machineID)
	if err != nil {
		utility.RedfishError(c, http.StatusNotFound, err.Error(), "ResourceNotFound")
		return
	}

	getManagersCallback(c, machine, "", func(targetManagers interface{}) {
		var response interface{}
		if response, err = GetManagerCollectionResponse(targetManagers, machineID); err != nil {
			utility.RedfishError(c, http.StatusInternalServerError, err.Error(), "InternalError")
			return
		}
		c.JSON(http.StatusOK, response)
	})
}

// ========== /MultiFish/v1/Platform/:machineId/Managers/:managerId ==========

// GET /MultiFish/v1/Platform/:machineId/Managers/:managerId - Get manager details
func GetManagerResponse(v interface{}, machineID string, managerID string) (gin.H, *utility.ResponseError) {
	return managerProviders.GetManagerResponse(v, machineID, managerID)
}
func getManager(c *gin.Context) {
	machineID := c.Param("machineId")
	managerID := c.Param("managerId")
	machine, err := platformMgr.GetMachine(machineID)
	if err != nil {
		utility.RedfishError(c, http.StatusNotFound, err.Error(), "ResourceNotFound")
		return
	}

	getManagersCallback(c, machine, managerID, func(targetManager interface{}) {
		response, respErr := GetManagerResponse(targetManager, machineID, managerID)
		if respErr != nil {
			utility.RedfishError(c, respErr.StatusCode, respErr.Error.Error(), respErr.Message)
			return
		}
		c.JSON(http.StatusOK, response)
	})
}

// PATCH /MultiFish/v1/Platform/:machineId/Managers/:managerId - Update manager (ServiceIdentification)
func PatchManager(v interface{}, patch interface{}) (*utility.ResponseError) {
	return managerProviders.PatchManager(v, patch)
}
func patchManager(c *gin.Context) {
	machineID := c.Param("machineId")
	managerID := c.Param("managerId")
	machine, err := platformMgr.GetMachine(machineID)
	if err != nil {
		utility.RedfishError(c, http.StatusNotFound, err.Error(), "ResourceNotFound")
		return
	}

	// Validate and bind the patch data to struct
	var updates redfishprovider.PatchManagerType
	if !utility.CheckAndBindPatchPayload(c, redfishprovider.ManagerAllowedPatchFields(), &updates) {
		return
	}

	getManagersCallback(c, machine, managerID, func(targetManager interface{}) {

		if respErr := managerProviders.PatchManager(targetManager, &updates); respErr != nil {
			utility.RedfishError(c, respErr.StatusCode, respErr.Error.Error(), respErr.Message)
		}

		c.JSON(http.StatusOK, gin.H{
			"@odata.id": fmt.Sprintf("/MultiFish/v1/Platform/%s/Managers/%s", machineID, managerID),
			"Message":   "Manager updated successfully",
		})
	})
}

// ========== /MultiFish/v1/Platform/:machineId/Managers/:managerId/Oem/OpenBmc/Profile ==========

// GET /MultiFish/v1/Platform/:machineId/Managers/:managerId/Oem/OpenBmc/Profile
func getProfile(c *gin.Context) {
	machineID := c.Param("machineId")
	managerID := c.Param("managerId")
	machine, err := platformMgr.GetMachine(machineID)
	if err != nil {
		utility.RedfishError(c, http.StatusNotFound, err.Error(), "ResourceNotFound")
		return
	}

	getManagersCallback(c, machine, managerID, func(targetManager interface{}) {
		var response interface{}
		var respErr *utility.ResponseError
		response, respErr = managerProviders.GetProfileResponse(targetManager, machineID, managerID)
		if respErr != nil {
			utility.RedfishError(c, respErr.StatusCode, respErr.Error.Error(), respErr.Message)
			return
		}
		c.JSON(http.StatusOK, response)
	})
}

// PATCH /MultiFish/v1/Platform/:machineId/Managers/:managerId/Oem/OpenBmc/Profile
func PatchProfile(v interface{}, patch extendprovider.PatchProfileType) (*utility.ResponseError) {
	return managerProviders.PatchProfile(v, patch)
}
func patchProfile(c *gin.Context) {
	machineID := c.Param("machineId")
	managerID := c.Param("managerId")
	machine, err := platformMgr.GetMachine(machineID)
	if err != nil {
		utility.RedfishError(c, http.StatusNotFound, err.Error(), "ResourceNotFound")
		return
	}

	// Validate and bind the patch data to struct
	var updates extendprovider.PatchProfileType
	if !utility.CheckAndBindPatchPayload(c, extendprovider.ProfileAllowedPatchFields(), &updates) {
		return
	}

	getManagersCallback(c, machine, managerID, func(targetManager interface{}) {

		if respErr := managerProviders.PatchProfile(targetManager, updates); respErr != nil {
			utility.RedfishError(c, respErr.StatusCode, fmt.Sprintf("Failed to patch profile: %v", respErr.Error), respErr.Message)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"@odata.id": fmt.Sprintf("/MultiFish/v1/Platform/%s/Managers/%s/Oem/OpenBmc/Profile", machineID, managerID),
			"Message":   "Profile updated successfully",
		})
	})
}

// ========== /MultiFish/v1/Platform/:machineId/Managers/:managerId/Oem/OpenBmc/FanController ==========

// GET /MultiFish/v1/Platform/:machineId/Managers/:managerId/Oem/OpenBmc/FanController
func GetFanControllerCollectionResponse(v interface{}, machineID string, managerID string) (gin.H, *utility.ResponseError) {
	return managerProviders.GetFanControllerCollectionResponse(v, machineID, managerID)
}
func getFanControllers(c *gin.Context) {
	machineID := c.Param("machineId")
	managerID := c.Param("managerId")
	machine, err := platformMgr.GetMachine(machineID)
	if err != nil {
		utility.RedfishError(c, http.StatusNotFound, err.Error(), "ResourceNotFound")
		return
	}

	getManagersCallback(c, machine, managerID, func(targetManager interface{}) {
		response, respErr := GetFanControllerCollectionResponse(targetManager, machineID, managerID)
		if respErr != nil {
			utility.RedfishError(c, respErr.StatusCode, respErr.Error.Error(), respErr.Message)
			return
		}
		c.JSON(http.StatusOK, response)
	})
}

// ========== /MultiFish/v1/Platform/:machineId/Managers/:managerId/Oem/OpenBmc/FanController/:fanControllerId ==========

// GET /MultiFish/v1/Platform/:machineId/Managers/:managerId/Oem/OpenBmc/FanController/:fanControllerId
func GetFanControllerResponse(v interface{}, machineID string, managerID string, fanID string) (gin.H, *utility.ResponseError) {
	return managerProviders.GetFanControllerResponse(v, machineID, managerID, fanID)
}
func getFanController(c *gin.Context) {
	machineID := c.Param("machineId")
	managerID := c.Param("managerId")	
	fanControllerID := c.Param("fanControllerId")
	machine, err := platformMgr.GetMachine(machineID)
	if err != nil {
		utility.RedfishError(c, http.StatusNotFound, err.Error(), "ResourceNotFound")
		return
	}

	getManagersCallback(c, machine, managerID, func(targetManager interface{}) {
		response, respErr := GetFanControllerResponse(targetManager, machineID, managerID, fanControllerID)
		if respErr != nil {
			utility.RedfishError(c, respErr.StatusCode, respErr.Error.Error(), respErr.Message)
			return
		}
		c.JSON(http.StatusOK, response)
	})
}

// PATCH /MultiFish/v1/Platform/:machineId/Managers/:managerId/Oem/OpenBmc/FanController/:fanControllerId
func PatchFanController(v interface{}, fanControllerID string, fcPatch *extendprovider.PatchFanControllerType) (*utility.ResponseError) {
	return managerProviders.PatchFanController(v, fanControllerID, fcPatch)
}
func patchFanController(c *gin.Context) {
	machineID := c.Param("machineId")
	managerID := c.Param("managerId")
	fanControllerID := c.Param("fanControllerId")
	machine, err := platformMgr.GetMachine(machineID)
	if err != nil {
		utility.RedfishError(c, http.StatusNotFound, err.Error(), "ResourceNotFound")
		return
	}

	// Validate and bind the patch data to struct
	var fcPatch extendprovider.PatchFanControllerType
	if !utility.CheckAndBindPatchPayload(c, extendprovider.FanControllerAllowedPatchFields(), &fcPatch) {
		return
	}

	getManagersCallback(c, machine, managerID, func(targetManager interface{}) {

		if respErr := PatchFanController(targetManager, fanControllerID, &fcPatch); respErr != nil {
			utility.RedfishError(c, respErr.StatusCode, respErr.Error.Error(), respErr.Message)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"@odata.id": fmt.Sprintf("/MultiFish/v1/Platform/%s/Managers/%s/Oem/OpenBmc/FanController/%s", machineID, managerID, fanControllerID),
			"Message":   "FanController updated successfully",
		})
	})
}

// ========== /MultiFish/v1/Platform/:machineId/Managers/:managerId/Oem/OpenBmc/FanZone ==========

// GET /MultiFish/v1/Platform/:machineId/Managers/:managerId/Oem/OpenBmc/FanZone
func GetFanZoneCollectionResponse(v interface{}, machineID string, managerID string) (gin.H, *utility.ResponseError) {
	return managerProviders.GetFanZoneCollectionResponse(v, machineID, managerID)
}
func getFanZones(c *gin.Context) {
	machineID := c.Param("machineId")
	managerID := c.Param("managerId")
	machine, err := platformMgr.GetMachine(machineID)
	if err != nil {
		utility.RedfishError(c, http.StatusNotFound, err.Error(), "ResourceNotFound")
		return
	}

	getManagersCallback(c, machine, managerID, func(targetManager interface{}) {
		response, respErr := GetFanZoneCollectionResponse(targetManager, machineID, managerID)
		if respErr != nil {
			utility.RedfishError(c, respErr.StatusCode, respErr.Error.Error(), respErr.Message)
			return
		}
		c.JSON(http.StatusOK, response)
	})
}

// ========== /MultiFish/v1/Platform/:machineId/Managers/:managerId/Oem/OpenBmc/FanZone/:fanZoneId ==========

// GET /MultiFish/v1/Platform/:machineId/Managers/:managerId/Oem/OpenBmc/FanZone/:fanZoneId
func GetFanZoneResponse(v interface{}, machineID string, managerID string, zoneID string) (gin.H, *utility.ResponseError) {
	return managerProviders.GetFanZoneResponse(v, machineID, managerID, zoneID)
}
func getFanZone(c *gin.Context) {
	machineID := c.Param("machineId")
	managerID := c.Param("managerId")	
	fanZoneID := c.Param("fanZoneId")
	machine, err := platformMgr.GetMachine(machineID)
	if err != nil {
		utility.RedfishError(c, http.StatusNotFound, err.Error(), "ResourceNotFound")
		return
	}

	getManagersCallback(c, machine, managerID, func(targetManager interface{}) {
		response, respErr := GetFanZoneResponse(targetManager, machineID, managerID, fanZoneID)
		if respErr != nil {
			utility.RedfishError(c, respErr.StatusCode, respErr.Error.Error(), respErr.Message)
			return
		}
		c.JSON(http.StatusOK, response)
	})
}

// PATCH /MultiFish/v1/Platform/:machineId/Managers/:managerId/Oem/OpenBmc/FanZone/:fanZoneId
func PatchFanZone(v interface{}, fanZoneID string, fzPatch *extendprovider.PatchFanZoneType) (*utility.ResponseError) {
	return managerProviders.PatchFanZone(v, fanZoneID, fzPatch)
}
func patchFanZone(c *gin.Context) {
	machineID := c.Param("machineId")
	managerID := c.Param("managerId")
	fanZoneID := c.Param("fanZoneId")
	machine, err := platformMgr.GetMachine(machineID)
	if err != nil {
		utility.RedfishError(c, http.StatusNotFound, err.Error(), "ResourceNotFound")
		return
	}

	// Validate and bind the patch data to struct
	var fzPatch extendprovider.PatchFanZoneType
	if !utility.CheckAndBindPatchPayload(c, extendprovider.FanZoneAllowedPatchFields(), &fzPatch) {
		return
	}

	getManagersCallback(c, machine, managerID, func(targetManager interface{}) {

		if respErr := PatchFanZone(targetManager, fanZoneID, &fzPatch); respErr != nil {
			utility.RedfishError(c, respErr.StatusCode, respErr.Error.Error(), respErr.Message)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"@odata.id": fmt.Sprintf("/MultiFish/v1/Platform/%s/Managers/%s/Oem/OpenBmc/FanZone/%s", machineID, managerID, fanZoneID),
			"Message":   "FanZone updated successfully",
		})
	})
}

// ========== /MultiFish/v1/Platform/:machineId/Managers/:managerId/Oem/OpenBmc/PidController ==========

// GET /MultiFish/v1/Platform/:machineId/Managers/:managerId/Oem/OpenBmc/PidController
func GetPidControllerCollectionResponse(v interface{}, machineID string, managerID string) (gin.H, *utility.ResponseError) {
	return managerProviders.GetPidControllerCollectionResponse(v, machineID, managerID)
}
func getPidControllers(c *gin.Context) {
	machineID := c.Param("machineId")
	managerID := c.Param("managerId")
	machine, err := platformMgr.GetMachine(machineID)
	if err != nil {
		utility.RedfishError(c, http.StatusNotFound, err.Error(), "ResourceNotFound")
		return
	}

	getManagersCallback(c, machine, managerID, func(targetManager interface{}) {
		response, respErr := GetPidControllerCollectionResponse(targetManager, machineID, managerID)
		if respErr != nil {
			utility.RedfishError(c, respErr.StatusCode, respErr.Error.Error(), respErr.Message)
			return
		}
		c.JSON(http.StatusOK, response)
	})
}

// ========== /MultiFish/v1/Platform/:machineId/Managers/:managerId/Oem/OpenBmc/PidController/:pidControllerId ==========

// GET /MultiFish/v1/Platform/:machineId/Managers/:managerId/Oem/OpenBmc/PidController/:pidControllerId
func GetPidControllerResponse(v interface{}, machineID string, managerID string, pidID string) (gin.H, *utility.ResponseError) {
	return managerProviders.GetPidControllerResponse(v, machineID, managerID, pidID)
}
func getPidController(c *gin.Context) {
	machineID := c.Param("machineId")
	managerID := c.Param("managerId")	
	pidControllerID := c.Param("pidControllerId")
	machine, err := platformMgr.GetMachine(machineID)
	if err != nil {
		utility.RedfishError(c, http.StatusNotFound, err.Error(), "ResourceNotFound")
		return
	}

	getManagersCallback(c, machine, managerID, func(targetManager interface{}) {
		response, respErr := GetPidControllerResponse(targetManager, machineID, managerID, pidControllerID)
		if respErr != nil {
			utility.RedfishError(c, respErr.StatusCode, respErr.Error.Error(), respErr.Message)
			return
		}
		c.JSON(http.StatusOK, response)
	})
}

// PATCH /MultiFish/v1/Platform/:machineId/Managers/:managerId/Oem/OpenBmc/PidController/:pidControllerId
func PatchPidController(v interface{}, pidControllerID string, pcPatch *extendprovider.PatchPidControllerType) (*utility.ResponseError) {
	return managerProviders.PatchPidController(v, pidControllerID, pcPatch)
}
func patchPidController(c *gin.Context) {
	machineID := c.Param("machineId")
	managerID := c.Param("managerId")
	pidControllerID := c.Param("pidControllerId")
	machine, err := platformMgr.GetMachine(machineID)
	if err != nil {
		utility.RedfishError(c, http.StatusNotFound, err.Error(), "ResourceNotFound")
		return
	}

	// Validate and bind the patch data to struct
	var pcPatch extendprovider.PatchPidControllerType
	if !utility.CheckAndBindPatchPayload(c, extendprovider.PidControllerAllowedPatchFields(), &pcPatch) {
		return
	}

	getManagersCallback(c, machine, managerID, func(targetManager interface{}) {

		if respErr := PatchPidController(targetManager, pidControllerID, &pcPatch); respErr != nil {
			utility.RedfishError(c, respErr.StatusCode, respErr.Error.Error(), respErr.Message)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"@odata.id": fmt.Sprintf("/MultiFish/v1/Platform/%s/Managers/%s/Oem/OpenBmc/PidController/%s", machineID, managerID, pidControllerID),
			"Message":   "PidController updated successfully",
		})
	})
}

// ========== Route Setup ==========

// managerRoutes sets up the manager-related routes
func managerRoutes(router *gin.Engine) {
	// Manager routes
	router.GET("/MultiFish/v1/Platform/:machineId/Managers", getManagers)
	router.GET("/MultiFish/v1/Platform/:machineId/Managers/:managerId", getManager)
	router.PATCH("/MultiFish/v1/Platform/:machineId/Managers/:managerId", patchManager)

	// Profile routes
	router.GET("/MultiFish/v1/Platform/:machineId/Managers/:managerId/Oem/OpenBmc/Profile", getProfile)
	router.PATCH("/MultiFish/v1/Platform/:machineId/Managers/:managerId/Oem/OpenBmc/Profile", patchProfile)

	// FanController routes
	router.GET("/MultiFish/v1/Platform/:machineId/Managers/:managerId/Oem/OpenBmc/FanController", getFanControllers)
	router.GET("/MultiFish/v1/Platform/:machineId/Managers/:managerId/Oem/OpenBmc/FanController/:fanControllerId", getFanController)
	router.PATCH("/MultiFish/v1/Platform/:machineId/Managers/:managerId/Oem/OpenBmc/FanController/:fanControllerId", patchFanController)

	// FanZone routes
	router.GET("/MultiFish/v1/Platform/:machineId/Managers/:managerId/Oem/OpenBmc/FanZone", getFanZones)
	router.GET("/MultiFish/v1/Platform/:machineId/Managers/:managerId/Oem/OpenBmc/FanZone/:fanZoneId", getFanZone)
	router.PATCH("/MultiFish/v1/Platform/:machineId/Managers/:managerId/Oem/OpenBmc/FanZone/:fanZoneId", patchFanZone)

	// PidController routes
	router.GET("/MultiFish/v1/Platform/:machineId/Managers/:managerId/Oem/OpenBmc/PidController", getPidControllers)
	router.GET("/MultiFish/v1/Platform/:machineId/Managers/:managerId/Oem/OpenBmc/PidController/:pidControllerId", getPidController)
	router.PATCH("/MultiFish/v1/Platform/:machineId/Managers/:managerId/Oem/OpenBmc/PidController/:pidControllerId", patchPidController)
}
