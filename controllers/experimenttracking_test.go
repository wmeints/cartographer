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

var _ = Describe("reconcileExperimentTracking", func() {
	It("Should deploy the experiment tracking component", func() {
		workspace := newTestWorkspace("test-experimenttracking")
		ctx := context.Background()

		err := k8sClient.Create(ctx, workspace)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() error {
			deploymentName := types.NamespacedName{
				Name:      fmt.Sprintf("%s-mlflow-server", workspace.GetName()),
				Namespace: workspace.GetNamespace(),
			}

			deployment := &appsv1.Deployment{}

			return k8sClient.Get(ctx, deploymentName, deployment)
		}, time.Minute, time.Second).Should(Succeed())

		Eventually(func() error {
			serviceName := types.NamespacedName{
				Name:      fmt.Sprintf("%s-mlflow-server", workspace.GetName()),
				Namespace: workspace.GetNamespace(),
			}

			service := &corev1.Service{}

			return k8sClient.Get(ctx, serviceName, service)
		}, time.Minute, time.Second).Should(Succeed())
	})
})
