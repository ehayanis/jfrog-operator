package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	cagipv1 "github.com/cagip/jfrog-operator/api/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ProjectReconciler reconciles a Project object
type ProjectReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cagip.github.com,resources=projects,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cagip.github.com,resources=projects/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cagip.github.com,resources=projects/finalizers,verbs=update

func (r *ProjectReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var project cagipv1.Project
	if err := r.Get(ctx, req.NamespacedName, &project); err != nil {
		if errors.IsNotFound(err) {
			log.Info("Project resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get Project")
		return ctrl.Result{}, err
	}

	// Prepare payload
	payload := map[string]string{
		"project":  project.Spec.Project,
		"entity":   project.Spec.Tenant,
		"techno":   "docker",
		"location": "intranet",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Error(err, "Failed to marshal payload")
		return ctrl.Result{}, err
	}

	// Send POST
	resp, err := http.Post("http://your-endpoint/api", "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Error(err, "Failed to send POST request")
		return ctrl.Result{}, err
	}
	defer resp.Body.Close()

	log.Info("POST sent successfully", "status", resp.StatusCode)
	return ctrl.Result{}, nil
}

func (r *ProjectReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cagipv1.Project{}).
		Complete(r)
}
