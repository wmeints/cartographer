package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	mlopsv1alpha1 "github.com/wmeints/cartographer/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *WorkspaceReconciler) reconcileWorkflowServer(ctx context.Context, workspace *mlopsv1alpha1.Workspace) error {
	logger := log.FromContext(ctx).WithValues(
		"workspace", workspace.GetName(),
		"namespace", workspace.GetNamespace())

	if err := r.reconcileWorkflowServerDeployment(ctx, logger, workspace); err != nil {
		return err
	}

	if err := r.reconcileWorkflowServerService(ctx, logger, workspace); err != nil {
		return err
	}

	if err := r.reconcileWorkflowAgents(ctx, logger, workspace); err != nil {
		return err
	}

	return nil
}

func (r *WorkspaceReconciler) reconcileWorkflowServerDeployment(ctx context.Context, logger logr.Logger, workspace *mlopsv1alpha1.Workspace) error {
	deploymentName := fmt.Sprintf("%s-orion-server", workspace.GetName())
	deployment := &appsv1.Deployment{}

	if err := r.Get(ctx, types.NamespacedName{Name: deploymentName, Namespace: workspace.GetNamespace()}, deployment); err != nil {
		if errors.IsNotFound(err) {
			return r.createWorkflowServerDeployment(ctx, workspace, logger)
		}

		logger.Error(err, "Failed to get deployment for workflow server")
		return err
	}

	return r.updateWorkflowServerDeployment(ctx, deployment, workspace, logger)
}

func (r *WorkspaceReconciler) updateWorkflowServerDeployment(ctx context.Context, deployment *appsv1.Deployment, workspace *mlopsv1alpha1.Workspace, logger logr.Logger) error {
	deployment.Spec.Replicas = workspace.Spec.Workflows.Controller.Replicas
	deployment.Spec.Template.Spec.Containers[0].Image = workspace.Spec.Workflows.Controller.Image

	if err := r.Update(ctx, deployment); err != nil {
		logger.Error(err, "Failed to scale deployment for workflow server")
		return err
	}

	return nil
}

func (r *WorkspaceReconciler) createWorkflowServerDeployment(ctx context.Context, workspace *mlopsv1alpha1.Workspace, logger logr.Logger) error {
	deployment := newWorkflowServerDeployment(workspace)

	if err := ctrl.SetControllerReference(workspace, deployment, r.Scheme); err != nil {
		logger.Error(err, "Failed to set controller reference for workflow server deployment")
		return err
	}

	if err := r.Create(ctx, deployment); err != nil {
		logger.Error(err, "Failed to create deployment for workflow server")
		return err
	}

	return nil
}

func (r *WorkspaceReconciler) reconcileWorkflowServerService(ctx context.Context, logger logr.Logger, workspace *mlopsv1alpha1.Workspace) error {
	serviceName := fmt.Sprintf("%s-orion-server", workspace.GetName())
	serviceLabels := newComponentLabels(workspace, "workflow-server")

	service := &corev1.Service{}

	if err := r.Get(ctx, types.NamespacedName{Name: serviceName, Namespace: workspace.GetNamespace()}, service); err != nil {
		if errors.IsNotFound(err) {
			service = newService(serviceName, workspace.GetNamespace(), serviceLabels)

			service.Spec.Ports = []corev1.ServicePort{
				{
					Name:       "http-orion",
					Protocol:   corev1.ProtocolTCP,
					Port:       4200,
					TargetPort: intstr.FromInt(4200),
				},
			}

			if err := ctrl.SetControllerReference(workspace, service, r.Scheme); err != nil {
				logger.Error(err, "Failed to set controller reference workflow server service")
				return err
			}

			if err := r.Create(ctx, service); err != nil {
				logger.Error(err, "Failed to create service for workflow server")
				return err
			}

			return nil
		}

		logger.Error(err, "Failed to get service for workflow server")
		return err
	}

	return nil
}

func (r *WorkspaceReconciler) reconcileWorkflowAgents(ctx context.Context, logger logr.Logger, workspace *mlopsv1alpha1.Workspace) error {
	for _, agentPoolSpec := range workspace.Spec.Workflows.Agents {
		statefulSetName := fmt.Sprintf("%s-agent-%s", workspace.GetName(), agentPoolSpec.Name)
		statefulSet := &appsv1.StatefulSet{}

		if err := r.Get(ctx, types.NamespacedName{Name: statefulSetName, Namespace: workspace.GetNamespace()}, statefulSet); err != nil {
			if errors.IsNotFound(err) {
				return r.createWorkflowAgentPool(ctx, &agentPoolSpec, workspace, logger)
			}

			logger.Error(err, "Failed to get statefulset for workflow agent pool")
			return err
		}

		if err := r.updateWorkflowAgentPool(ctx, statefulSet, agentPoolSpec, logger); err != nil {
			return err
		}
	}

	return nil
}

func (r *WorkspaceReconciler) updateWorkflowAgentPool(ctx context.Context, statefulSet *appsv1.StatefulSet, agentPoolSpec mlopsv1alpha1.WorkflowAgentPoolSpec, logger logr.Logger) error {
	statefulSetChanged := false

	if *statefulSet.Spec.Replicas != *agentPoolSpec.Replicas {
		statefulSet.Spec.Replicas = agentPoolSpec.Replicas
		statefulSetChanged = true
	}

	if statefulSet.Spec.Template.Spec.Containers[0].Image != agentPoolSpec.Image {
		statefulSet.Spec.Template.Spec.Containers[0].Image = agentPoolSpec.Image
		statefulSetChanged = true
	}

	if statefulSetChanged {
		if err := r.Update(ctx, statefulSet); err != nil {
			logger.Error(err, "Failed to update stateful set for workflow agent pool %s", agentPoolSpec.Name)
			return err
		}
	}

	return nil
}

func (r *WorkspaceReconciler) createWorkflowAgentPool(ctx context.Context, agentPoolSpec *mlopsv1alpha1.WorkflowAgentPoolSpec, workspace *mlopsv1alpha1.Workspace, logger logr.Logger) error {
	statefulSetName := fmt.Sprintf("%s-agent-%s", workspace.GetName(), agentPoolSpec.Name)
	statefulSetLabels := newComponentLabels(workspace, "workflow-agent")
	statefulSetLabels["mlops.aigency.com/pool"] = agentPoolSpec.Name

	container := newContainer("agent", agentPoolSpec.Image, agentPoolSpec.Resources)

	container.Command = []string{
		"prefect",
		"agent",
		"start",
		"-q",
		agentPoolSpec.Name,
	}

	container.Env = []corev1.EnvVar{
		{
			Name:  "PREFECT_API_URL",
			Value: fmt.Sprintf("http://%s-orion-server:4200/api", workspace.GetName()),
		},
		{
			Name:  "MLFLOW_TRACKING_URI",
			Value: fmt.Sprintf("http://%s-orion-server:4200/api", workspace.GetName()),
		},
	}

	statefulSet := newStatefulSet(workspace.GetNamespace(), statefulSetName, statefulSetLabels, agentPoolSpec.Replicas, container)

	if err := ctrl.SetControllerReference(workspace, statefulSet, r.Scheme); err != nil {
		logger.Error(err, "Failed to set controller reference for stateful set %s", agentPoolSpec.Name)
		return err
	}

	if err := r.Create(ctx, statefulSet); err != nil {
		logger.Error(err, "Failed to create stateful set for agent pool %s", agentPoolSpec.Name)
		return err
	}

	return nil
}

func newWorkflowServerDeployment(workspace *mlopsv1alpha1.Workspace) *appsv1.Deployment {
	deploymentName := fmt.Sprintf("%s-orion-server", workspace.GetName())
	deploymentLabels := newComponentLabels(workspace, "workflow-server")
	container := newContainer("orion", workspace.Spec.Workflows.Controller.Image, workspace.Spec.Workflows.Controller.Resources)

	container.Env = append(
		[]corev1.EnvVar{
			{
				Name:  "PREFECT_ORION_DATABASE_CONNECTION_URL",
				Value: "postgresql+asyncpg://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)",
			},
		},
		newDatabaseSecretEnvVars(workspace.Spec.Workflows.Controller.DatabaseConnectionSecret)...,
	)

	container.Ports = []corev1.ContainerPort{
		{
			Name:          "http-orion",
			ContainerPort: 4200,
		},
	}

	container.Command = []string{
		"prefect",
		"orion",
		"start",
	}

	deployment := newDeployment(
		workspace.GetNamespace(),
		deploymentName,
		deploymentLabels,
		workspace.Spec.Workflows.Controller.Replicas,
		container,
	)

	return deployment
}
