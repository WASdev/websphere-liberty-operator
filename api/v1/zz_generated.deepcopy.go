//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	"github.com/application-stacks/runtime-component-operator/common"
	routev1 "github.com/openshift/api/route/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DumpStatusVersions) DeepCopyInto(out *DumpStatusVersions) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DumpStatusVersions.
func (in *DumpStatusVersions) DeepCopy() *DumpStatusVersions {
	if in == nil {
		return nil
	}
	out := new(DumpStatusVersions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GithubLogin) DeepCopyInto(out *GithubLogin) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GithubLogin.
func (in *GithubLogin) DeepCopy() *GithubLogin {
	if in == nil {
		return nil
	}
	out := new(GithubLogin)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *License) DeepCopyInto(out *License) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new License.
func (in *License) DeepCopy() *License {
	if in == nil {
		return nil
	}
	out := new(License)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LicenseSimple) DeepCopyInto(out *LicenseSimple) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LicenseSimple.
func (in *LicenseSimple) DeepCopy() *LicenseSimple {
	if in == nil {
		return nil
	}
	out := new(LicenseSimple)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OAuth2Client) DeepCopyInto(out *OAuth2Client) {
	*out = *in
	if in.AccessTokenRequired != nil {
		in, out := &in.AccessTokenRequired, &out.AccessTokenRequired
		*out = new(bool)
		**out = **in
	}
	if in.AccessTokenSupported != nil {
		in, out := &in.AccessTokenSupported, &out.AccessTokenSupported
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OAuth2Client.
func (in *OAuth2Client) DeepCopy() *OAuth2Client {
	if in == nil {
		return nil
	}
	out := new(OAuth2Client)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OidcClient) DeepCopyInto(out *OidcClient) {
	*out = *in
	if in.UserInfoEndpointEnabled != nil {
		in, out := &in.UserInfoEndpointEnabled, &out.UserInfoEndpointEnabled
		*out = new(bool)
		**out = **in
	}
	if in.HostNameVerificationEnabled != nil {
		in, out := &in.HostNameVerificationEnabled, &out.HostNameVerificationEnabled
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OidcClient.
func (in *OidcClient) DeepCopy() *OidcClient {
	if in == nil {
		return nil
	}
	out := new(OidcClient)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OperatedResource) DeepCopyInto(out *OperatedResource) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OperatedResource.
func (in *OperatedResource) DeepCopy() *OperatedResource {
	if in == nil {
		return nil
	}
	out := new(OperatedResource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OperationStatusCondition) DeepCopyInto(out *OperationStatusCondition) {
	*out = *in
	if in.LastTransitionTime != nil {
		in, out := &in.LastTransitionTime, &out.LastTransitionTime
		*out = (*in).DeepCopy()
	}
	in.LastUpdateTime.DeepCopyInto(&out.LastUpdateTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OperationStatusCondition.
func (in *OperationStatusCondition) DeepCopy() *OperationStatusCondition {
	if in == nil {
		return nil
	}
	out := new(OperationStatusCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StatusCondition) DeepCopyInto(out *StatusCondition) {
	*out = *in
	if in.LastTransitionTime != nil {
		in, out := &in.LastTransitionTime, &out.LastTransitionTime
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StatusCondition.
func (in *StatusCondition) DeepCopy() *StatusCondition {
	if in == nil {
		return nil
	}
	out := new(StatusCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StatusEndpoint) DeepCopyInto(out *StatusEndpoint) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StatusEndpoint.
func (in *StatusEndpoint) DeepCopy() *StatusEndpoint {
	if in == nil {
		return nil
	}
	out := new(StatusEndpoint)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StatusVersions) DeepCopyInto(out *StatusVersions) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StatusVersions.
func (in *StatusVersions) DeepCopy() *StatusVersions {
	if in == nil {
		return nil
	}
	out := new(StatusVersions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TraceStatusVersions) DeepCopyInto(out *TraceStatusVersions) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TraceStatusVersions.
func (in *TraceStatusVersions) DeepCopy() *TraceStatusVersions {
	if in == nil {
		return nil
	}
	out := new(TraceStatusVersions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebSphereLibertyApplication) DeepCopyInto(out *WebSphereLibertyApplication) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebSphereLibertyApplication.
func (in *WebSphereLibertyApplication) DeepCopy() *WebSphereLibertyApplication {
	if in == nil {
		return nil
	}
	out := new(WebSphereLibertyApplication)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *WebSphereLibertyApplication) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebSphereLibertyApplicationAffinity) DeepCopyInto(out *WebSphereLibertyApplicationAffinity) {
	*out = *in
	if in.NodeAffinity != nil {
		in, out := &in.NodeAffinity, &out.NodeAffinity
		*out = new(corev1.NodeAffinity)
		(*in).DeepCopyInto(*out)
	}
	if in.PodAffinity != nil {
		in, out := &in.PodAffinity, &out.PodAffinity
		*out = new(corev1.PodAffinity)
		(*in).DeepCopyInto(*out)
	}
	if in.PodAntiAffinity != nil {
		in, out := &in.PodAntiAffinity, &out.PodAntiAffinity
		*out = new(corev1.PodAntiAffinity)
		(*in).DeepCopyInto(*out)
	}
	if in.NodeAffinityLabels != nil {
		in, out := &in.NodeAffinityLabels, &out.NodeAffinityLabels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Architecture != nil {
		in, out := &in.Architecture, &out.Architecture
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebSphereLibertyApplicationAffinity.
func (in *WebSphereLibertyApplicationAffinity) DeepCopy() *WebSphereLibertyApplicationAffinity {
	if in == nil {
		return nil
	}
	out := new(WebSphereLibertyApplicationAffinity)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebSphereLibertyApplicationAutoScaling) DeepCopyInto(out *WebSphereLibertyApplicationAutoScaling) {
	*out = *in
	if in.MinReplicas != nil {
		in, out := &in.MinReplicas, &out.MinReplicas
		*out = new(int32)
		**out = **in
	}
	if in.TargetCPUUtilizationPercentage != nil {
		in, out := &in.TargetCPUUtilizationPercentage, &out.TargetCPUUtilizationPercentage
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebSphereLibertyApplicationAutoScaling.
func (in *WebSphereLibertyApplicationAutoScaling) DeepCopy() *WebSphereLibertyApplicationAutoScaling {
	if in == nil {
		return nil
	}
	out := new(WebSphereLibertyApplicationAutoScaling)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebSphereLibertyApplicationDeployment) DeepCopyInto(out *WebSphereLibertyApplicationDeployment) {
	*out = *in
	if in.UpdateStrategy != nil {
		in, out := &in.UpdateStrategy, &out.UpdateStrategy
		*out = new(appsv1.DeploymentStrategy)
		(*in).DeepCopyInto(*out)
	}
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebSphereLibertyApplicationDeployment.
func (in *WebSphereLibertyApplicationDeployment) DeepCopy() *WebSphereLibertyApplicationDeployment {
	if in == nil {
		return nil
	}
	out := new(WebSphereLibertyApplicationDeployment)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebSphereLibertyApplicationList) DeepCopyInto(out *WebSphereLibertyApplicationList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]WebSphereLibertyApplication, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebSphereLibertyApplicationList.
func (in *WebSphereLibertyApplicationList) DeepCopy() *WebSphereLibertyApplicationList {
	if in == nil {
		return nil
	}
	out := new(WebSphereLibertyApplicationList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *WebSphereLibertyApplicationList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebSphereLibertyApplicationMonitoring) DeepCopyInto(out *WebSphereLibertyApplicationMonitoring) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Endpoints != nil {
		in, out := &in.Endpoints, &out.Endpoints
		*out = make([]monitoringv1.Endpoint, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebSphereLibertyApplicationMonitoring.
func (in *WebSphereLibertyApplicationMonitoring) DeepCopy() *WebSphereLibertyApplicationMonitoring {
	if in == nil {
		return nil
	}
	out := new(WebSphereLibertyApplicationMonitoring)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebSphereLibertyApplicationNetworkPolicy) DeepCopyInto(out *WebSphereLibertyApplicationNetworkPolicy) {
	*out = *in
	if in.Disable != nil {
		in, out := &in.Disable, &out.Disable
		*out = new(bool)
		**out = **in
	}
	if in.NamespaceLabels != nil {
		in, out := &in.NamespaceLabels, &out.NamespaceLabels
		*out = new(map[string]string)
		if **in != nil {
			in, out := *in, *out
			*out = make(map[string]string, len(*in))
			for key, val := range *in {
				(*out)[key] = val
			}
		}
	}
	if in.FromLabels != nil {
		in, out := &in.FromLabels, &out.FromLabels
		*out = new(map[string]string)
		if **in != nil {
			in, out := *in, *out
			*out = make(map[string]string, len(*in))
			for key, val := range *in {
				(*out)[key] = val
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebSphereLibertyApplicationNetworkPolicy.
func (in *WebSphereLibertyApplicationNetworkPolicy) DeepCopy() *WebSphereLibertyApplicationNetworkPolicy {
	if in == nil {
		return nil
	}
	out := new(WebSphereLibertyApplicationNetworkPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebSphereLibertyApplicationProbes) DeepCopyInto(out *WebSphereLibertyApplicationProbes) {
	*out = *in
	if in.Liveness != nil {
		in, out := &in.Liveness, &out.Liveness
		*out = new(corev1.Probe)
		(*in).DeepCopyInto(*out)
	}
	if in.Readiness != nil {
		in, out := &in.Readiness, &out.Readiness
		*out = new(corev1.Probe)
		(*in).DeepCopyInto(*out)
	}
	if in.Startup != nil {
		in, out := &in.Startup, &out.Startup
		*out = new(corev1.Probe)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebSphereLibertyApplicationProbes.
func (in *WebSphereLibertyApplicationProbes) DeepCopy() *WebSphereLibertyApplicationProbes {
	if in == nil {
		return nil
	}
	out := new(WebSphereLibertyApplicationProbes)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebSphereLibertyApplicationRoute) DeepCopyInto(out *WebSphereLibertyApplicationRoute) {
	*out = *in
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.CertificateSecretRef != nil {
		in, out := &in.CertificateSecretRef, &out.CertificateSecretRef
		*out = new(string)
		**out = **in
	}
	if in.Termination != nil {
		in, out := &in.Termination, &out.Termination
		*out = new(routev1.TLSTerminationType)
		**out = **in
	}
	if in.InsecureEdgeTerminationPolicy != nil {
		in, out := &in.InsecureEdgeTerminationPolicy, &out.InsecureEdgeTerminationPolicy
		*out = new(routev1.InsecureEdgeTerminationPolicyType)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebSphereLibertyApplicationRoute.
func (in *WebSphereLibertyApplicationRoute) DeepCopy() *WebSphereLibertyApplicationRoute {
	if in == nil {
		return nil
	}
	out := new(WebSphereLibertyApplicationRoute)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebSphereLibertyApplicationSSO) DeepCopyInto(out *WebSphereLibertyApplicationSSO) {
	*out = *in
	if in.OIDC != nil {
		in, out := &in.OIDC, &out.OIDC
		*out = make([]OidcClient, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Oauth2 != nil {
		in, out := &in.Oauth2, &out.Oauth2
		*out = make([]OAuth2Client, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Github != nil {
		in, out := &in.Github, &out.Github
		*out = new(GithubLogin)
		**out = **in
	}
	if in.MapToUserRegistry != nil {
		in, out := &in.MapToUserRegistry, &out.MapToUserRegistry
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebSphereLibertyApplicationSSO.
func (in *WebSphereLibertyApplicationSSO) DeepCopy() *WebSphereLibertyApplicationSSO {
	if in == nil {
		return nil
	}
	out := new(WebSphereLibertyApplicationSSO)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebSphereLibertyApplicationService) DeepCopyInto(out *WebSphereLibertyApplicationService) {
	*out = *in
	if in.Type != nil {
		in, out := &in.Type, &out.Type
		*out = new(corev1.ServiceType)
		**out = **in
	}
	if in.NodePort != nil {
		in, out := &in.NodePort, &out.NodePort
		*out = new(int32)
		**out = **in
	}
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.TargetPort != nil {
		in, out := &in.TargetPort, &out.TargetPort
		*out = new(int32)
		**out = **in
	}
	if in.CertificateSecretRef != nil {
		in, out := &in.CertificateSecretRef, &out.CertificateSecretRef
		*out = new(string)
		**out = **in
	}
	if in.Ports != nil {
		in, out := &in.Ports, &out.Ports
		*out = make([]corev1.ServicePort, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Bindable != nil {
		in, out := &in.Bindable, &out.Bindable
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebSphereLibertyApplicationService.
func (in *WebSphereLibertyApplicationService) DeepCopy() *WebSphereLibertyApplicationService {
	if in == nil {
		return nil
	}
	out := new(WebSphereLibertyApplicationService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebSphereLibertyApplicationServiceability) DeepCopyInto(out *WebSphereLibertyApplicationServiceability) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebSphereLibertyApplicationServiceability.
func (in *WebSphereLibertyApplicationServiceability) DeepCopy() *WebSphereLibertyApplicationServiceability {
	if in == nil {
		return nil
	}
	out := new(WebSphereLibertyApplicationServiceability)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebSphereLibertyApplicationSpec) DeepCopyInto(out *WebSphereLibertyApplicationSpec) {
	*out = *in
	out.License = in.License
	if in.PullPolicy != nil {
		in, out := &in.PullPolicy, &out.PullPolicy
		*out = new(corev1.PullPolicy)
		**out = **in
	}
	if in.PullSecret != nil {
		in, out := &in.PullSecret, &out.PullSecret
		*out = new(string)
		**out = **in
	}
	if in.ServiceAccountName != nil {
		in, out := &in.ServiceAccountName, &out.ServiceAccountName
		*out = new(string)
		**out = **in
	}
	if in.CreateKnativeService != nil {
		in, out := &in.CreateKnativeService, &out.CreateKnativeService
		*out = new(bool)
		**out = **in
	}
	if in.Expose != nil {
		in, out := &in.Expose, &out.Expose
		*out = new(bool)
		**out = **in
	}
	if in.ManageTLS != nil {
		in, out := &in.ManageTLS, &out.ManageTLS
		*out = new(bool)
		**out = **in
	}
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	if in.Autoscaling != nil {
		in, out := &in.Autoscaling, &out.Autoscaling
		*out = new(WebSphereLibertyApplicationAutoScaling)
		(*in).DeepCopyInto(*out)
	}
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = new(corev1.ResourceRequirements)
		(*in).DeepCopyInto(*out)
	}
	if in.Probes != nil {
		in, out := &in.Probes, &out.Probes
		*out = new(WebSphereLibertyApplicationProbes)
		(*in).DeepCopyInto(*out)
	}
	if in.Deployment != nil {
		in, out := &in.Deployment, &out.Deployment
		*out = new(WebSphereLibertyApplicationDeployment)
		(*in).DeepCopyInto(*out)
	}
	if in.StatefulSet != nil {
		in, out := &in.StatefulSet, &out.StatefulSet
		*out = new(WebSphereLibertyApplicationStatefulSet)
		(*in).DeepCopyInto(*out)
	}
	if in.Service != nil {
		in, out := &in.Service, &out.Service
		*out = new(WebSphereLibertyApplicationService)
		(*in).DeepCopyInto(*out)
	}
	if in.Route != nil {
		in, out := &in.Route, &out.Route
		*out = new(WebSphereLibertyApplicationRoute)
		(*in).DeepCopyInto(*out)
	}
	if in.Serviceability != nil {
		in, out := &in.Serviceability, &out.Serviceability
		*out = new(WebSphereLibertyApplicationServiceability)
		**out = **in
	}
	if in.SSO != nil {
		in, out := &in.SSO, &out.SSO
		*out = new(WebSphereLibertyApplicationSSO)
		(*in).DeepCopyInto(*out)
	}
	if in.Monitoring != nil {
		in, out := &in.Monitoring, &out.Monitoring
		*out = new(WebSphereLibertyApplicationMonitoring)
		(*in).DeepCopyInto(*out)
	}
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make([]corev1.EnvVar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.EnvFrom != nil {
		in, out := &in.EnvFrom, &out.EnvFrom
		*out = make([]corev1.EnvFromSource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Volumes != nil {
		in, out := &in.Volumes, &out.Volumes
		*out = make([]corev1.Volume, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.VolumeMounts != nil {
		in, out := &in.VolumeMounts, &out.VolumeMounts
		*out = make([]corev1.VolumeMount, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.InitContainers != nil {
		in, out := &in.InitContainers, &out.InitContainers
		*out = make([]corev1.Container, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.SidecarContainers != nil {
		in, out := &in.SidecarContainers, &out.SidecarContainers
		*out = make([]corev1.Container, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Affinity != nil {
		in, out := &in.Affinity, &out.Affinity
		*out = new(WebSphereLibertyApplicationAffinity)
		(*in).DeepCopyInto(*out)
	}
	if in.SecurityContext != nil {
		in, out := &in.SecurityContext, &out.SecurityContext
		*out = new(corev1.SecurityContext)
		(*in).DeepCopyInto(*out)
	}
	if in.NetworkPolicy != nil {
		in, out := &in.NetworkPolicy, &out.NetworkPolicy
		*out = new(WebSphereLibertyApplicationNetworkPolicy)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebSphereLibertyApplicationSpec.
func (in *WebSphereLibertyApplicationSpec) DeepCopy() *WebSphereLibertyApplicationSpec {
	if in == nil {
		return nil
	}
	out := new(WebSphereLibertyApplicationSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebSphereLibertyApplicationStatefulSet) DeepCopyInto(out *WebSphereLibertyApplicationStatefulSet) {
	*out = *in
	if in.UpdateStrategy != nil {
		in, out := &in.UpdateStrategy, &out.UpdateStrategy
		*out = new(appsv1.StatefulSetUpdateStrategy)
		(*in).DeepCopyInto(*out)
	}
	if in.Storage != nil {
		in, out := &in.Storage, &out.Storage
		*out = new(WebSphereLibertyApplicationStorage)
		(*in).DeepCopyInto(*out)
	}
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebSphereLibertyApplicationStatefulSet.
func (in *WebSphereLibertyApplicationStatefulSet) DeepCopy() *WebSphereLibertyApplicationStatefulSet {
	if in == nil {
		return nil
	}
	out := new(WebSphereLibertyApplicationStatefulSet)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebSphereLibertyApplicationStatus) DeepCopyInto(out *WebSphereLibertyApplicationStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]StatusCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Endpoints != nil {
		in, out := &in.Endpoints, &out.Endpoints
		*out = make([]StatusEndpoint, len(*in))
		copy(*out, *in)
	}
	if in.RouteAvailable != nil {
		in, out := &in.RouteAvailable, &out.RouteAvailable
		*out = new(bool)
		**out = **in
	}
	out.Versions = in.Versions
	if in.Binding != nil {
		in, out := &in.Binding, &out.Binding
		*out = new(corev1.LocalObjectReference)
		**out = **in
	}
	if in.References != nil {
		in, out := &in.References, &out.References
		*out = make(common.StatusReferences, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebSphereLibertyApplicationStatus.
func (in *WebSphereLibertyApplicationStatus) DeepCopy() *WebSphereLibertyApplicationStatus {
	if in == nil {
		return nil
	}
	out := new(WebSphereLibertyApplicationStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebSphereLibertyApplicationStorage) DeepCopyInto(out *WebSphereLibertyApplicationStorage) {
	*out = *in
	if in.VolumeClaimTemplate != nil {
		in, out := &in.VolumeClaimTemplate, &out.VolumeClaimTemplate
		*out = new(corev1.PersistentVolumeClaim)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebSphereLibertyApplicationStorage.
func (in *WebSphereLibertyApplicationStorage) DeepCopy() *WebSphereLibertyApplicationStorage {
	if in == nil {
		return nil
	}
	out := new(WebSphereLibertyApplicationStorage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebSphereLibertyDump) DeepCopyInto(out *WebSphereLibertyDump) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebSphereLibertyDump.
func (in *WebSphereLibertyDump) DeepCopy() *WebSphereLibertyDump {
	if in == nil {
		return nil
	}
	out := new(WebSphereLibertyDump)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *WebSphereLibertyDump) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebSphereLibertyDumpList) DeepCopyInto(out *WebSphereLibertyDumpList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]WebSphereLibertyDump, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebSphereLibertyDumpList.
func (in *WebSphereLibertyDumpList) DeepCopy() *WebSphereLibertyDumpList {
	if in == nil {
		return nil
	}
	out := new(WebSphereLibertyDumpList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *WebSphereLibertyDumpList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebSphereLibertyDumpSpec) DeepCopyInto(out *WebSphereLibertyDumpSpec) {
	*out = *in
	out.License = in.License
	if in.Include != nil {
		in, out := &in.Include, &out.Include
		*out = make([]WebSphereLibertyDumpInclude, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebSphereLibertyDumpSpec.
func (in *WebSphereLibertyDumpSpec) DeepCopy() *WebSphereLibertyDumpSpec {
	if in == nil {
		return nil
	}
	out := new(WebSphereLibertyDumpSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebSphereLibertyDumpStatus) DeepCopyInto(out *WebSphereLibertyDumpStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]OperationStatusCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	out.Versions = in.Versions
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebSphereLibertyDumpStatus.
func (in *WebSphereLibertyDumpStatus) DeepCopy() *WebSphereLibertyDumpStatus {
	if in == nil {
		return nil
	}
	out := new(WebSphereLibertyDumpStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebSphereLibertyTrace) DeepCopyInto(out *WebSphereLibertyTrace) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebSphereLibertyTrace.
func (in *WebSphereLibertyTrace) DeepCopy() *WebSphereLibertyTrace {
	if in == nil {
		return nil
	}
	out := new(WebSphereLibertyTrace)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *WebSphereLibertyTrace) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebSphereLibertyTraceList) DeepCopyInto(out *WebSphereLibertyTraceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]WebSphereLibertyTrace, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebSphereLibertyTraceList.
func (in *WebSphereLibertyTraceList) DeepCopy() *WebSphereLibertyTraceList {
	if in == nil {
		return nil
	}
	out := new(WebSphereLibertyTraceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *WebSphereLibertyTraceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebSphereLibertyTraceSpec) DeepCopyInto(out *WebSphereLibertyTraceSpec) {
	*out = *in
	out.License = in.License
	if in.MaxFileSize != nil {
		in, out := &in.MaxFileSize, &out.MaxFileSize
		*out = new(int32)
		**out = **in
	}
	if in.MaxFiles != nil {
		in, out := &in.MaxFiles, &out.MaxFiles
		*out = new(int32)
		**out = **in
	}
	if in.Disable != nil {
		in, out := &in.Disable, &out.Disable
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebSphereLibertyTraceSpec.
func (in *WebSphereLibertyTraceSpec) DeepCopy() *WebSphereLibertyTraceSpec {
	if in == nil {
		return nil
	}
	out := new(WebSphereLibertyTraceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebSphereLibertyTraceStatus) DeepCopyInto(out *WebSphereLibertyTraceStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]OperationStatusCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	out.OperatedResource = in.OperatedResource
	out.Versions = in.Versions
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebSphereLibertyTraceStatus.
func (in *WebSphereLibertyTraceStatus) DeepCopy() *WebSphereLibertyTraceStatus {
	if in == nil {
		return nil
	}
	out := new(WebSphereLibertyTraceStatus)
	in.DeepCopyInto(out)
	return out
}
