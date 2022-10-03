package utils

import (
	"github.com/application-stacks/runtime-component-operator/common"
)

// InternalOpConfig stored operator configuration
type InternalOpConfig map[string]string

const (
	// InternalOpConfigLicenseServiceEndpointURI endpoint URI for default IBM License Service instance
	InternalOpConfigLicenseServiceEndpointURI = "licenseServiceEndpointURI"

	// InternalOpConfigLicenseServiceEndpointScope endpoint scope for default IBM License Service instance
	InternalOpConfigLicenseServiceEndpointScope = "licenseServiceEndpointScope"
)

// InternalConfig stores operator configuration
var InternalConfig = InternalOpConfig{}

// DefaultInternalOpConfig returns default configuration
func DefaultInternalOpConfig() InternalOpConfig {
	cfg := InternalOpConfig{}
	cfg[InternalOpConfigLicenseServiceEndpointURI] = ""
	cfg[InternalOpConfigLicenseServiceEndpointScope] = string(common.StatusEndpointScopeInternal)
	return cfg
}
