package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	mlopsv1alpha1 "github.com/wmeints/cartographer/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	deploymentLabels := makeOrionServerLabels(workspace)

	deploymentImageName := workspace.Spec.Workflows.Image

	if deploymentImageName == "" {
		deploymentImageName = "prefecthq/prefect:2-latest"
	}

	databaseSecretName := workspace.Spec.Workflows.DatabaseConnectionSecret

	deployment := &appsv1.Deployment{}

	if err := r.Get(ctx, types.NamespacedName{Name: deploymentName, Namespace: workspace.GetNamespace()}, deployment); err != nil {
		if errors.IsNotFound(err) {
			deployment = &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      deploymentName,
					Namespace: workspace.GetNamespace(),
					Labels:    deploymentLabels,
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: deploymentLabels,
					},
					Replicas: workspace.Spec.Workflows.ControllerReplicas,
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Name:   deploymentName,
							Labels: deploymentLabels,
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "orion",
									Image: deploymentImageName,
									Command: []string{
										"prefect",
										"orion",
										"start",
									},
									Env: []corev1.EnvVar{
										{
											Name: "DB_HOST",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{
													Key: "host",
													LocalObjectReference: corev1.LocalObjectReference{
														Name: databaseSecretName,
													},
												},
											},
										},
										{
											Name: "DB_PORT",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{
													Key: "port",
													LocalObjectReference: corev1.LocalObjectReference{
														Name: databaseSecretName,
													},
												},
											},
										},
										{
											Name: "DB_USER",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{
													Key: "user",
													LocalObjectReference: corev1.LocalObjectReference{
														Name: databaseSecretName,
													},
												},
											},
										},
										{
											Name: "DB_PASS",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{
													Key: "password",
													LocalObjectReference: corev1.LocalObjectReference{
														Name: databaseSecretName,
													},
												},
											},
										},
										{
											Name:  "PREFECT_ORION_DATABASE_CONNECTION_URL",
											Value: "postgresql+asyncpg://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)",
										},
									},
									Resources: corev1.ResourceRequirements{
										Requests: makeResourceList("100m", "512Mi"),
										Limits:   makeResourceList("500m", "1Gi"),
									},
									Ports: []corev1.ContainerPort{
										{
											Name:          "http-orion",
											ContainerPort: 4200,
										},
									},
								},
							},
						},
					},
				},
			}

			if err := ctrl.SetControllerReference(workspace, deployment, r.Scheme); err != nil {
				logger.Error(err, "Failed to set controller reference for orion server deployment")
				return err
			}

			if err := r.Create(ctx, deployment); err != nil {
				logger.Error(err, "Failed to create deployment for orion server")
				return err
			}

			return nil
		}

		deployment.Spec.Replicas = workspace.Spec.Workflows.ControllerReplicas
		deployment.Spec.Template.Spec.Containers[0].Image = workspace.Spec.Workflows.Image

		if err := r.Update(ctx, deployment); err != nil {
			logger.Error(err, "Failed to scale deployment for orion server")
			return err
		}

		logger.Error(err, "Failed to get deployment for the workflow server")
		return err
	}

	return nil
}

func (r *WorkspaceReconciler) reconcileWorkflowServerService(ctx context.Context, logger logr.Logger, workspace *mlopsv1alpha1.Workspace) error {
	serviceName := fmt.Sprintf("%s-orion-server", workspace.GetName())
	serviceLabels := makeOrionServerLabels(workspace)

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
							Name:       "http-orion",
							Protocol:   corev1.ProtocolTCP,
							Port:       4200,
							TargetPort: intstr.FromInt(4200),
						},
					},
				},
			}

			if err := ctrl.SetControllerReference(workspace, service, r.Scheme); err != nil {
				logger.Error(err, "Failed to set controller reference for service")
				return err
			}

			if err := r.Create(ctx, service); err != nil {
				logger.Error(err, "Failed to create service for orion server")
				return err
			}

			return nil
		}

		logger.Error(err, "Failed to get service for orion server")
		return err
	}

	return nil
}

func (r *WorkspaceReconciler) reconcileWorkflowAgents(ctx context.Context, logger logr.Logger, workspace *mlopsv1alpha1.Workspace) error {

	for _, agentPoolSpec := range workspace.Spec.Workflows.Agents {
		statefulSetLabels := map[string]string{
			"mlops.aigency.com/environment": workspace.GetName(),
			"mlops.aigency.com/component":   "prefect-agent",
			"mlops.aigency.com/pool":        agentPoolSpec.Name,
		}

		statefulSetName := fmt.Sprintf("%s-agent-%s", workspace.GetName(), agentPoolSpec.Name)
		statefulSetImageName := agentPoolSpec.Image

		if statefulSetImageName == "" {
			statefulSetImageName = "prefecthq/prefect:2-latest"
		}

		statefulSet := &appsv1.StatefulSet{}

		if err := r.Get(ctx, types.NamespacedName{Name: statefulSetName, Namespace: workspace.GetNamespace()}, statefulSet); err != nil {
			if errors.IsNotFound(err) {
				statefulSet = &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      statefulSetName,
						Namespace: workspace.GetNamespace(),
						Labels:    statefulSetLabels,
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: agentPoolSpec.Replicas,
						Selector: &metav1.LabelSelector{
							MatchLabels: statefulSetLabels,
						},
						UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
							Type: appsv1.RollingUpdateStatefulSetStrategyType,
						},
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: statefulSetLabels,
							},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "agent",
										Image: statefulSetImageName,
										Env: []corev1.EnvVar{
											{
												Name:  "PREFECT_API_URL",
												Value: fmt.Sprintf("http://%s-orion-server:4200/api", workspace.GetName()),
											},
											{
												Name:  "MLFLOW_TRACKING_URI",
												Value: fmt.Sprintf("http://%s-orion-server:4200/api", workspace.GetName()),
											},
										},
										Resources: corev1.ResourceRequirements{
											Requests: agentPoolSpec.Resources.Requests,
											Limits:   agentPoolSpec.Resources.Limits,
										},
									},
								},
							},
						},
					},
				}

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
					logger.Error(err, "Failed to update stateful set for agent pool %s", agentPoolSpec.Name)
					return err
				}

				return nil
			}

			logger.Error(err, "Failed to get stateful set for agent pool %s", agentPoolSpec.Name)
			return err
		}
	}

	return nil
}

func makeResourceList(cpuAmount string, memoryAmount string) corev1.ResourceList {
	return corev1.ResourceList{
		"cpu":    resource.MustParse(cpuAmount),
		"memory": resource.MustParse(memoryAmount),
	}
}

func makeOrionServerLabels(workspace *mlopsv1alpha1.Workspace) map[string]string {
	return map[string]string{
		"mlops.aigency.com/environment": workspace.GetName(),
		"mlops.aigency.com/component":   "orion-server",
	}
}
