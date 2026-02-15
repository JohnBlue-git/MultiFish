package utility

// MaskPassword returns a masked version of the password
func MaskPassword(password string) string {
	if password == "" {
		return ""
	}
	return "******"
}

// MaskMachinePassword creates a copy of MachineConfig with masked password
// Note: This function accepts interface{} to avoid circular dependencies
// The caller should pass a struct with a Password field
func MaskSensitiveData(data interface{}) interface{} {
	// This is a marker function to document the security pattern
	// Actual masking should be done at the caller level
	return data
}
