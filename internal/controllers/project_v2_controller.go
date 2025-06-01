package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	cagipv2 "github.com/cagip/jfrog-operator/api/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ProjectV2Reconciler reconciles a Project object
type ProjectV2Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cagip.github.com,resources=projects,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cagip.github.com,resources=projects/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cagip.github.com,resources=projects/finalizers,verbs=update

func (r *ProjectV2Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// 1. Fetch the Project v2 resource
	var project cagipv2.Project
	if err := r.Get(ctx, req.NamespacedName, &project); err != nil {
		if errors.IsNotFound(err) {
			log.Info("Project v2 resource not found. Ignoring since it must have been deleted.")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get Project v2 resource")
		return ctrl.Result{}, err
	}

	// 2. Get token from secret
	secret := &corev1.Secret{}
	secretName := types.NamespacedName{
		Name:      "api-auth-token", // <- change if needed
		Namespace: req.Namespace,    // assumes secret is in same namespace
	}
	if err := r.Get(ctx, secretName, secret); err != nil {
		log.Error(err, "Failed to fetch API auth token secret")
		return ctrl.Result{}, err
	}

	tokenBytes, ok := secret.Data["token"]
	if !ok {
		log.Error(nil, "Token key 'token' not found in secret")
		return ctrl.Result{}, fmt.Errorf("missing token in secret")
	}
	token := string(tokenBytes)

	// 3. Prepare payload
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

	// 4. Send POST request with Authorization header
	reqBody := bytes.NewBuffer(jsonPayload)
	httpReq, err := http.NewRequest("POST", "http://your-endpoint/api", reqBody)
	if err != nil {
		log.Error(err, "Failed to create HTTP request")
		return ctrl.Result{}, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		log.Error(err, "POST request failed")
		return ctrl.Result{}, err
	}
	defer resp.Body.Close()

	log.Info("POST request sent", "statusCode", resp.StatusCode)
	return ctrl.Result{}, nil
}

func (r *ProjectV2Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cagipv2.Project{}).
		Complete(r)
}
