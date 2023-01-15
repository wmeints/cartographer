package v1alpha1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/pointer"
)

var _ = Describe("Defaulting webhook", func() {
	It("Should set the default values for experiment tracking", func() {
		workspace := &Workspace{}
		workspace.Default()

		Expect(workspace.Spec.ExperimentTracking).To(Equal(ExperimentTrackingComponentSpec{
			Replicas: pointer.Int32Ptr(1),
			Image:    "willemmeints/mlflow:2.1.1",
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

	It("Should set the default values for the workflows", func() {
		workspace := &Workspace{}
		workspace.Default()

		Expect(workspace.Spec.Workflows).To(Equal(WorkflowComponentSpec{
			Controller: WorkflowControllerSpec{
				Replicas: pointer.Int32(1),
				Image:    "prefecthq/prefect:2-latest",
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
				Image:    "prefecthq/prefect:2-latest",
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
			Image:    "willemmeints/mlflow:2.1.1",
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

var _ = Describe("Validation webhook", func() {
	Context("Create", func() {
		It("Should validate the databaseConnectionSecret variable", func() {
			workspace := &Workspace{
				Spec: WorkspaceSpec{
					ExperimentTracking: ExperimentTrackingComponentSpec{
						DatabaseConnectionSecret: "test-connection",
					},
					Workflows: WorkflowComponentSpec{
						Controller: WorkflowControllerSpec{
							DatabaseConnectionSecret: "",
						},
					},
				},
			}

			err := workspace.ValidateCreate()

			Expect(err).To(HaveOccurred())
		})

		It("Should validate the databaseConnectionSecret variable", func() {
			workspace := &Workspace{
				Spec: WorkspaceSpec{
					ExperimentTracking: ExperimentTrackingComponentSpec{
						DatabaseConnectionSecret: "",
					},
					Workflows: WorkflowComponentSpec{
						Controller: WorkflowControllerSpec{
							DatabaseConnectionSecret: "test-connection",
						},
					},
				},
			}

			err := workspace.ValidateCreate()

			Expect(err).To(HaveOccurred())
		})

		It("Should work for valid workspaces", func() {
			workspace := &Workspace{
				Spec: WorkspaceSpec{
					ExperimentTracking: ExperimentTrackingComponentSpec{
						DatabaseConnectionSecret: "test-connection",
					},
					Workflows: WorkflowComponentSpec{
						Controller: WorkflowControllerSpec{
							DatabaseConnectionSecret: "test-connection",
						},
					},
				},
			}

			err := workspace.ValidateCreate()

			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("Update", func() {
		It("Should validate the databaseConnectionSecret variable", func() {
			oldWorkspace := &Workspace{
				Spec: WorkspaceSpec{
					ExperimentTracking: ExperimentTrackingComponentSpec{
						DatabaseConnectionSecret: "test-connection",
					},
					Workflows: WorkflowComponentSpec{
						Controller: WorkflowControllerSpec{
							DatabaseConnectionSecret: "test-connection",
						},
					},
				},
			}

			workspace := &Workspace{
				Spec: WorkspaceSpec{
					ExperimentTracking: ExperimentTrackingComponentSpec{
						DatabaseConnectionSecret: "test-connection",
					},
					Workflows: WorkflowComponentSpec{
						Controller: WorkflowControllerSpec{
							DatabaseConnectionSecret: "",
						},
					},
				},
			}

			err := workspace.ValidateUpdate(oldWorkspace)

			Expect(err).To(HaveOccurred())
		})

		It("Should validate the databaseConnectionSecret variable", func() {
			oldWorkspace := &Workspace{
				Spec: WorkspaceSpec{
					ExperimentTracking: ExperimentTrackingComponentSpec{
						DatabaseConnectionSecret: "test-connection",
					},
					Workflows: WorkflowComponentSpec{
						Controller: WorkflowControllerSpec{
							DatabaseConnectionSecret: "test-connection",
						},
					},
				},
			}

			workspace := &Workspace{
				Spec: WorkspaceSpec{
					ExperimentTracking: ExperimentTrackingComponentSpec{
						DatabaseConnectionSecret: "",
					},
					Workflows: WorkflowComponentSpec{
						Controller: WorkflowControllerSpec{
							DatabaseConnectionSecret: "test-connection",
						},
					},
				},
			}

			err := workspace.ValidateUpdate(oldWorkspace)

			Expect(err).To(HaveOccurred())
		})

		It("Should work for valid workspaces", func() {
			oldWorkspace := &Workspace{
				Spec: WorkspaceSpec{
					ExperimentTracking: ExperimentTrackingComponentSpec{
						DatabaseConnectionSecret: "test-connection",
					},
					Workflows: WorkflowComponentSpec{
						Controller: WorkflowControllerSpec{
							DatabaseConnectionSecret: "test-connection",
						},
					},
				},
			}

			workspace := &Workspace{
				Spec: WorkspaceSpec{
					ExperimentTracking: ExperimentTrackingComponentSpec{
						DatabaseConnectionSecret: "test-connection-2",
					},
					Workflows: WorkflowComponentSpec{
						Controller: WorkflowControllerSpec{
							DatabaseConnectionSecret: "test-connection-2",
						},
					},
				},
			}

			err := workspace.ValidateUpdate(oldWorkspace)

			Expect(err).NotTo(HaveOccurred())
		})
	})
})
