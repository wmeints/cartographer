package controllers

import (
	"context"
	"fmt"

	ray "github.com/ray-project/kuberay/ray-operator/apis/ray/v1alpha1"
	mlopsv1alpha1 "github.com/wmeints/cartographer/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *WorkspaceReconciler) reconcileRayCluster(ctx context.Context, workspace *mlopsv1alpha1.Workspace) error {
	logger := log.FromContext(ctx).WithValues(
		"workspace", workspace.GetName(),
		"namespace", workspace.GetNamespace())

	rayCluster := &ray.RayCluster{}

	rayClusterLabels := map[string]string{
		"mlops.aigency.com/workspace": workspace.GetName(),
		"mlops.aigency.com/component": "ray-cluster",
	}

	if err := r.Get(ctx, types.NamespacedName{Name: workspace.Name, Namespace: workspace.Namespace}, rayCluster); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Ray cluster not found. Creating new ray cluster for workspace")

			rayContainerImage := fmt.Sprintf("rayproject/ray:%s", workspace.Spec.Compute.RayVersion)

			rayCluster = &ray.RayCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      workspace.Name,
					Namespace: workspace.Namespace,
				},
				Spec: ray.RayClusterSpec{
					RayVersion: workspace.Spec.Compute.RayVersion,
					HeadGroupSpec: ray.HeadGroupSpec{
						EnableIngress: pointer.Bool(false),
						Replicas:      workspace.Spec.Compute.Controller.Replicas,
						RayStartParams: map[string]string{
							"dashboard-host": "0.0.0.0",
							"block":          "true",
						},
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: rayClusterLabels,
							},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:      "ray-head",
										Image:     rayContainerImage,
										Resources: workspace.Spec.Compute.Controller.Resources,
										Ports: []corev1.ContainerPort{
											{
												Name:          "tcp-gcs",
												ContainerPort: 6379,
											},
											{
												Name:          "http-dashboard",
												ContainerPort: 8265,
											},
											{
												Name:          "tcp-client",
												ContainerPort: 10001,
											},
										},
										Lifecycle: &corev1.Lifecycle{
											PreStop: &corev1.LifecycleHandler{
												Exec: &corev1.ExecAction{
													Command: []string{
														"/bin/sh", "-c", "ray stop",
													},
												},
											},
										},
									},
								},
							},
						},
					},
					WorkerGroupSpecs: []ray.WorkerGroupSpec{},
				},
			}

			if err := ctrl.SetControllerReference(workspace, rayCluster, r.Scheme); err != nil {
				logger.Error(err, "Failed to set controller reference for ray cluster")
				return err
			}

			if err := r.Create(ctx, rayCluster); err != nil {
				logger.Error(err, "Failed to create ray cluster for workspace")
				return err
			}

			return nil
		}

		logger.Error(err, "Failed to get ray cluster for workspace")
		return err
	}

	return nil
}
