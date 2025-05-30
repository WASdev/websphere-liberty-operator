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

package controller

import (
	"context"
	"os"
	"time"

	"github.com/WASdev/websphere-liberty-operator/utils"
	oputils "github.com/application-stacks/runtime-component-operator/utils"
	"github.com/go-logr/logr"

	webspherelibertyv1 "github.com/WASdev/websphere-liberty-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ReconcileWebSphereLibertyDump reconciles a WebSphereLibertyDump object
type ReconcileWebSphereLibertyDump struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	Client     client.Client
	Scheme     *runtime.Scheme
	Recorder   record.EventRecorder
	RestConfig *rest.Config
	Log        logr.Logger
}

// +kubebuilder:rbac:groups=liberty.websphere.ibm.com,resources=webspherelibertydumps;webspherelibertydumps/status;webspherelibertydumps/finalizers,verbs=get;list;watch;create;update;patch;delete,namespace=websphere-liberty-operator
// +kubebuilder:rbac:groups=core,resources=pods;pods/exec,verbs=get;list;watch;create;update;patch;delete,namespace=websphere-liberty-operator

// Reconcile reads that state of the cluster for a WebSphereLibertyDump object and makes changes based on the state read
// and what is in the WebsphereLibertyDump.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.

func (r *ReconcileWebSphereLibertyDump) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	reqLogger := r.Log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling WebSphereLibertyDump")

	// Fetch the WebSphereLibertyDump instance
	instance := &webspherelibertyv1.WebSphereLibertyDump{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	//do not reconcile if the dump already started
	oc := webspherelibertyv1.GetOperationCondtion(instance.Status.Conditions, webspherelibertyv1.OperationStatusConditionTypeStarted)
	if oc != nil && oc.Status == corev1.ConditionTrue {
		return reconcile.Result{}, err
	}

	//check if Pod exists and running
	pod := &corev1.Pod{}

	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.PodName, Namespace: request.Namespace}, pod)
	if err != nil || pod.Status.Phase != corev1.PodRunning {
		//handle error
		message := "Failed to find pod " + instance.Spec.PodName + " in namespace " + request.Namespace
		reqLogger.Error(err, message)
		r.Recorder.Event(instance, "Warning", "ProcessingError", message)
		c := webspherelibertyv1.OperationStatusCondition{
			Type:    webspherelibertyv1.OperationStatusConditionTypeStarted,
			Status:  corev1.ConditionFalse,
			Reason:  "Error",
			Message: "Failed to find a pod or pod is not in running state",
		}
		instance.Status.Conditions = webspherelibertyv1.SetOperationCondtion(instance.Status.Conditions, c)
		instance.Status.ObservedGeneration = instance.GetObjectMeta().GetGeneration()
		instance.Status.Versions.Reconciled = utils.OperandVersion
		r.Client.Status().Update(context.TODO(), instance)
		return reconcile.Result{}, nil
	}

	currentTime := time.Now()
	dumpFolder := "/serviceability/" + pod.Namespace + "/" + pod.Name
	dumpFileName := dumpFolder + "/" + "dump_" + currentTime.UTC().Format("2006.01.02_15.04.05") + "_utc.zip"
	dumpCmd := "mkdir -p " + dumpFolder + " &&  server dump --archive=" + dumpFileName
	if len(instance.Spec.Include) > 0 {
		dumpCmd += " --include="
		for i := range instance.Spec.Include {
			dumpCmd += string(instance.Spec.Include[i]) + ","
		}
	}

	c := webspherelibertyv1.OperationStatusCondition{
		Type:   webspherelibertyv1.OperationStatusConditionTypeStarted,
		Status: corev1.ConditionTrue,
	}

	instance.Status.Conditions = webspherelibertyv1.SetOperationCondtion(instance.Status.Conditions, c)
	r.Client.Status().Update(context.TODO(), instance)

	_, err = utils.ExecuteCommandInContainer(r.RestConfig, pod.Name, pod.Namespace, "app", []string{"/bin/sh", "-c", dumpCmd})
	if err != nil {
		//handle error
		reqLogger.Error(err, "Execute dump cmd failed ", "cmd", dumpCmd)
		r.Recorder.Event(instance, "Warning", "ProcessingError", err.Error())
		c = webspherelibertyv1.OperationStatusCondition{
			Type:    webspherelibertyv1.OperationStatusConditionTypeCompleted,
			Status:  corev1.ConditionFalse,
			Reason:  "Error",
			Message: err.Error(),
		}
		instance.Status.Conditions = webspherelibertyv1.SetOperationCondtion(instance.Status.Conditions, c)
		instance.Status.ObservedGeneration = instance.GetObjectMeta().GetGeneration()
		instance.Status.Versions.Reconciled = utils.OperandVersion
		r.Client.Status().Update(context.TODO(), instance)
		return reconcile.Result{}, nil

	}

	c = webspherelibertyv1.OperationStatusCondition{
		Type:   webspherelibertyv1.OperationStatusConditionTypeCompleted,
		Status: corev1.ConditionTrue,
	}

	instance.Status.Conditions = webspherelibertyv1.SetOperationCondtion(instance.Status.Conditions, c)
	instance.Status.DumpFile = dumpFileName
	instance.Status.ObservedGeneration = instance.GetObjectMeta().GetGeneration()
	instance.Status.Versions.Reconciled = utils.OperandVersion
	r.Client.Status().Update(context.TODO(), instance)
	return reconcile.Result{}, nil
}

func (r *ReconcileWebSphereLibertyDump) SetupWithManager(mgr ctrl.Manager) error {

	watchNamespaces, err := oputils.GetWatchNamespaces()
	if err != nil {
		r.Log.Error(err, "Failed to get watch namespace")
		os.Exit(1)
	}

	watchNamespacesMap := make(map[string]bool)
	for _, ns := range watchNamespaces {
		watchNamespacesMap[ns] = true
	}
	isClusterWide := len(watchNamespacesMap) == 1 && watchNamespacesMap[""]

	r.Log.V(1).Info("Adding a new controller", "watchNamespaces", watchNamespaces, "isClusterWide", isClusterWide)

	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration() && (isClusterWide || watchNamespacesMap[e.ObjectOld.GetNamespace()])
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
	return ctrl.NewControllerManagedBy(mgr).For(&webspherelibertyv1.WebSphereLibertyDump{}, builder.WithPredicates(pred)).WithOptions(controller.Options{
		MaxConcurrentReconciles: 1,
	}).Complete(r)

}
