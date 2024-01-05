package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func log(msg string, params ...interface{}) {
	fmt.Printf(msg+"\n", params...)
}

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	ctx := context.Background()
	pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

	// Create the target namespace
	log("Creating namespace")
	err = createNamespace(ctx, clientset)
	if err != nil && !errors.IsAlreadyExists(err) {
		panic(err)
	}

	// Create the central-db PVC
	log("Creating DB PVC")
	err = createCentralDbPvc(ctx, clientset)
	if err != nil && !errors.IsAlreadyExists(err) {
		panic(err)
	}

	// Create the secrets
	log("Creating admin password")
	err = createAdminPassword(ctx, clientset)
	if err != nil && !errors.IsAlreadyExists(err) {
		panic(err)
	}

	// Create the central-db deployment
	log("Creating central DB deployment")
	err = createCentralDbDeployment(ctx, clientset)
	if err != nil {
		panic(err)
	}

	// Create the central deployment
	log("Creating central deployment")
	err = createCentralDeployment(ctx, clientset)
	if err != nil {
		panic(err)
	}
}

func createNamespace(ctx context.Context, client *kubernetes.Clientset) error {
	ns := v1.Namespace{}
	ns.SetName("stackrox")
	_, err := client.CoreV1().Namespaces().Create(ctx, &ns, metav1.CreateOptions{})

	return err
}

func createCentralDbPvc(ctx context.Context, client *kubernetes.Clientset) error {
	pvc := v1.PersistentVolumeClaim{
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{"ReadWriteOnce"},
			Resources: v1.VolumeResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		},
	}
	pvc.SetName("central-db")
	_, err := client.CoreV1().PersistentVolumeClaims("stackrox").Create(ctx, &pvc, metav1.CreateOptions{})

	return err
}

func createAdminPassword(ctx context.Context, client *kubernetes.Clientset) error {
	secret := v1.Secret{
		StringData: map[string]string{
			"password": "letmein",
		},
	}
	secret.SetName("admin-pass")
	_, err := client.CoreV1().Secrets("stackrox").Create(ctx, &secret, metav1.CreateOptions{})

	return err
}

type VolumeDefAndMount struct {
	Name      string
	MountPath string
	ReadOnly  bool
	Volume    v1.Volume
}

func (v VolumeDefAndMount) Apply(c *v1.Container, spec *v1.PodSpec) {
	c.VolumeMounts = append(c.VolumeMounts, v1.VolumeMount{
		Name:      v.Name,
		MountPath: v.MountPath,
		ReadOnly:  v.ReadOnly,
	})
	v.Volume.Name = v.Name
	spec.Volumes = append(spec.Volumes, v.Volume)
}

func createCentralDbDeployment(ctx context.Context, client *kubernetes.Clientset) error {
	deployment := apps.Deployment{
		Spec: apps.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "central-db",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "central-db",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{
						Name:  "central-db",
						Image: "quay.io/stackrox-io/central-db:latest",
						Env: []v1.EnvVar{
							{
								Name:  "POSTGRES_HOST_AUTH_METHOD",
								Value: "password",
							},
							{
								Name:  "PGDATA",
								Value: "/var/lib/postgresql/data/pgdata",
							},
						},
					}},
					InitContainers: []v1.Container{{
						Name:    "init-db",
						Image:   "quay.io/stackrox-io/central-db:latest",
						Command: []string{"init-entrypoint.sh"},
						Env: []v1.EnvVar{
							{
								Name:  "PGDATA",
								Value: "/var/lib/postgresql/data/pgdata",
							},
						},
						VolumeMounts: []v1.VolumeMount{
							{
								Name:      "disk",
								MountPath: "/var/lib/postgresql/data",
							},
						},
					}},
				},
			},
		},
	}
	volumeMounts := []VolumeDefAndMount{
		{
			Name:      "config-volume",
			MountPath: "/etc/stackrox.d/config/",
			Volume: v1.Volume{
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "central-db-config",
						},
					},
				},
			},
		},
		{
			Name:      "disk",
			MountPath: "/var/lib/postgresql/data",
			Volume: v1.Volume{
				VolumeSource: v1.VolumeSource{
					PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
						ClaimName: "central-db",
					},
				},
			},
		},
		{
			Name:      "central-db-tls-volume",
			MountPath: "/run/secrets/stackrox.io/certs",
			Volume: v1.Volume{
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: "central-db-tls",
						Items: []v1.KeyToPath{
							{
								Key:  "cert.pem",
								Path: "server.crt",
							},
							{
								Key:  "key.pem",
								Path: "server.key",
							},
							{
								Key:  "ca.pem",
								Path: "root.crt",
							},
						},
					},
				},
			},
		},
		{
			Name:      "shared-memory",
			MountPath: "/dev/shm",
			Volume: v1.Volume{
				VolumeSource: v1.VolumeSource{
					EmptyDir: &v1.EmptyDirVolumeSource{},
				},
			},
		},
	}

	dbPasswd := VolumeDefAndMount{
		Name:      "central-db-password",
		MountPath: "/run/secrets/stackrox.io/secrets",
		Volume: v1.Volume{
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: "central-db-password",
				},
			},
		},
	}

	dbPasswd.Apply(&deployment.Spec.Template.Spec.Containers[0], &deployment.Spec.Template.Spec)

	for _, v := range volumeMounts {
		v.Apply(&deployment.Spec.Template.Spec.Containers[0], &deployment.Spec.Template.Spec)
	}

	deployment.SetName("central-db")
	_, err := client.AppsV1().Deployments("stackrox").Create(ctx, &deployment, metav1.CreateOptions{})

	return err
}

func createCentralDeployment(ctx context.Context, client *kubernetes.Clientset) error {
	deployment := apps.Deployment{
		Spec: apps.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "central",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "central",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{
						Name:    "central",
						Image:   "quay.io/stackrox-io/main:latest",
						Command: []string{"/stackrox/central-entrypoint.sh"},
						Env: []v1.EnvVar{
							{
								Name: "ROX_NAMESPACE",
								ValueFrom: &v1.EnvVarSource{
									FieldRef: &v1.ObjectFieldSelector{
										FieldPath: "metadata.namespace",
									},
								},
							},
						},
					}},
				},
			},
		},
	}

	trueBool := true
	volumeMounts := []VolumeDefAndMount{
		{
			Name:      "varlog",
			MountPath: "/var/log/stackrox/",
			Volume: v1.Volume{
				VolumeSource: v1.VolumeSource{
					EmptyDir: &v1.EmptyDirVolumeSource{},
				},
			},
		},
		{
			Name:      "central-tmp-volume",
			MountPath: "/tmp",
			Volume: v1.Volume{
				VolumeSource: v1.VolumeSource{
					EmptyDir: &v1.EmptyDirVolumeSource{},
				},
			},
		},
		{
			Name:      "central-etc-ssl-volume",
			MountPath: "/etc/ssl",
			Volume: v1.Volume{
				VolumeSource: v1.VolumeSource{
					EmptyDir: &v1.EmptyDirVolumeSource{},
				},
			},
		},
		{
			Name:      "central-etc-pki-volume",
			MountPath: "/etc/pki/ca-trust",
			Volume: v1.Volume{
				VolumeSource: v1.VolumeSource{
					EmptyDir: &v1.EmptyDirVolumeSource{},
				},
			},
		},
		{
			Name:      "central-certs-volume",
			MountPath: "/run/secrets/stackrox.io/certs/",
			ReadOnly:  true,
			Volume: v1.Volume{
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: "central-tls",
					},
				},
			},
		},
		{
			Name:      "central-default-tls-cert-volume",
			MountPath: "/run/secrets/stackrox.io/default-tls-cert/",
			ReadOnly:  true,
			Volume: v1.Volume{
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: "central-default-tls-cert",
						Optional:   &trueBool,
					},
				},
			},
		},
		{
			Name:      "central-htpasswd-volume",
			MountPath: "/run/secrets/stackrox.io/htpasswd/",
			ReadOnly:  true,
			Volume: v1.Volume{
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: "central-htpasswd",
						Optional:   &trueBool,
					},
				},
			},
		},
		{
			Name:      "central-jwt-volume",
			MountPath: "/run/secrets/stackrox.io/jwt/",
			ReadOnly:  true,
			Volume: v1.Volume{
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: "central-tls",
						Items: []v1.KeyToPath{
							{
								Key:  "jwt-key.pem",
								Path: "jwt-key.pem",
							},
						},
					},
				},
			},
		},
		{
			Name:      "additional-ca-volume",
			MountPath: "/usr/local/share/ca-certificates/",
			ReadOnly:  true,
			Volume: v1.Volume{
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: "additional-ca",
						Optional:   &trueBool,
					},
				},
			},
		},
		{
			Name:      "central-license-volume",
			MountPath: "/run/secrets/stackrox.io/central-license/",
			ReadOnly:  true,
			Volume: v1.Volume{
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: "central-license",
						Optional:   &trueBool,
					},
				},
			},
		},
		{
			Name:      "central-config-volume",
			MountPath: "/etc/stackrox",
			Volume: v1.Volume{
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "central-config",
						},
					},
				},
			},
		},
		{
			Name:      "proxy-config-volume",
			MountPath: "/run/secrets/stackrox.io/proxy-config/",
			ReadOnly:  true,
			Volume: v1.Volume{
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: "proxy-config",
						Optional:   &trueBool,
					},
				},
			},
		},
		{
			Name:      "endpoints-config-volume",
			MountPath: "/etc/stackrox.d/endpoints/",
			ReadOnly:  true,
			Volume: v1.Volume{
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "central-endpoints",
						},
					},
				},
			},
		},
	}

	for _, v := range volumeMounts {
		v.Apply(&deployment.Spec.Template.Spec.Containers[0], &deployment.Spec.Template.Spec)
	}

	deployment.SetName("central")

	_, err := client.AppsV1().Deployments("stackrox").Create(ctx, &deployment, metav1.CreateOptions{})

	return err
}
