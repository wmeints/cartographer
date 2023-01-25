package controllers

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ray "github.com/ray-project/kuberay/ray-operator/apis/ray/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("reconcileComputeCluster", func() {
	It("Should deploy the compute cluster", func() {
		workspace := newTestWorkspace("test-compute")
		ctx := context.Background()

		err := k8sClient.Create(ctx, workspace)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() error {
			clusterName := types.NamespacedName{
				Name:      fmt.Sprintf("%s-ray", workspace.GetName()),
				Namespace: workspace.GetNamespace(),
			}

			computeCluster := &ray.RayCluster{}

			return k8sClient.Get(ctx, clusterName, computeCluster)
		}, time.Minute, time.Second).Should(Succeed())
	})
})
