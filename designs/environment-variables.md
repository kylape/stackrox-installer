# Environment Variable Configuration System

## Overview

This design adds the ability to set arbitrary environment variables on any particular container in the StackRox installer. The system follows the StackRox Helm chart customization pattern, providing flexible targeting options to apply environment variables at global, deployment, and container levels with clear precedence rules.

## Goals

- Enable setting custom environment variables on any container
- Follow StackRox Helm chart customization patterns for consistency
- Support different scoping levels (global, deployment, container)
- Maintain backward compatibility with existing configurations
- Provide clear precedence rules to resolve conflicts
- Prevent accidental override of critical system variables

## Configuration Schema

### YAML Structure (StackRox Helm Pattern)

```yaml
# installer.yaml
customize:
  envVars:
    # Global env vars applied to all containers
    "":
      - name: "DEPLOYMENT_ENV"
        value: "production"
      - name: "LOG_FORMAT"
        value: "json"
    
    # Deployment-specific configuration
    central:
      # Deployment-level env vars (apply to all containers in deployment)
      envVars:
        - name: "CENTRAL_LOG_LEVEL"
          value: "debug"
        - name: "MAX_CONCURRENT_SCANS"
          value: "10"
      
      # Container-specific env vars (using /containerName syntax)
      "/central":
        - name: "CENTRAL_LOG_LEVEL"
          value: "trace"  # Overrides deployment setting
        - name: "MEMORY_LIMIT"
          valueFrom:
            resourceFieldRef:
              resource: "limits.memory"
      
      "/nodejs":
        - name: "NODE_ENV"
          value: "development"
    
    sensor:
      envVars:
        - name: "SENSOR_LOG_LEVEL"
          value: "info"
      
      "/sensor":
        - name: "MAX_CONNECTIONS"
          value: "1000"
    
    collector:
      envVars:
        - name: "COLLECTOR_LOG_LEVEL"
          value: "warn"
      
      "/collector":
        - name: "COLLECTION_METHOD"
          value: "CORE_BPF"
      
      "/compliance":
        - name: "COMPLIANCE_INTERVAL"
          value: "24h"
```

### Go Struct Definition

```go
type CustomizeConfig struct {
    EnvVars map[string]interface{} `yaml:"envVars"`
}

type Config struct {
    // ... existing fields ...
    Customize CustomizeConfig `yaml:"customize"`
}
```

## Precedence Rules

Environment variables are merged with the following precedence (higher number wins):

1. **Global env vars** (lowest priority) - `customize.envVars.""`
2. **Deployment-level env vars** - `customize.envVars.{deployment}.envVars`
3. **Container-specific env vars** - `customize.envVars.{deployment}./{container}`
4. **Default/hardcoded env vars** (highest priority) - System-defined variables

This ensures that:
- More specific configurations override general ones
- System-critical variables cannot be accidentally overridden
- Users can still override most application-level settings

## Alignment with StackRox Helm Charts

This design follows the exact pattern used in StackRox Helm charts:

### StackRox Helm Chart Structure
```yaml
# values.yaml
customize:
  envVars: {}  # Global
  central:
    envVars: {}  # Deployment-level
  scanner:
    envVars: {}  # Deployment-level
```

### Container-Specific Pattern
StackRox Helm uses `/containerName` syntax for container-specific customizations:
```yaml
customize:
  central:
    envVars:
      "/central": []     # Central container only
      "/nodejs": []      # NodeJS container only
```

This installer implementation follows the same pattern, providing seamless migration path to full Helm charts.

## Implementation Details

### Core Functions

1. **GetEnvVarsForContainer()** - Merges env vars according to precedence rules
2. **getEnvVarsFromInterface()** - Extracts env vars from YAML interface{}
3. **convertToEnvVars()** - Converts YAML data to v1.EnvVar structs
4. **isValidEnvVarName()** - Validates and filters protected variables

### Security Features

- **Protected Variable Filtering**: Prevents override of critical system variables:
  - `ROX_*` - StackRox system variables
  - `KUBERNETES_*` - Kubernetes system variables  
  - `PATH`, `HOME`, `USER`, etc. - Standard system variables

### Supported Environment Variable Types

```yaml
# Simple value
- name: "LOG_LEVEL"
  value: "debug"

# Field reference
- name: "POD_NAME"
  valueFrom:
    fieldRef:
      fieldPath: "metadata.name"

# Resource reference
- name: "MEMORY_LIMIT"
  valueFrom:
    resourceFieldRef:
      resource: "limits.memory"
```

## Usage Examples

### Basic Global Configuration

```yaml
# installer.yaml
customize:
  envVars:
    "":
      - name: "DEPLOYMENT_ENV"
        value: "production"
      - name: "LOG_FORMAT"
        value: "json"
```

### Deployment-Specific Configuration

```yaml
# installer.yaml
customize:
  envVars:
    central:
      envVars:
        - name: "CENTRAL_LOG_LEVEL"
          value: "debug"
        - name: "MAX_CONCURRENT_SCANS"
          value: "10"
    
    sensor:
      envVars:
        - name: "SENSOR_LOG_LEVEL"
          value: "info"
```

### Fine-Grained Container Control

```yaml
# installer.yaml
customize:
  envVars:
    central:
      "/central":
        - name: "MEMORY_LIMIT_OVERRIDE"
          value: "8Gi"
      "/nodejs":
        - name: "NODE_ENV"
          value: "development"
    
    collector:
      "/collector":
        - name: "COLLECTION_METHOD"
          value: "CORE_BPF"
      "/compliance":
        - name: "COMPLIANCE_INTERVAL"
          value: "24h"
```

### Mixed Configuration with Precedence

```yaml
# installer.yaml
customize:
  envVars:
    "":
      - name: "CLUSTER_ENV"
        value: "staging"
    
    central:
      envVars:
        - name: "LOG_LEVEL"
          value: "debug"
      
      "/central":
        - name: "LOG_LEVEL"  # This overrides the deployment setting
          value: "trace"
```

## Migration from Original Design

The implementation maintains backward compatibility while following StackRox patterns:

**Original Design** → **StackRox-Aligned Design**
```yaml
# Old structure (would still work for simple cases)
envVars:
  global: []
  generators:
    central: []

# New structure (recommended, StackRox-aligned)
customize:
  envVars:
    "": []           # Global
    central:         # Deployment
      envVars: []    # Deployment-level
      "/central": [] # Container-specific
```

## Benefits

1. **StackRox Consistency**: Follows exact pattern used in StackRox Helm charts
2. **Migration Path**: Natural progression from installer to full Helm deployment
3. **Familiar UX**: StackRox users already understand this pattern
4. **Extensibility**: Room for future customization types (labels, annotations, etc.)
5. **Flexibility**: Support for multiple targeting scopes
6. **Safety**: Protected system variables with validation
7. **Backward Compatibility**: No breaking changes to existing configs

## Security Considerations

- Validation prevents override of critical system environment variables
- Environment variable values are not encrypted - avoid storing secrets
- Consider using Kubernetes secrets for sensitive values instead
- Validate environment variable names follow security best practices

## Future Enhancements

Since this follows the StackRox Helm pattern, future enhancements can include:

- **Full Customize Support**: `labels`, `annotations`, `podLabels`, `podAnnotations`
- **Kubernetes Integration**: ConfigMaps and Secrets support
- **Template Variables**: Environment variable interpolation
- **Conditional Logic**: Context-based environment variables
- **Migration Tools**: Automated conversion to full Helm charts