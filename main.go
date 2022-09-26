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

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	odlmv1alpha1 "github.com/IBM/operand-deployment-lifecycle-manager/api/v1alpha1"
	webspherelibertyv1 "github.com/WASdev/websphere-liberty-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"

	"github.com/WASdev/websphere-liberty-operator/controllers"
	lutils "github.com/WASdev/websphere-liberty-operator/utils"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	clientcfg "sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/application-stacks/runtime-component-operator/common"
	oputils "github.com/application-stacks/runtime-component-operator/utils"
	certmanagerv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	prometheusv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(webspherelibertyv1.AddToScheme(scheme))

	utilruntime.Must(routev1.AddToScheme(scheme))

	utilruntime.Must(prometheusv1.AddToScheme(scheme))

	utilruntime.Must(imagev1.AddToScheme(scheme))

	utilruntime.Must(servingv1.AddToScheme(scheme))

	utilruntime.Must(certmanagerv1.AddToScheme(scheme))

	utilruntime.Must(odlmv1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	// see https://github.com/operator-framework/operator-sdk/issues/1813
	leaseDuration := 30 * time.Second
	renewDeadline := 20 * time.Second

	watchNamespace, err := getWatchNamespace()
	if err != nil {
		setupLog.Error(err, "unable to get WatchNamespace, "+
			"the manager will watch and manage resources in all Namespaces")
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "7111f50b.websphere.ibm.com",
		LeaseDuration:          &leaseDuration,
		RenewDeadline:          &renewDeadline,
		Namespace:              watchNamespace,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.ReconcileWebSphereLiberty{
		ReconcilerBase: oputils.NewReconcilerBase(mgr.GetAPIReader(), mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("websphere-liberty-operator")),
		Log:            ctrl.Log.WithName("controllers").WithName("WebSphereLibertyApplication"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "WebSphereLibertyApplication")
		os.Exit(1)
	}
	if err = (&controllers.ReconcileWebSphereLibertyDump{
		Log:        ctrl.Log.WithName("controllers").WithName("WebSphereLibertyDump"),
		Client:     mgr.GetClient(),
		Scheme:     mgr.GetScheme(),
		RestConfig: mgr.GetConfig(),
		Recorder:   mgr.GetEventRecorderFor("websphere-liberty-operator"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "WebSphereLibertyDump")
		os.Exit(1)
	}
	if err = (&controllers.ReconcileWebSphereLibertyTrace{
		Log:        ctrl.Log.WithName("controllers").WithName("WebSphereLibertyTrace"),
		Client:     mgr.GetClient(),
		Scheme:     mgr.GetScheme(),
		RestConfig: mgr.GetConfig(),
		Recorder:   mgr.GetEventRecorderFor("websphere-liberty-operator"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "WebSphereLibertyTrace")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	// The helper functions below are called before the manager is started, so the normal client isn't setup properly
	client, clerr := client.New(clientcfg.GetConfigOrDie(), client.Options{Scheme: scheme})
	if clerr != nil {
		setupLog.Error(clerr, "Couldn't create a new client")
		return
	}

	operatorNamespace, _ := oputils.GetOperatorNamespace()
	if operatorNamespace != "" {
		configMapName := controllers.OperatorName
		createWebSphereLibertyConfigMap(client, operatorNamespace, configMapName)

		// Install OperandRequest if CRD exists
		operandRequestErrorMessage := "Couldn't find the OperandRequest custom resource, verify that IBM Common Services is installed then reinstall the operator."
		if ok, err := isGroupVersionSupported(mgr.GetConfig(), odlmv1alpha1.SchemeBuilder.GroupVersion.String(), "OperandRequest"); err != nil {
			setupLog.Error(err, operandRequestErrorMessage)
		} else if !ok {
			setupLog.Info(operandRequestErrorMessage)
		} else {
			createOperandRequest(client, operatorNamespace, configMapName)
		}
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

// getWatchNamespace returns the Namespace the operator should be watching for changes
func getWatchNamespace() (string, error) {
	// WatchNamespaceEnvVar is the constant for env variable WATCH_NAMESPACE
	// which specifies the Namespace to watch.
	// An empty value means the operator is running with cluster scope.
	var watchNamespaceEnvVar = "WATCH_NAMESPACE"

	ns, found := os.LookupEnv(watchNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", watchNamespaceEnvVar)
	}
	return ns, nil
}

func createWebSphereLibertyConfigMap(client client.Client, operatorNamespace string, mapName string) {
	configMap := &corev1.ConfigMap{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: mapName, Namespace: operatorNamespace}, configMap)
	if err != nil {
		setupLog.Error(err, "The operator config map was not found. Attempting to create it")
	} else {
		setupLog.Info("Existing operator config map was found")
		// Update the config map to support WL specific key-value pairs, if missing
		lutils.LoadFromWebSphereLibertyConfigMap(&common.Config, configMap)
		return
	}

	newConfigMap := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: mapName, Namespace: operatorNamespace}}
	// The config map doesn't exist, so need to initialize the default config data, and then
	// store it in a new map
	common.Config = lutils.DefaultOpConfig()
	_, cerr := controllerutil.CreateOrUpdate(context.TODO(), client, newConfigMap, func() error {
		newConfigMap.Data = common.Config
		return nil
	})
	if cerr != nil {
		setupLog.Error(cerr, "Couldn't create config map in namespace "+operatorNamespace)
	} else {
		setupLog.Info("Operator config map created in namespace " + operatorNamespace)
	}
}

func createOperandRequest(client client.Client, requestNamespace string, mapName string) {
	var operands []odlmv1alpha1.Operand
	operands = append(operands, odlmv1alpha1.Operand{Name: "ibm-licensing-operator"})

	// Generate operand request instance
	var requestInstance *odlmv1alpha1.OperandRequest
	requestInstance = &odlmv1alpha1.OperandRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mapName,
			Namespace: requestNamespace,
			Labels: map[string]string{
				"app.kubernetes.io/instance":   "operand-deployment-lifecycle-manager",
				"app.kubernetes.io/managed-by": "operand-deployment-lifecycle-manager",
				"app.kubernetes.io/name":       "operand-deployment-lifecycle-manager",
			},
		},
		Spec: odlmv1alpha1.OperandRequestSpec{
			Requests: []odlmv1alpha1.Request{
				{
					Registry:          common.Config[lutils.OpConfigLicenseServiceRegistry],
					RegistryNamespace: common.Config[lutils.OpConfigLicenseServiceRegistryNamespace],
					Operands:          operands,
				},
			},
		},
	}

	// Create operand request
	_, cerr := controllerutil.CreateOrUpdate(context.TODO(), client, requestInstance, func() error {
		return nil
	})
	if cerr != nil {
		setupLog.Error(cerr, "Couldn't create operand request in namespace "+requestNamespace)
	} else {
		setupLog.Info("Operand request created in namespace " + requestNamespace)
	}
}

func isGroupVersionSupported(restConfig *rest.Config, groupVersion string, kind string) (bool, error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return false, err
	}

	res, err := discoveryClient.ServerResourcesForGroupVersion(groupVersion)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}

		return false, err
	}

	for _, v := range res.APIResources {
		if v.Kind == kind {
			return true, nil
		}
	}
	return false, nil
}
