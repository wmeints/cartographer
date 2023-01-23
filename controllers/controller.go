/*
Copyright 2023 Willem Meints.

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

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mlopsv1alpha1 "github.com/wmeints/cartographer/api/v1alpha1"
)

// WorkspaceReconciler reconciles a Workspace object
type WorkspaceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=mlops.aigency.com,resources=workspaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mlops.aigency.com,resources=workspaces/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mlops.aigency.com,resources=workspaces/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments;statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=postgres-operator.crunchydata.com,resources=postgresclusters,verbs=get;list;watch;create;update;patch;delete

// Reconcile matches the expected state of the workspace against the cluster state.
// It automatically updates the cluster state if there's a mismatch.
func (r *WorkspaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	workspace := &mlopsv1alpha1.Workspace{}
	workspace.Default()

	if err := r.Get(ctx, req.NamespacedName, workspace); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Workspace not found. Skipping reconciliation.", "workspaceName", req.Name, "namespace", req.Namespace)
			return ctrl.Result{}, nil
		}

		logger.Error(err, "Failed to get the workspace",
			"workspaceName", workspace.GetName(),
			"namespace", workspace.GetNamespace())

		return ctrl.Result{}, err
	}

	if err := r.reconcilePostgresCluster(ctx, workspace); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileExperimentTracking(ctx, workspace); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileWorkflowServer(ctx, workspace); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mlopsv1alpha1.Workspace{}).
		Complete(r)
}
