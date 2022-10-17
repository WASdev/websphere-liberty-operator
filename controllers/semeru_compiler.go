package controllers

import (
	wlv1 "github.com/WASdev/websphere-liberty-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)
// Create the Deployment and Service objects for a Semeru Compiler used by a Websphere Liberty Application
func (r *ReconcileWebSphereLiberty) reconcileSemeruCompiler(la *wlv1.WebSphereLibertyApplication) (error, string) {
	compilerMeta := metav1.ObjectMeta{
		Name:      la.GetName() + `-semeru-compiler`,
		Namespace: la.GetNamespace(),
	}
	// Create the Semeru Service object
	semsvc := &corev1.Service{ObjectMeta: compilerMeta}
	err := r.CreateOrUpdate(semsvc, la, func() error {
		createSemeruService(semsvc, la)
		return nil
	})
	if err != nil {
		return err, "Failed to reconcile the Semeru Compiler Service"
	}
	return nil, ""
}

func createSemeruService(svc *corev1.Service, la *wlv1.WebSphereLibertyApplication) {
	var port int32 = 38400
	var timeout int32 = 86400
	svc.Labels = getLabels(la)
	svc.Spec.Selector = getSelectors(la)
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
func getSelectors(la *wlv1.WebSphereLibertyApplication) map[string]string {
	requiredSelector := make(map[string]string)
	requiredSelector["app.kubernetes.io/component"] = `semeru-compiler`
	requiredSelector["app.kubernetes.io/instance"] = la.GetName() + `-semeru-compiler`
	requiredSelector["app.kubernetes.io/part-of"] = la.GetName()
	return requiredSelector
}
// Create the Labels map for a Semeru Compiler
func getLabels(la *wlv1.WebSphereLibertyApplication) map[string]string {
	requiredLabels := make(map[string]string)
	requiredLabels["app.kubernetes.io/name"] = la.GetName() + `-semeru-compiler`
	requiredLabels["app.kubernetes.io/instance"] = la.GetName() + `-semeru-compiler`
	requiredLabels["app.kubernetes.io/managed-by"] = `websphere-liberty-operator`
	requiredLabels["app.kubernetes.io/component"] = `semeru-compiler`
	requiredLabels["app.kubernetes.io/part-of"] = la.GetName()
	return requiredLabels
}
