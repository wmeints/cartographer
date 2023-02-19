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

			err := k8sClient.Get(ctx, deploymentName, deployment)

			if err != nil {
				return err
			}

			if deployment.OwnerReferences[0].Name != workspace.GetName() {
				return fmt.Errorf("expected owner reference to be %s, got %s", workspace.GetName(), deployment.OwnerReferences[0].Name)
			}

			return nil
		}, time.Minute, time.Second).Should(Succeed())

		Eventually(func() error {
			service := &corev1.Service{}
			serviceName := types.NamespacedName{
				Name:      fmt.Sprintf("%s-orion-server", workspace.GetName()),
				Namespace: "test-namespace",
			}

			err := k8sClient.Get(ctx, serviceName, service)

			if err != nil {
				return err
			}

			if service.OwnerReferences[0].Name != workspace.GetName() {
				return fmt.Errorf("expected owner reference to be %s, got %s", workspace.GetName(), service.OwnerReferences[0].Name)
			}

			return nil
		}, time.Minute, time.Second).Should(Succeed())

		Eventually(func() error {
			statefulSetName := types.NamespacedName{
				Name:      fmt.Sprintf("%s-agent-%s", workspace.GetName(), "test"),
				Namespace: "test-namespace",
			}

			statefulSet := &appsv1.StatefulSet{}

			err := k8sClient.Get(ctx, statefulSetName, statefulSet)

			if err != nil {
				return err
			}

			if statefulSet.OwnerReferences[0].Name != workspace.GetName() {
				return fmt.Errorf("expected owner reference to be %s, got %s", workspace.GetName(), statefulSet.OwnerReferences[0].Name)
			}

			return nil
		}, time.Minute, time.Second).Should(Succeed())
	})

	It("Should scale the workflow agents", func() {
		ctx := context.Background()

		workspace := createWorkspaceAndWaitForAgentPoolDeployment(ctx, "test-workflow-agents-scale")

		workspace.Spec.Workflows.Agents[0].Replicas = pointer.Int32(2)

		err := k8sClient.Update(ctx, workspace)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() error {
			statefulSet, err := getWorkflowAgentPool(workspace, "test")

			if err != nil {
				return err
			}

			if *statefulSet.Spec.Replicas != 2 {
				return fmt.Errorf("expected 2 replicas, got %d", *statefulSet.Spec.Replicas)
			}

			return nil
		}, time.Minute, time.Second).Should(Succeed())
	})

	It("Should update the image of the workflow agents", func() {
		ctx := context.Background()

		workspace := createWorkspaceAndWaitForAgentPoolDeployment(ctx, "test-workflow-agents-image")

		workspace.Spec.Workflows.Agents[0].Image = "willemmeints/workflow-agent:unknown"

		err := k8sClient.Update(ctx, workspace)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() error {
			statefulSet, err := getWorkflowAgentPool(workspace, "test")

			if err != nil {
				return err
			}

			if statefulSet.Spec.Template.Spec.Containers[0].Image != "willemmeints/workflow-agent:unknown" {
				return fmt.Errorf("expected image to be 'willemmeints/workflow-agent:unknown', got %s", statefulSet.Spec.Template.Spec.Containers[0].Image)
			}

			return nil
		}, time.Minute, time.Second).Should(Succeed())
	})

	It("Should vertically scale workflow agents", func() {
		ctx := context.Background()
		workspace := createWorkspaceAndWaitForAgentPoolDeployment(ctx, "test-workflow-agents-vertical-scaling")

		workspace.Spec.Workflows.Agents[0].Resources = corev1.ResourceRequirements{
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
			statefulSet, err := getWorkflowAgentPool(workspace, "test")

			if err != nil {
				return err
			}

			if statefulSet.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().String() != "2" {
				return fmt.Errorf("expected cpu limit to be '2', got %s", statefulSet.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().String())
			}

			if statefulSet.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().String() != "2Gi" {
				return fmt.Errorf("expected memory limit to be '2Gi', got %s", statefulSet.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().String())
			}

			if statefulSet.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().String() != "1" {
				return fmt.Errorf("expected cpu request to be '1', got %s", statefulSet.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().String())
			}

			if statefulSet.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().String() != "1Gi" {
				return fmt.Errorf("expected memory request to be '1Gi', got %s", statefulSet.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().String())
			}

			return nil
		}, time.Minute, time.Second).Should(Succeed())
	})

	It("Should horizontally scale workflow controllers", func() {
		ctx := context.Background()
		workspace := createWorkspaceAndWaitForAgentPoolDeployment(ctx, "test-workflow-controller-replicas")

		workspace.Spec.Workflows.Controller.Replicas = pointer.Int32(2)

		err := k8sClient.Update(ctx, workspace)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() error {
			typedDeploymentName := types.NamespacedName{
				Name:      fmt.Sprintf("%s-orion-server", workspace.GetName()),
				Namespace: "test-namespace",
			}

			deployment := appsv1.Deployment{}
			err := k8sClient.Get(ctx, typedDeploymentName, &deployment)

			if err != nil {
				return err
			}

			if *deployment.Spec.Replicas != 2 {
				return fmt.Errorf("expected 2 replicas, got %d", *deployment.Spec.Replicas)
			}

			return nil
		}, time.Minute, time.Second).Should(Succeed())
	})

	It("Should update the image of the workflow controller", func() {
		ctx := context.Background()
		workspace := createWorkspaceAndWaitForAgentPoolDeployment(ctx, "test-workflow-controller-image")

		workspace.Spec.Workflows.Controller.Image = "willemmeints/workflow-controller:unknown"

		err := k8sClient.Update(ctx, workspace)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() error {
			typedDeploymentName := types.NamespacedName{
				Name:      fmt.Sprintf("%s-orion-server", workspace.GetName()),
				Namespace: "test-namespace",
			}

			deployment := appsv1.Deployment{}
			err := k8sClient.Get(ctx, typedDeploymentName, &deployment)

			if err != nil {
				return err
			}

			if deployment.Spec.Template.Spec.Containers[0].Image != "willemmeints/workflow-controller:unknown" {
				return fmt.Errorf("expected image to be 'willemmeints/workflow-controller:unknown', got %s", deployment.Spec.Template.Spec.Containers[0].Image)
			}

			return nil
		}, time.Minute, time.Second).Should(Succeed())
	})

	It("Should vertically scale workflow controllers", func() {
		ctx := context.Background()
		workspace := createWorkspaceAndWaitForAgentPoolDeployment(ctx, "test-workflow-controller-resources")

		workspace.Spec.Workflows.Controller.Resources = corev1.ResourceRequirements{
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
			typedDeploymentName := types.NamespacedName{
				Name:      fmt.Sprintf("%s-orion-server", workspace.GetName()),
				Namespace: "test-namespace",
			}

			deployment := appsv1.Deployment{}
			err := k8sClient.Get(ctx, typedDeploymentName, &deployment)

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
})

func getWorkflowAgentPool(workspace *mlopsv1alpha1.Workspace, poolName string) (*appsv1.StatefulSet, error) {
	statefulSetName := fmt.Sprintf("%s-agent-%s", workspace.GetName(), poolName)

	typedStatefulSetName := types.NamespacedName{
		Name:      statefulSetName,
		Namespace: "test-namespace",
	}

	statefulSet := &appsv1.StatefulSet{}

	err := k8sClient.Get(ctx, typedStatefulSetName, statefulSet)

	if err != nil {
		return nil, err
	}

	return statefulSet, nil
}

func createWorkspaceAndWaitForAgentPoolDeployment(ctx context.Context, name string) *mlopsv1alpha1.Workspace {
	workspace := newTestWorkspace(name)

	err := k8sClient.Create(ctx, workspace)
	Expect(err).NotTo(HaveOccurred())

	Eventually(func() error {
		_, err := getWorkflowAgentPool(workspace, "test")
		return err
	})

	return workspace
}
