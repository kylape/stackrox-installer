package manifest

import (
	"context"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func init() {
	securedCluster = append(securedCluster, CentralProxyGenerator{})
}

type CentralProxyGenerator struct{}

func (g CentralProxyGenerator) Name() string {
	return "CentralProxy"
}

func (g CentralProxyGenerator) Exportable() bool {
	return true
}

func (g CentralProxyGenerator) Generate(ctx context.Context, m *manifestGenerator) ([]Resource, error) {
	return []Resource{
		g.genServiceAccount(m),
		g.genClusterRole(m),
		g.genClusterRoleBinding(m),
		g.genDeployment(m),
		g.genService(m),
	}, nil
}

func (g CentralProxyGenerator) genServiceAccount(m *manifestGenerator) Resource {
	sa := v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "central-proxy",
			Namespace: m.Config.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":     "central-proxy",
				"app.kubernetes.io/instance": "stackrox",
				"app.kubernetes.io/part-of":  "stackrox",
			},
		},
	}
	sa.SetGroupVersionKind(v1.SchemeGroupVersion.WithKind("ServiceAccount"))
	return Resource{
		Object: &sa,
	}
}

func (g CentralProxyGenerator) genClusterRole(m *manifestGenerator) Resource {
	cr := rbac.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "central-proxy",
			Labels: map[string]string{
				"app.kubernetes.io/name":     "central-proxy",
				"app.kubernetes.io/instance": "stackrox",
				"app.kubernetes.io/part-of":  "stackrox",
			},
		},
		Rules: []rbac.PolicyRule{
			{
				APIGroups: []string{"authorization.k8s.io"},
				Resources: []string{"subjectaccessreviews"},
				Verbs:     []string{"create"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"users"},
				Verbs:     []string{"get"},
			},
			{
				APIGroups: []string{"security.stackrox.io"},
				Resources: []string{"images", "vulnerabilities", "policies", "violations"},
				Verbs:     []string{"list", "get"},
			},
		},
	}
	cr.SetGroupVersionKind(rbac.SchemeGroupVersion.WithKind("ClusterRole"))
	return Resource{
		Object: &cr,
	}
}

func (g CentralProxyGenerator) genClusterRoleBinding(m *manifestGenerator) Resource {
	crb := rbac.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "central-proxy",
			Labels: map[string]string{
				"app.kubernetes.io/name":     "central-proxy",
				"app.kubernetes.io/instance": "stackrox",
				"app.kubernetes.io/part-of":  "stackrox",
			},
		},
		RoleRef: rbac.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "central-proxy",
		},
		Subjects: []rbac.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "central-proxy",
				Namespace: m.Config.Namespace,
			},
		},
	}
	crb.SetGroupVersionKind(rbac.SchemeGroupVersion.WithKind("ClusterRoleBinding"))
	return Resource{
		Object: &crb,
	}
}

func (g CentralProxyGenerator) genDeployment(m *manifestGenerator) Resource {
	replicas := int32(1)
	
	deployment := apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "central-proxy",
			Namespace: m.Config.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":     "central-proxy",
				"app.kubernetes.io/instance": "stackrox",
				"app.kubernetes.io/part-of":  "stackrox",
			},
		},
		Spec: apps.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "central-proxy",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":                        "central-proxy",
						"app.kubernetes.io/name":     "central-proxy",
						"app.kubernetes.io/instance": "stackrox",
						"app.kubernetes.io/part-of":  "stackrox",
					},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: "central-proxy",
					SecurityContext: &v1.PodSecurityContext{
						RunAsNonRoot: &[]bool{true}[0],
						RunAsUser:    &[]int64{4000}[0],
						FSGroup:      &[]int64{4000}[0],
					},
					Containers: []v1.Container{
						{
							Name:  "central-proxy",
							Image: m.Config.Images.CentralProxy,
							Ports: []v1.ContainerPort{
								{
									Name:          "https",
									ContainerPort: 8080,
									Protocol:      v1.ProtocolTCP,
								},
							},
							Env: []v1.EnvVar{
								{
									Name:  "CENTRAL_PROXY_PORT",
									Value: "8080",
								},
								{
									Name:  "CENTRAL_ENDPOINT",
									Value: "central." + m.Config.Namespace + ".svc:443",
								},
								{
									Name:  "CERTIFICATE_PATH",
									Value: "/etc/ssl/certs",
								},
								{
									Name: "NAMESPACE",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("100m"),
									v1.ResourceMemory: resource.MustParse("128Mi"),
								},
								Limits: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("500m"),
									v1.ResourceMemory: resource.MustParse("512Mi"),
								},
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "central-proxy-certs",
									MountPath: "/etc/ssl/certs",
									ReadOnly:  true,
								},
							},
							LivenessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									HTTPGet: &v1.HTTPGetAction{
										Path:   "/health",
										Port:   intstr.FromInt(8080),
										Scheme: v1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: 30,
								PeriodSeconds:       10,
								TimeoutSeconds:      5,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
							ReadinessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									HTTPGet: &v1.HTTPGetAction{
										Path:   "/ready",
										Port:   intstr.FromInt(8080),
										Scheme: v1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       5,
								TimeoutSeconds:      3,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
							SecurityContext: &v1.SecurityContext{
								AllowPrivilegeEscalation: &[]bool{false}[0],
								ReadOnlyRootFilesystem:   &[]bool{true}[0],
								RunAsNonRoot:             &[]bool{true}[0],
								RunAsUser:                &[]int64{4000}[0],
								Capabilities: &v1.Capabilities{
									Drop: []v1.Capability{"ALL"},
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "central-proxy-certs",
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: "tls-cert-central-proxy",
								},
							},
						},
					},
				},
			},
		},
	}
	deployment.SetGroupVersionKind(apps.SchemeGroupVersion.WithKind("Deployment"))
	return Resource{
		Object: &deployment,
	}
}

func (g CentralProxyGenerator) genService(m *manifestGenerator) Resource {
	service := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "central-proxy",
			Namespace: m.Config.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":     "central-proxy",
				"app.kubernetes.io/instance": "stackrox",
				"app.kubernetes.io/part-of":  "stackrox",
			},
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"app": "central-proxy",
			},
			Ports: []v1.ServicePort{
				{
					Name:       "https",
					Port:       8080,
					Protocol:   v1.ProtocolTCP,
					TargetPort: intstr.FromString("https"),
				},
			},
			Type: v1.ServiceTypeClusterIP,
		},
	}
	service.SetGroupVersionKind(v1.SchemeGroupVersion.WithKind("Service"))
	return Resource{
		Object: &service,
	}
}