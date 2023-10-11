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
	"context"
	"fmt"
	"os"
	"strings"

	v1 "k8s.io/api/batch/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/rest"

	"github.com/application-stacks/runtime-component-operator/common"
	"github.com/go-logr/logr"

	"github.com/WASdev/websphere-liberty-operator/utils"
	lutils "github.com/WASdev/websphere-liberty-operator/utils"
	oputils "github.com/application-stacks/runtime-component-operator/utils"

	webspherelibertyv1 "github.com/WASdev/websphere-liberty-operator/api/v1"

	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	imageutil "github.com/openshift/library-go/pkg/image/imageutil"
	"github.com/pkg/errors"
	prometheusv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	OperatorName      = "websphere-liberty-operator"
	OperatorShortName = "wlo"
)

// ReconcileWebSphereLiberty reconciles a WebSphereLibertyApplication object
type ReconcileWebSphereLiberty struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	oputils.ReconcilerBase
	RestConfig      *rest.Config
	Log             logr.Logger
	watchNamespaces []string
}

const applicationFinalizer = "finalizer.webspherelibertyapps.liberty.websphere.ibm.com"

// +kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,resourceNames=restricted,verbs=use,namespace=websphere-liberty-operator
// +kubebuilder:rbac:groups=liberty.websphere.ibm.com,resources=webspherelibertyapplications;webspherelibertyapplications/status;webspherelibertyapplications/finalizers,verbs=get;list;watch;create;update;patch;delete,namespace=websphere-liberty-operator
// +kubebuilder:rbac:groups=apps,resources=deployments;statefulsets,verbs=get;list;watch;create;update;delete,namespace=websphere-liberty-operator
// +kubebuilder:rbac:groups=apps,resources=deployments/finalizers;statefulsets,verbs=update,namespace=websphere-liberty-operator
// +kubebuilder:rbac:groups=core,resources=services;secrets;serviceaccounts;configmaps;persistentvolumeclaims,verbs=get;list;watch;create;update;delete,namespace=websphere-liberty-operator
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;delete,namespace=websphere-liberty-operator
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;rolebindings,verbs=get;list;watch;create;update;delete,namespace=websphere-liberty-operator
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;delete,namespace=websphere-liberty-operator
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses;networkpolicies,verbs=get;list;watch;create;update;delete,namespace=websphere-liberty-operator
// +kubebuilder:rbac:groups=route.openshift.io,resources=routes;routes/custom-host,verbs=get;list;watch;create;update;delete,namespace=websphere-liberty-operator
// +kubebuilder:rbac:groups=image.openshift.io,resources=imagestreams;imagestreamtags,verbs=get;list;watch,namespace=websphere-liberty-operator
// +kubebuilder:rbac:groups=serving.knative.dev,resources=services,verbs=get;list;watch;create;update;delete,namespace=websphere-liberty-operator
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;list;watch;create;update;delete,namespace=websphere-liberty-operator
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificates;issuers,verbs=get;list;watch;create;update;delete,namespace=websphere-liberty-operator

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ReconcileWebSphereLiberty) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	reqLogger := r.Log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconcile WebSphereLibertyApplication - starting")
	ns, err := oputils.GetOperatorNamespace()
	if err != nil {
		reqLogger.Info("Failed to get operator namespace, error: " + err.Error())
	}

	// When running the operator locally, `ns` will be empty string
	if ns == "" {
		// Since this method can be called directly from unit test, populate `watchNamespaces`.
		if r.watchNamespaces == nil {
			r.watchNamespaces, err = oputils.GetWatchNamespaces()
			if err != nil {
				reqLogger.Error(err, "Error getting watch namespace")
				return reconcile.Result{}, err
			}
		}
		// If the operator is running locally, use the first namespace in the `watchNamespaces`
		// `watchNamespaces` must have at least one item
		ns = r.watchNamespaces[0]
	}

	configMap, err := r.GetOpConfigMap(OperatorName, ns)
	if err != nil {
		reqLogger.Info("Failed to get websphere-liberty-operator config map, error: " + err.Error())
		oputils.CreateConfigMap(OperatorName)
	} else {
		common.Config.LoadFromConfigMap(configMap)
	}

	// Fetch the WebSphereLiberty instance
	instance := &webspherelibertyv1.WebSphereLibertyApplication{}
	var ba common.BaseComponent = instance
	err = r.GetClient().Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if kerrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	isKnativeSupported, err := r.IsGroupVersionSupported(servingv1.SchemeGroupVersion.String(), "Service")
	if err != nil {
		r.ManageError(err, common.StatusConditionTypeReconciled, instance)
	} else if !isKnativeSupported && instance.Spec.CreateKnativeService != nil && *instance.Spec.CreateKnativeService {
		reqLogger.V(1).Info(fmt.Sprintf("%s is not supported on the cluster", servingv1.SchemeGroupVersion.String()))
	}

	// Check if there is an existing Deployment, Statefulset or Knative service by this name
	// not managed by this operator
	err = oputils.CheckForNameConflicts("WebSphereLibertyApplication", instance.Name, instance.Namespace, r.GetClient(), request, isKnativeSupported)
	if err != nil {
		return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
	}

	// Check if the WebSphereLibertyApplication instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isInstanceMarkedToBeDeleted := instance.GetDeletionTimestamp() != nil
	if isInstanceMarkedToBeDeleted {
		if lutils.Contains(instance.GetFinalizers(), applicationFinalizer) {
			// Run finalization logic for applicationFinalizer. If the finalization logic fails, don't remove the
			// finalizer so that we can retry during the next reconciliation.
			if err := r.finalizeWebSphereLibertyApplication(reqLogger, instance, instance.Name+"-serviceability", instance.Namespace); err != nil {
				return reconcile.Result{}, err
			}

			// Remove applicationFinalizer. Once all finalizers have been removed, the object will be deleted.
			instance.SetFinalizers(lutils.Remove(instance.GetFinalizers(), applicationFinalizer))
			err := r.GetClient().Update(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
		return reconcile.Result{}, nil
	}

	// Add finalizer for this CR
	if !lutils.Contains(instance.GetFinalizers(), applicationFinalizer) {
		if err := r.addFinalizer(reqLogger, instance); err != nil {
			return reconcile.Result{}, err
		}
	}
	instance.Initialize()

	_, err = oputils.Validate(instance)
	// If there's any validation error, don't bother with requeuing
	if err != nil {
		reqLogger.Error(err, "Error validating WebSphereLibertyApplication")
		r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		return reconcile.Result{}, nil
	}

	_, err = lutils.Validate(instance)
	// If there's any validation error, don't bother with requeuing
	if err != nil {
		reqLogger.Error(err, "Error validating WebSphereLibertyApplication")
		r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		return reconcile.Result{}, nil
	}

	if r.IsOpenShift() {
		// The order of items passed to the MergeMaps matters here! Annotations from GetOpenShiftAnnotations have higher importance. Otherwise,
		// it is not possible to override converted annotations.
		instance.Annotations = oputils.MergeMaps(instance.Annotations, oputils.GetOpenShiftAnnotations(instance))
	}

	err = r.GetClient().Update(context.TODO(), instance)
	if err != nil {
		reqLogger.Error(err, "Error updating WebSphereLibertyApplication")
		return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
	}

	// if currentGen == 1 {
	// 	return reconcile.Result{}, nil
	// }

	defaultMeta := metav1.ObjectMeta{
		Name:      instance.Name,
		Namespace: instance.Namespace,
	}

	imageReferenceOld := instance.Status.ImageReference
	instance.Status.ImageReference = instance.Spec.ApplicationImage
	if r.IsOpenShift() {
		image, err := imageutil.ParseDockerImageReference(instance.Spec.ApplicationImage)
		if err == nil {
			isTag := &imagev1.ImageStreamTag{}
			isTagName := imageutil.JoinImageStreamTag(image.Name, image.Tag)
			isTagNamespace := image.Namespace
			if isTagNamespace == "" {
				isTagNamespace = instance.Namespace
			}
			key := types.NamespacedName{Name: isTagName, Namespace: isTagNamespace}
			err = r.GetAPIReader().Get(context.Background(), key, isTag)
			// Call ManageError only if the error type is not found or is not forbidden. Forbidden could happen
			// when the operator tries to call GET for ImageStreamTags on a namespace that doesn't exists (e.g.
			// cannot get imagestreamtags.image.openshift.io in the namespace "navidsh": no RBAC policy matched)
			if err == nil {
				image := isTag.Image
				if image.DockerImageReference != "" {
					instance.Status.ImageReference = image.DockerImageReference
				}
			} else if err != nil && !kerrors.IsNotFound(err) && !kerrors.IsForbidden(err) && !strings.Contains(isTagName, "/") {
				return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
			}
		}
	}
	if imageReferenceOld != instance.Status.ImageReference {
		// Trigger a new Semeru Cloud Compiler generation
		createNewSemeruGeneration(instance)

		// If the shared LTPA keys was not generated from the last application image, restart the key generation process
		if r.isLTPAKeySharingEnabled(instance) {
			if err := r.restartLTPAKeysGeneration(instance); err != nil {
				reqLogger.Error(err, "Error restarting the LTPA keys generation process")
				return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
			}
		}

		reqLogger.Info("Updating status.imageReference", "status.imageReference", instance.Status.ImageReference)
		err = r.UpdateStatus(instance)
		if err != nil {
			reqLogger.Error(err, "Error updating WebSphere Liberty application status")
			return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}
	}

	if oputils.GetServiceAccountName(instance) == "" {
		serviceAccount := &corev1.ServiceAccount{ObjectMeta: defaultMeta}
		err = r.CreateOrUpdate(serviceAccount, instance, func() error {
			return oputils.CustomizeServiceAccount(serviceAccount, instance, r.GetClient())
		})
		if err != nil {
			reqLogger.Error(err, "Failed to reconcile ServiceAccount")
			return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}
	} else {
		serviceAccount := &corev1.ServiceAccount{ObjectMeta: defaultMeta}
		err = r.DeleteResource(serviceAccount)
		if err != nil {
			reqLogger.Error(err, "Failed to delete ServiceAccount")
			return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}
	}

	// Check if the ServiceAccount has a valid pull secret before creating the deployment/statefulset
	// or setting up knative. Otherwise the pods can go into an ImagePullBackOff loop
	saErr := oputils.ServiceAccountPullSecretExists(instance, r.GetClient())
	if saErr != nil {
		return r.ManageError(saErr, common.StatusConditionTypeReconciled, instance)
	}

	// Check if SemeruCloudCompiler is enabled before reconciling the Semeru Compiler deployment and service.
	// Otherwise, delete the Semeru Compiler deployment and service.
	message := "Start Semeru Compiler reconcile"
	reqLogger.Info(message)
	err, message, areCompletedSemeruInstancesMarkedToBeDeleted := r.reconcileSemeruCompiler(instance)
	if err != nil {
		reqLogger.Error(err, message)
		return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
	}
	// If semeru compiler is enabled, make sure its ready
	if r.isSemeruEnabled(instance) {
		message = "Check Semeru Compiler resources ready"
		reqLogger.Info(message)
		err = r.areSemeruCompilerResourcesReady(instance)
		if err != nil {
			reqLogger.Error(err, message)
			return r.ManageError(err, common.StatusConditionTypeResourcesReady, instance)
		}
	}

	if instance.Spec.CreateKnativeService != nil && *instance.Spec.CreateKnativeService {
		// Clean up non-Knative resources
		resources := []client.Object{
			&corev1.Service{ObjectMeta: defaultMeta},
			&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: instance.Name + "-headless", Namespace: instance.Namespace}},
			&appsv1.Deployment{ObjectMeta: defaultMeta},
			&appsv1.StatefulSet{ObjectMeta: defaultMeta},
			&autoscalingv1.HorizontalPodAutoscaler{ObjectMeta: defaultMeta},
			&networkingv1.NetworkPolicy{ObjectMeta: defaultMeta},
		}
		err = r.DeleteResources(resources)
		if err != nil {
			reqLogger.Error(err, "Failed to clean up non-Knative resources")
			return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}

		if ok, _ := r.IsGroupVersionSupported(networkingv1.SchemeGroupVersion.String(), "Ingress"); ok {
			r.DeleteResource(&networkingv1.Ingress{ObjectMeta: defaultMeta})
		}

		if r.IsOpenShift() {
			route := &routev1.Route{ObjectMeta: defaultMeta}
			err = r.DeleteResource(route)
			if err != nil {
				reqLogger.Error(err, "Failed to clean up non-Knative resource Route")
				return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
			}
		}
		if isKnativeSupported {
			reqLogger.Info("Knative is supported and Knative Service is enabled")
			ksvc := &servingv1.Service{ObjectMeta: defaultMeta}
			err = r.CreateOrUpdate(ksvc, instance, func() error {
				oputils.CustomizeKnativeService(ksvc, instance)
				return nil
			})

			if err != nil {
				reqLogger.Error(err, "Failed to reconcile Knative Service")
				return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
			}

			instance.Status.Versions.Reconciled = lutils.OperandVersion
			reqLogger.Info("Reconcile WebSphereLibertyApplication - completed")
			return r.ManageSuccess(common.StatusConditionTypeReconciled, instance)
		}
		return r.ManageError(errors.New("failed to reconcile Knative service as operator could not find Knative CRDs"), common.StatusConditionTypeReconciled, instance)
	}

	if isKnativeSupported {
		ksvc := &servingv1.Service{ObjectMeta: defaultMeta}
		err = r.DeleteResource(ksvc)
		if err != nil {
			reqLogger.Error(err, "Failed to delete Knative Service")
			r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}
	}

	useCertmanager, err := r.GenerateSvcCertSecret(ba, OperatorShortName, "WebSphere Liberty Operator", OperatorName)
	if err != nil {
		reqLogger.Error(err, "Failed to reconcile CertManager Certificate")
		return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
	}
	if ba.GetService().GetCertificateSecretRef() != nil {
		ba.GetStatus().SetReference(common.StatusReferenceCertSecretName, *ba.GetService().GetCertificateSecretRef())
	}

	svc := &corev1.Service{ObjectMeta: defaultMeta}
	err = r.CreateOrUpdate(svc, instance, func() error {
		oputils.CustomizeService(svc, ba)
		svc.Annotations = oputils.MergeMaps(svc.Annotations, instance.Spec.Service.Annotations)
		if !useCertmanager && r.IsOpenShift() {
			oputils.AddOCPCertAnnotation(ba, svc)
		}
		monitoringEnabledLabelName := getMonitoringEnabledLabelName(ba)
		if instance.Spec.Monitoring != nil {
			svc.Labels[monitoringEnabledLabelName] = "true"
		} else {
			delete(svc.Labels, monitoringEnabledLabelName)
		}
		return nil
	})
	if err != nil {
		reqLogger.Error(err, "Failed to reconcile Service")
		return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
	}

	if (ba.GetManageTLS() == nil || *ba.GetManageTLS()) &&
		ba.GetStatus().GetReferences()[common.StatusReferenceCertSecretName] == "" {
		return r.ManageError(errors.New("Failed to generate TLS certificate. Ensure cert-manager is installed and running"),
			common.StatusConditionTypeReconciled, instance)
	}

	networkPolicy := &networkingv1.NetworkPolicy{ObjectMeta: defaultMeta}
	if np := instance.Spec.NetworkPolicy; np == nil || np != nil && !np.IsDisabled() {
		err = r.CreateOrUpdate(networkPolicy, instance, func() error {
			oputils.CustomizeNetworkPolicy(networkPolicy, r.IsOpenShift(), instance)
			return nil
		})
		if err != nil {
			reqLogger.Error(err, "Failed to reconcile network policy")
			return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}
	} else {
		if err := r.DeleteResource(networkPolicy); err != nil {
			reqLogger.Error(err, "Failed to delete network policy")
			return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}
	}

	if instance.Spec.Serviceability != nil {
		if instance.Spec.Serviceability.VolumeClaimName != "" {
			pvcName := instance.Spec.Serviceability.VolumeClaimName
			err := r.GetClient().Get(context.TODO(), types.NamespacedName{Name: pvcName, Namespace: instance.Namespace}, &corev1.PersistentVolumeClaim{})
			if err != nil && kerrors.IsNotFound(err) {
				reqLogger.Error(err, "Failed to find PersistentVolumeClaim "+pvcName+" in namespace "+instance.Namespace)
				return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
			}
		} else {
			err = r.CreateOrUpdate(lutils.CreateServiceabilityPVC(instance), nil, func() error {
				return nil
			})
			if err != nil {
				reqLogger.Error(err, "Failed to create PersistentVolumeClaim for Serviceability")
				return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
			}
		}
	} else {
		r.deletePVC(reqLogger, instance.Name+"-serviceability", instance.Namespace)
	}

	err = r.ReconcileBindings(instance)
	if err != nil {
		return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
	}

	err, message, ltpaSecretName := r.reconcileLTPAKeysSharing(instance, defaultMeta)
	if err != nil {
		reqLogger.Error(err, message)
		return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
	}

	if instance.Spec.StatefulSet != nil {
		// Delete Deployment if exists
		deploy := &appsv1.Deployment{ObjectMeta: defaultMeta}
		err = r.DeleteResource(deploy)

		if err != nil {
			reqLogger.Error(err, "Failed to delete Deployment")
			return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}
		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: instance.Name + "-headless", Namespace: instance.Namespace}}
		err = r.CreateOrUpdate(svc, instance, func() error {
			oputils.CustomizeService(svc, instance)
			svc.Spec.ClusterIP = corev1.ClusterIPNone
			svc.Spec.Type = corev1.ServiceTypeClusterIP
			return nil
		})
		if err != nil {
			reqLogger.Error(err, "Failed to reconcile headless Service")
			return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}

		statefulSet := &appsv1.StatefulSet{ObjectMeta: defaultMeta}
		err = r.CreateOrUpdate(statefulSet, instance, func() error {
			oputils.CustomizeStatefulSet(statefulSet, instance)
			oputils.CustomizePodSpec(&statefulSet.Spec.Template, instance)
			oputils.CustomizePersistence(statefulSet, instance)
			if err := lutils.CustomizeLibertyEnv(&statefulSet.Spec.Template, instance, r.GetClient()); err != nil {
				reqLogger.Error(err, "Failed to reconcile Liberty env, error: "+err.Error())
				return err
			}

			statefulSet.Spec.Template.Spec.Containers[0].Args = r.getSemeruJavaOptions(instance)

			if err := oputils.CustomizePodWithSVCCertificate(&statefulSet.Spec.Template, instance, r.GetClient()); err != nil {
				return err
			}
			lutils.CustomizeLibertyAnnotations(&statefulSet.Spec.Template, instance)
			lutils.CustomizeLicenseAnnotations(&statefulSet.Spec.Template, instance)
			if instance.Spec.SSO != nil {
				err = lutils.CustomizeEnvSSO(&statefulSet.Spec.Template, instance, r.GetClient(), r.IsOpenShift())
				if err != nil {
					reqLogger.Error(err, "Failed to reconcile Single sign-on configuration")
					return err
				}
			}
			lutils.ConfigureServiceability(&statefulSet.Spec.Template, instance)
			semeruCertVolume := getSemeruCertVolume(instance)
			if r.isSemeruEnabled(instance) && semeruCertVolume != nil {
				statefulSet.Spec.Template.Spec.Volumes = append(statefulSet.Spec.Template.Spec.Volumes, *semeruCertVolume)
				statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts = append(statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts,
					getSemeruCertVolumeMount(instance))
				semeruTLSSecretName := instance.Status.SemeruCompiler.TLSSecretName
				err := lutils.AddSecretResourceVersionAsEnvVar(&statefulSet.Spec.Template, instance, r.GetClient(),
					semeruTLSSecretName, "SEMERU_TLS")
				if err != nil {
					return err
				}
			}

			if r.isLTPAKeySharingEnabled(instance) && len(ltpaSecretName) > 0 {
				lutils.ConfigureLTPA(&statefulSet.Spec.Template, instance, OperatorShortName)
				err := lutils.AddSecretResourceVersionAsEnvVar(&statefulSet.Spec.Template, instance, r.GetClient(), ltpaSecretName, "LTPA")
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			reqLogger.Error(err, "Failed to reconcile StatefulSet")
			return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}

	} else {
		// Delete StatefulSet if exists
		statefulSet := &appsv1.StatefulSet{ObjectMeta: defaultMeta}
		err = r.DeleteResource(statefulSet)
		if err != nil {
			reqLogger.Error(err, "Failed to delete Statefulset")
			return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}

		// Delete StatefulSet if exists
		headlesssvc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: instance.Name + "-headless", Namespace: instance.Namespace}}
		err = r.DeleteResource(headlesssvc)

		if err != nil {
			reqLogger.Error(err, "Failed to delete headless Service")
			return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}
		deploy := &appsv1.Deployment{ObjectMeta: defaultMeta}
		err = r.CreateOrUpdate(deploy, instance, func() error {
			oputils.CustomizeDeployment(deploy, instance)
			oputils.CustomizePodSpec(&deploy.Spec.Template, instance)
			if err := lutils.CustomizeLibertyEnv(&deploy.Spec.Template, instance, r.GetClient()); err != nil {
				reqLogger.Error(err, "Failed to reconcile Liberty env, error: "+err.Error())
				return err
			}
			deploy.Spec.Template.Spec.Containers[0].Args = r.getSemeruJavaOptions(instance)

			if err := oputils.CustomizePodWithSVCCertificate(&deploy.Spec.Template, instance, r.GetClient()); err != nil {
				return err
			}
			lutils.CustomizeLibertyAnnotations(&deploy.Spec.Template, instance)
			lutils.CustomizeLicenseAnnotations(&deploy.Spec.Template, instance)
			if instance.Spec.SSO != nil {
				err = lutils.CustomizeEnvSSO(&deploy.Spec.Template, instance, r.GetClient(), r.IsOpenShift())
				if err != nil {
					reqLogger.Error(err, "Failed to reconcile Single sign-on configuration")
					return err
				}
			}

			lutils.ConfigureServiceability(&deploy.Spec.Template, instance)
			semeruCertVolume := getSemeruCertVolume(instance)
			if r.isSemeruEnabled(instance) && semeruCertVolume != nil {
				deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, *semeruCertVolume)
				deploy.Spec.Template.Spec.Containers[0].VolumeMounts = append(deploy.Spec.Template.Spec.Containers[0].VolumeMounts,
					getSemeruCertVolumeMount(instance))
				semeruTLSSecretName := instance.Status.SemeruCompiler.TLSSecretName
				err := lutils.AddSecretResourceVersionAsEnvVar(&deploy.Spec.Template, instance, r.GetClient(),
					semeruTLSSecretName, "SEMERU_TLS")
				if err != nil {
					return err
				}
			}

			if r.isLTPAKeySharingEnabled(instance) && len(ltpaSecretName) > 0 {
				lutils.ConfigureLTPA(&deploy.Spec.Template, instance, OperatorShortName)
				err := lutils.AddSecretResourceVersionAsEnvVar(&deploy.Spec.Template, instance, r.GetClient(), ltpaSecretName, "LTPA")
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			reqLogger.Error(err, "Failed to reconcile Deployment")
			return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}

	}

	if instance.Spec.Autoscaling != nil {
		hpa := &autoscalingv1.HorizontalPodAutoscaler{ObjectMeta: defaultMeta}
		err = r.CreateOrUpdate(hpa, instance, func() error {
			oputils.CustomizeHPA(hpa, instance)
			return nil
		})

		if err != nil {
			reqLogger.Error(err, "Failed to reconcile HorizontalPodAutoscaler")
			return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}
	} else {
		hpa := &autoscalingv1.HorizontalPodAutoscaler{ObjectMeta: defaultMeta}
		err = r.DeleteResource(hpa)
		if err != nil {
			reqLogger.Error(err, "Failed to delete HorizontalPodAutoscaler")
			return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}
	}

	if ok, err := r.IsGroupVersionSupported(routev1.SchemeGroupVersion.String(), "Route"); err != nil {
		reqLogger.Error(err, fmt.Sprintf("Failed to check if %s is supported", routev1.SchemeGroupVersion.String()))
		r.ManageError(err, common.StatusConditionTypeReconciled, instance)
	} else if ok {
		if instance.Spec.Expose != nil && *instance.Spec.Expose {
			route := &routev1.Route{ObjectMeta: defaultMeta}
			err = r.CreateOrUpdate(route, instance, func() error {
				key, cert, caCert, destCACert, err := r.GetRouteTLSValues(ba)
				if err != nil {
					return err
				}
				oputils.CustomizeRoute(route, instance, key, cert, caCert, destCACert)

				return nil
			})
			if err != nil {
				reqLogger.Error(err, "Failed to reconcile Route")
				return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
			}

		} else {
			route := &routev1.Route{ObjectMeta: defaultMeta}
			err = r.DeleteResource(route)
			if err != nil {
				reqLogger.Error(err, "Failed to delete Route")
				return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
			}
		}
	} else {

		if ok, err := r.IsGroupVersionSupported(networkingv1.SchemeGroupVersion.String(), "Ingress"); err != nil {
			reqLogger.Error(err, fmt.Sprintf("Failed to check if %s is supported", networkingv1.SchemeGroupVersion.String()))
			r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		} else if ok {
			if instance.Spec.Expose != nil && *instance.Spec.Expose {
				ing := &networkingv1.Ingress{ObjectMeta: defaultMeta}
				err = r.CreateOrUpdate(ing, instance, func() error {
					oputils.CustomizeIngress(ing, instance)
					return nil
				})
				if err != nil {
					reqLogger.Error(err, "Failed to reconcile Ingress")
					return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
				}
			} else {
				ing := &networkingv1.Ingress{ObjectMeta: defaultMeta}
				err = r.DeleteResource(ing)
				if err != nil {
					reqLogger.Error(err, "Failed to delete Ingress")
					return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
				}
			}
		}
	}

	if ok, err := r.IsGroupVersionSupported(prometheusv1.SchemeGroupVersion.String(), "ServiceMonitor"); err != nil {
		reqLogger.Error(err, fmt.Sprintf("Failed to check if %s is supported", prometheusv1.SchemeGroupVersion.String()))
		r.ManageError(err, common.StatusConditionTypeReconciled, instance)
	} else if ok {
		if instance.Spec.Monitoring != nil && (instance.Spec.CreateKnativeService == nil || !*instance.Spec.CreateKnativeService) {
			// Validate the monitoring endpoints' configuration before creating/updating the ServiceMonitor
			if err := oputils.ValidatePrometheusMonitoringEndpoints(instance, r.GetClient(), instance.GetNamespace()); err != nil {
				return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
			}
			sm := &prometheusv1.ServiceMonitor{ObjectMeta: defaultMeta}
			err = r.CreateOrUpdate(sm, instance, func() error {
				oputils.CustomizeServiceMonitor(sm, instance)
				return nil
			})
			if err != nil {
				reqLogger.Error(err, "Failed to reconcile ServiceMonitor")
				return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
			}
		} else {
			sm := &prometheusv1.ServiceMonitor{ObjectMeta: defaultMeta}
			err = r.DeleteResource(sm)
			if err != nil {
				reqLogger.Error(err, "Failed to delete ServiceMonitor")
				return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
			}
		}

	} else {
		reqLogger.V(1).Info(fmt.Sprintf("%s is not supported", prometheusv1.SchemeGroupVersion.String()))
	}

	// Delete completed Semeru instances because all pods now point to the newest Semeru service
	if areCompletedSemeruInstancesMarkedToBeDeleted && r.isWebSphereLibertyApplicationReady(instance) {
		if err := r.deleteCompletedSemeruInstances(instance); err != nil {
			reqLogger.Error(err, "Failed to delete completed Semeru instance")
			return r.ManageError(err, common.StatusConditionTypeReconciled, instance)
		}
	}

	instance.Status.Versions.Reconciled = lutils.OperandVersion
	reqLogger.Info("Reconcile WebSphereLibertyApplication - completed")
	return r.ManageSuccess(common.StatusConditionTypeReconciled, instance)
}

func (r *ReconcileWebSphereLiberty) isWebSphereLibertyApplicationReady(ba common.BaseComponent) bool {
	if r.CheckApplicationStatus(ba) == corev1.ConditionTrue {
		statusCondition := ba.GetStatus().GetCondition(common.StatusConditionTypeReady)
		return statusCondition != nil && statusCondition.GetMessage() == common.StatusConditionTypeReadyMessage
	}
	return false
}

func (r *ReconcileWebSphereLiberty) generateLTPAKeys(instance *webspherelibertyv1.WebSphereLibertyApplication, defaultMeta metav1.ObjectMeta) {
	// Create ReadWriteMany persistent volume to share amongst Liberty pods in a single WebSphereLibertyApplication
	ltpaPVC := &corev1.PersistentVolumeClaim{}
	ltpaPVC.Name = instance.GetName() + "-ltpa-token-pvc"
	ltpaPVC.Namespace = instance.GetNamespace()
	r.CreateOrUpdate(ltpaPVC, instance, func() error {
		utils.CustomizeLTPAPersistentVolumeClaim(ltpaPVC, instance)
		return nil
	})

	// Initialize LTPA resources
	ltpaConfigMap := &corev1.ConfigMap{}
	ltpaConfigMap.Name = instance.GetName() + "-ltpa-server-xml"
	ltpaConfigMap.Namespace = instance.GetNamespace()

	ltpaTokenJob := &v1.Job{}
	ltpaTokenJob.Name = instance.GetName() + "-ltpa-token-job"
	ltpaTokenJob.Namespace = instance.GetNamespace()

	deletePropagationBackground := metav1.DeletePropagationBackground

	ltpaSecret := &corev1.Secret{}
	ltpaSecret.Name = instance.GetName() + "-ltpa-secret"
	ltpaSecret.Namespace = instance.GetNamespace()

	// If the LTPA Secret does not exist, generate a new LTPA token
	err := r.GetClient().Get(context.TODO(), types.NamespacedName{Name: ltpaSecret.Name, Namespace: ltpaSecret.Namespace}, ltpaSecret)
	if err != nil && kerrors.IsNotFound(err) {
		// Delete existing Liberty apps because a new LTPA password needs to be loaded into the pod by initContainer
		var libertyApps client.Object
		if instance.Spec.StatefulSet != nil {
			libertyApps = &appsv1.StatefulSet{ObjectMeta: defaultMeta}

		} else {
			libertyApps = &appsv1.Deployment{ObjectMeta: defaultMeta}
		}
		r.DeleteResource(libertyApps)

		// Clear all LTPA-related resources from a prior reconcile
		r.DeleteResource(ltpaConfigMap)
		r.GetClient().Delete(context.TODO(), ltpaTokenJob, &client.DeleteOptions{PropagationPolicy: &deletePropagationBackground})

		ltpaServiceAccount := &corev1.ServiceAccount{}
		ltpaServiceAccount.Name = instance.GetName() + "-ltpa"
		ltpaServiceAccount.Namespace = instance.GetNamespace()

		// Generate LTPA Role/RoleBinding and ServiceAccount
		ltpaRole := &rbacv1.Role{}
		ltpaRole.Name = instance.GetName() + "-ltpa-role"
		ltpaRole.Namespace = instance.GetNamespace()
		ltpaRole.Rules = []rbacv1.PolicyRule{
			{
				Verbs:         []string{"create", "get", "list", "watch"},
				APIGroups:     []string{""},
				Resources:     []string{"secrets"},
				ResourceNames: []string{ltpaSecret.Name},
			},
		}
		r.CreateOrUpdate(ltpaRole, instance, func() error {
			return nil
		})

		ltpaRoleBinding := &rbacv1.RoleBinding{}
		ltpaRoleBinding.Name = instance.GetName() + "-ltpa-rolebinding"
		ltpaRoleBinding.Namespace = instance.GetNamespace()
		ltpaRoleBinding.Subjects = []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      ltpaServiceAccount.Name,
				Namespace: instance.GetNamespace(),
			},
		}
		ltpaRoleBinding.RoleRef = rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     ltpaRole.Name,
		}
		r.CreateOrUpdate(ltpaRoleBinding, instance, func() error {
			return nil
		})
		r.CreateOrUpdate(ltpaServiceAccount, instance, func() error {
			return nil
		})

		// Generate an LTPA token in the shared volume
		r.CreateOrUpdate(ltpaTokenJob, instance, func() error {
			utils.CustomizeLTPAJob(ltpaTokenJob, instance, ltpaSecret.Name, ltpaServiceAccount.Name)
			return nil
		})
	}

	err = r.GetClient().Get(context.TODO(), types.NamespacedName{Name: ltpaSecret.Name, Namespace: ltpaSecret.Namespace}, ltpaSecret)
	if err == nil {
		// Create the server.xml ConfigMap if it doesn't exist
		configMapErr := r.GetClient().Get(context.TODO(), types.NamespacedName{Name: ltpaConfigMap.Name, Namespace: ltpaConfigMap.Namespace}, ltpaConfigMap)
		if configMapErr != nil && kerrors.IsNotFound(configMapErr) {
			// Generate ConfigMap for the server.xml
			ltpaConfigMap := &corev1.ConfigMap{}
			ltpaConfigMap.Name = instance.GetName() + "-ltpa-server-xml"
			ltpaConfigMap.Namespace = instance.GetNamespace()
			r.CreateOrUpdate(ltpaConfigMap, instance, func() error {
				utils.CustomizeLTPAServerXML(ltpaConfigMap, instance, string(ltpaSecret.Data["password"]))
				return nil
			})
			// Cleanup the Job
			r.GetClient().Delete(context.TODO(), ltpaTokenJob, &client.DeleteOptions{PropagationPolicy: &deletePropagationBackground})
		}
	}
}

func (r *ReconcileWebSphereLiberty) SetupWithManager(mgr ctrl.Manager) error {

	mgr.GetFieldIndexer().IndexField(context.Background(), &webspherelibertyv1.WebSphereLibertyApplication{}, indexFieldImageStreamName, func(obj client.Object) []string {
		instance := obj.(*webspherelibertyv1.WebSphereLibertyApplication)
		image, err := imageutil.ParseDockerImageReference(instance.Spec.ApplicationImage)
		if err == nil {
			imageNamespace := image.Namespace
			if imageNamespace == "" {
				imageNamespace = instance.Namespace
			}
			fullName := fmt.Sprintf("%s/%s", imageNamespace, image.Name)
			return []string{fullName}
		}
		return nil
	})

	watchNamespaces, err := oputils.GetWatchNamespaces()
	if err != nil {
		r.Log.Error(err, "Failed to get watch namespace")
		os.Exit(1)
	}

	watchNamespacesMap := make(map[string]bool)
	for _, ns := range watchNamespaces {
		watchNamespacesMap[ns] = true
	}
	isClusterWide := oputils.IsClusterWide(watchNamespaces)

	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration() && (isClusterWide || watchNamespacesMap[e.ObjectNew.GetNamespace()])
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return isClusterWide || watchNamespacesMap[e.Object.GetNamespace()]
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return isClusterWide || watchNamespacesMap[e.Object.GetNamespace()]
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return isClusterWide || watchNamespacesMap[e.Object.GetNamespace()]
		},
	}

	predSubResource := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return (isClusterWide || watchNamespacesMap[e.ObjectOld.GetNamespace()])
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return isClusterWide || watchNamespacesMap[e.Object.GetNamespace()]
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}

	predSubResWithGenCheck := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
			return (isClusterWide || watchNamespacesMap[e.ObjectOld.GetNamespace()]) && e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return isClusterWide || watchNamespacesMap[e.Object.GetNamespace()]
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}

	b := ctrl.NewControllerManagedBy(mgr).For(&webspherelibertyv1.WebSphereLibertyApplication{}, builder.WithPredicates(pred)).
		Owns(&corev1.Service{}, builder.WithPredicates(predSubResource)).
		Owns(&corev1.Secret{}, builder.WithPredicates(predSubResource)).
		Owns(&appsv1.Deployment{}, builder.WithPredicates(predSubResWithGenCheck)).
		Owns(&appsv1.StatefulSet{}, builder.WithPredicates(predSubResWithGenCheck)).
		Owns(&autoscalingv1.HorizontalPodAutoscaler{}, builder.WithPredicates(predSubResource))

	ok, _ := r.IsGroupVersionSupported(routev1.SchemeGroupVersion.String(), "Route")
	if ok {
		b = b.Owns(&routev1.Route{}, builder.WithPredicates(predSubResource))
	}
	ok, _ = r.IsGroupVersionSupported(networkingv1.SchemeGroupVersion.String(), "Ingress")
	if ok {
		b = b.Owns(&networkingv1.Ingress{}, builder.WithPredicates(predSubResource))
	}
	ok, _ = r.IsGroupVersionSupported(servingv1.SchemeGroupVersion.String(), "Service")
	if ok {
		b = b.Owns(&servingv1.Service{}, builder.WithPredicates(predSubResource))
	}
	ok, _ = r.IsGroupVersionSupported(prometheusv1.SchemeGroupVersion.String(), "ServiceMonitor")
	if ok {
		b = b.Owns(&prometheusv1.ServiceMonitor{}, builder.WithPredicates(predSubResource))
	}
	ok, _ = r.IsGroupVersionSupported(imagev1.SchemeGroupVersion.String(), "ImageStream")
	if ok {
		b = b.Watches(&source.Kind{Type: &imagev1.ImageStream{}}, &EnqueueRequestsForCustomIndexField{
			Matcher: &ImageStreamMatcher{
				Klient:          mgr.GetClient(),
				WatchNamespaces: watchNamespaces,
			},
		})
	}
	return b.Complete(r)
}

func getMonitoringEnabledLabelName(ba common.BaseComponent) string {
	return "monitor." + ba.GetGroupName() + "/enabled"
}

func (r *ReconcileWebSphereLiberty) finalizeWebSphereLibertyApplication(reqLogger logr.Logger, wlapp *webspherelibertyv1.WebSphereLibertyApplication, pvcName string, pvcNamespace string) error {
	r.deletePVC(reqLogger, pvcName, pvcNamespace)
	return nil
}

func (r *ReconcileWebSphereLiberty) addFinalizer(reqLogger logr.Logger, wlapp *webspherelibertyv1.WebSphereLibertyApplication) error {
	reqLogger.Info("Adding Finalizer for WebSphereLibertyApplication")
	wlapp.SetFinalizers(append(wlapp.GetFinalizers(), applicationFinalizer))

	// Update CR
	err := r.GetClient().Update(context.TODO(), wlapp)
	if err != nil {
		reqLogger.Error(err, "Failed to update WebSphereLibertyApplication with finalizer")
		return err
	}

	return nil
}

func (r *ReconcileWebSphereLiberty) deletePVC(reqLogger logr.Logger, pvcName string, pvcNamespace string) {
	pvc := &corev1.PersistentVolumeClaim{}
	err := r.GetClient().Get(context.TODO(), types.NamespacedName{Name: pvcName, Namespace: pvcNamespace}, pvc)
	if err == nil {
		if pvc.Status.Phase != "Bound" {
			reqLogger.Info("Deleting dangling PVC that is not in Bound state")
			err = r.DeleteResource(pvc)
			if err != nil {
				reqLogger.Error(err, "Failed to delete dangling PersistentVolumeClaim for Serviceability")
			}
		}
	}
}
