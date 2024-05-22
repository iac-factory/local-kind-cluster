package main

import (
	"context"
	"crypto/tls"
	"flag"
	"log/slog"
	"os"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"installer.operators.ethr.gg/installer/api/v1alpha1"
	installeroperatorsethrggv1alpha1 "installer.operators.ethr.gg/installer/api/v1alpha1"
	"installer.operators.ethr.gg/installer/internal/controller"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

var VERSION string = "development" // production builds have VERSION dynamically linked.

func init() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	slog.Debug("Initialization", slog.Group("variable", slog.String("name", "VERSION"), slog.String("value", VERSION)))
	if e := os.Setenv("VERSION", VERSION); e != nil {
		panic(e)
	}

	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	utilruntime.Must(installeroperatorsethrggv1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	ctx := context.Background()

	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var secureMetrics bool
	var enableHTTP2 bool
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&secureMetrics, "metrics-secure", false,
		"If set the metrics endpoint is served securely")
	flag.BoolVar(&enableHTTP2, "enable-http2", false,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// if the enable-http2 flag is false (the default), http/2 should be disabled
	// due to its vulnerabilities. More specifically, disabling http/2 will
	// prevent from being vulnerable to the HTTP/2 Stream Cancellation and
	// Rapid Reset CVEs. For more information see:
	// - https://github.com/advisories/GHSA-qppj-fm5r-hxr3
	// - https://github.com/advisories/GHSA-4374-p667-p6c8
	disableHTTP2 := func(c *tls.Config) {
		setupLog.Info("disabling http/2")
		c.NextProtos = []string{"http/1.1"}
	}

	tlsOpts := []func(*tls.Config){}
	if !enableHTTP2 {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}

	webhookServer := webhook.NewServer(webhook.Options{
		TLSOpts: tlsOpts,
	})

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress:   ":8080",
			SecureServing: secureMetrics,
			TLSOpts:       tlsOpts,
		},
		WebhookServer:                 webhookServer,
		HealthProbeBindAddress:        probeAddr,
		LeaderElection:                enableLeaderElection,
		LeaderElectionID:              "3614bd78.analytics.operators.ethr.gg",
		LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controller.HelmReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Report")
		os.Exit(1)
	}
	if err = (&controller.KustomizeReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Kustomize")
		os.Exit(1)
	}
	if err = (&controller.ManifestReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Manifest")
		os.Exit(1)
	}
	if err = (&controller.GitOpsReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GitOps")
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

	var information = manager.RunnableFunc(func(ctx context.Context) error {
		slog.InfoContext(ctx, "Attempting to Establish Authenticated Session", slog.String("version", VERSION))

		configuration := mgr.GetConfig()
		clientset, e := kubernetes.NewForConfig(configuration)
		if e != nil {
			return e
		}

		pods, e := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
		if e != nil {
			return e
		}

		slog.InfoContext(ctx, "Total Pods on Cluster", slog.Int("value", len(pods.Items)))

		return nil
	})

	var initialize = manager.RunnableFunc(func(ctx context.Context) error {
		{
			t := "Manifest"
			name := "flux"
			namespace := "installer-system"

			url := "https://github.com/fluxcd/flux2/releases/latest/download/install.yaml"

			object := &v1alpha1.Manifest{}

			if e := mgr.GetClient().Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, object); e != nil && errors.IsNotFound(e) {
				slog.InfoContext(ctx, "Executing Operator First-Time Setup", slog.String("version", VERSION), slog.String("type", t), slog.String("name", name), slog.String("namespace", namespace))

				object.SetURL(url)
				object.SetName(name)
				object.SetType(v1alpha1.Standard)
				object.SetNamespace(namespace)

				if e := mgr.GetClient().Create(ctx, object); e != nil {
					slog.ErrorContext(ctx, "Unexpected Error While Attempting to Create ", slog.String("version", VERSION), slog.String("type", t), slog.String("name", name), slog.String("namespace", namespace), slog.String("error", e.Error()))

					return e
				}

				slog.InfoContext(ctx, "Successfully Established an Operator Setup CR", slog.String("version", VERSION), slog.String("type", t), slog.String("name", name), slog.String("namespace", namespace))
			} else if e != nil {
				slog.ErrorContext(ctx, "Error While Attempting to Retrieve Primary Installer", slog.String("version", VERSION), slog.String("Type", t), slog.String("error", e.Error()))

				return e
			}
		}

		return nil
	})

	_ = information
	// if e := mgr.Add(information); e != nil {
	// 	slog.ErrorContext(ctx, "Unable to Add Information Callable To Operator", slog.String("error", e.Error()))
	//
	// 	panic(e)
	// }

	if e := mgr.Add(initialize); e != nil {
		slog.ErrorContext(ctx, "Unable to Add Initialization Callable To Operator", slog.String("error", e.Error()))

		panic(e)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
