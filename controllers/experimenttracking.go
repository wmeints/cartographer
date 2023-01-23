package controllers

import (
	"context"
	"fmt"
	"reflect"

	mlopsv1alpha1 "github.com/wmeints/cartographer/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *WorkspaceReconciler) reconcileExperimentTracking(ctx context.Context, workspace *mlopsv1alpha1.Workspace) error {
	if err := r.reconcileExperimentTrackingDeployment(ctx, workspace); err != nil {
		return err
	}

	if err := r.reconcileExperimentTrackingService(ctx, workspace); err != nil {
		return err
	}

	return nil
}

func (r *WorkspaceReconciler) reconcileExperimentTrackingDeployment(ctx context.Context, workspace *mlopsv1alpha1.Workspace) error {
	logger := log.FromContext(ctx).WithValues("workspace", workspace.GetName(), "namespace", workspace.GetNamespace())

	deploymentName := fmt.Sprintf("%s-mlflow-server", workspace.GetName())
	deploymentLabels := newComponentLabels(workspace, "experiment-tracking")

	deployment := &appsv1.Deployment{}

	if err := r.Get(ctx, types.NamespacedName{Name: deploymentName, Namespace: workspace.GetNamespace()}, deployment); err != nil {
		if errors.IsNotFound(err) {
			databaseSecretName := fmt.Sprintf("%s-pguser-mlflow", workspace.GetName())

			container := newContainer(
				"mlflow",
				workspace.Spec.ExperimentTracking.Image,
				workspace.Spec.ExperimentTracking.Resources,
			)

			container.Env = newDatabaseSecretEnvVars(databaseSecretName)

			container.Ports = []corev1.ContainerPort{
				{
					Name:          "http-mlflow",
					ContainerPort: 5000,
				},
			}

			deployment = newDeployment(
				workspace.GetNamespace(),
				deploymentName,
				deploymentLabels,
				workspace.Spec.ExperimentTracking.Replicas,
				container,
			)

			if err := ctrl.SetControllerReference(workspace, deployment, r.Scheme); err != nil {
				logger.Error(err, "Failed to set controller reference for deployment of experiment tracking server")
				return err
			}

			if err := r.Create(ctx, deployment); err != nil {
				logger.Error(err, "Failed to create deployment for experiment tracking server")
				return err
			}

			return nil
		}

		deploymentChanged := false

		if haveReplicasChanged(deployment, workspace) {
			deployment.Spec.Replicas = workspace.Spec.ExperimentTracking.Replicas
			deploymentChanged = true
		}

		if isApplicationImageChanged(deployment, workspace) {
			deployment.Spec.Template.Spec.Containers[0].Image = workspace.Spec.ExperimentTracking.Image
			deploymentChanged = true
		}

		if haveResourceLimitsChanged(deployment, workspace) {
			deployment.Spec.Template.Spec.Containers[0].Resources.Limits = workspace.Spec.ExperimentTracking.Resources.Limits
			deploymentChanged = true
		}

		if haveResourceRequestsChanged(deployment, workspace) {
			deployment.Spec.Template.Spec.Containers[0].Resources.Requests = workspace.Spec.ExperimentTracking.Resources.Requests
			deploymentChanged = true
		}

		if deploymentChanged {
			if err := r.Update(ctx, deployment); err != nil {
				logger.Error(err, "Failed to update deployment for experiment tracking server")
				return err
			}
		}

		logger.Error(err, "Failed to get deployment")
		return err
	}

	return nil
}

func (r *WorkspaceReconciler) reconcileExperimentTrackingService(ctx context.Context, workspace *mlopsv1alpha1.Workspace) error {
	logger := log.FromContext(ctx).WithValues("workspace", workspace.GetName(), "namespace", workspace.GetNamespace())

	serviceLabels := newComponentLabels(workspace, "experiment-tracking")
	serviceName := fmt.Sprintf("%s-mlflow-server", workspace.GetName())

	service := &corev1.Service{}

	if err := r.Get(ctx, types.NamespacedName{Name: serviceName, Namespace: workspace.GetNamespace()}, service); err != nil {
		if errors.IsNotFound(err) {
			service = &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      serviceName,
					Namespace: workspace.GetNamespace(),
					Labels:    serviceLabels,
				},
				Spec: corev1.ServiceSpec{
					Type:     corev1.ServiceTypeClusterIP,
					Selector: serviceLabels,
					Ports: []corev1.ServicePort{
						{
							Name:       "http-mlflow",
							Port:       5000,
							TargetPort: intstr.FromInt(5000),
						},
					},
				},
			}

			if err := ctrl.SetControllerReference(workspace, service, r.Scheme); err != nil {
				logger.Error(err, "Failed to set controller reference for service of experiment tracking server")
				return err
			}

			if err := r.Create(ctx, service); err != nil {
				logger.Error(err, "Failed to create service for experiment tracking server")
				return err
			}

			return nil
		}

		return err
	}

	return nil
}

func haveReplicasChanged(deployment *appsv1.Deployment, workspace *mlopsv1alpha1.Workspace) bool {
	return *deployment.Spec.Replicas != *workspace.Spec.ExperimentTracking.Replicas
}

func isApplicationImageChanged(deployment *appsv1.Deployment, workspace *mlopsv1alpha1.Workspace) bool {
	return deployment.Spec.Template.Spec.Containers[0].Image != workspace.Spec.ExperimentTracking.Image
}

func haveResourceLimitsChanged(deployment *appsv1.Deployment, workspace *mlopsv1alpha1.Workspace) bool {
	return !reflect.DeepEqual(deployment.Spec.Template.Spec.Containers[0].Resources.Limits, workspace.Spec.ExperimentTracking.Resources.Limits)
}

func haveResourceRequestsChanged(deployment *appsv1.Deployment, workspace *mlopsv1alpha1.Workspace) bool {
	return !reflect.DeepEqual(deployment.Spec.Template.Spec.Containers[0].Resources.Requests, workspace.Spec.ExperimentTracking.Resources.Requests)
}
