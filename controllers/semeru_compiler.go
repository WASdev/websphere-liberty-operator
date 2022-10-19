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
	wlv1 "github.com/WASdev/websphere-liberty-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	SemeruLabelNameSuffix = "-semeru-compiler"
	SemeruLabelName       = "semeru-compiler"
	JitServer             = "jitserver"
)

// Create the Deployment and Service objects for a Semeru Compiler used by a Websphere Liberty Application
func (r *ReconcileWebSphereLiberty) reconcileSemeruCompiler(la *wlv1.WebSphereLibertyApplication) (error, string) {
	compilerMeta := metav1.ObjectMeta{
		Name:      la.GetName() + SemeruLabelNameSuffix,
		Namespace: la.GetNamespace(),
	}
	//Deployment
	semeruDeployment := &appsv1.Deployment{ObjectMeta: compilerMeta}
	err := r.CreateOrUpdate(semeruDeployment, la, func() error {
		r.reconcileSemeruDeployment(la, semeruDeployment)
		return nil
	})
	if err != nil {
		return err, "Failed to reconcile Deployment : " + semeruDeployment.Name
	}
	// Create the Semeru Service object
	semsvc := &corev1.Service{ObjectMeta: compilerMeta}
	err = r.CreateOrUpdate(semsvc, la, func() error {
		reconcileSemeruService(semsvc, la)
		return nil
	})
	if err != nil {
		return err, "Failed to reconcile the Semeru Compiler Service"
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

	deploy.Spec.Template = corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:   wlva.Name + SemeruLabelNameSuffix,
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
							corev1.ResourceMemory: resource.MustParse("1200Mi"),
							corev1.ResourceCPU:    resource.MustParse("1000m"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("1200Mi"),
							corev1.ResourceCPU:    resource.MustParse("8000m"),
						},
					},
					Env: []corev1.EnvVar{
						{Name: "OPENJ9_JAVA_OPTIONS", Value: "-XX:+JITServerLogConnections"},
					},
				},
			},
		},
	}
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
