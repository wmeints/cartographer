package controllers

import (
	mlopsv1alpha1 "github.com/wmeints/cartographer/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
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
	}
}

func newComponentLabels(workspace *mlopsv1alpha1.Workspace, componentName string) map[string]string {
	return map[string]string{
		"mlops.aigency.com/environment": workspace.GetName(),
		"mlops.aigency.com/component":   componentName,
	}
}
