package controllers

import (
	"context"
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
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

	if err := r.Get(ctx, types.NamespacedName{Name: workspace.Name, Namespace: workspace.Namespace}, rayCluster); err != nil {
		if errors.IsNotFound(err) {
			return r.createComputeCluster(ctx, workspace, logger)
		}

		logger.Error(err, "Failed to get ray cluster for workspace")
		return err
	}

	return r.updateComputeCluster(ctx, workspace, rayCluster, logger)
}

func (r *WorkspaceReconciler) updateComputeCluster(ctx context.Context, workspace *mlopsv1alpha1.Workspace, rayCluster *ray.RayCluster, logger logr.Logger) error {
	if rayCluster.Spec.HeadGroupSpec.Replicas != workspace.Spec.Compute.Controller.Replicas {
		rayCluster.Spec.HeadGroupSpec.Replicas = workspace.Spec.Compute.Controller.Replicas
	}

	if rayCluster.Spec.HeadGroupSpec.Template.Spec.Containers[0].Image != workspace.Spec.Compute.Controller.Image {
		rayCluster.Spec.HeadGroupSpec.Template.Spec.Containers[0].Image = workspace.Spec.Compute.Controller.Image
	}

	if !reflect.DeepEqual(rayCluster.Spec.HeadGroupSpec.Template.Spec.Containers[0].Resources, workspace.Spec.Compute.Controller.Resources) {
		rayCluster.Spec.HeadGroupSpec.Template.Spec.Containers[0].Resources = workspace.Spec.Compute.Controller.Resources
	}

	if !reflect.DeepEqual(rayCluster.Spec.WorkerGroupSpecs, newWorkerGroups(workspace)) {
		workerGroups := newWorkerGroups(workspace)
		rayCluster.Spec.WorkerGroupSpecs = workerGroups
	}

	if err := r.Update(ctx, rayCluster); err != nil {
		logger.Error(err, "Failed to update compute cluster for the workspace")
		return err
	}

	return nil
}

func (r *WorkspaceReconciler) createComputeCluster(ctx context.Context, workspace *mlopsv1alpha1.Workspace, logger logr.Logger) error {
	logger.Info("Compute cluster not found. Creating new ray cluster for workspace")

	controllerSpec := newRayClusterController(workspace)
	workerGroups := newWorkerGroups(workspace)

	rayCluster := &ray.RayCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workspace.Name,
			Namespace: workspace.Namespace,
		},
		Spec: ray.RayClusterSpec{
			RayVersion:       workspace.Spec.Compute.RayVersion,
			HeadGroupSpec:    controllerSpec,
			WorkerGroupSpecs: workerGroups,
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

func newRayClusterController(workspace *mlopsv1alpha1.Workspace) ray.HeadGroupSpec {
	rayClusterLabels := map[string]string{
		"mlops.aigency.com/workspace": workspace.GetName(),
		"mlops.aigency.com/component": "ray-controller",
	}

	controllerSpec := ray.HeadGroupSpec{
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
						Image:     workspace.Spec.Compute.Controller.Image,
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
	}

	return controllerSpec
}

func newWorkerGroups(workspace *mlopsv1alpha1.Workspace) []ray.WorkerGroupSpec {
	workerGroups := []ray.WorkerGroupSpec{}

	for _, workerGroup := range workspace.Spec.Compute.WorkerPools {
		workerGroupLabels := map[string]string{
			"mlops.aigency.com/workspace": workspace.GetName(),
			"mlops.aigency.com/component": "ray-worker",
			"mlops.aigency.com/pool":      workerGroup.Name,
		}

		workerGroup := ray.WorkerGroupSpec{
			GroupName:   workerGroup.Name,
			Replicas:    workerGroup.MinReplicas,
			MinReplicas: workerGroup.MinReplicas,
			MaxReplicas: workerGroup.MaxReplicas,
			RayStartParams: map[string]string{
				"block": "true",
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: workerGroupLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:      "ray-worker",
							Image:     workerGroup.Image,
							Resources: workerGroup.Resources,
							Lifecycle: &corev1.Lifecycle{
								PreStop: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{
										Command: []string{"/bin/sh", "-c", "ray stop"},
									},
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "MLFLOW_TRACKING_URI",
									Value: fmt.Sprintf("http://%s-orion-server:4200/api", workspace.GetName()),
								},
							},
						},
					},
					InitContainers: []corev1.Container{
						{
							Name:    "ray-worker-init",
							Image:   "busbox:1.28",
							Command: []string{"sh", "-c", "until nslookup $RAY_IP.$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace).svc.cluster.local; do echo waiting for K8s Service $RAY_IP; sleep 2; done"},
						},
					},
				},
			},
		}

		workerGroups = append(workerGroups, workerGroup)
	}

	return workerGroups
}
