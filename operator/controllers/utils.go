package controllers

import (
	mlopsv1alpha1 "github.com/wmeints/cartographer/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newDatabaseSecretEnvVars(databaseSecretName string) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name: "DB_HOST",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: "host",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: databaseSecretName,
					},
				},
			},
		},
		{
			Name: "DB_PORT",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: "port",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: databaseSecretName,
					},
				},
			},
		},
		{
			Name: "DB_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: "user",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: databaseSecretName,
					},
				},
			},
		},
		{
			Name: "DB_PASS",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: "password",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: databaseSecretName,
					},
				},
			},
		},
		{
			Name: "DB_NAME",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: "dbname",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: databaseSecretName,
					},
				},
			},
		},
	}
}

func newComponentLabels(workspace *mlopsv1alpha1.Workspace, componentName string) map[string]string {
	return map[string]string{
		"mlops.aigency.com/environment": workspace.GetName(),
		"mlops.aigency.com/component":   componentName,
	}
}

func newContainer(name string, image string, resources corev1.ResourceRequirements) corev1.Container {
	return corev1.Container{
		Name:            name,
		Image:           image,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Resources:       resources,
	}
}

func newDeployment(namespaceName string, deploymentName string, deploymentLabels map[string]string, replicas *int32, container corev1.Container) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: namespaceName,
			Labels:    deploymentLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: deploymentLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: deploymentLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						container,
					},
				},
			},
		},
	}
}

func newStatefulSet(namespaceName string, statefulSetName string, statefulSetLabels map[string]string, replicas *int32, container corev1.Container) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      statefulSetName,
			Namespace: namespaceName,
			Labels:    statefulSetLabels,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: statefulSetLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: statefulSetLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						container,
					},
				},
			},
		},
	}
}

func newService(name string, namespace string, serviceLabels map[string]string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: serviceLabels,
		},
	}
}
