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
			Storage: mlopsv1alpha1.WorkspaceStorageSpec{
				DatabaseStorage:       resource.MustParse("1Gi"),
				DatabaseBackupStorage: resource.MustParse("1Gi"),
			},
			Workflows: mlopsv1alpha1.WorkflowComponentSpec{
				Controller: mlopsv1alpha1.WorkflowControllerSpec{
					Replicas: pointer.Int32(1),
					Image:    "willemmeints/workflow-controller:latest",
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
						Image:    "willemmeints/workflow-agent:latest",
						Replicas: pointer.Int32(1),
					},
				},
			},
			ExperimentTracking: mlopsv1alpha1.ExperimentTrackingComponentSpec{
				Image:    "willemmeints/experiment-tracking:latest",
				Replicas: pointer.Int32(1),
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
