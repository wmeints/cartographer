package controllers

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ray "github.com/ray-project/kuberay/ray-operator/apis/ray/v1alpha1"
	mlopsv1alpha1 "github.com/wmeints/cartographer/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
)

var _ = Describe("reconcileComputeCluster", func() {
	It("Should deploy the compute cluster", func() {
		ctx := context.Background()
		_ = createWorkspaceAndWaitForRayCluster(ctx, "test-compute")
	})

	It("Should update the image for the ray controller", func() {
		ctx := context.Background()
		workspace := createWorkspaceAndWaitForRayCluster(ctx, "test-compute-image")

		workspace.Spec.Compute.Controller.Image = "test-image"

		err := k8sClient.Update(ctx, workspace)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() error {
			computeCluster, err := getRayCluster(workspace)

			if err != nil {
				return err
			}

			if computeCluster.Spec.HeadGroupSpec.Template.Spec.Containers[0].Image != "test-image" {
				return fmt.Errorf("image not updated")
			}

			return nil
		}, time.Minute, time.Second).Should(Succeed())
	})

	It("Should update the replicas for the ray controller", func() {
		ctx := context.Background()
		workspace := createWorkspaceAndWaitForRayCluster(ctx, "test-compute-replicas")

		workspace.Spec.Compute.Controller.Replicas = pointer.Int32(2)

		err := k8sClient.Update(ctx, workspace)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() error {
			computeCluster, err := getRayCluster(workspace)

			if err != nil {
				return err
			}

			if *computeCluster.Spec.HeadGroupSpec.Replicas != 2 {
				return fmt.Errorf("image not updated")
			}

			return nil
		}, time.Minute, time.Second).Should(Succeed())
	})

	It("Should update the replicas of the worker pools", func() {
		ctx := context.Background()
		workspace := createWorkspaceAndWaitForRayCluster(ctx, "test-compute-workers-replicas")

		workspace.Spec.Compute.WorkerPools[0].MinReplicas = pointer.Int32(2)
		workspace.Spec.Compute.WorkerPools[0].MaxReplicas = pointer.Int32(2)

		err := k8sClient.Update(ctx, workspace)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() error {
			computeCluster, err := getRayCluster(workspace)

			if err != nil {
				return err
			}

			if *computeCluster.Spec.WorkerGroupSpecs[0].Replicas != 2 {
				return fmt.Errorf("Invalid number of replicas for the worker pool")
			}

			if *computeCluster.Spec.WorkerGroupSpecs[0].MinReplicas != 2 {
				return fmt.Errorf("Invalid number of min-replicas for the worker pool")
			}

			if *computeCluster.Spec.WorkerGroupSpecs[0].MaxReplicas != 2 {
				return fmt.Errorf("Invalid number of max-replicas for the worker pool")
			}

			return nil
		}, time.Minute, time.Second).Should(Succeed())
	})

	It("Should update the image of the worker pools", func() {
		ctx := context.Background()
		workspace := createWorkspaceAndWaitForRayCluster(ctx, "test-compute-workers-image")

		workspace.Spec.Compute.WorkerPools[0].Image = "test-image"

		err := k8sClient.Update(ctx, workspace)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() error {
			computeCluster, err := getRayCluster(workspace)

			if err != nil {
				return err
			}

			if computeCluster.Spec.WorkerGroupSpecs[0].Template.Spec.Containers[0].Image != "test-image" {
				return fmt.Errorf("Invalid image for the worker pool")
			}

			return nil
		}, time.Minute, time.Second).Should(Succeed())
	})
})

func createWorkspaceAndWaitForRayCluster(ctx context.Context, name string) *mlopsv1alpha1.Workspace {
	workspace := newTestWorkspace(name)

	err := k8sClient.Create(ctx, workspace)
	Expect(err).NotTo(HaveOccurred())

	Eventually(func() error {
		_, err := getRayCluster(workspace)
		return err
	}, time.Minute, time.Second).Should(Succeed())

	return workspace
}

func getRayCluster(workspace *mlopsv1alpha1.Workspace) (*ray.RayCluster, error) {
	typedClusterName := types.NamespacedName{
		Name:      fmt.Sprintf("%s-ray", workspace.GetName()),
		Namespace: workspace.GetNamespace(),
	}

	computeCluster := &ray.RayCluster{}

	err := k8sClient.Get(context.Background(), typedClusterName, computeCluster)

	if err != nil {
		return nil, err
	}

	return computeCluster, nil
}
