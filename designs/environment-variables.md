# Environment Variable Configuration System

## Overview

This design adds the ability to set arbitrary environment variables on any particular container in the StackRox installer. The system provides flexible targeting options to apply environment variables to groups of containers (per pod, per generator group, or globally) with clear precedence rules.

## Goals

- Enable setting custom environment variables on any container
- Support different scoping levels (global, generator, pod, container)
- Maintain backward compatibility with existing configurations
- Provide clear precedence rules to resolve conflicts
- Prevent accidental override of critical system variables

## Configuration Schema

### YAML Structure

```yaml
envVars:
  # Global env vars applied to all containers
  global:
    - name: "CUSTOM_GLOBAL_VAR"
      value: "global-value"
    - name: "DEPLOYMENT_ENV"
      value: "production"
  
  # Generator-specific env vars (central, sensor, etc.)
  generators:
    central:
      - name: "CENTRAL_CUSTOM_VAR"
        value: "central-value"
      - name: "LOG_LEVEL"
        value: "debug"
    sensor:
      - name: "SENSOR_CUSTOM_VAR"
        value: "sensor-value"
    admission_control:
      - name: "ADMISSION_LOG_LEVEL"
        value: "info"
  
  # Pod-specific env vars
  pods:
    central:
      - name: "POD_SPECIFIC_VAR"
        value: "pod-value"
    sensor:
      - name: "SENSOR_POD_VAR"
        value: "sensor-pod-value"
  
  # Container-specific env vars (most granular)
  containers:
    "central":
      - name: "CONTAINER_SPECIFIC_VAR"
        value: "container-value"
    "sensor":
      - name: "MAX_CONNECTIONS"
        value: "1000"
```

### Go Struct Definition

```go
type EnvVarConfig struct {
    Global     []v1.EnvVar            `yaml:"global"`
    Generators map[string][]v1.EnvVar `yaml:"generators"`
    Pods       map[string][]v1.EnvVar `yaml:"pods"`
    Containers map[string][]v1.EnvVar `yaml:"containers"`
}

type Config struct {
    // ... existing fields ...
    EnvVars EnvVarConfig `yaml:"envVars"`
}
```

## Precedence Rules

Environment variables are merged with the following precedence (higher number wins):

1. **Global env vars** (lowest priority)
2. **Generator-specific env vars**
3. **Pod-specific env vars**
4. **Container-specific env vars**
5. **Default/hardcoded env vars** (highest priority)

This ensures that:
- More specific configurations override general ones
- System-critical variables cannot be accidentally overridden
- Users can still override most application-level settings

## Implementation Plan

### Phase 1: Core Infrastructure

1. **Extend Config struct** in `installer/manifest/config.go`
   - Add `EnvVars EnvVarConfig` field
   - Update `DefaultConfig` with empty env var configuration

2. **Add helper functions** in `installer/manifest/manifest.go`
   - `mergeEnvVars()` - merges env vars according to precedence rules
   - `getEnvVarsForContainer()` - collects all applicable env vars for a container
   - `validateEnvVars()` - validates env var names and prevents system overrides

### Phase 2: Generator Integration

3. **Update each generator** to accept and apply custom env vars:
   - `central.go` - central container
   - `sensor.go` - sensor container
   - `admission_control.go` - admission controller container
   - `scanner.go`, `scanner_db.go`, etc. - other components

4. **Modify container creation logic**
   - Update deployment/pod specs to include merged env vars
   - Ensure existing hardcoded env vars take precedence

### Phase 3: Validation and Safety

5. **Add validation**
   - Prevent override of critical system variables (ROX_*, KUBERNETES_*, etc.)
   - Validate env var names follow Kubernetes conventions
   - Check for reserved/conflicting names

6. **Testing**
   - Unit tests for merge logic
   - Integration tests with sample configurations
   - Verify precedence rules work correctly

## Usage Examples

### Basic Global Configuration

```yaml
# installer.yaml
envVars:
  global:
    - name: "DEPLOYMENT_ENV"
      value: "production"
    - name: "LOG_FORMAT"
      value: "json"
```

### Component-Specific Configuration

```yaml
# installer.yaml
envVars:
  generators:
    central:
      - name: "CENTRAL_LOG_LEVEL"
        value: "debug"
      - name: "MAX_CONCURRENT_SCANS"
        value: "10"
    sensor:
      - name: "SENSOR_LOG_LEVEL"
        value: "info"
```

### Fine-Grained Container Control

```yaml
# installer.yaml
envVars:
  containers:
    central:
      - name: "MEMORY_LIMIT_OVERRIDE"
        value: "8Gi"
    sensor:
      - name: "COLLECTION_INTERVAL"
        value: "30s"
```

### Mixed Configuration

```yaml
# installer.yaml
envVars:
  global:
    - name: "CLUSTER_ENV"
      value: "staging"
  generators:
    central:
      - name: "LOG_LEVEL"
        value: "debug"
  containers:
    central:
      - name: "LOG_LEVEL"  # This overrides the generator setting
        value: "trace"
```

## Benefits

1. **Flexibility**: Support for multiple targeting scopes
2. **Safety**: Protected system variables with validation
3. **Simplicity**: Intuitive YAML configuration
4. **Backward Compatibility**: No breaking changes to existing configs
5. **Extensibility**: Easy to add new scoping levels in the future

## Security Considerations

- Validation prevents override of critical system environment variables
- Environment variable values are not encrypted - avoid storing secrets
- Consider using Kubernetes secrets for sensitive values instead
- Validate environment variable names follow security best practices

## Future Enhancements

- Support for environment variable templates/interpolation
- Integration with Kubernetes ConfigMaps and Secrets
- Environment variable inheritance from parent scopes
- Conditional environment variables based on deployment context