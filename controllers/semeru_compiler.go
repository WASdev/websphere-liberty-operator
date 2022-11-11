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

package controllers

import (
	"fmt"
	"strings"
	"time"

	wlv1 "github.com/WASdev/websphere-liberty-operator/api/v1"
	wlutils "github.com/WASdev/websphere-liberty-operator/utils"
	"github.com/application-stacks/runtime-component-operator/common"
	utils "github.com/application-stacks/runtime-component-operator/utils"
	certmanagerv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	certmanagermetav1 "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	SemeruLabelNameSuffix = "-semeru-compiler"
	SemeruLabelName       = "semeru-compiler"
	JitServer             = "jitserver"
)

// Create the Deployment and Service objects for a Semeru Compiler used by a Websphere Liberty Application
func (r *ReconcileWebSphereLiberty) reconcileSemeruCompiler(wlva *wlv1.WebSphereLibertyApplication) (error, string) {

	compilerMeta := metav1.ObjectMeta{
		Name:      wlva.GetName() + SemeruLabelNameSuffix,
		Namespace: wlva.GetNamespace(),
	}

	semeruCloudCompiler := wlva.GetSemeruCloudCompiler()
	if semeruCloudCompiler == nil {
		semsvc := &corev1.Service{ObjectMeta: compilerMeta}
		semeruDeployment := &appsv1.Deployment{ObjectMeta: compilerMeta}
		err := r.DeleteResources([]client.Object{semsvc, semeruDeployment})
		if err != nil {
			return err, "Failed to delete Semeru Compiler resources"
		}
		wlva.Status.SemeruCompiler = nil
		return nil, ""
	}

	// Create the Semeru Service object
	semsvc := &corev1.Service{ObjectMeta: compilerMeta}
	tlsSecretName := ""
	err := r.CreateOrUpdate(semsvc, wlva, func() error {
		reconcileSemeruService(semsvc, wlva)
		if r.IsOpenShift() {
			if _, ok := semsvc.Annotations["service.beta.openshift.io/serving-cert-secret-name"]; !ok {
				if _, ok = semsvc.Annotations["service.alpha.openshift.io/serving-cert-secret-name"]; !ok {
					if semsvc.Annotations == nil {
						semsvc.Annotations = map[string]string{}
					}
					tlsSecretName = semsvc.GetName() + "-tls-ocp"
					semsvc.Annotations["service.beta.openshift.io/serving-cert-secret-name"] = tlsSecretName
					if wlva.Status.SemeruCompiler == nil {
						wlva.Status.SemeruCompiler = &wlv1.SemeruCompilerStatus{}
					}
					wlva.Status.SemeruCompiler.TLSSecretName = tlsSecretName
				}
			}
		}
		return nil
	})
	if err != nil {
		return err, "Failed to reconcile the Semeru Compiler Service"
	}

	//create certmanager issuer and certificate if necessary
	if !r.IsOpenShift() {
		err = r.GenerateCMIssuer(wlva.Namespace, "wlo", "WebSphere Liberty Operator", "websphere-liberty-operator")
		if err != nil {
			return err, "Failed to reconcile Certificate Issuer"
		}
		err = r.reconcileSemeruCMCertificate(wlva)
		if err != nil {
			return err, "Failed to reconcile Semeru Compiler Certificate"
		}
	}

	//Deployment
	semeruDeployment := &appsv1.Deployment{ObjectMeta: compilerMeta}
	err = r.CreateOrUpdate(semeruDeployment, wlva, func() error {
		r.reconcileSemeruDeployment(wlva, semeruDeployment)
		return nil
	})
	if err != nil {
		return err, "Failed to reconcile Deployment : " + semeruDeployment.Name
	}
	return nil, ""
}

func (r *ReconcileWebSphereLiberty) reconcileSemeruDeployment(wlva *wlv1.WebSphereLibertyApplication, deploy *appsv1.Deployment) {
	deploy.Labels = getLabels(wlva)
	deploy.Spec.Strategy.Type = appsv1.RecreateDeploymentStrategyType
	replicas := int32(1)
	deploy.Spec.Replicas = &replicas

	if deploy.Spec.Selector == nil {
		deploy.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: getSelectors(wlva),
		}
	}
	// Get Semeru resources config
	semeruCloudCompiler := wlva.GetSemeruCloudCompiler()
	instanceResources := semeruCloudCompiler.Resources

	requestsMemory := getQuantityFromRequestsOrDefault(instanceResources, corev1.ResourceMemory, "1200Mi")
	requestsCPU := getQuantityFromRequestsOrDefault(instanceResources, corev1.ResourceCPU, "1000m")
	limitsMemory := getQuantityFromLimitsOrDefault(instanceResources, corev1.ResourceMemory, "1200Mi")
	limitsCPU := getQuantityFromLimitsOrDefault(instanceResources, corev1.ResourceCPU, "8000m")

	// Liveness probe
	livenessProbe := corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromInt(38400),
			},
		},
		InitialDelaySeconds: 10,
		PeriodSeconds:       10,
	}

	// Readiness probe
	readinessProbe := corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromInt(38400),
			},
		},
		InitialDelaySeconds: 5,
		PeriodSeconds:       5,
	}
	// Get Semeru resources config

	deploy.Spec.Template = corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: getLabels(wlva),
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            JitServer,
					Image:           wlva.Status.GetImageReference(),
					ImagePullPolicy: *wlva.GetPullPolicy(),
					Command:         []string{"jitserver"},
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: 38400,
							Protocol:      corev1.ProtocolTCP,
						},
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceMemory: requestsMemory,
							corev1.ResourceCPU:    requestsCPU,
						},
						Limits: corev1.ResourceList{
							corev1.ResourceMemory: limitsMemory,
							corev1.ResourceCPU:    limitsCPU,
						},
					},
					Env: []corev1.EnvVar{
						{Name: "OPENJ9_JAVA_OPTIONS", Value: "-XX:+JITServerLogConnections" +
							" -XX:JITServerSSLKey=/etc/x509/certs/tls.key" +
							" -XX:JITServerSSLCert=/etc/x509/certs/tls.crt"},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "certs",
							ReadOnly:  true,
							MountPath: "/etc/x509/certs",
						},
					},
					LivenessProbe:  &livenessProbe,
					ReadinessProbe: &readinessProbe,
				},
			},
			Volumes: []corev1.Volume{{
				Name: "certs",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: wlva.Status.SemeruCompiler.TLSSecretName,
					},
				},
			}},
		},
	}

	// Copy the service account from the WebSphereLibertyApplcation CR
	if wlva.GetServiceAccountName() != nil && *wlva.GetServiceAccountName() != "" {
		deploy.Spec.Template.Spec.ServiceAccountName = *wlva.GetServiceAccountName()
	} else {
		deploy.Spec.Template.Spec.ServiceAccountName = wlva.GetName()
	}

	// This ensures that the semeru pod(s) are updated if the service account is updated
	saRV := wlva.GetStatus().GetReferences()[common.StatusReferenceSAResourceVersion]
	if saRV != "" {
		deploy.Spec.Template.Spec.Containers[0].Env = append(deploy.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: "SA_RESOURCE_VERSION", Value: saRV})
	}

	// Copy the securityContext from the WebSphereLibertyApplcation CR
	deploy.Spec.Template.Spec.Containers[0].SecurityContext = utils.GetSecurityContext(wlva)

	wlutils.AddSecretResourceVersionAsEnvVar(&deploy.Spec.Template, wlva, r.GetClient(), wlva.Status.SemeruCompiler.TLSSecretName, "TLS")
}

func reconcileSemeruService(svc *corev1.Service, wlva *wlv1.WebSphereLibertyApplication) {
	var port int32 = 38400
	var timeout int32 = 86400
	svc.Labels = getLabels(wlva)
	svc.Spec.Selector = getSelectors(wlva)
	if len(svc.Spec.Ports) == 0 {
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{})
	}
	svc.Spec.Ports[0].Protocol = corev1.ProtocolTCP
	svc.Spec.Ports[0].Port = port
	svc.Spec.Ports[0].TargetPort = intstr.FromInt(int(port))
	svc.Spec.SessionAffinity = corev1.ServiceAffinityClientIP
	svc.Spec.SessionAffinityConfig = &corev1.SessionAffinityConfig{
		ClientIP: &corev1.ClientIPConfig{
			TimeoutSeconds: &timeout,
		},
	}

	if wlva.Status.SemeruCompiler == nil {
		wlva.Status.SemeruCompiler = &wlv1.SemeruCompilerStatus{}
	}
	wlva.Status.SemeruCompiler.ServiceHostname = svc.GetName() + "." + svc.GetNamespace() + ".svc"

}

func (r *ReconcileWebSphereLiberty) reconcileSemeruCMCertificate(wlva *wlv1.WebSphereLibertyApplication) error {
	svcCert := &certmanagerv1.Certificate{}
	svcCert.Name = wlva.GetName() + SemeruLabelNameSuffix
	svcCert.Namespace = wlva.GetNamespace()

	err := r.CreateOrUpdate(svcCert, wlva, func() error {
		svcCert.Labels = wlva.GetLabels()
		svcCert.Spec.IssuerRef = certmanagermetav1.ObjectReference{
			Name: "wlo-ca-issuer",
		}
		svcCert.Spec.SecretName = wlva.GetName() + SemeruLabelNameSuffix + "-tls-cm"
		svcCert.Spec.DNSNames = make([]string, 2)
		svcCert.Spec.DNSNames[0] = wlva.GetName() + SemeruLabelNameSuffix + "." + wlva.Namespace + ".svc"
		svcCert.Spec.DNSNames[1] = wlva.GetName() + SemeruLabelNameSuffix + "." + wlva.Namespace + ".svc.cluster.local"
		svcCert.Spec.CommonName = svcCert.Spec.DNSNames[0]
		duration, err := time.ParseDuration(common.Config[common.OpConfigCMCertDuration])
		if err != nil {
			return err
		}
		svcCert.Spec.Duration = &metav1.Duration{Duration: duration}
		return nil
	})
	if err != nil {
		return err
	}
	if wlva.Status.SemeruCompiler == nil {
		wlva.Status.SemeruCompiler = &wlv1.SemeruCompilerStatus{}
	}
	wlva.Status.SemeruCompiler.TLSSecretName = svcCert.Spec.SecretName
	return nil
}

// Create the Selector map for a Semeru Compiler
func getSelectors(wlva *wlv1.WebSphereLibertyApplication) map[string]string {
	requiredSelector := make(map[string]string)
	requiredSelector["app.kubernetes.io/component"] = SemeruLabelName
	requiredSelector["app.kubernetes.io/instance"] = wlva.GetName() + SemeruLabelNameSuffix
	requiredSelector["app.kubernetes.io/part-of"] = wlva.GetName()
	return requiredSelector
}

// Create the Labels map for a Semeru Compiler
func getLabels(wlva *wlv1.WebSphereLibertyApplication) map[string]string {
	requiredLabels := make(map[string]string)
	requiredLabels["app.kubernetes.io/name"] = wlva.GetName() + SemeruLabelNameSuffix
	requiredLabels["app.kubernetes.io/instance"] = wlva.GetName() + SemeruLabelNameSuffix
	requiredLabels["app.kubernetes.io/managed-by"] = OperatorName
	requiredLabels["app.kubernetes.io/component"] = SemeruLabelName
	requiredLabels["app.kubernetes.io/part-of"] = wlva.GetName()
	return requiredLabels
}

// Returns quantity at resourceRequirements.Requests[resourceName] if it exists, otherwise return the parsed defaultQuantity
func getQuantityFromRequestsOrDefault(resourceRequirements *corev1.ResourceRequirements, resourceName corev1.ResourceName, defaultQuantity string) resource.Quantity {
	if resourceRequirements != nil && resourceRequirements.Requests != nil {
		if mapValue, ok := resourceRequirements.Requests[resourceName]; ok {
			return mapValue
		}
	}
	return resource.MustParse(defaultQuantity)
}

// Returns quantity at resourceRequirements.Limits[resourceName] if it exists, otherwise return the parsed defaultQuantity
func getQuantityFromLimitsOrDefault(resourceRequirements *corev1.ResourceRequirements, resourceName corev1.ResourceName, defaultQuantity string) resource.Quantity {
	if resourceRequirements != nil && resourceRequirements.Limits != nil {
		if mapValue, ok := resourceRequirements.Limits[resourceName]; ok {
			return mapValue
		}
	}
	return resource.MustParse(defaultQuantity)
}

func getSemeruCertVolumeMount(wlva *wlv1.WebSphereLibertyApplication) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      "semeru-certs",
		MountPath: "/etc/x509/semeru-certs",
		ReadOnly:  true,
	}
}
func getSemeruCertVolume(wlva *wlv1.WebSphereLibertyApplication) *corev1.Volume {
	if wlva.Status.SemeruCompiler == nil || wlva.Status.SemeruCompiler.TLSSecretName == "" ||
		strings.HasSuffix(wlva.Status.SemeruCompiler.TLSSecretName, "-ocp") {
		return nil
	}
	return &corev1.Volume{
		Name: "semeru-certs",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: wlva.Status.SemeruCompiler.TLSSecretName,
			},
		},
	}
}

func getSemeruJavaOptions(instance *wlv1.WebSphereLibertyApplication) []string {
	if instance.GetSemeruCloudCompiler() != nil {
		certificateLocation := "/etc/x509/semeru-certs/ca.crt"
		if instance.Status.SemeruCompiler != nil && strings.HasSuffix(instance.Status.SemeruCompiler.TLSSecretName, "-ocp") {
			certificateLocation = "/var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt"
		}
		jitServerAddress := instance.Status.SemeruCompiler.ServiceHostname
		jitSeverOptions := fmt.Sprintf("-XX:+UseJITServer -XX:+JITServerLogConnections -XX:JITServerAddress=%v -XX:JITServerSSLRootCerts=%v",
			jitServerAddress, certificateLocation)

		args := []string{
			"/bin/bash",
			"-c",
			"export OPENJ9_JAVA_OPTIONS=\"$OPENJ9_JAVA_OPTIONS " +
				jitSeverOptions +
				"\" && server run",
		}
		return args
	}
	return nil
}
