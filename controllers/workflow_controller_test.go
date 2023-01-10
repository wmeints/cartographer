package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	mlopsv1alpha1 "github.com/wmeints/cartographer/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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
					DatabaseConnectionSecret: "test-secret",
					ControllerReplicas:       pointer.Int32(1),
					Agents: []mlopsv1alpha1.WorkflowAgentPoolSpec{
						{
							Name:     "test",
							Replicas: pointer.Int32(1),
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
