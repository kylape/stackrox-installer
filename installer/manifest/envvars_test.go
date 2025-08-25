package manifest

import (
	"testing"

	v1 "k8s.io/api/core/v1"
)

func TestGetEnvVarsForContainer_StackRoxPattern(t *testing.T) {
	// Create a test configuration following StackRox Helm pattern
	config := &Config{
		Customize: CustomizeConfig{
			// Global env vars directly under customize.envVars
			EnvVars: []interface{}{
				map[string]interface{}{
					"name":  "DEPLOYMENT_ENV",
					"value": "test",
				},
				map[string]interface{}{
					"name":  "LOG_FORMAT",
					"value": "json",
				},
			},
			// Central deployment customization (true StackRox pattern)
			Central: &DeploymentCustomization{
				EnvVars: map[string]interface{}{
					// Deployment-level env vars under "envVars" key
					"envVars": []interface{}{
						map[string]interface{}{
							"name":  "CENTRAL_LOG_LEVEL",
							"value": "debug",
						},
						map[string]interface{}{
							"name":  "DEPLOYMENT_ENV",
							"value": "staging", // Overrides global
						},
					},
					// Container-specific env vars using /containerName syntax in same map
					"/central": []interface{}{
						map[string]interface{}{
							"name":  "CENTRAL_LOG_LEVEL",
							"value": "trace", // Overrides deployment setting
						},
						map[string]interface{}{
							"name":  "CONTAINER_SPECIFIC",
							"value": "container-value",
						},
					},
					"/nodejs": []interface{}{
						map[string]interface{}{
							"name":  "NODE_ENV",
							"value": "development",
						},
					},
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

		// Test precedence rules: Global < Deployment < Container-specific < Existing
		tests := []struct {
			name     string
			key      string
			expected string
		}{
			{"Global variable should be overridden by deployment", "DEPLOYMENT_ENV", "staging"},
			{"Container variable should override deployment", "CENTRAL_LOG_LEVEL", "trace"},
			{"Existing variable should override custom", "CONTAINER_SPECIFIC", "existing-value"},
			{"Existing system variable should be preserved", "ROX_HOTRELOAD", "true"},
			{"Global variable should be present", "LOG_FORMAT", "json"},
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

	t.Run("TestContainerSpecificEnvVars", func(t *testing.T) {
		// Test nodejs container gets its specific env vars
		nodejsEnvVars := GetEnvVarsForContainer(config, "central", "central", "nodejs", []v1.EnvVar{})

		envMap := make(map[string]string)
		for _, envVar := range nodejsEnvVars {
			envMap[envVar.Name] = envVar.Value
		}

		// Should have global, deployment, and nodejs-specific vars
		expectedVars := map[string]string{
			"DEPLOYMENT_ENV":    "staging",     // Global overridden by deployment
			"LOG_FORMAT":        "json",        // Global
			"CENTRAL_LOG_LEVEL": "debug",       // Deployment-specific
			"NODE_ENV":          "development", // Container-specific
		}

		for key, expectedValue := range expectedVars {
			if value, exists := envMap[key]; !exists {
				t.Errorf("Expected environment variable %q to exist for nodejs container", key)
			} else if value != expectedValue {
				t.Errorf("Expected %q = %q for nodejs container, got %q", key, expectedValue, value)
			}
		}
	})
}

func TestProtectedVariableValidation(t *testing.T) {
	testConfig := &Config{
		Customize: CustomizeConfig{
			EnvVars: []interface{}{
				map[string]interface{}{
					"name":  "ROX_PROTECTED",
					"value": "should-be-blocked",
				},
				map[string]interface{}{
					"name":  "KUBERNETES_SERVICE_HOST",
					"value": "should-be-blocked",
				},
				map[string]interface{}{
					"name":  "PATH",
					"value": "should-be-blocked",
				},
				map[string]interface{}{
					"name":  "CUSTOM_VAR",
					"value": "should-be-allowed",
				},
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
		Customize: CustomizeConfig{
			EnvVars: []interface{}{},
			Other:   make(map[string]*DeploymentCustomization),
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

func TestNonExistentDeployments(t *testing.T) {
	config := &Config{
		Customize: CustomizeConfig{
			EnvVars: []interface{}{
				map[string]interface{}{
					"name":  "GLOBAL_VAR",
					"value": "global-value",
				},
			},
			Sensor: &DeploymentCustomization{
				EnvVars: map[string]interface{}{
					"envVars": []interface{}{
						map[string]interface{}{
							"name":  "SENSOR_VAR",
							"value": "sensor-value",
						},
					},
				},
			},
		},
	}

	t.Run("TestNonExistentDeploymentsOnlyReturnGlobalAndExisting", func(t *testing.T) {
		existingEnvVars := []v1.EnvVar{
			{Name: "EXISTING_VAR", Value: "existing-value"},
		}

		result := GetEnvVarsForContainer(config, "nonexistent_deployment", "nonexistent_pod", "nonexistent_container", existingEnvVars)

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

		// Should not have variables from other deployments
		unexpectedVars := []string{"SENSOR_VAR"}
		for _, unexpectedVar := range unexpectedVars {
			if _, exists := envMap[unexpectedVar]; exists {
				t.Errorf("Unexpected variable %q should not exist", unexpectedVar)
			}
		}
	})
}

func TestEnvVarWithValueFrom(t *testing.T) {
	config := &Config{
		Customize: CustomizeConfig{
			EnvVars: []interface{}{
				map[string]interface{}{
					"name": "POD_NAME",
					"valueFrom": map[string]interface{}{
						"fieldRef": map[string]interface{}{
							"fieldPath": "metadata.name",
						},
					},
				},
				map[string]interface{}{
					"name": "MEMORY_LIMIT",
					"valueFrom": map[string]interface{}{
						"resourceFieldRef": map[string]interface{}{
							"resource": "limits.memory",
						},
					},
				},
			},
		},
	}

	t.Run("TestValueFromFieldRef", func(t *testing.T) {
		result := GetEnvVarsForContainer(config, "test", "test", "test", []v1.EnvVar{})

		var podNameVar *v1.EnvVar
		var memoryLimitVar *v1.EnvVar

		for _, envVar := range result {
			if envVar.Name == "POD_NAME" {
				podNameVar = &envVar
			}
			if envVar.Name == "MEMORY_LIMIT" {
				memoryLimitVar = &envVar
			}
		}

		if podNameVar == nil {
			t.Fatal("POD_NAME environment variable not found")
		}

		if podNameVar.ValueFrom == nil || podNameVar.ValueFrom.FieldRef == nil {
			t.Error("POD_NAME should have valueFrom.fieldRef")
		} else if podNameVar.ValueFrom.FieldRef.FieldPath != "metadata.name" {
			t.Errorf("Expected fieldPath 'metadata.name', got '%s'", podNameVar.ValueFrom.FieldRef.FieldPath)
		}

		if memoryLimitVar == nil {
			t.Fatal("MEMORY_LIMIT environment variable not found")
		}

		if memoryLimitVar.ValueFrom == nil || memoryLimitVar.ValueFrom.ResourceFieldRef == nil {
			t.Error("MEMORY_LIMIT should have valueFrom.resourceFieldRef")
		} else if memoryLimitVar.ValueFrom.ResourceFieldRef.Resource != "limits.memory" {
			t.Errorf("Expected resource 'limits.memory', got '%s'", memoryLimitVar.ValueFrom.ResourceFieldRef.Resource)
		}
	})
}