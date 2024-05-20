package controller

import (
	"context"
	"log/slog"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"installer.operators.ethr.gg/installer/api/v1alpha1"
)

// KustomizeReconciler reconciles a Kustomize object
type KustomizeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=installer.operators.ethr.gg,resources=kustomizes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=installer.operators.ethr.gg,resources=kustomizes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=installer.operators.ethr.gg,resources=kustomizes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Kustomize object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *KustomizeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	version := os.Getenv("VERSION")

	_ = log.FromContext(ctx)

	slog.DebugContext(ctx, "Beginning Reconciliation of Kustomize Controller", slog.String("version", version), slog.Group("request", slog.String("name", req.Name)))

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KustomizeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Kustomize{}).
		Complete(r)
}
