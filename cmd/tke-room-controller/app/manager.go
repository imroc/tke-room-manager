package app

import (
	"os"

	"github.com/spf13/viper"

	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	gamev1alpha1 "github.com/imroc/tke-room-manager/api/v1alpha1"
	"github.com/imroc/tke-room-manager/pkg/manager"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"github.com/imroc/tke-room-manager/internal/controller"
	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(gamev1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func runManager() {
	metricsAddr := viper.GetString(metricsBindAddress)
	probeAddr := viper.GetString(healthProbeBindAddress)
	enableLeaderElection := viper.GetBool(leaderElect)

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(zapOptions)))

	mgr, err := ctrl.NewManager(
		ctrl.GetConfigOrDie(),
		manager.GetOptions(scheme, metricsAddr, probeAddr, enableLeaderElection),
	)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	go func() {
		if err := runApiServer(mgr); err != nil {
			setupLog.Error(err, "unable to start API server")
			os.Exit(1)
		}
	}()

	if err = (&controller.RoomReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Room")
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

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
