package manifest

import (
	"context"
	"testing"

	v1 "k8s.io/api/core/v1"
)

func TestNamespaceGenerator_Name(t *testing.T) {
	gen := NamespaceGenerator{}
	expected := "Namespace"
	if gen.Name() != expected {
		t.Errorf("NamespaceGenerator.Name() = %s, want %s", gen.Name(), expected)
	}
}

func TestNamespaceGenerator_Exportable(t *testing.T) {
	gen := NamespaceGenerator{}
	if !gen.Exportable() {
		t.Errorf("NamespaceGenerator.Exportable() = false, want true")
	}
}

func TestNamespaceGenerator_Generate(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
	}{
		{
			name:      "default namespace",
			namespace: "stackrox",
		},
		{
			name:      "custom namespace",
			namespace: "my-custom-namespace",
		},
		{
			name:      "test namespace",
			namespace: "test-ns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NamespaceGenerator{}
			config := &Config{
				Namespace: tt.namespace,
			}
			manifestGen := &manifestGenerator{
				Config: config,
			}

			ctx := context.Background()
			resources, err := gen.Generate(ctx, manifestGen)

			if err != nil {
				t.Errorf("NamespaceGenerator.Generate() error = %v", err)
				return
			}

			if len(resources) != 1 {
				t.Errorf("NamespaceGenerator.Generate() returned %d resources, want 1", len(resources))
				return
			}

			resource := resources[0]

			// Check resource properties
			if resource.Name != tt.namespace {
				t.Errorf("Resource.Name = %s, want %s", resource.Name, tt.namespace)
			}

			if resource.IsUpdateable {
				t.Errorf("Resource.IsUpdateable = true, want false")
			}

			if !resource.ClusterScoped {
				t.Errorf("Resource.ClusterScoped = false, want true")
			}

			// Check the Kubernetes object
			ns, ok := resource.Object.(*v1.Namespace)
			if !ok {
				t.Errorf("Resource.Object is not a *v1.Namespace, got %T", resource.Object)
				return
			}

			if ns.Name != tt.namespace {
				t.Errorf("Namespace.Name = %s, want %s", ns.Name, tt.namespace)
			}

			// Check GroupVersionKind
			expectedGVK := v1.SchemeGroupVersion.WithKind("Namespace")
			if ns.GroupVersionKind() != expectedGVK {
				t.Errorf("Namespace.GroupVersionKind() = %v, want %v", ns.GroupVersionKind(), expectedGVK)
			}
		})
	}
}

func TestNamespaceGenerator_GenerateWithEmptyNamespace(t *testing.T) {
	gen := NamespaceGenerator{}
	config := &Config{
		Namespace: "",
	}
	manifestGen := &manifestGenerator{
		Config: config,
	}

	ctx := context.Background()
	resources, err := gen.Generate(ctx, manifestGen)

	if err != nil {
		t.Errorf("NamespaceGenerator.Generate() error = %v", err)
		return
	}

	if len(resources) != 1 {
		t.Errorf("NamespaceGenerator.Generate() returned %d resources, want 1", len(resources))
		return
	}

	resource := resources[0]
	ns, ok := resource.Object.(*v1.Namespace)
	if !ok {
		t.Errorf("Resource.Object is not a *v1.Namespace")
		return
	}

	// Should create namespace with empty name (which would be invalid in Kubernetes)
	if ns.Name != "" {
		t.Errorf("Expected empty namespace name, got %s", ns.Name)
	}
}

func TestNamespaceGenerator_GenerateMultipleCalls(t *testing.T) {
	// Test that multiple calls to Generate produce identical results
	gen := NamespaceGenerator{}
	config := &Config{
		Namespace: "test-namespace",
	}
	manifestGen := &manifestGenerator{
		Config: config,
	}

	ctx := context.Background()

	// First call
	resources1, err1 := gen.Generate(ctx, manifestGen)
	if err1 != nil {
		t.Errorf("First Generate() call error = %v", err1)
		return
	}

	// Second call
	resources2, err2 := gen.Generate(ctx, manifestGen)
	if err2 != nil {
		t.Errorf("Second Generate() call error = %v", err2)
		return
	}

	// Compare results
	if len(resources1) != len(resources2) {
		t.Errorf("Different number of resources: first=%d, second=%d", len(resources1), len(resources2))
		return
	}

	if len(resources1) == 0 {
		t.Errorf("No resources generated")
		return
	}

	resource1 := resources1[0]
	resource2 := resources2[0]

	if resource1.Name != resource2.Name {
		t.Errorf("Resource names differ: first=%s, second=%s", resource1.Name, resource2.Name)
	}

	if resource1.IsUpdateable != resource2.IsUpdateable {
		t.Errorf("Resource IsUpdateable differs: first=%t, second=%t", resource1.IsUpdateable, resource2.IsUpdateable)
	}

	if resource1.ClusterScoped != resource2.ClusterScoped {
		t.Errorf("Resource ClusterScoped differs: first=%t, second=%t", resource1.ClusterScoped, resource2.ClusterScoped)
	}

	ns1, ok1 := resource1.Object.(*v1.Namespace)
	ns2, ok2 := resource2.Object.(*v1.Namespace)

	if !ok1 || !ok2 {
		t.Errorf("Resources are not Namespace objects")
		return
	}

	if ns1.Name != ns2.Name {
		t.Errorf("Namespace names differ: first=%s, second=%s", ns1.Name, ns2.Name)
	}

	if ns1.GroupVersionKind() != ns2.GroupVersionKind() {
		t.Errorf("Namespace GVKs differ: first=%v, second=%v", ns1.GroupVersionKind(), ns2.GroupVersionKind())
	}
}

func TestNamespaceGenerator_InterfaceCompliance(t *testing.T) {
	// Test that NamespaceGenerator implements the Generator interface
	var _ Generator = NamespaceGenerator{}
}

func TestNamespaceGenerator_ResourceValidation(t *testing.T) {
	gen := NamespaceGenerator{}
	config := &Config{
		Namespace: "validation-test",
	}
	manifestGen := &manifestGenerator{
		Config: config,
	}

	ctx := context.Background()
	resources, err := gen.Generate(ctx, manifestGen)

	if err != nil {
		t.Errorf("Generate() error = %v", err)
		return
	}

	if len(resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(resources))
		return
	}

	resource := resources[0]

	// Validate that the resource has all required fields
	if resource.Object == nil {
		t.Errorf("Resource.Object is nil")
	}

	if resource.Name == "" {
		t.Errorf("Resource.Name is empty")
	}

	// For namespace resources, should be cluster-scoped and not updateable
	if resource.IsUpdateable {
		t.Errorf("Namespace should not be updateable")
	}

	if !resource.ClusterScoped {
		t.Errorf("Namespace should be cluster-scoped")
	}

	// Validate the namespace object itself
	ns, ok := resource.Object.(*v1.Namespace)
	if !ok {
		t.Errorf("Resource.Object should be *v1.Namespace, got %T", resource.Object)
		return
	}

	if ns.Name == "" {
		t.Errorf("Namespace.Name should not be empty")
	}

	// Check that GroupVersionKind is set
	gvk := ns.GroupVersionKind()
	if gvk.Group != "" {
		t.Errorf("Expected empty group for core API, got %s", gvk.Group)
	}
	if gvk.Version != "v1" {
		t.Errorf("Expected version v1, got %s", gvk.Version)
	}
	if gvk.Kind != "Namespace" {
		t.Errorf("Expected kind Namespace, got %s", gvk.Kind)
	}
}