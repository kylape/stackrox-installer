package manifest

import (
	"testing"

	v1 "k8s.io/api/core/v1"
)

func TestGetEnvVarsForContainer(t *testing.T) {
	// Create a test configuration with environment variables
	config := &Config{
		EnvVars: EnvVarConfig{
			Global: []v1.EnvVar{
				{Name: "GLOBAL_VAR", Value: "global-value"},
				{Name: "DEPLOYMENT_ENV", Value: "test"},
			},
			Generators: map[string][]v1.EnvVar{
				"central": {
					{Name: "CENTRAL_LOG_LEVEL", Value: "debug"},
					{Name: "GLOBAL_VAR", Value: "generator-override"},
				},
			},
			Pods: map[string][]v1.EnvVar{
				"central": {
					{Name: "POD_MEMORY_LIMIT", Value: "8Gi"},
				},
			},
			Containers: map[string][]v1.EnvVar{
				"central": {
					{Name: "CENTRAL_LOG_LEVEL", Value: "trace"}, // This should override generator setting
					{Name: "CONTAINER_SPECIFIC", Value: "container-value"},
				},
			},
		},
	}

	// Test environment variable merging for central container
	existingEnvVars := []v1.EnvVar{
		{Name: "ROX_HOTRELOAD", Value: "true"},
		{Name: "CONTAINER_SPECIFIC", Value: "existing-value"}, // This should override custom config
	}

	t.Run("TestEnvironmentVariablePrecedence", func(t *testing.T) {
		mergedEnvVars := GetEnvVarsForContainer(config, "central", "central", "central", existingEnvVars)

		// Convert to map for easier testing
		envMap := make(map[string]string)
		for _, envVar := range mergedEnvVars {
			envMap[envVar.Name] = envVar.Value
		}

		// Test precedence rules
		tests := []struct {
			name     string
			key      string
			expected string
		}{
			{"Global variable should be overridden by generator", "GLOBAL_VAR", "generator-override"},
			{"Container variable should override generator", "CENTRAL_LOG_LEVEL", "trace"},
			{"Existing variable should override custom", "CONTAINER_SPECIFIC", "existing-value"},
			{"Existing system variable should be preserved", "ROX_HOTRELOAD", "true"},
			{"Global variable should be present", "DEPLOYMENT_ENV", "test"},
			{"Pod-specific variable should be present", "POD_MEMORY_LIMIT", "8Gi"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got, exists := envMap[tt.key]; !exists {
					t.Errorf("Expected environment variable %q to exist", tt.key)
				} else if got != tt.expected {
					t.Errorf("Expected %q = %q, got %q", tt.key, tt.expected, got)
				}
			})
		}
	})
}

func TestProtectedVariableValidation(t *testing.T) {
	testConfig := &Config{
		EnvVars: EnvVarConfig{
			Global: []v1.EnvVar{
				{Name: "ROX_PROTECTED", Value: "should-be-blocked"},
				{Name: "KUBERNETES_SERVICE_HOST", Value: "should-be-blocked"},
				{Name: "PATH", Value: "should-be-blocked"},
				{Name: "CUSTOM_VAR", Value: "should-be-allowed"},
			},
		},
	}

	t.Run("TestProtectedVariablesFiltered", func(t *testing.T) {
		validatedEnvVars := GetEnvVarsForContainer(testConfig, "test", "test", "test", []v1.EnvVar{})

		// Convert to map for easier testing
		envMap := make(map[string]string)
		for _, envVar := range validatedEnvVars {
			envMap[envVar.Name] = envVar.Value
		}

		// Test that protected variables are filtered out
		protectedVars := []string{"ROX_PROTECTED", "KUBERNETES_SERVICE_HOST", "PATH"}
		for _, protectedVar := range protectedVars {
			if _, exists := envMap[protectedVar]; exists {
				t.Errorf("Protected variable %q should be filtered out", protectedVar)
			}
		}

		// Test that allowed variables are preserved
		if value, exists := envMap["CUSTOM_VAR"]; !exists {
			t.Error("Custom variable CUSTOM_VAR should be allowed")
		} else if value != "should-be-allowed" {
			t.Errorf("Expected CUSTOM_VAR = 'should-be-allowed', got %q", value)
		}
	})
}

func TestEmptyConfiguration(t *testing.T) {
	emptyConfig := &Config{
		EnvVars: EnvVarConfig{
			Global:     []v1.EnvVar{},
			Generators: make(map[string][]v1.EnvVar),
			Pods:       make(map[string][]v1.EnvVar),
			Containers: make(map[string][]v1.EnvVar),
		},
	}

	t.Run("TestEmptyConfigurationReturnsExistingOnly", func(t *testing.T) {
		existingEnvVars := []v1.EnvVar{
			{Name: "EXISTING_VAR", Value: "existing-value"},
		}

		result := GetEnvVarsForContainer(emptyConfig, "test", "test", "test", existingEnvVars)

		if len(result) != 1 {
			t.Errorf("Expected 1 environment variable, got %d", len(result))
		}

		if result[0].Name != "EXISTING_VAR" || result[0].Value != "existing-value" {
			t.Errorf("Expected EXISTING_VAR = 'existing-value', got %s = %s", result[0].Name, result[0].Value)
		}
	})
}

func TestNonExistentScopes(t *testing.T) {
	config := &Config{
		EnvVars: EnvVarConfig{
			Global: []v1.EnvVar{
				{Name: "GLOBAL_VAR", Value: "global-value"},
			},
			Generators: map[string][]v1.EnvVar{
				"other_generator": {
					{Name: "OTHER_VAR", Value: "other-value"},
				},
			},
			Pods: map[string][]v1.EnvVar{
				"other_pod": {
					{Name: "OTHER_POD_VAR", Value: "other-pod-value"},
				},
			},
			Containers: map[string][]v1.EnvVar{
				"other_container": {
					{Name: "OTHER_CONTAINER_VAR", Value: "other-container-value"},
				},
			},
		},
	}

	t.Run("TestNonExistentScopesOnlyReturnGlobalAndExisting", func(t *testing.T) {
		existingEnvVars := []v1.EnvVar{
			{Name: "EXISTING_VAR", Value: "existing-value"},
		}

		result := GetEnvVarsForContainer(config, "nonexistent_generator", "nonexistent_pod", "nonexistent_container", existingEnvVars)

		// Convert to map for easier testing
		envMap := make(map[string]string)
		for _, envVar := range result {
			envMap[envVar.Name] = envVar.Value
		}

		// Should only have global and existing variables
		if len(envMap) != 2 {
			t.Errorf("Expected 2 environment variables, got %d", len(envMap))
		}

		expectedVars := map[string]string{
			"GLOBAL_VAR":   "global-value",
			"EXISTING_VAR": "existing-value",
		}

		for key, expectedValue := range expectedVars {
			if value, exists := envMap[key]; !exists {
				t.Errorf("Expected variable %q to exist", key)
			} else if value != expectedValue {
				t.Errorf("Expected %q = %q, got %q", key, expectedValue, value)
			}
		}

		// Should not have variables from other scopes
		unexpectedVars := []string{"OTHER_VAR", "OTHER_POD_VAR", "OTHER_CONTAINER_VAR"}
		for _, unexpectedVar := range unexpectedVars {
			if _, exists := envMap[unexpectedVar]; exists {
				t.Errorf("Unexpected variable %q should not exist", unexpectedVar)
			}
		}
	})
}