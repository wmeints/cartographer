package v1alpha1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/pointer"
)

var _ = Describe("Defaulting webhook", func() {
	It("Should set the default values for the workflows", func() {
		workspace := &Workspace{}
		workspace.Default()

		Expect(workspace.Spec.Workflows).To(Equal(WorkflowComponentSpec{
			Controller: WorkflowControllerSpec{
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
			Agents: []WorkflowAgentPoolSpec{},
		}))
	})

	It("Should set the default values for the agents", func() {
		workspace := &Workspace{
			Spec: WorkspaceSpec{
				Workflows: WorkflowComponentSpec{
					Controller: WorkflowControllerSpec{},
					Agents: []WorkflowAgentPoolSpec{
						{
							Name:     "test-agent",
							Replicas: pointer.Int32(1),
						},
					},
				},
			},
		}

		workspace.Default()

		Expect(workspace.Spec.Workflows.Agents).To(Equal([]WorkflowAgentPoolSpec{
			{
				Name:     "test-agent",
				Replicas: pointer.Int32(1),
				Image:    "willemmeints/workflow-agent:latest",
				Resources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("2"),
						corev1.ResourceMemory: resource.MustParse("16Gi"),
					},
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("500m"),
						corev1.ResourceMemory: resource.MustParse("2Gi"),
					},
				},
			},
		}))
	})

	It("Should set the default values for experiment tracking", func() {
		workspace := &Workspace{}

		workspace.Default()

		Expect(workspace.Spec.ExperimentTracking).To(Equal(ExperimentTrackingComponentSpec{
			Image:    "willemmeints/experiment-tracking:latest",
			Replicas: pointer.Int32(1),
			Resources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("512Mi"),
				},
			},
		}))
	})
})
