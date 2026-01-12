# StackRox Installer Configuration Gap Analysis

## Problem Statement

The current StackRox installer provides a minimal configuration interface that covers less than 5% of the deployment options available through the official StackRox Central and SecuredCluster CRDs. This significantly limits enterprise adoption and prevents the installer from being used in production environments.

## Current State Analysis

### Existing Configuration Coverage
The current installer supports only basic configuration:

```go
type Config struct {
	Action               string `yaml:"action"`
	ApplyNetworkPolicies bool   `yaml:"applyNetworkPolicies"`
	CRS                  CRS    `yaml:"crs"`
	CertPath             string `yaml:"certPath"`
	DevMode              bool   `yaml:"devMode"`
	Images               Images `yaml:"images"`
	ImageArchitecture    string `yaml:"imageArchitecture"`
	Namespace            string `yaml:"namespace"`
	ScannerV4            bool   `yaml:"scannerV4"`
}
```

**What's Missing**: 120+ configuration categories that are essential for enterprise deployments.

## Comprehensive Gap Analysis

### Critical Priority (P0) - Enterprise Blockers

#### 1. Resource Management & Sizing
**Current State**: No resource configuration
**Missing Features**:
- CPU/memory limits and requests for all components
- Autoscaling configuration for Scanner components
- Resource allocation for databases
- Performance tuning parameters

**Impact**: Cannot deploy in production environments with proper resource allocation.

**Example Configuration Needed**:
```yaml
resources:
  central:
    requests: { cpu: "2", memory: "4Gi" }
    limits: { cpu: "4", memory: "8Gi" }
  centralDB:
    requests: { cpu: "1", memory: "2Gi" }
    limits: { cpu: "2", memory: "4Gi" }
  scanner:
    autoScaling:
      enabled: true
      minReplicas: 2
      maxReplicas: 10
      targetCPUUtilization: 70
```

#### 2. Storage & Persistence Configuration
**Current State**: No storage configuration
**Missing Features**:
- PVC size and storage class selection
- External storage integration
- Backup and recovery configuration
- Database persistence settings

**Impact**: Cannot configure enterprise storage requirements or ensure data persistence.

**Example Configuration Needed**:
```yaml
storage:
  central:
    persistence:
      size: "100Gi"
      storageClassName: "fast-ssd"
      accessMode: "ReadWriteOnce"
  centralDB:
    persistence:
      size: "200Gi"
      storageClassName: "database-storage"
```

#### 3. Network Exposure & Load Balancing
**Current State**: No network exposure configuration
**Missing Features**:
- LoadBalancer configuration with static IPs
- NodePort exposure settings
- OpenShift Route configuration
- Ingress controller integration

**Impact**: Cannot expose services for external access or integrate with enterprise networking.

**Example Configuration Needed**:
```yaml
networking:
  central:
    exposure:
      type: "LoadBalancer"
      loadBalancer:
        ip: "192.168.1.100"
        port: 443
      annotations:
        service.beta.kubernetes.io/aws-load-balancer-type: "nlb"
```

#### 4. Database Configuration
**Current State**: No database configuration options
**Missing Features**:
- External database connection support
- Connection pool sizing
- Database credential management
- PostgreSQL configuration overrides

**Impact**: Cannot use external databases for HA deployments or meet enterprise database requirements.

**Example Configuration Needed**:
```yaml
database:
  central:
    external:
      enabled: true
      connectionString: "postgresql://central-db.company.com:5432/central"
      passwordSecret: "central-db-credentials"
    connectionPool:
      minConnections: 10
      maxConnections: 90
```

### High Priority (P1) - Operational Excellence

#### 5. Node Placement & Scheduling
**Current State**: No node placement controls
**Missing Features**:
- Node selectors for component placement
- Tolerations for tainted nodes
- Host aliases for custom DNS
- Anti-affinity rules

**Impact**: Cannot control where components run or ensure proper distribution across nodes.

#### 6. Security & Access Control
**Current State**: Basic certificate path only
**Missing Features**:
- Admin password management
- Image pull secrets for private registries
- Custom TLS certificates
- RBAC configuration

**Impact**: Cannot meet enterprise security requirements or deploy in air-gapped environments.

#### 7. Monitoring & Observability
**Current State**: No monitoring configuration
**Missing Features**:
- Prometheus endpoint exposure
- OpenShift monitoring integration
- Telemetry configuration
- Log aggregation settings

**Impact**: Cannot integrate with enterprise monitoring and observability systems.

### Medium Priority (P2) - Advanced Features

#### 8. Admission Control Configuration
**Current State**: Basic admission control deployment only
**Missing Features**:
- Bypass mechanisms (break-glass)
- Inline image scanning configuration
- Timeout settings
- Replica count for HA

#### 9. Scanner Configuration
**Current State**: Basic Scanner V4 enable/disable
**Missing Features**:
- Collection method selection (EBPF vs kernel module)
- Local scanner deployment
- Vulnerability database updates
- Scan result retention

#### 10. Cluster-Specific Settings
**Current State**: Basic namespace configuration
**Missing Features**:
- Cluster labels and metadata
- Central endpoint configuration
- Audit log collection
- Compliance scanning settings

### Low Priority (P3) - Customization

#### 11. Customization Framework
**Current State**: No customization support
**Missing Features**:
- Custom annotations and labels
- Environment variable injection
- Kubernetes resource overlays
- Declarative configuration management

#### 12. Air-gapped Deployment Support
**Current State**: Basic image override only
**Missing Features**:
- Connectivity policy (online/offline)
- Registry mirror configuration
- Custom CA certificate management
- Offline vulnerability database

## Implementation Roadmap

### Phase 1: Critical Enterprise Features (8-12 weeks)
**Objective**: Enable basic enterprise deployments

1. **Resource Management** (2-3 weeks)
   - Add CPU/memory configuration for all components
   - Implement resource validation
   - Add autoscaling configuration

2. **Storage Configuration** (2-3 weeks)
   - Add PVC size and storage class options
   - Implement persistence configuration
   - Add backup configuration

3. **Network Exposure** (2-3 weeks)
   - Add LoadBalancer configuration
   - Implement ingress options
   - Add service annotation support

4. **Database Configuration** (2-3 weeks)
   - Add external database support
   - Implement connection pooling
   - Add credential management

### Phase 2: Operational Excellence (6-8 weeks)
**Objective**: Enable production-ready deployments

1. **Node Placement** (2 weeks)
   - Add node selectors and tolerations
   - Implement affinity rules

2. **Security Configuration** (2-3 weeks)
   - Add TLS certificate management
   - Implement image pull secrets
   - Add RBAC configuration

3. **Monitoring Integration** (2-3 weeks)
   - Add Prometheus configuration
   - Implement telemetry settings
   - Add observability options

### Phase 3: Advanced Features (8-10 weeks)
**Objective**: Enable advanced enterprise scenarios

1. **Admission Control** (2-3 weeks)
2. **Scanner Configuration** (2-3 weeks)  
3. **Customization Framework** (3-4 weeks)

## Configuration Structure Proposal

### Enhanced Config Structure
```go
type EnhancedConfig struct {
    // Existing basic fields
    Action               string `yaml:"action"`
    Namespace            string `yaml:"namespace"`
    DevMode              bool   `yaml:"devMode"`
    ScannerV4            bool   `yaml:"scannerV4"`
    Images               Images `yaml:"images"`
    
    // New configuration sections
    Resources    ResourceConfig     `yaml:"resources,omitempty"`
    Storage      StorageConfig      `yaml:"storage,omitempty"`
    Networking   NetworkConfig      `yaml:"networking,omitempty"`
    Security     SecurityConfig     `yaml:"security,omitempty"`
    Monitoring   MonitoringConfig   `yaml:"monitoring,omitempty"`
    Database     DatabaseConfig     `yaml:"database,omitempty"`
    Placement    PlacementConfig    `yaml:"placement,omitempty"`
    
    // Advanced sections (Phase 2-3)
    Scanner      ScannerConfig      `yaml:"scanner,omitempty"`
    Admission    AdmissionConfig    `yaml:"admission,omitempty"`
    Customization CustomConfig     `yaml:"customization,omitempty"`
}
```

### Configuration Profiles
To maintain simplicity while adding power, implement configuration profiles:

```yaml
# Development profile - minimal config
profile: "development"
namespace: "stackrox"
images:
  central: "quay.io/stackrox/main:latest"

---
# Production profile - full enterprise config
profile: "production"
namespace: "stackrox"
resources:
  central:
    requests: { cpu: "2", memory: "4Gi" }
    limits: { cpu: "4", memory: "8Gi" }
storage:
  central:
    size: "100Gi"
    storageClassName: "fast-ssd"
networking:
  central:
    exposure:
      type: "LoadBalancer"
      loadBalancer:
        ip: "192.168.1.100"
```

## Impact Assessment

### Enterprise Deployment Scenarios Currently Impossible

1. **High Availability Production Deployment**
   - Multi-replica configurations ❌
   - Load balancer setup ❌
   - External database ❌
   - Resource sizing ❌

2. **Air-gapped Enterprise Environment**
   - Registry overrides ❌
   - Offline mode ❌
   - Custom CA certificates ❌
   - Image pull secrets ❌

3. **Multi-tenant Cluster Deployment**
   - Node selectors ❌
   - Resource quotas ❌
   - Network isolation ❌
   - Security contexts ❌

4. **Compliance/Regulated Environment**
   - Security constraints ❌
   - Audit logging ❌
   - Policy enforcement ❌
   - Monitoring integration ❌

## Success Metrics

### Phase 1 Success Criteria
- [ ] Support for production resource sizing
- [ ] External storage integration
- [ ] LoadBalancer exposure working
- [ ] External database connectivity

### Long-term Success Criteria
- [ ] Feature parity with 80% of CRD options
- [ ] Enterprise customer adoption
- [ ] Air-gapped deployment capability
- [ ] Multi-cluster deployment support

## Risk Assessment

### High Risk
- **Complexity Creep**: Adding too many options too quickly
- **Breaking Changes**: Incompatible config format changes
- **Maintenance Burden**: Supporting extensive configuration matrix

### Mitigation Strategies
- Implement configuration validation
- Use configuration profiles for simplicity
- Maintain backward compatibility
- Comprehensive testing automation

## Conclusion

The current installer configuration represents a significant gap that prevents enterprise adoption. The proposed enhancement plan provides a structured approach to bridge this gap while maintaining usability for development scenarios.

**Immediate Priority**: Begin Phase 1 implementation focusing on the 4 critical enterprise blockers identified above.

**Long-term Vision**: Create a comprehensive yet user-friendly configuration system that rivals the operator's capabilities while providing simple defaults for common scenarios.