package manifest

import (
	"os"
	"reflect"
	"strings"
	"testing"
	"path/filepath"
)

func TestDefaultConfig(t *testing.T) {
	expected := Config{
		Namespace:            "stackrox",
		ScannerV4:            false,
		DevMode:              false,
		ApplyNetworkPolicies: false,
		CertPath:             "./certs",
		ImageArchitecture:    "single",
		Customize: CustomizeConfig{
			EnvVars: []interface{}{},
			Other:   make(map[string]*DeploymentCustomization),
		},
		Images: Images{
			AdmissionControl: localStackroxImage,
			Sensor:           localStackroxImage,
			Collector:        localStackroxImage,
			Compliance:       localStackroxImage,
			ConfigController: localStackroxImage,
			Central:          localStackroxImage,
			Scanner:          localStackroxImage,
			ScannerV4:        localStackroxImage,
			VSOCKListener:    localStackroxImage,
			CentralDB:        localDbImage,
			ScannerDB:        localDbImage,
			ScannerV4DB:      localDbImage,
		},
	}

	if !reflect.DeepEqual(DefaultConfig, expected) {
		t.Errorf("DefaultConfig does not match expected values")
		t.Errorf("Got: %+v", DefaultConfig)
		t.Errorf("Expected: %+v", expected)
	}
}

func TestReadConfig_EmptyFilename(t *testing.T) {
	cfg, err := ReadConfig("")
	if err != nil {
		t.Errorf("ReadConfig(\"\") returned error: %v", err)
	}

	if cfg == nil {
		t.Errorf("ReadConfig(\"\") returned nil config")
		return
	}

	// Should return a copy of DefaultConfig
	if !reflect.DeepEqual(*cfg, DefaultConfig) {
		t.Errorf("ReadConfig(\"\") did not return DefaultConfig")
	}

	// Verify it's a copy, not the same instance
	cfg.Namespace = "modified"
	if DefaultConfig.Namespace == "modified" {
		t.Errorf("ReadConfig(\"\") returned reference to DefaultConfig instead of copy")
	}
}

func TestReadConfig_NonExistentFile(t *testing.T) {
	_, err := ReadConfig("/nonexistent/file.yaml")
	if err == nil {
		t.Errorf("ReadConfig() should return error for non-existent file")
	}
}

func TestReadConfig_ValidFile(t *testing.T) {
	// Create a temporary config file
	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	configContent := `namespace: test-namespace
scannerV4: true
devMode: true
applyNetworkPolicies: true
certPath: /custom/certs
images:
  central: custom-central:latest
  sensor: custom-sensor:latest
crs:
  portForward: true`

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	cfg, err := ReadConfig(tmpFile.Name())
	if err != nil {
		t.Errorf("ReadConfig() returned error: %v", err)
	}

	if cfg == nil {
		t.Errorf("ReadConfig() returned nil config")
		return
	}

	// Check specific values that should be overridden
	if cfg.Namespace != "test-namespace" {
		t.Errorf("Expected namespace 'test-namespace', got '%s'", cfg.Namespace)
	}
	if !cfg.ScannerV4 {
		t.Errorf("Expected ScannerV4 to be true")
	}
	if !cfg.DevMode {
		t.Errorf("Expected DevMode to be true")
	}
	if !cfg.ApplyNetworkPolicies {
		t.Errorf("Expected ApplyNetworkPolicies to be true")
	}
	if cfg.CertPath != "/custom/certs" {
		t.Errorf("Expected certPath '/custom/certs', got '%s'", cfg.CertPath)
	}
	if cfg.Images.Central != "custom-central:latest" {
		t.Errorf("Expected central image 'custom-central:latest', got '%s'", cfg.Images.Central)
	}
	if cfg.Images.Sensor != "custom-sensor:latest" {
		t.Errorf("Expected sensor image 'custom-sensor:latest', got '%s'", cfg.Images.Sensor)
	}
	if !cfg.CRS.PortForward {
		t.Errorf("Expected CRS.PortForward to be true")
	}

	// Check that unspecified values retain defaults
	if cfg.Images.Collector != localStackroxImage {
		t.Errorf("Expected collector image to retain default '%s', got '%s'", localStackroxImage, cfg.Images.Collector)
	}
}

func TestLoad_ValidYAML(t *testing.T) {
	yamlContent := `namespace: test-ns
scannerV4: true
images:
  central: test-central:v1.0`

	reader := strings.NewReader(yamlContent)
	cfg, err := load(reader)

	if err != nil {
		t.Errorf("load() returned error: %v", err)
	}

	if cfg == nil {
		t.Errorf("load() returned nil config")
		return
	}

	if cfg.Namespace != "test-ns" {
		t.Errorf("Expected namespace 'test-ns', got '%s'", cfg.Namespace)
	}
	if !cfg.ScannerV4 {
		t.Errorf("Expected ScannerV4 to be true")
	}
	if cfg.Images.Central != "test-central:v1.0" {
		t.Errorf("Expected central image 'test-central:v1.0', got '%s'", cfg.Images.Central)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	invalidYaml := `namespace: test
invalid: [unclosed array`

	reader := strings.NewReader(invalidYaml)
	_, err := load(reader)

	if err == nil {
		t.Errorf("load() should return error for invalid YAML")
	}

	expectedPrefix := "malformed yaml:"
	if !strings.HasPrefix(err.Error(), expectedPrefix) {
		t.Errorf("Expected error to start with '%s', got: %s", expectedPrefix, err.Error())
	}
}

func TestLoad_UnknownFields(t *testing.T) {
	yamlWithUnknownField := `namespace: test
unknownField: value
scannerV4: true`

	reader := strings.NewReader(yamlWithUnknownField)
	_, err := load(reader)

	if err == nil {
		t.Errorf("load() should return error for unknown fields")
	}

	if !strings.Contains(err.Error(), "malformed yaml:") {
		t.Errorf("Expected error about malformed yaml, got: %s", err.Error())
	}
}

func TestLoad_EmptyContent(t *testing.T) {
	reader := strings.NewReader("")
	_, err := load(reader)

	// Empty content should result in an EOF error when trying to decode YAML
	if err == nil {
		t.Errorf("load() should return error for empty content (EOF)")
	}

	if !strings.Contains(err.Error(), "malformed yaml:") {
		t.Errorf("Expected error about malformed yaml, got: %s", err.Error())
	}
}

func TestConfigConstants(t *testing.T) {
	expectedLocalStackroxImage := "localhost:5001/stackrox/stackrox:latest"
	expectedLocalDbImage := "localhost:5001/stackrox/db:latest"

	if localStackroxImage != expectedLocalStackroxImage {
		t.Errorf("localStackroxImage = %s, want %s", localStackroxImage, expectedLocalStackroxImage)
	}

	if localDbImage != expectedLocalDbImage {
		t.Errorf("localDbImage = %s, want %s", localDbImage, expectedLocalDbImage)
	}
}

func TestReadConfig_RealFileExample(t *testing.T) {
	// Test with the actual installer.yaml if it exists
	installerYamlPath := filepath.Join("..", "..", "installer.yaml")
	if _, err := os.Stat(installerYamlPath); os.IsNotExist(err) {
		t.Skip("installer.yaml not found, skipping real file test")
	}

	cfg, err := ReadConfig(installerYamlPath)
	if err != nil {
		t.Errorf("ReadConfig() failed to read installer.yaml: %v", err)
	}

	if cfg == nil {
		t.Errorf("ReadConfig() returned nil for installer.yaml")
		return
	}

	// Basic validation - should have a namespace
	if cfg.Namespace == "" {
		t.Errorf("installer.yaml should specify a namespace")
	}
}

func TestConfigTypes(t *testing.T) {
	// Test that Config struct can be properly marshaled/unmarshaled
	original := Config{
		Namespace: "test",
		ScannerV4: true,
		DevMode:   false,
		Images: Images{
			Central: "test-central:latest",
			Sensor:  "test-sensor:latest",
		},
		CRS: CRS{
			PortForward: true,
		},
	}

	// Convert to YAML and back
	yamlContent := `namespace: test
scannerV4: true
devMode: false
images:
  central: test-central:latest
  sensor: test-sensor:latest
crs:
  portForward: true`

	reader := strings.NewReader(yamlContent)
	cfg, err := load(reader)
	if err != nil {
		t.Errorf("Failed to load config: %v", err)
	}

	// Compare key fields
	if cfg.Namespace != original.Namespace {
		t.Errorf("Namespace mismatch: got %s, want %s", cfg.Namespace, original.Namespace)
	}
	if cfg.ScannerV4 != original.ScannerV4 {
		t.Errorf("ScannerV4 mismatch: got %t, want %t", cfg.ScannerV4, original.ScannerV4)
	}
	if cfg.DevMode != original.DevMode {
		t.Errorf("DevMode mismatch: got %t, want %t", cfg.DevMode, original.DevMode)
	}
	if cfg.Images.Central != original.Images.Central {
		t.Errorf("Central image mismatch: got %s, want %s", cfg.Images.Central, original.Images.Central)
	}
	if cfg.CRS.PortForward != original.CRS.PortForward {
		t.Errorf("CRS.PortForward mismatch: got %t, want %t", cfg.CRS.PortForward, original.CRS.PortForward)
	}
}