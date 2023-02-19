package controllers

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	mlopsv1alpha1 "github.com/wmeints/cartographer/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
)

var _ = Describe("reconcileExperimentTracking", func() {
	It("Should deploy the experiment tracking component", func() {
		ctx := context.Background()
		workspace := createWorkspaceAndWaitForExperimentTrackingDeployment(ctx, "test-experimenttracking")

		Eventually(func() error {
			serviceName := types.NamespacedName{
				Name:      fmt.Sprintf("%s-mlflow-server", workspace.GetName()),
				Namespace: workspace.GetNamespace(),
			}

			service := &corev1.Service{}

			return k8sClient.Get(ctx, serviceName, service)
		}, time.Minute, time.Second).Should(Succeed())
	})

	It("Should update the experiment tracking component image", func() {
		ctx := context.Background()
		workspace := createWorkspaceAndWaitForExperimentTrackingDeployment(ctx, "test-experimenttracking-image")

		workspace.Spec.ExperimentTracking.Image = "willemmeints/experimenttracking:unknown"

		err := k8sClient.Update(ctx, workspace)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() error {
			deployment, err := getExperimentTrackingDeployment(workspace)

			if err != nil {
				return err
			}

			if deployment.Spec.Template.Spec.Containers[0].Image != "willemmeints/experimenttracking:unknown" {
				return fmt.Errorf("expected image to be 'willemmeints/experimenttracking:unknown', got %s", deployment.Spec.Template.Spec.Containers[0].Image)
			}

			return nil
		}, time.Minute, time.Second).Should(Succeed())
	})

	It(("Should vertically scale experiment tracking component"), func() {
		ctx := context.Background()
		workspace := createWorkspaceAndWaitForExperimentTrackingDeployment(ctx, "test-experimenttracking-resources")

		workspace.Spec.ExperimentTracking.Resources = corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("2"),
				corev1.ResourceMemory: resource.MustParse("2Gi"),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("1Gi"),
			},
		}

		err := k8sClient.Update(ctx, workspace)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() error {
			deployment, err := getExperimentTrackingDeployment(workspace)

			if err != nil {
				return err
			}

			if deployment.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().String() != "2" {
				return fmt.Errorf("expected cpu limit to be '2', got %s", deployment.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().String())
			}

			if deployment.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().String() != "2Gi" {
				return fmt.Errorf("expected memory limit to be '2Gi', got %s", deployment.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().String())
			}

			if deployment.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().String() != "1" {
				return fmt.Errorf("expected cpu request to be '1', got %s", deployment.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().String())
			}

			if deployment.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().String() != "1Gi" {
				return fmt.Errorf("expected memory request to be '1Gi', got %s", deployment.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().String())
			}

			return nil
		}, time.Minute, time.Second).Should(Succeed())
	})

	It("Should horizontally scale experiment tracking component", func() {
		ctx := context.Background()
		workspace := createWorkspaceAndWaitForExperimentTrackingDeployment(ctx, "test-experimenttracking-replicas")

		workspace.Spec.ExperimentTracking.Replicas = pointer.Int32(2)

		err := k8sClient.Update(ctx, workspace)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() error {
			deployment, err := getExperimentTrackingDeployment(workspace)

			if err != nil {
				return err
			}

			if *deployment.Spec.Replicas != 2 {
				return fmt.Errorf("expected replicas to be '2', got %d", *deployment.Spec.Replicas)
			}

			return nil
		}, time.Minute, time.Second).Should(Succeed())
	})
})

func getExperimentTrackingDeployment(workspace *mlopsv1alpha1.Workspace) (*appsv1.Deployment, error) {
	deploymentName := types.NamespacedName{
		Name:      fmt.Sprintf("%s-mlflow-server", workspace.GetName()),
		Namespace: workspace.GetNamespace(),
	}

	deployment := &appsv1.Deployment{}

	return deployment, k8sClient.Get(context.Background(), deploymentName, deployment)
}

func createWorkspaceAndWaitForExperimentTrackingDeployment(ctx context.Context, name string) *mlopsv1alpha1.Workspace {
	workspace := newTestWorkspace(name)

	err := k8sClient.Create(ctx, workspace)
	Expect(err).NotTo(HaveOccurred())

	Eventually(func() error {
		_, err := getExperimentTrackingDeployment(workspace)
		return err
	})

	return workspace
}
