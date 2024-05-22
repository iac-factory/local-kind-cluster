package controller

import (
	"context"
	"log/slog"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	installeroperatorsethrggv1alpha1 "installer.operators.ethr.gg/installer/api/v1alpha1"
)

// HelmReconciler reconciles a Helm object
type HelmReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=installer.operators.ethr.gg,resources=helms,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=installer.operators.ethr.gg,resources=helms/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=installer.operators.ethr.gg,resources=helms/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Helm object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *HelmReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	version := os.Getenv("VERSION")

	_ = log.FromContext(ctx)

	slog.DebugContext(ctx, "Beginning Reconciliation of Helm Controller", slog.String("version", version), slog.Group("request", slog.String("name", req.Name)))

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HelmReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&installeroperatorsethrggv1alpha1.Helm{}).
		Complete(r)
}
