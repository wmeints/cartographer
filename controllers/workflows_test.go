package controllers

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("reconcileWorkflowServer", Ordered, func() {
	It("Should deploy the workflow component", func() {
		ctx := context.Background()

		workspace := newTestWorkspace("test-workflows")

		err := k8sClient.Create(ctx, workspace)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() error {
			deployment := &appsv1.Deployment{}
			deploymentName := types.NamespacedName{
				Name:      fmt.Sprintf("%s-orion-server", workspace.GetName()),
				Namespace: "test-namespace",
			}

			return k8sClient.Get(ctx, deploymentName, deployment)
		}, time.Minute, time.Second).Should(Succeed())

		Eventually(func() error {
			service := &corev1.Service{}
			serviceName := types.NamespacedName{
				Name:      fmt.Sprintf("%s-orion-server", workspace.GetName()),
				Namespace: "test-namespace",
			}

			return k8sClient.Get(ctx, serviceName, service)
		}, time.Minute, time.Second).Should(Succeed())

		Eventually(func() error {
			statefulSetName := types.NamespacedName{
				Name:      fmt.Sprintf("%s-agent-%s", workspace.GetName(), "test"),
				Namespace: "test-namespace",
			}

			statefulSet := &appsv1.StatefulSet{}

			return k8sClient.Get(ctx, statefulSetName, statefulSet)
		}, time.Minute, time.Second).Should(Succeed())
	})
})
