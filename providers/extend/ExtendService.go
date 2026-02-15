package extendprovider

import (
	"github.com/stmcginnis/gofish"
)

// ExtendService extends the gofish.Service with custom methods.
type ExtendService struct {
	*gofish.Service
}

// NewExtendService creates a new ExtendService from a gofish.APIClient.
func NewExtendService(client *gofish.APIClient) *ExtendService {
	return &ExtendService{
		Service: client.Service,
	}
}
