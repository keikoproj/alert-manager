/*
Copyright 2025 Keikoproj authors.

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
	"os"

	"github.com/keikoproj/alert-manager/internal/config"
	"github.com/keikoproj/alert-manager/internal/controllers/common"
	"github.com/keikoproj/alert-manager/pkg/k8s"
	"github.com/keikoproj/alert-manager/pkg/log"
	"github.com/keikoproj/alert-manager/pkg/wavefront"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	wf "github.com/WavefrontHQ/go-wavefront-management-api"
	alertmanagerv1alpha1 "github.com/keikoproj/alert-manager/api/v1alpha1"
	configcommon "github.com/keikoproj/alert-manager/internal/config/common"
	"github.com/keikoproj/alert-manager/internal/controllers"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(alertmanagerv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8082", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	log.New()
	ctx := context.Background()
	log := log.Logger(ctx, "main", "setup")

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress:    metricsAddr,
			SecureServing:  true,
			FilterProvider: filters.WithAuthenticationAndAuthorization,
		},
		WebhookServer:          webhook.NewServer(webhook.Options{Port: 9443}),
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "0cecb213.keikoproj.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	k8sSelfClient := k8s.NewK8sSelfClientDoOrDie()
	recorder := k8sSelfClient.SetUpEventHandler(ctx)

	// Get the config map
	// retrieve k8s secret
	// Call for wavefront new client
	wfTokenSecret, err := k8sSelfClient.GetK8sSecret(ctx, config.Props.WavefrontAPITokenSecretName(), configcommon.AlertManagerNamespaceName)
	if err != nil {
		log.Error(err, "unable to get wavefront api token secret")
		os.Exit(1)
	}
	wfToken, ok := wfTokenSecret.Data[config.Props.WavefrontAPITokenSecretName()]
	if !ok {
		log.Error(err, "unable to get wavefront api token from secret")
		os.Exit(1)
	}
	wavefront.ApiToken = string(wfToken)
	wfClient, wfErr := wavefront.NewClient(ctx, &wf.Config{
		Address: config.Props.WavefrontAPIUrl(),
		Token:   string(wfToken),
	})
	if wfErr != nil {
		log.Error(wfErr, "unable to create wavefront client")
		os.Exit(1)
	}

	if err = (&controllers.WavefrontAlertReconciler{
		Client:          mgr.GetClient(),
		Log:             log.WithValues("controllers", "WavefrontAlert"),
		Scheme:          mgr.GetScheme(),
		Recorder:        recorder,
		WavefrontClient: wfClient,
		CommonClient: &common.Client{
			Client:   mgr.GetClient(),
			Recorder: recorder,
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "WavefrontAlert")
		os.Exit(1)
	}

	if err = (&controllers.AlertsConfigReconciler{
		Client:          mgr.GetClient(),
		Log:             log.WithValues("controllers", "AlertsConfig"),
		Scheme:          mgr.GetScheme(),
		Recorder:        recorder,
		WavefrontClient: wfClient,
		CommonClient: &common.Client{
			Client:   mgr.GetClient(),
			Recorder: recorder,
		},
	}).SetupWithManager(mgr); err != nil {
		log.Error(err, "unable to create controller", "controller", "AlertsConfig")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "problem running manager")
		os.Exit(1)
	}
}
