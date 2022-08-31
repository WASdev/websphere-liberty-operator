/*
  Copyright contributors to the WASdev project.

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
*/

package v1

import (
	"time"

	"github.com/application-stacks/runtime-component-operator/common"
	routev1 "github.com/openshift/api/route/v1"
	prometheusv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Defines the desired state of WebSphereLibertyApplication.
type WebSphereLibertyApplicationSpec struct {

	// +operator-sdk:csv:customresourcedefinitions:order=1,type=spec,displayName="License",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	License License `json:"license"`

	// Application image to deploy.
	// +operator-sdk:csv:customresourcedefinitions:order=1,type=spec,displayName="Application Image",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	ApplicationImage string `json:"applicationImage"`

	// Name of the application. Defaults to the name of this custom resource.
	// +operator-sdk:csv:customresourcedefinitions:order=2,type=spec,displayName="Application Name",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	ApplicationName string `json:"applicationName,omitempty"`

	// Version of the application.
	// +operator-sdk:csv:customresourcedefinitions:order=3,type=spec,displayName="Application Version",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	ApplicationVersion string `json:"applicationVersion,omitempty"`

	// Policy for pulling container images. Defaults to IfNotPresent.
	// +operator-sdk:csv:customresourcedefinitions:order=4,type=spec,displayName="Pull Policy",xDescriptors="urn:alm:descriptor:com.tectonic.ui:imagePullPolicy"
	PullPolicy *corev1.PullPolicy `json:"pullPolicy,omitempty"`

	// Name of the Secret to use to pull images from the specified repository. It is not required if the cluster is configured with a global image pull secret.
	// +operator-sdk:csv:customresourcedefinitions:order=5,type=spec,displayName="Pull Secret",xDescriptors="urn:alm:descriptor:io.kubernetes:Secret"
	PullSecret *string `json:"pullSecret,omitempty"`

	// Name of the service account to use for deploying the application. A service account is automatically created if it's not specified.
	// +operator-sdk:csv:customresourcedefinitions:order=6,type=spec,displayName="Service Account Name",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	ServiceAccountName *string `json:"serviceAccountName,omitempty"`

	// Create Knative resources and use Knative serving.
	// +operator-sdk:csv:customresourcedefinitions:order=7,type=spec,displayName="Create Knative Service",xDescriptors="urn:alm:descriptor:com.tectonic.ui:booleanSwitch"
	CreateKnativeService *bool `json:"createKnativeService,omitempty"`

	// Expose the application externally via a Route, a Knative Route or an Ingress resource.
	// +operator-sdk:csv:customresourcedefinitions:order=8,type=spec,displayName="Expose",xDescriptors="urn:alm:descriptor:com.tectonic.ui:booleanSwitch"
	Expose *bool `json:"expose,omitempty"`

	// Enable management of TLS certificates. Defaults to true.
	// +operator-sdk:csv:customresourcedefinitions:order=8,type=spec,displayName="Manage TLS",xDescriptors="urn:alm:descriptor:com.tectonic.ui:booleanSwitch"
	ManageTLS *bool `json:"manageTLS,omitempty"`

	// Number of pods to create. Defaults to 1. Not applicable when .spec.autoscaling or .spec.createKnativeService is specified.
	// +operator-sdk:csv:customresourcedefinitions:order=9,type=spec,displayName="Replicas",xDescriptors="urn:alm:descriptor:com.tectonic.ui:podCount"
	Replicas *int32 `json:"replicas,omitempty"`

	// +operator-sdk:csv:customresourcedefinitions:order=10,type=spec,displayName="Auto Scaling"
	Autoscaling *WebSphereLibertyApplicationAutoScaling `json:"autoscaling,omitempty"`

	// Resource requests and limits for the application container.
	// +operator-sdk:csv:customresourcedefinitions:order=11,type=spec,displayName="Resource Requirements",xDescriptors="urn:alm:descriptor:com.tectonic.ui:resourceRequirements"
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// +operator-sdk:csv:customresourcedefinitions:order=12,type=spec,displayName="Probes"
	Probes *WebSphereLibertyApplicationProbes `json:"probes,omitempty"`

	// +operator-sdk:csv:customresourcedefinitions:order=13,type=spec,displayName="Deployment"
	Deployment *WebSphereLibertyApplicationDeployment `json:"deployment,omitempty"`

	// +operator-sdk:csv:customresourcedefinitions:order=14,type=spec,displayName="StatefulSet"
	StatefulSet *WebSphereLibertyApplicationStatefulSet `json:"statefulSet,omitempty"`

	// +operator-sdk:csv:customresourcedefinitions:order=15,type=spec,displayName="Service"
	Service *WebSphereLibertyApplicationService `json:"service,omitempty"`

	// +operator-sdk:csv:customresourcedefinitions:order=16,type=spec,displayName="Route"
	Route *WebSphereLibertyApplicationRoute `json:"route,omitempty"`

	// +operator-sdk:csv:customresourcedefinitions:order=17,type=spec,displayName="Network Policy"
	NetworkPolicy *WebSphereLibertyApplicationNetworkPolicy `json:"networkPolicy,omitempty"`

	// +operator-sdk:csv:customresourcedefinitions:order=18,type=spec,displayName="Serviceability"
	Serviceability *WebSphereLibertyApplicationServiceability `json:"serviceability,omitempty"`

	// +operator-sdk:csv:customresourcedefinitions:order=19,type=spec,displayName="Single Sign-On"
	SSO *WebSphereLibertyApplicationSSO `json:"sso,omitempty"`

	// +operator-sdk:csv:customresourcedefinitions:order=20,type=spec,displayName="Monitoring"
	Monitoring *WebSphereLibertyApplicationMonitoring `json:"monitoring,omitempty"`

	// An array of environment variables for the application container.
	// +listType=map
	// +listMapKey=name
	// +operator-sdk:csv:customresourcedefinitions:order=21,type=spec,displayName="Environment Variables"
	Env []corev1.EnvVar `json:"env,omitempty"`

	// List of sources to populate environment variables in the application container.
	// +listType=atomic
	// +operator-sdk:csv:customresourcedefinitions:order=22,type=spec,displayName="Environment Variables from Sources"
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty"`

	// Represents a volume with data that is accessible to the application container.
	// +listType=map
	// +listMapKey=name
	// +operator-sdk:csv:customresourcedefinitions:order=23,type=spec,displayName="Volumes"
	Volumes []corev1.Volume `json:"volumes,omitempty"`

	// Represents where to mount the volumes into the application container.
	// +listType=atomic
	// +operator-sdk:csv:customresourcedefinitions:order=24,type=spec,displayName="Volume Mounts"
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`

	// List of containers to run before other containers in a pod.
	// +listType=map
	// +listMapKey=name
	// +operator-sdk:csv:customresourcedefinitions:order=25,type=spec,displayName="Init Containers"
	InitContainers []corev1.Container `json:"initContainers,omitempty"`

	// List of sidecar containers. These are additional containers to be added to the pods.
	// +listType=map
	// +listMapKey=name
	// +operator-sdk:csv:customresourcedefinitions:order=26,type=spec,displayName="Sidecar Containers"
	SidecarContainers []corev1.Container `json:"sidecarContainers,omitempty"`

	// +operator-sdk:csv:customresourcedefinitions:order=27,type=spec,displayName="Affinity"
	Affinity *WebSphereLibertyApplicationAffinity `json:"affinity,omitempty"`

	// Security context for the application container.
	// +operator-sdk:csv:customresourcedefinitions:order=28,type=spec,displayName="Security Context"
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`
}

// License information is required.
type License struct {
	// Product edition. Defaults to IBM WebSphere Application Server. Other options: IBM WebSphere Application Server Liberty Core, IBM WebSphere Application Server Network Deployment
	// +operator-sdk:csv:customresourcedefinitions:order=100,type=spec,displayName="Edition"
	Edition LicenseEdition `json:"edition,omitempty"`

	// Entitlement source for the product. Defaults to Standalone. Other options: IBM Cloud Pak for Applications, IBM WebSphere Application Server Family Edition, IBM WebSphere Hybrid Edition
	// +operator-sdk:csv:customresourcedefinitions:order=101,type=spec,displayName="Product Entitlement Source"
	ProductEntitlementSource LicenseEntitlement `json:"productEntitlementSource,omitempty"`

	// Charge metric code. Defaults to Virtual Processor Core (VPC). Other option: Processor Value Unit (PVU)
	// +operator-sdk:csv:customresourcedefinitions:order=102,type=spec,displayName="Metric"
	Metric LicenseMetric `json:"metric,omitempty"`

	// I represent that the software in the above-referenced application container includes the IBM Program referenced below and I accept the terms of the license agreement corresponding
	// to the version of IBM Program in the application container by setting this value to true. See https://ibm.biz/was-license for the license agreements applicable to this IBM Program
	// +operator-sdk:csv:customresourcedefinitions:order=103,type=spec,displayName="Accept License",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:checkbox"}
	// +kubebuilder:validation:Enum:=true
	Accept bool `json:"accept"`
}

// Defines the possible values for charge metric codes
// +kubebuilder:validation:Enum=Virtual Processor Core (VPC);Processor Value Unit (PVU)
type LicenseMetric string

const (
	// Metric Virtual Processor Core (VPC)
	LicenseMetricVPC LicenseMetric = "Virtual Processor Core (VPC)"
	// Metric Processor Value Unit (PVU)
	LicenseMetricPVU LicenseMetric = "Processor Value Unit (PVU)"
)

// Defines the possible values for editions
// +kubebuilder:validation:Enum=IBM WebSphere Application Server;IBM WebSphere Application Server Liberty Core;IBM WebSphere Application Server Network Deployment
type LicenseEdition string

const (
	// Edition IBM WebSphere Application Server
	LicenseEditionBase LicenseEdition = "IBM WebSphere Application Server"
	// Edition IBM WebSphere Application Server Liberty Core
	LicenseEditionCore LicenseEdition = "IBM WebSphere Application Server Liberty Core"
	// Edition IBM WebSphere Application Server Network Deployment
	LicenseEditionND LicenseEdition = "IBM WebSphere Application Server Network Deployment"
)

// Defines the possible values for product entitlement source
// +kubebuilder:validation:Enum=Standalone;IBM Cloud Pak for Applications;IBM WebSphere Application Server Family Edition;IBM WebSphere Hybrid Edition
type LicenseEntitlement string

const (
	// Entitlement source Standalone
	LicenseEntitlementStandalone LicenseEntitlement = "Standalone"
	// Entitlement source IBM Cloud Pak for Applications
	LicenseEntitlementCP4Apps LicenseEntitlement = "IBM Cloud Pak for Applications"
	// Entitlement source IBM WebSphere Application Server Family Edition
	LicenseEntitlementFamilyEdition LicenseEntitlement = "IBM WebSphere Application Server Family Edition"
	// Entitlement source IBM WebSphere Hybrid Edition
	LicenseEntitlementWSHE LicenseEntitlement = "IBM WebSphere Hybrid Edition"
)

// Define health checks on application container to determine whether it is alive or ready to receive traffic
type WebSphereLibertyApplicationProbes struct {
	// Periodic probe of container liveness. Container will be restarted if the probe fails.
	// +operator-sdk:csv:customresourcedefinitions:order=49,type=spec,displayName="Liveness Probe"
	Liveness *corev1.Probe `json:"liveness,omitempty"`

	// Periodic probe of container service readiness. Container will be removed from service endpoints if the probe fails.
	// +operator-sdk:csv:customresourcedefinitions:order=50,type=spec,displayName="Readiness Probe"
	Readiness *corev1.Probe `json:"readiness,omitempty"`

	// Probe to determine successful initialization. If specified, other probes are not executed until this completes successfully.
	// +operator-sdk:csv:customresourcedefinitions:order=51,type=spec,displayName="Startup Probe"
	Startup *corev1.Probe `json:"startup,omitempty"`
}

// Configure pods to run on particular Nodes.
type WebSphereLibertyApplicationAffinity struct {
	// Controls which nodes the pod are scheduled to run on, based on labels on the node.
	// +operator-sdk:csv:customresourcedefinitions:order=37,type=spec,displayName="Node Affinity",xDescriptors="urn:alm:descriptor:com.tectonic.ui:nodeAffinity"
	NodeAffinity *corev1.NodeAffinity `json:"nodeAffinity,omitempty"`

	// Controls the nodes the pod are scheduled to run on, based on labels on the pods that are already running on the node.
	// +operator-sdk:csv:customresourcedefinitions:order=38,type=spec,displayName="Pod Affinity",xDescriptors="urn:alm:descriptor:com.tectonic.ui:podAffinity"
	PodAffinity *corev1.PodAffinity `json:"podAffinity,omitempty"`

	// Enables the ability to prevent running a pod on the same node as another pod.
	// +operator-sdk:csv:customresourcedefinitions:order=39,type=spec,displayName="Pod Anti Affinity",xDescriptors="urn:alm:descriptor:com.tectonic.ui:podAntiAffinity"
	PodAntiAffinity *corev1.PodAntiAffinity `json:"podAntiAffinity,omitempty"`

	// A YAML object that contains a set of required labels and their values.
	// +operator-sdk:csv:customresourcedefinitions:order=40,type=spec,displayName="Node Affinity Labels",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	NodeAffinityLabels map[string]string `json:"nodeAffinityLabels,omitempty"`

	// An array of architectures to be considered for deployment. Their position in the array indicates preference.
	// +listType=set
	Architecture []string `json:"architecture,omitempty"`
}

// Configures the desired resource consumption of pods.
type WebSphereLibertyApplicationAutoScaling struct {
	// Required field for autoscaling. Upper limit for the number of pods that can be set by the autoscaler. Parameter .spec.resources.requests.cpu must also be specified.
	// +kubebuilder:validation:Minimum=1
	// +operator-sdk:csv:customresourcedefinitions:order=1,type=spec,displayName="Max Replicas",xDescriptors="urn:alm:descriptor:com.tectonic.ui:number"
	MaxReplicas int32 `json:"maxReplicas,omitempty"`

	// Lower limit for the number of pods that can be set by the autoscaler.
	// +operator-sdk:csv:customresourcedefinitions:order=2,type=spec,displayName="Min Replicas",xDescriptors="urn:alm:descriptor:com.tectonic.ui:number"
	MinReplicas *int32 `json:"minReplicas,omitempty"`

	// Target average CPU utilization, represented as a percentage of requested CPU, over all the pods.
	// +operator-sdk:csv:customresourcedefinitions:order=3,type=spec,displayName="Target CPU Utilization Percentage",xDescriptors="urn:alm:descriptor:com.tectonic.ui:number"
	TargetCPUUtilizationPercentage *int32 `json:"targetCPUUtilizationPercentage,omitempty"`
}

// Configures parameters for the network service of pods.
type WebSphereLibertyApplicationService struct {
	// The port exposed by the container.
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:validation:Minimum=1
	// +operator-sdk:csv:customresourcedefinitions:order=9,type=spec,displayName="Service Port",xDescriptors="urn:alm:descriptor:com.tectonic.ui:number"
	Port int32 `json:"port,omitempty"`

	// +operator-sdk:csv:customresourcedefinitions:order=10,type=spec,displayName="Service Type",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Type *corev1.ServiceType `json:"type,omitempty"`

	// Node proxies this port into your service.
	// +kubebuilder:validation:Maximum=32767
	// +kubebuilder:validation:Minimum=30000
	// +operator-sdk:csv:customresourcedefinitions:order=11,type=spec,displayName="Node Port",xDescriptors="urn:alm:descriptor:com.tectonic.ui:number"
	NodePort *int32 `json:"nodePort,omitempty"`

	// The name for the port exposed by the container.
	// +operator-sdk:csv:customresourcedefinitions:order=12,type=spec,displayName="Port Name",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	PortName string `json:"portName,omitempty"`

	// Annotations to be added to the service.
	// +operator-sdk:csv:customresourcedefinitions:order=13,type=spec,displayName="Service Annotations",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Annotations map[string]string `json:"annotations,omitempty"`

	// The port that the operator assigns to containers inside pods. Defaults to the value of .spec.service.port.
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:validation:Minimum=1
	// +operator-sdk:csv:customresourcedefinitions:order=14,type=spec,displayName="Target Port",xDescriptors="urn:alm:descriptor:com.tectonic.ui:number"
	TargetPort *int32 `json:"targetPort,omitempty"`

	// A name of a secret that already contains TLS key, certificate and CA to be mounted in the pod. The following keys are valid in the secret: ca.crt, tls.crt, and tls.key.
	// +operator-sdk:csv:customresourcedefinitions:order=15,type=spec,displayName="Certificate Secret Reference",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	CertificateSecretRef *string `json:"certificateSecretRef,omitempty"`

	// An array consisting of service ports.
	// +operator-sdk:csv:customresourcedefinitions:order=16,type=spec
	Ports []corev1.ServicePort `json:"ports,omitempty"`

	// Expose the application as a bindable service. Defaults to false.
	// +operator-sdk:csv:customresourcedefinitions:order=17,type=spec,displayName="Bindable",xDescriptors="urn:alm:descriptor:com.tectonic.ui:booleanSwitch"
	Bindable *bool `json:"bindable,omitempty"`
}

// Defines the network policy
type WebSphereLibertyApplicationNetworkPolicy struct {
	// Disable the creation of the network policy. Defaults to false.
	// +operator-sdk:csv:customresourcedefinitions:order=52,type=spec,displayName="Disable",xDescriptors="urn:alm:descriptor:com.tectonic.ui:booleanSwitch"
	Disable *bool `json:"disable,omitempty"`

	// Specify the labels of namespaces that incoming traffic is allowed from.
	// +operator-sdk:csv:customresourcedefinitions:order=53,type=spec,displayName="Namespace Labels",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	NamespaceLabels *map[string]string `json:"namespaceLabels,omitempty"`

	// Specify the labels of pod(s) that incoming traffic is allowed from.
	// +operator-sdk:csv:customresourcedefinitions:order=54,type=spec,displayName="From Labels",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	FromLabels *map[string]string `json:"fromLabels,omitempty"`
}

// Defines the desired state and cycle of applications.
type WebSphereLibertyApplicationDeployment struct {
	// Specifies the strategy to replace old deployment pods with new pods.
	// +operator-sdk:csv:customresourcedefinitions:order=21,type=spec,displayName="Deployment Update Strategy",xDescriptors="urn:alm:descriptor:com.tectonic.ui:updateStrategy"
	UpdateStrategy *appsv1.DeploymentStrategy `json:"updateStrategy,omitempty"`

	// Annotations to be added only to the Deployment and resources owned by the Deployment
	Annotations map[string]string `json:"annotations,omitempty"`
}

// Defines the desired state and cycle of stateful applications.
type WebSphereLibertyApplicationStatefulSet struct {
	// Specifies the strategy to replace old StatefulSet pods with new pods.
	// +operator-sdk:csv:customresourcedefinitions:order=23,type=spec,displayName="StatefulSet Update Strategy"
	UpdateStrategy *appsv1.StatefulSetUpdateStrategy `json:"updateStrategy,omitempty"`

	// +operator-sdk:csv:customresourcedefinitions:order=24,type=spec,displayName="Storage"
	Storage *WebSphereLibertyApplicationStorage `json:"storage,omitempty"`

	// Annotations to be added only to the StatefulSet and resources owned by the StatefulSet.
	Annotations map[string]string `json:"annotations,omitempty"`
}

// Defines settings of persisted storage for StatefulSets.
type WebSphereLibertyApplicationStorage struct {
	// A convenient field to set the size of the persisted storage.
	// +kubebuilder:validation:Pattern=^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$
	// +operator-sdk:csv:customresourcedefinitions:order=25,type=spec,displayName="Storage Size",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Size string `json:"size,omitempty"`

	// A convenient field to request the storage class of the persisted storage. The name can not be specified or updated after the storage is created.
	// +kubebuilder:validation:Pattern=.+
	// +operator-sdk:csv:customresourcedefinitions:order=26,type=spec,displayName="Storage Class Name",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	ClassName string `json:"className,omitempty"`

	// The directory inside the container where this persisted storage will be bound to.
	// +operator-sdk:csv:customresourcedefinitions:order=27,type=spec,displayName="Storage Mount Path",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	MountPath string `json:"mountPath,omitempty"`

	// A YAML object that represents a volumeClaimTemplate component of a StatefulSet.
	// +operator-sdk:csv:customresourcedefinitions:order=28,type=spec,displayName="Storage Volume Claim Template",xDescriptors="urn:alm:descriptor:com.tectonic.ui:PersistentVolumeClaim"
	VolumeClaimTemplate *corev1.PersistentVolumeClaim `json:"volumeClaimTemplate,omitempty"`
}

// Specifies parameters for Service Monitor.
type WebSphereLibertyApplicationMonitoring struct {
	// Labels to set on ServiceMonitor.
	// +operator-sdk:csv:customresourcedefinitions:order=34,type=spec,displayName="Monitoring Labels",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Labels map[string]string `json:"labels,omitempty"`

	// A YAML snippet representing an array of Endpoint component from ServiceMonitor.
	// +listType=atomic
	// +operator-sdk:csv:customresourcedefinitions:order=35,type=spec,displayName="Monitoring Endpoints",xDescriptors="urn:alm:descriptor:com.tectonic.ui:endpointList"
	Endpoints []prometheusv1.Endpoint `json:"endpoints,omitempty"`
}

// Specifies serviceability-related operations, such as gathering server memory dumps and server traces.
type WebSphereLibertyApplicationServiceability struct {
	// A convenient field to request the size of the persisted storage to use for serviceability.
	// +kubebuilder:validation:Pattern=^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$
	Size string `json:"size,omitempty"`

	// The name of the PersistentVolumeClaim resource you created to be used for serviceability.
	// +kubebuilder:validation:Pattern=.+
	VolumeClaimName string `json:"volumeClaimName,omitempty"`

	// A convenient field to request the StorageClassName of the persisted storage to use for serviceability.
	// +kubebuilder:validation:Pattern=.+
	StorageClassName string `json:"storageClassName,omitempty"`
}

// Configures the ingress resource.
type WebSphereLibertyApplicationRoute struct {

	// Annotations to be added to the Route.
	// +operator-sdk:csv:customresourcedefinitions:order=42,type=spec,displayName="Route Annotations",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Annotations map[string]string `json:"annotations,omitempty"`

	// Hostname to be used for the Route.
	// +operator-sdk:csv:customresourcedefinitions:order=43,type=spec,displayName="Route Host",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Host string `json:"host,omitempty"`

	// Path to be used for Route.
	// +operator-sdk:csv:customresourcedefinitions:order=44,type=spec,displayName="Route Path",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Path string `json:"path,omitempty"`

	// Path type to be used for Ingress. This does not apply to Route on OpenShift.
	// +operator-sdk:csv:customresourcedefinitions:order=44,type=spec,displayName="Path Type",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:select:Exact", "urn:alm:descriptor:com.tectonic.ui:select:Prefix", "urn:alm:descriptor:com.tectonic.ui:select:ImplementationSpecific"}
	PathType networkingv1.PathType `json:"pathType,omitempty"`

	// A name of a secret that already contains TLS key, certificate and CA to be used in the route. It can also contain destination CA certificate. The following keys are valid in the secret: ca.crt, destCA.crt, tls.crt, and tls.key.
	// +operator-sdk:csv:customresourcedefinitions:order=45,type=spec,displayName="Certificate Secret Reference",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	CertificateSecretRef *string `json:"certificateSecretRef,omitempty"`

	// TLS termination policy. Can be one of edge, reencrypt and passthrough.
	// +operator-sdk:csv:customresourcedefinitions:order=46,type=spec,displayName="Termination",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:select:edge", "urn:alm:descriptor:com.tectonic.ui:select:reencrypt", "urn:alm:descriptor:com.tectonic.ui:select:passthrough"}
	Termination *routev1.TLSTerminationType `json:"termination,omitempty"`

	// HTTP traffic policy with TLS enabled. Can be one of Allow, Redirect and None.
	// +operator-sdk:csv:customresourcedefinitions:order=47,type=spec,displayName="Insecure Edge Termination Policy",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:select:Allow", "urn:alm:descriptor:com.tectonic.ui:select:Redirect", "urn:alm:descriptor:com.tectonic.ui:select:None"}
	InsecureEdgeTerminationPolicy *routev1.InsecureEdgeTerminationPolicyType `json:"insecureEdgeTerminationPolicy,omitempty"`
}

// Defines the observed state of WebSphereLibertyApplication.
type WebSphereLibertyApplicationStatus struct {
	// +listType=atomic
	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="Status Conditions",xDescriptors="urn:alm:descriptor:io.kubernetes.conditions"
	Conditions     []StatusCondition `json:"conditions,omitempty"`
	Endpoints      []StatusEndpoint  `json:"endpoints,omitempty"`
	RouteAvailable *bool             `json:"routeAvailable,omitempty"`
	ImageReference string            `json:"imageReference,omitempty"`
	Versions       StatusVersions    `json:"versions,omitempty"`

	// +operator-sdk:csv:customresourcedefinitions:order=61,type=status,displayName="Service Binding"
	Binding *corev1.LocalObjectReference `json:"binding,omitempty"`

	References common.StatusReferences `json:"references,omitempty"`
}

// Defines possible status conditions.
type StatusCondition struct {
	LastTransitionTime *metav1.Time           `json:"lastTransitionTime,omitempty"`
	Reason             string                 `json:"reason,omitempty"`
	Message            string                 `json:"message,omitempty"`
	Status             corev1.ConditionStatus `json:"status,omitempty"`
	Type               StatusConditionType    `json:"type,omitempty"`
}

// Defines the type of status condition.
type StatusConditionType string

// Reports endpoint information.
type StatusEndpoint struct {
	Name  string              `json:"name,omitempty"`
	Scope StatusEndpointScope `json:"scope,omitempty"`
	Type  string              `json:"type,omitempty"`
	// Exposed URI of the application endpoint
	// +operator-sdk:csv:customresourcedefinitions:order=60,type=status,displayName="Application",xDescriptors={"urn:alm:descriptor:org.w3:link"}
	URI string `json:"uri,omitempty"`
}

// Defines the scope of endpoint information in status.
type StatusEndpointScope string

const (
	// Status Condition Types
	StatusConditionTypeReconciled     StatusConditionType = "Reconciled"
	StatusConditionTypeResourcesReady StatusConditionType = "ResourcesReady"
	StatusConditionTypeReady          StatusConditionType = "Ready"

	// Status Endpoint Scopes
	StatusEndpointScopeExternal StatusEndpointScope = "External"
	StatusEndpointScopeInternal StatusEndpointScope = "Internal"
)

type StatusVersions struct {
	Reconciled string `json:"reconciled,omitempty"`
}

// +kubebuilder:resource:path=webspherelibertyapplications,scope=Namespaced,shortName=wlapp;wlapps
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Image",type="string",JSONPath=".spec.applicationImage",priority=0,description="Absolute name of the deployed image containing registry and tag"
// +kubebuilder:printcolumn:name="Exposed",type="boolean",JSONPath=".spec.expose",priority=0,description="Specifies whether deployment is exposed externally via default Route"
// +kubebuilder:printcolumn:name="Reconciled",type="string",JSONPath=".status.conditions[?(@.type=='Reconciled')].status",priority=0,description="Status of the reconcile condition"
// +kubebuilder:printcolumn:name="ReconciledReason",type="string",JSONPath=".status.conditions[?(@.type=='Reconciled')].reason",priority=1,description="Reason for the failure of reconcile condition"
// +kubebuilder:printcolumn:name="ReconciledMessage",type="string",JSONPath=".status.conditions[?(@.type=='Reconciled')].message",priority=1,description="Failure message from reconcile condition"
// +kubebuilder:printcolumn:name="ResourcesReady",type="string",JSONPath=".status.conditions[?(@.type=='ResourcesReady')].status",priority=0,description="Status of the resource ready condition"
// +kubebuilder:printcolumn:name="ResourcesReadyReason",type="string",JSONPath=".status.conditions[?(@.type=='ResourcesReady')].reason",priority=1,description="Reason for the failure of resource ready condition"
// +kubebuilder:printcolumn:name="ResourcesReadyMessage",type="string",JSONPath=".status.conditions[?(@.type=='ResourcesReady')].message",priority=1,description="Failure message from resource ready condition"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status",priority=0,description="Status of the component ready condition"
// +kubebuilder:printcolumn:name="ReadyReason",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].reason",priority=1,description="Reason for the failure of component ready condition"
// +kubebuilder:printcolumn:name="ReadyMessage",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].message",priority=1,description="Failure message from component ready condition"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",priority=0,description="Age of the resource"
// +operator-sdk:csv:customresourcedefinitions:displayName="WebSphereLibertyApplication",resources={{Deployment,v1},{Service,v1},{StatefulSet,v1},{Route,v1},{HorizontalPodAutoscaler,v1},{ServiceAccount,v1},{Secret,v1},{NetworkPolicy,v1}}

// Represents the deployment of a WebSphere Liberty application. Documentation: For more information about installation parameters, see https://ibm.biz/wlo-crs. License: By installing this product, you accept the license terms at https://ibm.biz/was-license.
type WebSphereLibertyApplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WebSphereLibertyApplicationSpec   `json:"spec,omitempty"`
	Status WebSphereLibertyApplicationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WebSphereLibertyApplicationList contains a list of WebSphereLibertyApplication
type WebSphereLibertyApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WebSphereLibertyApplication `json:"items"`
}

// Specifies the configuration for Single Sign-On (SSO) providers to authenticate with.
type WebSphereLibertyApplicationSSO struct {
	// +listType=atomic
	// +operator-sdk:csv:customresourcedefinitions:order=1,type=spec,displayName="OIDC"
	OIDC []OidcClient `json:"oidc,omitempty"`

	// +listType=atomic
	// +operator-sdk:csv:customresourcedefinitions:order=2,type=spec,displayName="OAuth2"
	Oauth2 []OAuth2Client `json:"oauth2,omitempty"`

	// +operator-sdk:csv:customresourcedefinitions:order=3,type=spec,displayName="GitHub"
	Github *GithubLogin `json:"github,omitempty"`

	// Common parameters for all SSO providers

	// Specifies a callback protocol, host and port number.
	// +operator-sdk:csv:customresourcedefinitions:order=4,type=spec,displayName="Redirect to RP Host and Port"
	RedirectToRPHostAndPort string `json:"redirectToRPHostAndPort,omitempty"`

	// Specifies whether to map a user identifier to a registry user. This parameter applies to all providers.
	// +operator-sdk:csv:customresourcedefinitions:order=5,type=spec,displayName="Map to User Registry",xDescriptors="urn:alm:descriptor:com.tectonic.ui:booleanSwitch"
	MapToUserRegistry *bool `json:"mapToUserRegistry,omitempty"`
}

// Represents configuration for an OpenID Connect (OIDC) client.
type OidcClient struct {
	// The unique ID for the provider. Default value is oidc.
	// +operator-sdk:csv:customresourcedefinitions:order=1,type=spec,displayName="ID"
	ID string `json:"id,omitempty"`

	// Specifies a discovery endpoint URL for the OpenID Connect provider. Required field.
	// +operator-sdk:csv:customresourcedefinitions:order=2,type=spec
	DiscoveryEndpoint string `json:"discoveryEndpoint"`

	// Specifies the name of the claim. Use its value as the user group membership.
	// +operator-sdk:csv:customresourcedefinitions:order=3,type=spec
	GroupNameAttribute string `json:"groupNameAttribute,omitempty"`

	// Specifies the name of the claim. Use its value as the authenticated user principal.
	// +operator-sdk:csv:customresourcedefinitions:order=4,type=spec
	UserNameAttribute string `json:"userNameAttribute,omitempty"`

	// The name of the social login configuration for display.
	// +operator-sdk:csv:customresourcedefinitions:order=5,type=spec
	DisplayName string `json:"displayName,omitempty"`

	// Specifies whether the UserInfo endpoint is contacted.
	// +operator-sdk:csv:customresourcedefinitions:order=6,type=spec,displayName="User Info Endpoint Enabled",xDescriptors="urn:alm:descriptor:com.tectonic.ui:booleanSwitch"
	UserInfoEndpointEnabled *bool `json:"userInfoEndpointEnabled,omitempty"`

	// Specifies the name of the claim. Use its value as the subject realm.
	// +operator-sdk:csv:customresourcedefinitions:order=7,type=spec
	RealmNameAttribute string `json:"realmNameAttribute,omitempty"`

	// Specifies one or more scopes to request.
	// +operator-sdk:csv:customresourcedefinitions:order=8,type=spec
	Scope string `json:"scope,omitempty"`

	// Specifies the required authentication method.
	// +operator-sdk:csv:customresourcedefinitions:order=9,type=spec
	TokenEndpointAuthMethod string `json:"tokenEndpointAuthMethod,omitempty"`

	// Specifies whether to enable host name verification when the client contacts the provider.
	// +operator-sdk:csv:customresourcedefinitions:order=10,type=spec,displayName="Host Name Verification Enabled",xDescriptors="urn:alm:descriptor:com.tectonic.ui:booleanSwitch"
	HostNameVerificationEnabled *bool `json:"hostNameVerificationEnabled,omitempty"`
}

// Represents configuration for an OAuth2 client.
type OAuth2Client struct {
	// Specifies the unique ID for the provider. The default value is oauth2.
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="ID"
	ID string `json:"id,omitempty"`

	// Specifies a token endpoint URL for the OAuth 2.0 provider. Required field.
	TokenEndpoint string `json:"tokenEndpoint"`

	// Specifies an authorization endpoint URL for the OAuth 2.0 provider. Required field.
	AuthorizationEndpoint string `json:"authorizationEndpoint"`

	// Specifies the name of the claim. Use its value as the user group membership
	GroupNameAttribute string `json:"groupNameAttribute,omitempty"`

	// Specifies the name of the claim. Use its value as the authenticated user principal.
	UserNameAttribute string `json:"userNameAttribute,omitempty"`

	// The name of the social login configuration for display.
	DisplayName string `json:"displayName,omitempty"`

	// Specifies the name of the claim. Use its value as the subject realm.
	RealmNameAttribute string `json:"realmNameAttribute,omitempty"`

	// Specifies the realm name for this social media.
	RealmName string `json:"realmName,omitempty"`

	// Specifies one or more scopes to request.
	Scope string `json:"scope,omitempty"`

	// Specifies the required authentication method.
	TokenEndpointAuthMethod string `json:"tokenEndpointAuthMethod,omitempty"`

	// Name of the header to use when an OAuth access token is forwarded.
	AccessTokenHeaderName string `json:"accessTokenHeaderName,omitempty"`

	// Determines whether the access token that is provided in the request is used for authentication.
	// +operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors="urn:alm:descriptor:com.tectonic.ui:booleanSwitch"
	AccessTokenRequired *bool `json:"accessTokenRequired,omitempty"`

	// Determines whether to support access token authentication if an access token is provided in the request.
	// +operator-sdk:csv:customresourcedefinitions:type=spec,xDescriptors="urn:alm:descriptor:com.tectonic.ui:booleanSwitch"
	AccessTokenSupported *bool `json:"accessTokenSupported,omitempty"`

	// Indicates which specification to use for the user API.
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="User API Type"
	UserApiType string `json:"userApiType,omitempty"`

	// The URL for retrieving the user information.
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="User API"
	UserApi string `json:"userApi,omitempty"`
}

// Represents configuration for social login using GitHub.
type GithubLogin struct {
	// Specifies the host name of your enterprise GitHub.
	Hostname string `json:"hostname,omitempty"`
}

func init() {
	SchemeBuilder.Register(&WebSphereLibertyApplication{}, &WebSphereLibertyApplicationList{})
}

// GetApplicationImage returns application image
func (cr *WebSphereLibertyApplication) GetApplicationImage() string {
	return cr.Spec.ApplicationImage
}

// GetPullPolicy returns image pull policy
func (cr *WebSphereLibertyApplication) GetPullPolicy() *corev1.PullPolicy {
	return cr.Spec.PullPolicy
}

// GetPullSecret returns secret name for docker registry credentials
func (cr *WebSphereLibertyApplication) GetPullSecret() *string {
	return cr.Spec.PullSecret
}

// GetServiceAccountName returns service account name
func (cr *WebSphereLibertyApplication) GetServiceAccountName() *string {
	return cr.Spec.ServiceAccountName
}

// GetReplicas returns number of replicas
func (cr *WebSphereLibertyApplication) GetReplicas() *int32 {
	return cr.Spec.Replicas
}

func (cr *WebSphereLibertyApplication) GetProbes() common.BaseComponentProbes {
	if cr.Spec.Probes == nil {
		return nil
	}
	return cr.Spec.Probes
}

// GetLivenessProbe returns liveness probe
func (p *WebSphereLibertyApplicationProbes) GetLivenessProbe() *corev1.Probe {
	return p.Liveness
}

// GetReadinessProbe returns readiness probe
func (p *WebSphereLibertyApplicationProbes) GetReadinessProbe() *corev1.Probe {
	return p.Readiness
}

// GetStartupProbe returns startup probe
func (p *WebSphereLibertyApplicationProbes) GetStartupProbe() *corev1.Probe {
	return p.Startup
}

// GetDefaultLivenessProbe returns default values for liveness probe
func (p *WebSphereLibertyApplicationProbes) GetDefaultLivenessProbe(ba common.BaseComponent) *corev1.Probe {
	return common.GetDefaultMicroProfileLivenessProbe(ba)
}

// GetDefaultReadinessProbe returns default values for readiness probe
func (p *WebSphereLibertyApplicationProbes) GetDefaultReadinessProbe(ba common.BaseComponent) *corev1.Probe {
	return common.GetDefaultMicroProfileReadinessProbe(ba)
}

// GetDefaultStartupProbe returns default values for startup probe
func (p *WebSphereLibertyApplicationProbes) GetDefaultStartupProbe(ba common.BaseComponent) *corev1.Probe {
	return common.GetDefaultMicroProfileStartupProbe(ba)
}

// GetVolumes returns volumes slice
func (cr *WebSphereLibertyApplication) GetVolumes() []corev1.Volume {
	return cr.Spec.Volumes
}

// GetVolumeMounts returns volume mounts slice
func (cr *WebSphereLibertyApplication) GetVolumeMounts() []corev1.VolumeMount {
	return cr.Spec.VolumeMounts
}

// GetResourceConstraints returns resource constraints
func (cr *WebSphereLibertyApplication) GetResourceConstraints() *corev1.ResourceRequirements {
	return cr.Spec.Resources
}

// GetExpose returns expose flag
func (cr *WebSphereLibertyApplication) GetExpose() *bool {
	return cr.Spec.Expose
}

// GetManageTLS returns deployment's node and pod affinity settings
func (cr *WebSphereLibertyApplication) GetManageTLS() *bool {
	return cr.Spec.ManageTLS
}

// GetEnv returns slice of environment variables
func (cr *WebSphereLibertyApplication) GetEnv() []corev1.EnvVar {
	return cr.Spec.Env
}

// GetEnvFrom returns slice of environment variables from source
func (cr *WebSphereLibertyApplication) GetEnvFrom() []corev1.EnvFromSource {
	return cr.Spec.EnvFrom
}

// GetCreateKnativeService returns flag that toggles Knative service
func (cr *WebSphereLibertyApplication) GetCreateKnativeService() *bool {
	return cr.Spec.CreateKnativeService
}

// GetAutoscaling returns autoscaling settings
func (cr *WebSphereLibertyApplication) GetAutoscaling() common.BaseComponentAutoscaling {
	if cr.Spec.Autoscaling == nil {
		return nil
	}
	return cr.Spec.Autoscaling
}

// GetStorage returns storage settings
func (ss *WebSphereLibertyApplicationStatefulSet) GetStorage() common.BaseComponentStorage {
	if ss.Storage == nil {
		return nil
	}
	return ss.Storage
}

// GetService returns service settings
func (cr *WebSphereLibertyApplication) GetService() common.BaseComponentService {
	if cr.Spec.Service == nil {
		return nil
	}
	return cr.Spec.Service
}

// GetNetworkPolicy returns network policy settings
func (cr *WebSphereLibertyApplication) GetNetworkPolicy() common.BaseComponentNetworkPolicy {
	return cr.Spec.NetworkPolicy
}

// GetApplicationVersion returns application version
func (cr *WebSphereLibertyApplication) GetApplicationVersion() string {
	return cr.Spec.ApplicationVersion
}

// GetApplicationName returns Application name
func (cr *WebSphereLibertyApplication) GetApplicationName() string {
	return cr.Spec.ApplicationName
}

// GetMonitoring returns monitoring settings
func (cr *WebSphereLibertyApplication) GetMonitoring() common.BaseComponentMonitoring {
	if cr.Spec.Monitoring == nil {
		return nil
	}
	return cr.Spec.Monitoring
}

// GetStatus returns WebSphereLibertyApplication status
func (cr *WebSphereLibertyApplication) GetStatus() common.BaseComponentStatus {
	return &cr.Status
}

// GetInitContainers returns list of init containers
func (cr *WebSphereLibertyApplication) GetInitContainers() []corev1.Container {
	return cr.Spec.InitContainers
}

// GetSidecarContainers returns list of sidecar containers
func (cr *WebSphereLibertyApplication) GetSidecarContainers() []corev1.Container {
	return cr.Spec.SidecarContainers
}

// GetGroupName returns group name to be used in labels and annotation
func (cr *WebSphereLibertyApplication) GetGroupName() string {
	return "liberty.websphere.ibm.com"
}

// GetRoute returns route
func (cr *WebSphereLibertyApplication) GetRoute() common.BaseComponentRoute {
	if cr.Spec.Route == nil {
		return nil
	}
	return cr.Spec.Route
}

// GetAffinity returns deployment's node and pod affinity settings
func (cr *WebSphereLibertyApplication) GetAffinity() common.BaseComponentAffinity {
	if cr.Spec.Affinity == nil {
		return nil
	}
	return cr.Spec.Affinity
}

// GetDeployment returns deployment settings
func (cr *WebSphereLibertyApplication) GetDeployment() common.BaseComponentDeployment {
	if cr.Spec.Deployment == nil {
		return nil
	}
	return cr.Spec.Deployment
}

// GetDeploymentStrategy returns deployment strategy struct
func (cr *WebSphereLibertyApplicationDeployment) GetDeploymentUpdateStrategy() *appsv1.DeploymentStrategy {
	return cr.UpdateStrategy
}

// GetAnnotations returns annotations to be added only to the Deployment and its child resources
func (rcd *WebSphereLibertyApplicationDeployment) GetAnnotations() map[string]string {
	return rcd.Annotations
}

// GetStatefulSet returns statefulSet settings
func (cr *WebSphereLibertyApplication) GetStatefulSet() common.BaseComponentStatefulSet {
	if cr.Spec.StatefulSet == nil {
		return nil
	}
	return cr.Spec.StatefulSet
}

// GetStatefulSetUpdateStrategy returns statefulSet strategy struct
func (cr *WebSphereLibertyApplicationStatefulSet) GetStatefulSetUpdateStrategy() *appsv1.StatefulSetUpdateStrategy {
	return cr.UpdateStrategy
}

// GetAnnotations returns annotations to be added only to the StatefulSet and its child resources
func (rcss *WebSphereLibertyApplicationStatefulSet) GetAnnotations() map[string]string {
	return rcss.Annotations
}

// GetImageReference returns Docker image reference to be deployed by the CR
func (s *WebSphereLibertyApplicationStatus) GetImageReference() string {
	return s.ImageReference
}

// SetImageReference sets Docker image reference on the status portion of the CR
func (s *WebSphereLibertyApplicationStatus) SetImageReference(imageReference string) {
	s.ImageReference = imageReference
}

// GetBinding returns BindingStatus representing binding status
func (s *WebSphereLibertyApplicationStatus) GetBinding() *corev1.LocalObjectReference {
	return s.Binding
}

// SetBinding sets BindingStatus representing binding status
func (s *WebSphereLibertyApplicationStatus) SetBinding(r *corev1.LocalObjectReference) {
	s.Binding = r
}

func (s *WebSphereLibertyApplicationStatus) GetReferences() common.StatusReferences {
	if s.References == nil {
		s.References = make(common.StatusReferences)
	}
	return s.References
}

func (s *WebSphereLibertyApplicationStatus) SetReferences(refs common.StatusReferences) {
	s.References = refs
}

func (s *WebSphereLibertyApplicationStatus) SetReference(name string, value string) {
	if s.References == nil {
		s.References = make(common.StatusReferences)
	}
	s.References[name] = value
}

// GetMinReplicas returns minimum replicas
func (a *WebSphereLibertyApplicationAutoScaling) GetMinReplicas() *int32 {
	return a.MinReplicas
}

// GetMaxReplicas returns maximum replicas
func (a *WebSphereLibertyApplicationAutoScaling) GetMaxReplicas() int32 {
	return a.MaxReplicas
}

// GetTargetCPUUtilizationPercentage returns target cpu usage
func (a *WebSphereLibertyApplicationAutoScaling) GetTargetCPUUtilizationPercentage() *int32 {
	return a.TargetCPUUtilizationPercentage
}

// GetSize returns pesistent volume size
func (s *WebSphereLibertyApplicationStorage) GetSize() string {
	return s.Size
}

// GetClassName returns persistent volume ClassName
func (s *WebSphereLibertyApplicationStorage) GetClassName() string {
	return s.ClassName
}

// GetMountPath returns mount path for persistent volume
func (s *WebSphereLibertyApplicationStorage) GetMountPath() string {
	return s.MountPath
}

// GetVolumeClaimTemplate returns a template representing requested persistent volume
func (s *WebSphereLibertyApplicationStorage) GetVolumeClaimTemplate() *corev1.PersistentVolumeClaim {
	return s.VolumeClaimTemplate
}

// GetAnnotations returns a set of annotations to be added to the service
func (s *WebSphereLibertyApplicationService) GetAnnotations() map[string]string {
	return s.Annotations
}

// GetServiceability returns serviceability
func (cr *WebSphereLibertyApplication) GetServiceability() *WebSphereLibertyApplicationServiceability {
	return cr.Spec.Serviceability
}

// GetSize returns pesistent volume size for Serviceability
func (s *WebSphereLibertyApplicationServiceability) GetSize() string {
	return s.Size
}

// GetVolumeClaimName returns the name of custom PersistentVolumeClaim (PVC) for Serviceability. Must be in the same namespace as the WebSphereLibertyApplication.
func (s *WebSphereLibertyApplicationServiceability) GetVolumeClaimName() string {
	return s.VolumeClaimName
}

// GetPort returns service port
func (s *WebSphereLibertyApplicationService) GetPort() int32 {
	if s != nil && s.Port != 0 {
		return s.Port
	}
	return 9443
}

// GetNodePort returns service nodePort
func (s *WebSphereLibertyApplicationService) GetNodePort() *int32 {
	if s.NodePort == nil {
		return nil
	}
	return s.NodePort
}

// GetTargetPort returns the internal target port for containers
func (s *WebSphereLibertyApplicationService) GetTargetPort() *int32 {
	return s.TargetPort
}

// GetPortName returns name of service port
func (s *WebSphereLibertyApplicationService) GetPortName() string {
	return s.PortName
}

// GetType returns service type
func (s *WebSphereLibertyApplicationService) GetType() *corev1.ServiceType {
	return s.Type
}

// GetPorts returns a list of service ports
func (s *WebSphereLibertyApplicationService) GetPorts() []corev1.ServicePort {
	return s.Ports
}

// GetCertificateSecretRef returns a secret reference with a certificate
func (s *WebSphereLibertyApplicationService) GetCertificateSecretRef() *string {
	return s.CertificateSecretRef
}

// GetBindable returns whether the application should be exposable as a service
func (s *WebSphereLibertyApplicationService) GetBindable() *bool {
	return s.Bindable
}

// GetNamespaceLabels returns the namespace selector labels that should be used for the ingress rule
func (np *WebSphereLibertyApplicationNetworkPolicy) GetNamespaceLabels() map[string]string {
	if np == nil || np.NamespaceLabels == nil {
		return nil
	}
	return *np.NamespaceLabels
}

// GetFromLabels returns the pod selector labels that should be used for the ingress rule
func (np *WebSphereLibertyApplicationNetworkPolicy) GetFromLabels() map[string]string {
	if np == nil || np.FromLabels == nil {
		return nil
	}
	return *np.FromLabels
}

// IsDisabled returns whether the network policy should be created or not
func (np *WebSphereLibertyApplicationNetworkPolicy) IsDisabled() bool {
	return np != nil && np.Disable != nil && *np.Disable
}

// GetLabels returns labels to be added on ServiceMonitor
func (m *WebSphereLibertyApplicationMonitoring) GetLabels() map[string]string {
	return m.Labels
}

// GetEndpoints returns endpoints to be added to ServiceMonitor
func (m *WebSphereLibertyApplicationMonitoring) GetEndpoints() []prometheusv1.Endpoint {
	return m.Endpoints
}

// GetAnnotations returns route annotations
func (r *WebSphereLibertyApplicationRoute) GetAnnotations() map[string]string {
	return r.Annotations
}

// GetCertificateSecretRef returns a secret reference with a certificate
func (r *WebSphereLibertyApplicationRoute) GetCertificateSecretRef() *string {
	return r.CertificateSecretRef
}

// GetTermination returns terminatation of the route's TLS
func (r *WebSphereLibertyApplicationRoute) GetTermination() *routev1.TLSTerminationType {
	return r.Termination
}

// GetInsecureEdgeTerminationPolicy returns terminatation of the route's TLS
func (r *WebSphereLibertyApplicationRoute) GetInsecureEdgeTerminationPolicy() *routev1.InsecureEdgeTerminationPolicyType {
	return r.InsecureEdgeTerminationPolicy
}

// GetHost returns hostname to be used by the route
func (r *WebSphereLibertyApplicationRoute) GetHost() string {
	return r.Host
}

// GetPath returns path to use for the route
func (r *WebSphereLibertyApplicationRoute) GetPath() string {
	return r.Path
}

// GetPathType returns pathType to use for the route
func (r *WebSphereLibertyApplicationRoute) GetPathType() networkingv1.PathType {
	return r.PathType
}

// GetNodeAffinity returns node affinity
func (a *WebSphereLibertyApplicationAffinity) GetNodeAffinity() *corev1.NodeAffinity {
	return a.NodeAffinity
}

// GetPodAffinity returns pod affinity
func (a *WebSphereLibertyApplicationAffinity) GetPodAffinity() *corev1.PodAffinity {
	return a.PodAffinity
}

// GetPodAntiAffinity returns pod anti-affinity
func (a *WebSphereLibertyApplicationAffinity) GetPodAntiAffinity() *corev1.PodAntiAffinity {
	return a.PodAntiAffinity
}

// GetArchitecture returns list of architecture names
func (a *WebSphereLibertyApplicationAffinity) GetArchitecture() []string {
	return a.Architecture
}

// GetNodeAffinityLabels returns list of architecture names
func (a *WebSphereLibertyApplicationAffinity) GetNodeAffinityLabels() map[string]string {
	return a.NodeAffinityLabels
}

// GetSecurityContext returns container security context
func (cr *WebSphereLibertyApplication) GetSecurityContext() *corev1.SecurityContext {
	return cr.Spec.SecurityContext
}

// Initialize sets default values
func (cr *WebSphereLibertyApplication) Initialize() {
	if cr.Spec.PullPolicy == nil {
		pp := corev1.PullIfNotPresent
		cr.Spec.PullPolicy = &pp
	}

	if cr.Spec.Resources == nil {
		cr.Spec.Resources = &corev1.ResourceRequirements{}
	}

	// Default applicationName to cr.Name, if a user sets createAppDefinition to true but doesn't set applicationName
	if cr.Spec.ApplicationName == "" {
		if cr.Labels != nil && cr.Labels["app.kubernetes.io/part-of"] != "" {
			cr.Spec.ApplicationName = cr.Labels["app.kubernetes.io/part-of"]
		} else {
			cr.Spec.ApplicationName = cr.Name
		}
	}

	if cr.Labels != nil {
		cr.Labels["app.kubernetes.io/part-of"] = cr.Spec.ApplicationName
	}

	// This is to handle when there is no service in the CR
	if cr.Spec.Service == nil {
		cr.Spec.Service = &WebSphereLibertyApplicationService{}
	}

	if cr.Spec.Service.Type == nil {
		st := corev1.ServiceTypeClusterIP
		cr.Spec.Service.Type = &st
	}

	if cr.Spec.Service.Port == 0 {
		if cr.Spec.ManageTLS == nil || *cr.Spec.ManageTLS {
			cr.Spec.Service.Port = 9443

		} else {
			cr.Spec.Service.Port = 9080
		}
	}

	// If TargetPorts on Serviceports are not set, default them to the Port value in the CR
	numOfAdditionalPorts := len(cr.GetService().GetPorts())
	for i := 0; i < numOfAdditionalPorts; i++ {
		if cr.Spec.Service.Ports[i].TargetPort.String() == "0" {
			cr.Spec.Service.Ports[i].TargetPort = intstr.FromInt(int(cr.Spec.Service.Ports[i].Port))
		}
	}

	if cr.Spec.License.Edition == "" {
		cr.Spec.License.Edition = LicenseEditionBase
	}
	if cr.Spec.License.ProductEntitlementSource == "" {
		cr.Spec.License.ProductEntitlementSource = LicenseEntitlementStandalone
	}
	if cr.Spec.License.Metric == "" {
		if cr.Spec.License.ProductEntitlementSource == LicenseEntitlementWSHE || cr.Spec.License.ProductEntitlementSource == LicenseEntitlementCP4Apps {
			cr.Spec.License.Metric = LicenseMetricVPC
		} else if cr.Spec.License.ProductEntitlementSource == LicenseEntitlementStandalone || cr.Spec.License.ProductEntitlementSource == LicenseEntitlementFamilyEdition {
			cr.Spec.License.Metric = LicenseMetricPVU
		}
	}
}

// GetLabels returns set of labels to be added to all resources
func (cr *WebSphereLibertyApplication) GetLabels() map[string]string {
	labels := map[string]string{
		"app.kubernetes.io/instance":     cr.Name,
		"app.kubernetes.io/name":         cr.Name,
		"app.kubernetes.io/managed-by":   "websphere-liberty-operator",
		"app.kubernetes.io/component":    "backend",
		"app.kubernetes.io/part-of":      cr.Spec.ApplicationName,
		common.GetComponentNameLabel(cr): cr.Name,
	}

	if cr.Spec.ApplicationVersion != "" {
		labels["app.kubernetes.io/version"] = cr.Spec.ApplicationVersion
	}

	for key, value := range cr.Labels {
		if key != "app.kubernetes.io/instance" {
			labels[key] = value
		}
	}

	return labels
}

// GetAnnotations returns set of annotations to be added to all resources
func (cr *WebSphereLibertyApplication) GetAnnotations() map[string]string {
	annotations := map[string]string{}
	for k, v := range cr.Annotations {
		annotations[k] = v
	}
	delete(annotations, "kubectl.kubernetes.io/last-applied-configuration")
	return annotations
}

// GetType returns status condition type
func (c *StatusCondition) GetType() common.StatusConditionType {
	return convertToCommonStatusConditionType(c.Type)
}

// SetType returns status condition type
func (c *StatusCondition) SetType(ct common.StatusConditionType) {
	c.Type = convertFromCommonStatusConditionType(ct)
}

// GetLastTransitionTime returns time of last status change
func (c *StatusCondition) GetLastTransitionTime() *metav1.Time {
	return c.LastTransitionTime
}

// SetLastTransitionTime sets time of last status change
func (c *StatusCondition) SetLastTransitionTime(t *metav1.Time) {
	c.LastTransitionTime = t
}

// GetMessage returns condition's message
func (c *StatusCondition) GetMessage() string {
	return c.Message
}

// SetMessage sets condition's message
func (c *StatusCondition) SetMessage(m string) {
	c.Message = m
}

// GetReason returns condition's message
func (c *StatusCondition) GetReason() string {
	return c.Reason
}

// SetReason sets condition's reason
func (c *StatusCondition) SetReason(r string) {
	c.Reason = r
}

// GetStatus returns condition's status
func (c *StatusCondition) GetStatus() corev1.ConditionStatus {
	return c.Status
}

// SetStatus sets condition's status
func (c *StatusCondition) SetStatus(s corev1.ConditionStatus) {
	c.Status = s
}

// SetConditionFields sets status condition fields
func (c *StatusCondition) SetConditionFields(message string, reason string, status corev1.ConditionStatus) common.StatusCondition {
	c.Message = message
	c.Reason = reason
	c.Status = status
	return c
}

// NewCondition returns new condition
func (s *WebSphereLibertyApplicationStatus) NewCondition(ct common.StatusConditionType) common.StatusCondition {
	c := &StatusCondition{}
	c.Type = convertFromCommonStatusConditionType(ct)
	return c
}

// GetConditions returns slice of conditions
func (s *WebSphereLibertyApplicationStatus) GetConditions() []common.StatusCondition {
	var conditions = make([]common.StatusCondition, len(s.Conditions))
	for i := range s.Conditions {
		conditions[i] = &s.Conditions[i]
	}
	return conditions
}

// GetCondition returns status condition with status condition type
func (s *WebSphereLibertyApplicationStatus) GetCondition(t common.StatusConditionType) common.StatusCondition {
	for i := range s.Conditions {
		if s.Conditions[i].GetType() == t {
			return &s.Conditions[i]
		}
	}
	return nil
}

// SetCondition sets status condition
func (s *WebSphereLibertyApplicationStatus) SetCondition(c common.StatusCondition) {
	condition := &StatusCondition{}
	found := false
	for i := range s.Conditions {
		if s.Conditions[i].GetType() == c.GetType() {
			condition = &s.Conditions[i]
			found = true
			break
		}
	}

	if condition.GetStatus() != c.GetStatus() || condition.GetMessage() != c.GetMessage() {
		condition.SetLastTransitionTime(&metav1.Time{Time: time.Now()})
	}

	condition.SetReason(c.GetReason())
	condition.SetMessage(c.GetMessage())
	condition.SetStatus(c.GetStatus())
	condition.SetType(c.GetType())
	if !found {
		s.Conditions = append(s.Conditions, *condition)
	}
}

func convertToCommonStatusConditionType(c StatusConditionType) common.StatusConditionType {
	switch c {
	case StatusConditionTypeReconciled:
		return common.StatusConditionTypeReconciled
	case StatusConditionTypeResourcesReady:
		return common.StatusConditionTypeResourcesReady
	case StatusConditionTypeReady:
		return common.StatusConditionTypeReady
	default:
		panic(c)
	}
}

func convertFromCommonStatusConditionType(c common.StatusConditionType) StatusConditionType {
	switch c {
	case common.StatusConditionTypeReconciled:
		return StatusConditionTypeReconciled
	case common.StatusConditionTypeResourcesReady:
		return StatusConditionTypeResourcesReady
	case common.StatusConditionTypeReady:
		return StatusConditionTypeReady
	default:
		panic(c)
	}
}

// GetEndpointName returns endpoint name in status
func (e *StatusEndpoint) GetEndpointName() string {
	return e.Name
}

// SetEndpointName sets endpoint name in status
func (e *StatusEndpoint) SetEndpointName(n string) {
	e.Name = n
}

// GetEndpointScope returns endpoint scope in status
func (e *StatusEndpoint) GetEndpointScope() common.StatusEndpointScope {
	return convertToCommonStatusEndpointScope(e.Scope)
}

// SetEndpointScope sets endpoint scope in status
func (e *StatusEndpoint) SetEndpointScope(s common.StatusEndpointScope) {
	e.Scope = convertFromCommonStatusEndpointScope(s)
}

// GetEndpointType returns endpoint type in status
func (e *StatusEndpoint) GetEndpointType() string {
	return e.Type
}

// SetEndpointType sets endpoint type in status
func (e *StatusEndpoint) SetEndpointType(t string) {
	e.Type = t
}

// GetEndpointUri returns endpoint uri in status
func (e *StatusEndpoint) GetEndpointUri() string {
	return e.URI
}

// SetEndpointUri sets endpoint uri in status
func (e *StatusEndpoint) SetEndpointUri(u string) {
	e.URI = u
}

// SetStatusEndpointFields sets endpoint information fields
func (e *StatusEndpoint) SetStatusEndpointFields(eScope common.StatusEndpointScope, eType string, eUri string) common.StatusEndpoint {
	e.Scope = convertFromCommonStatusEndpointScope(eScope)
	e.Type = eType
	e.URI = eUri
	return e
}

// RemoveEndpoint removes endpoint in status
func (s *WebSphereLibertyApplicationStatus) RemoveStatusEndpoint(endpointName string) {
	endpoints := s.Endpoints
	for i, ep := range endpoints {
		if ep.GetEndpointName() == endpointName {
			s.Endpoints = append(endpoints[:i], endpoints[i+1:]...)
			break
		}
	}
}

// NewStatusEndpoint returns new endpoint information
func (s *WebSphereLibertyApplicationStatus) NewStatusEndpoint(endpointName string) common.StatusEndpoint {
	e := &StatusEndpoint{}
	e.Name = endpointName
	return e
}

// GetStatusEndpoint returns endpoint information with endpoint name
func (s *WebSphereLibertyApplicationStatus) GetStatusEndpoint(endpointName string) common.StatusEndpoint {
	for i := range s.Endpoints {
		if s.Endpoints[i].GetEndpointName() == endpointName {
			return &s.Endpoints[i]
		}
	}
	return nil
}

// SetStatusEndpoint sets endpoint in status
func (s *WebSphereLibertyApplicationStatus) SetStatusEndpoint(c common.StatusEndpoint) {
	endpoint := &StatusEndpoint{}
	found := false
	for i := range s.Endpoints {
		if s.Endpoints[i].GetEndpointName() == c.GetEndpointName() {
			endpoint = &s.Endpoints[i]
			found = true
			break
		}
	}

	endpoint.SetEndpointName(c.GetEndpointName())
	endpoint.SetEndpointScope(c.GetEndpointScope())
	endpoint.SetEndpointType(c.GetEndpointType())
	endpoint.SetEndpointUri(c.GetEndpointUri())
	if !found {
		s.Endpoints = append(s.Endpoints, *endpoint)
	}
}

func convertToCommonStatusEndpointScope(c StatusEndpointScope) common.StatusEndpointScope {
	switch c {
	case StatusEndpointScopeExternal:
		return common.StatusEndpointScopeExternal
	case StatusEndpointScopeInternal:
		return common.StatusEndpointScopeInternal
	default:
		panic(c)
	}
}

func convertFromCommonStatusEndpointScope(c common.StatusEndpointScope) StatusEndpointScope {
	switch c {
	case common.StatusEndpointScopeExternal:
		return StatusEndpointScopeExternal
	case common.StatusEndpointScopeInternal:
		return StatusEndpointScopeInternal
	default:
		panic(c)
	}
}
