package controllers

import (
	mlopsv1alpha1 "github.com/wmeints/cartographer/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

func newTestWorkspace(workspaceName string) *mlopsv1alpha1.Workspace {
	return &mlopsv1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workspaceName,
			Namespace: "test-namespace",
		},
		Spec: mlopsv1alpha1.WorkspaceSpec{
			Workflows: mlopsv1alpha1.WorkflowComponentSpec{
				Controller: mlopsv1alpha1.WorkflowControllerSpec{
					DatabaseConnectionSecret: "test-secret",
					Replicas:                 pointer.Int32(1),
					Image:                    "prefecthq/prefect:2-latest",
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("1Gi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("200Mi"),
						},
					},
				},
				Agents: []mlopsv1alpha1.WorkflowAgentPoolSpec{
					{
						Name:     "test",
						Replicas: pointer.Int32(1),
					},
				},
			},
			ExperimentTracking: mlopsv1alpha1.ExperimentTrackingComponentSpec{
				Image:                    "willemmeints/mlflow:2.1.1",
				DatabaseConnectionSecret: "test-secret",
				Resources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1"),
						corev1.ResourceMemory: resource.MustParse("1Gi"),
					},
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("100m"),
						corev1.ResourceMemory: resource.MustParse("200Mi"),
					},
				},
			},
		},
	}
}
