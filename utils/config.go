package utils

import (
	"github.com/application-stacks/runtime-component-operator/common"
	corev1 "k8s.io/api/core/v1"
)

const (
	// OpConfigLicenseServiceRegistry default registry for IBM License Service install
	OpConfigLicenseServiceRegistry = "licenseServiceRegistry"

	// OpConfigLicenseServiceRegistryNamespace default registry namespace for IBM License Service install
	OpConfigLicenseServiceRegistryNamespace = "licenseServiceRegistryNamespace"
)

// LoadFromWebSphereLibertyConfigMap creates a config out of kubernetes config map
func LoadFromWebSphereLibertyConfigMap(oc *common.OpConfig, cm *corev1.ConfigMap) {
	for k, v := range DefaultOpConfig() {
		(*oc)[k] = v
	}

	for k, v := range cm.Data {
		(*oc)[k] = v
	}
}

// DefaultOpConfig returns default configuration
func DefaultOpConfig() common.OpConfig {
	cfg := common.DefaultOpConfig()
	cfg[OpConfigLicenseServiceRegistry] = "common-service"
	cfg[OpConfigLicenseServiceRegistryNamespace] = "ibm-common-services"
	return cfg
}
