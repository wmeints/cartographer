package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	mlopsv1alpha1 "github.com/wmeints/cartographer/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
)

var _ = Describe("reconcileWorkflowServer", Ordered, func() {
	It("Should deploy the workflow component", func() {
		ctx := context.Background()

		workspace := &mlopsv1alpha1.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
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

		err := k8sClient.Create(ctx, workspace)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() error {
			deployment := &appsv1.Deployment{}
			deploymentName := types.NamespacedName{Name: "test-orion-server", Namespace: "test-namespace"}

			return k8sClient.Get(ctx, deploymentName, deployment)
		}, time.Minute, time.Second).Should(Succeed())

		Eventually(func() error {
			service := &corev1.Service{}
			serviceName := types.NamespacedName{Name: "test-orion-server", Namespace: "test-namespace"}

			return k8sClient.Get(ctx, serviceName, service)
		}, time.Minute, time.Second).Should(Succeed())
	})
})
