package manifest

import (
	"context"
	"reflect"
	"testing"

	"k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid config",
			config: &Config{
				Namespace: "stackrox",
			},
			wantError: false,
		},
		{
			name: "empty namespace",
			config: &Config{
				Namespace: "",
			},
			wantError: true,
			errorMsg:  "Invalid namespace: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := New(tt.config, nil, nil)
			if tt.wantError {
				if err == nil {
					t.Errorf("New() expected error but got none")
				}
				if err != nil && err.Error() != tt.errorMsg {
					t.Errorf("New() error = %v, want %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("New() unexpected error = %v", err)
				}
				if m == nil {
					t.Errorf("New() returned nil manifestGenerator")
				}
				if m != nil && m.Config != tt.config {
					t.Errorf("New() config not set correctly")
				}
			}
		})
	}
}

// Mock generators for testing
type mockGenerator struct {
	name     string
	priority int
}

type mockOrderedGenerator struct {
	mockGenerator
}

func (m mockGenerator) Generate(ctx context.Context, mg *manifestGenerator) ([]Resource, error) {
	return nil, nil
}
func (m mockGenerator) Name() string     { return m.name }
func (m mockGenerator) Exportable() bool { return true }

func (m mockOrderedGenerator) Priority() int { return m.priority }

func TestSortGeneratorsByPriority(t *testing.T) {

	gen1 := mockGenerator{name: "gen1"}
	gen2 := mockOrderedGenerator{mockGenerator{name: "gen2", priority: -1}}
	gen3 := mockOrderedGenerator{mockGenerator{name: "gen3", priority: 5}}
	gen4 := mockGenerator{name: "gen4"}

	generators := []Generator{gen1, gen2, gen3, gen4}
	sorted := sortGeneratorsByPriority(generators)

	// Expected order: gen2 (priority -1), gen1 (priority 0), gen4 (priority 0), gen3 (priority 5)
	expectedOrder := []string{"gen2", "gen1", "gen4", "gen3"}
	
	if len(sorted) != len(expectedOrder) {
		t.Errorf("sortGeneratorsByPriority() returned %d generators, want %d", len(sorted), len(expectedOrder))
	}

	for i, gen := range sorted {
		if gen.Name() != expectedOrder[i] {
			t.Errorf("sortGeneratorsByPriority() generator at index %d = %s, want %s", i, gen.Name(), expectedOrder[i])
		}
	}
}

func TestGenServiceAccount(t *testing.T) {
	name := "test-service-account"
	resource := genServiceAccount(name)

	if resource.Name != name {
		t.Errorf("genServiceAccount() name = %s, want %s", resource.Name, name)
	}

	if resource.IsUpdateable {
		t.Errorf("genServiceAccount() IsUpdateable = true, want false")
	}

	if resource.ClusterScoped {
		t.Errorf("genServiceAccount() ClusterScoped = true, want false")
	}

	sa, ok := resource.Object.(*v1.ServiceAccount)
	if !ok {
		t.Errorf("genServiceAccount() object is not a ServiceAccount")
	}

	if sa.Name != name {
		t.Errorf("genServiceAccount() ServiceAccount name = %s, want %s", sa.Name, name)
	}

	expectedGVK := v1.SchemeGroupVersion.WithKind("ServiceAccount")
	if sa.GroupVersionKind() != expectedGVK {
		t.Errorf("genServiceAccount() GVK = %v, want %v", sa.GroupVersionKind(), expectedGVK)
	}
}

func TestGenService(t *testing.T) {
	name := "test-service"
	ports := []v1.ServicePort{
		{
			Name: "http",
			Port: 8080,
		},
	}
	resource := genService(name, ports)

	if resource.Name != name {
		t.Errorf("genService() name = %s, want %s", resource.Name, name)
	}

	if !resource.IsUpdateable {
		t.Errorf("genService() IsUpdateable = false, want true")
	}

	if resource.ClusterScoped {
		t.Errorf("genService() ClusterScoped = true, want false")
	}

	svc, ok := resource.Object.(*v1.Service)
	if !ok {
		t.Errorf("genService() object is not a Service")
	}

	if svc.Name != name {
		t.Errorf("genService() Service name = %s, want %s", svc.Name, name)
	}

	if !reflect.DeepEqual(svc.Spec.Ports, ports) {
		t.Errorf("genService() Service ports = %v, want %v", svc.Spec.Ports, ports)
	}

	expectedSelector := map[string]string{"app": name}
	if !reflect.DeepEqual(svc.Spec.Selector, expectedSelector) {
		t.Errorf("genService() Service selector = %v, want %v", svc.Spec.Selector, expectedSelector)
	}
}

func TestGenClusterRoleBinding(t *testing.T) {
	serviceAccountName := "test-sa"
	roleName := "test-role"
	namespace := "test-ns"
	
	resource := genClusterRoleBinding(serviceAccountName, roleName, namespace)
	
	expectedName := "test-ns-test-sa-test-role"
	if resource.Name != expectedName {
		t.Errorf("genClusterRoleBinding() name = %s, want %s", resource.Name, expectedName)
	}

	if !resource.IsUpdateable {
		t.Errorf("genClusterRoleBinding() IsUpdateable = false, want true")
	}

	if !resource.ClusterScoped {
		t.Errorf("genClusterRoleBinding() ClusterScoped = false, want true")
	}

	crb, ok := resource.Object.(*rbacv1.ClusterRoleBinding)
	if !ok {
		t.Errorf("genClusterRoleBinding() object is not a ClusterRoleBinding")
	}

	if crb.Name != expectedName {
		t.Errorf("genClusterRoleBinding() ClusterRoleBinding name = %s, want %s", crb.Name, expectedName)
	}

	if crb.RoleRef.Name != roleName {
		t.Errorf("genClusterRoleBinding() RoleRef name = %s, want %s", crb.RoleRef.Name, roleName)
	}

	if len(crb.Subjects) != 1 {
		t.Errorf("genClusterRoleBinding() subjects count = %d, want 1", len(crb.Subjects))
	}

	if crb.Subjects[0].Name != serviceAccountName {
		t.Errorf("genClusterRoleBinding() subject name = %s, want %s", crb.Subjects[0].Name, serviceAccountName)
	}

	if crb.Subjects[0].Namespace != namespace {
		t.Errorf("genClusterRoleBinding() subject namespace = %s, want %s", crb.Subjects[0].Namespace, namespace)
	}
}

func TestGenRole(t *testing.T) {
	name := "test-role"
	rules := []rbacv1.PolicyRule{
		{
			APIGroups: []string{""},
			Resources: []string{"pods"},
			Verbs:     []string{"get", "list"},
		},
	}
	
	resource := genRole(name, rules)
	
	if resource.Name != name {
		t.Errorf("genRole() name = %s, want %s", resource.Name, name)
	}

	if !resource.IsUpdateable {
		t.Errorf("genRole() IsUpdateable = false, want true")
	}

	if resource.ClusterScoped {
		t.Errorf("genRole() ClusterScoped = true, want false")
	}

	role, ok := resource.Object.(*rbacv1.Role)
	if !ok {
		t.Errorf("genRole() object is not a Role")
	}

	if role.Name != name {
		t.Errorf("genRole() Role name = %s, want %s", role.Name, name)
	}

	if !reflect.DeepEqual(role.Rules, rules) {
		t.Errorf("genRole() Role rules = %v, want %v", role.Rules, rules)
	}
}

func TestToGVR(t *testing.T) {
	tests := []struct {
		name string
		gvk  schema.GroupVersionKind
		want schema.GroupVersionResource
	}{
		{
			name: "core api",
			gvk:  schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
			want: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"},
		},
		{
			name: "apps api",
			gvk:  schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
			want: schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
		},
		{
			name: "rbac api",
			gvk:  schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRole"},
			want: schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toGVR(tt.gvk)
			if got != tt.want {
				t.Errorf("toGVR() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRestrictedSecurityContext(t *testing.T) {
	user := int64(1000)
	sc := RestrictedSecurityContext(user)

	if sc.RunAsUser == nil || *sc.RunAsUser != user {
		t.Errorf("RestrictedSecurityContext() RunAsUser = %v, want %d", sc.RunAsUser, user)
	}

	if sc.RunAsGroup == nil || *sc.RunAsGroup != user {
		t.Errorf("RestrictedSecurityContext() RunAsGroup = %v, want %d", sc.RunAsGroup, user)
	}

	if sc.AllowPrivilegeEscalation == nil || *sc.AllowPrivilegeEscalation {
		t.Errorf("RestrictedSecurityContext() AllowPrivilegeEscalation = %v, want false", sc.AllowPrivilegeEscalation)
	}

	if sc.RunAsNonRoot == nil || !*sc.RunAsNonRoot {
		t.Errorf("RestrictedSecurityContext() RunAsNonRoot = %v, want true", sc.RunAsNonRoot)
	}

	if sc.SeccompProfile == nil || sc.SeccompProfile.Type != v1.SeccompProfileTypeRuntimeDefault {
		t.Errorf("RestrictedSecurityContext() SeccompProfile.Type = %v, want %v", sc.SeccompProfile.Type, v1.SeccompProfileTypeRuntimeDefault)
	}

	if sc.Capabilities == nil || len(sc.Capabilities.Drop) != 1 || sc.Capabilities.Drop[0] != "ALL" {
		t.Errorf("RestrictedSecurityContext() Capabilities.Drop = %v, want [ALL]", sc.Capabilities.Drop)
	}
}

func TestHotloadCommand(t *testing.T) {
	tests := []struct {
		name    string
		cmd     string
		devMode bool
		want    []string
	}{
		{
			name:    "dev mode enabled",
			cmd:     "my-command",
			devMode: true,
			want:    []string{"sh", "-c", "while true; do my-command; sleep 5; done"},
		},
		{
			name:    "dev mode disabled",
			cmd:     "my-command",
			devMode: false,
			want:    []string{"my-command"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{DevMode: tt.devMode}
			got := hotloadCommand(tt.cmd, config)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("hotloadCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVolumeDefAndMount_Apply(t *testing.T) {
	vdm := VolumeDefAndMount{
		Name:      "test-volume",
		MountPath: "/test/path",
		ReadOnly:  true,
		Volume: v1.Volume{
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		},
	}

	container := &v1.Container{}
	podSpec := &v1.PodSpec{}

	vdm.Apply(container, podSpec)

	// Check volume mount was added to container
	if len(container.VolumeMounts) != 1 {
		t.Errorf("Apply() container.VolumeMounts length = %d, want 1", len(container.VolumeMounts))
	}

	vm := container.VolumeMounts[0]
	if vm.Name != vdm.Name {
		t.Errorf("Apply() VolumeMount name = %s, want %s", vm.Name, vdm.Name)
	}
	if vm.MountPath != vdm.MountPath {
		t.Errorf("Apply() VolumeMount mountPath = %s, want %s", vm.MountPath, vdm.MountPath)
	}
	if vm.ReadOnly != vdm.ReadOnly {
		t.Errorf("Apply() VolumeMount readOnly = %v, want %v", vm.ReadOnly, vdm.ReadOnly)
	}

	// Check volume was added to pod spec
	if len(podSpec.Volumes) != 1 {
		t.Errorf("Apply() podSpec.Volumes length = %d, want 1", len(podSpec.Volumes))
	}

	vol := podSpec.Volumes[0]
	if vol.Name != vdm.Name {
		t.Errorf("Apply() Volume name = %s, want %s", vol.Name, vdm.Name)
	}
}

func TestConstants(t *testing.T) {
	if ReadOnlyMode != 0640 {
		t.Errorf("ReadOnlyMode = %d, want 0640", ReadOnlyMode)
	}

	if PostgresUser != 70 {
		t.Errorf("PostgresUser = %d, want 70", PostgresUser)
	}

	if ScannerUser != 65534 {
		t.Errorf("ScannerUser = %d, want 65534", ScannerUser)
	}

	expectedTwoGigs := resource.MustParse("2Gi")
	if !TwoGigs.Equal(expectedTwoGigs) {
		t.Errorf("TwoGigs = %v, want %v", TwoGigs, expectedTwoGigs)
	}
}