package controller

import (
	"context"
	"log/slog"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"installer.operators.ethr.gg/installer/api/v1alpha1"
	"installer.operators.ethr.gg/installer/internal/installation"
)

// ManifestReconciler reconciles a Manifest object
type ManifestReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups="*",resources="*",verbs="*"

// +kubebuilder:rbac:groups=installer.operators.ethr.gg,resources=manifests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=installer.operators.ethr.gg,resources=manifests/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=installer.operators.ethr.gg,resources=manifests/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *ManifestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	slog.DebugContext(ctx, "Beginning Reconciliation of Manifest Controller", slog.String("version", os.Getenv("VERSION")), slog.Group("request", slog.String("name", req.Name)))

	var object v1alpha1.Manifest
	if e := r.Get(ctx, req.NamespacedName, &object); e != nil {
		slog.WarnContext(ctx, "Unable to Fetch Manifest Type", slog.String("name", object.Name), slog.String("error", e.Error()))

		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(e)
	}

	// --> evaluate necessity to process custom-resource; check status for zeroth or error value(s)
	if !(object.Status.Processed) || (object.Status.Error != nil) || (object.Status.Total == -1) {
		var installer = installation.New()

		slog.DebugContext(ctx, "Executing Installation(s) on Manifest's Custom-Resource", slog.String("name", object.Name))
		if e := installer.Manifest(ctx, &object); e != nil {
			slog.ErrorContext(ctx, "Unable to Execute Manifest", slog.String("name", object.Name), slog.String("error", e.Error()))

			object.SetError(e) // --> given the installer doesn't update status, the error must be assigned and updated in the controller
			if e := r.Status().Update(ctx, &object); e != nil {
				// --> upon update error, it's superfluous to overwrite the error on the custom-resource; however, it's still important to log
				slog.ErrorContext(ctx, "Unable to Update Manifest Status After Failed Installation", slog.String("name", object.Name), slog.String("error", e.Error()))

				return ctrl.Result{}, e
			}

			return ctrl.Result{}, e
		}
	}

	// --> only if no errors were evaluated can the status be updated
	slog.DebugContext(ctx, "Updating the Manifest's Custom-Resource Status", slog.String("name", object.Name))
	if e := r.Status().Update(ctx, &object); e != nil {
		slog.ErrorContext(ctx, "Unable to Update Kustomize Status After Successful Installation", slog.String("name", object.Name), slog.String("error", e.Error()))

		return ctrl.Result{}, e
	}

	/***
	resource "helm_release" "kyverno" {
	  name       = "kyverno"
	  chart      = "kyverno"
	  repository = "https://kyverno.github.io/kyverno"
	  namespace = "kyverno"

	  version = "3.2.2"

	  create_namespace = true
	  skip_crds = false

	  atomic  = true
	  wait    = true
	  timeout = 900

	  cleanup_on_fail = true

	  dependency_update = true
	  force_update = true
	}
	*/

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ManifestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Manifest{}).
		Complete(r)
}
