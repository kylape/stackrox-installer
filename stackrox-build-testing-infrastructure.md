# StackRox Build, Installation, and Testing Infrastructure

## Overview

StackRox is a Kubernetes-native security platform composed of multiple services that work together to provide runtime protection, vulnerability management, and compliance monitoring. The project involves complex build processes, deployment strategies, and testing approaches across multiple components.

## Current Architecture

### Core Components

- **Central**: Main control plane service providing API, UI, and policy management
- **Sensor**: Kubernetes cluster agent for runtime monitoring and enforcement
- **Scanner**: Vulnerability scanning service for images and nodes
- **Collector**: eBPF/kernel module for runtime data collection
- **Compliance**: Compliance scanning and reporting service

### Repository Structure

```
stackrox/
├── central/          # Central service implementation
├── sensor/           # Sensor agent implementation
├── scanner/          # Vulnerability scanner
├── collector/        # Runtime data collection
├── ui/               # Web interface
├── image/            # Container image definitions
├── deploy/           # Deployment scripts and configurations
├── tests/            # Test suites and utilities
├── qa-tests-backend/ # Integration test suite (Groovy)
└── .openshift-ci/    # CI/CD infrastructure
```

## Build System

### Current Build Process

The build system uses GNU Make with Docker-based compilation for consistency across environments.

#### Key Build Targets

```bash
# Dependency management
make deps                    # Download and prepare dependencies
make proto-generated-srcs    # Generate protobuf sources
make go-generated-srcs       # Generate Go sources

# Component builds
make main-build-dockerized   # Build Central service
make sensor-build-dockerized # Build Sensor agent
make scanner-build           # Build Scanner service
make collector-modules       # Build Collector components

# Image creation
make main-image             # Create Central container image
make sensor-image           # Create Sensor container image
make scanner-image          # Create Scanner container image
```

#### Fast Development Builds

For development iteration, the build system provides optimized targets:

```bash
make fast-central          # Quick Central rebuild (~3-5 minutes)
make fast-sensor           # Quick Sensor rebuild (~3-5 minutes)
make fast-scanner          # Quick Scanner rebuild
```

### Build Environment

- **Language**: Go 1.24+ for core services
- **Frontend**: Node.js/React for UI components
- **Build tools**: Docker, Make, Protocol Buffers
- **CI container**: `quay.io/stackrox-io/apollo-ci:stackrox-test-0.4.9`

## Installation and Deployment

### Deployment Methods

#### Helm Charts (Current Primary Method)

StackRox uses Helm charts with a meta-templating system:

```bash
# Generate Helm charts from meta-templates
roxctl helm output central-services --image-defaults=development_build --output-dir=./chart

# Deploy Central services
helm install central ./chart/central-services -f values.yaml

# Deploy Secured Cluster services
helm install secured-cluster ./chart/secured-cluster-services -f sensor-values.yaml
```

#### Direct Deployment Scripts

Simplified deployment scripts for development:

```bash
# Kubernetes deployment
./deploy/k8s/deploy-local.sh

# OpenShift deployment
./deploy/openshift/deploy-local.sh
```

### Environment Configuration

Key environment variables for deployment customization:

| Variable | Values | Purpose |
|----------|--------|---------|
| `COLLECTION_METHOD` | `core_bpf` | Collector data collection method |
| `ROX_HOTRELOAD` | `true/false` | Enable local binary hot-reloading |
| `LOAD_BALANCER` | `route/lb` | Central service exposure method |
| `STORAGE` | `none/pvc` | Central database persistence |
| `ROX_LOCAL_SOURCE_PATH` | `path` | Local source path for KIND development |

### Development Environment Support

#### KIND Integration

For local Kubernetes development:

```bash
export ROX_LOCAL_SOURCE_PATH="/path/to/stackrox"
export ROX_HOTRELOAD=true
./deploy/k8s/deploy-local.sh
```

#### Hot-Reload Capability

Development workflow support for rapid iteration:

```bash
# Build and hot-reload specific components
make fast-central
./dev-tools/enable-hotreload.sh central

# Components: central, sensor, scanner, compliance
```

## Testing Infrastructure

### Testing Pyramid Structure

#### Unit Tests

- **Location**: `*_test.go` files throughout codebase
- **Count**: ~1,518 test files
- **Execution**: `make go-unit-tests`
- **Current duration**: Over 1 minute (performance issue)
- **Dependencies**: Minimal external dependencies

#### Integration Tests

- **Groovy-based**: `qa-tests-backend/` directory (~168 test files)
- **Go-based**: Component integration tests
- **Database tests**: PostgreSQL integration testing
- **Execution time**: 30 minutes to 4+ hours depending on scope

#### End-to-End Tests

- **Location**: `tests/e2e/` and `.openshift-ci/`
- **Execution environment**: CI container with cloud infrastructure
- **Dependencies**: Vault credentials, cloud provider access
- **Test categories**: QA, UI, compliance, upgrade, compatibility

### Current Testing Challenges

#### Local E2E Testing Complexity

E2E testing currently requires:

- Red Hat employee vault access for credentials
- CI container execution (`apollo-ci:stackrox-test-0.4.9`)
- Cloud provider infrastructure (GCP, AWS, Azure)
- Complex environment setup and teardown

#### Test Execution Environment

```bash
# Local E2E testing (simplified)
./tests/e2e/run-e2e-tests.sh qa

# Requires:
# - VAULT_TOKEN environment variable
# - Pre-configured Kubernetes cluster
# - Docker daemon access
# - Network access to external services
```

### CI/CD Infrastructure

#### OpenShift CI Integration

- **Configuration**: `.openshift-ci/` directory
- **Job types**: GKE, EKS, OpenShift variants
- **Workflow**: begin.sh → dispatch.sh → job execution → end.sh
- **Artifact collection**: Logs, database dumps, performance data

#### Test Categories in CI

- **QA tests**: Comprehensive functional testing
- **UI tests**: Frontend automation testing
- **Compliance tests**: Security compliance validation
- **Upgrade tests**: Version migration testing
- **Performance tests**: Scale and load testing

## Development Workflow

### Current Developer Experience

#### Typical Development Cycle

```bash
# 1. Code changes
# 2. Component build
make fast-central                          # 3-5 minutes

# 3. Deployment update
./dev-tools/enable-hotreload.sh central    # 30 seconds

# 4. Manual testing
kubectl port-forward svc/central 8443:443

# 5. Unit test execution (optional due to speed)
make go-unit-tests                         # 1+ minutes

# 6. Integration testing via PR/CI
```

#### Multi-Component Development

For changes affecting multiple components:

```bash
# Build multiple components
make fast-central && make fast-sensor

# Coordinate deployments
./deploy/k8s/deploy-local.sh

# Manual version coordination required
```

### Development Environment Tooling

#### Container-Based Development

The CI container provides consistent tooling:

- **Kubernetes tools**: kubectl, oc, helm
- **Cloud CLIs**: gcloud, aws
- **Build tools**: Go toolchain, Node.js, Docker
- **Testing tools**: bats, gradle, PostgreSQL clients
- **Development utilities**: vault, shellcheck, yq

#### Local Development Scripts

Supporting scripts for development workflow:

- `dev-tools/enable-hotreload.sh`: Component hot-reloading
- `dev-tools/helmdiff.sh`: Helm chart comparison
- `dev-tools/upgrade-dev-secured-cluster.sh`: Cluster updates

## Configuration and Customization

### Image Flavor System

StackRox supports multiple image flavors:

- **development_build**: Development images with debug symbols
- **rhacs**: Red Hat Advanced Cluster Security branded images
- **opensource**: Open source distribution images

### Profile-Based Deployment

Different deployment profiles for various scenarios:

- **Development**: Fast iteration, debug-enabled, local storage
- **Testing**: Integration testing, external dependencies
- **Production**: Performance-optimized, security-hardened
- **Demo**: Simplified setup, sample data

## Infrastructure Dependencies

### External Services

- **Image registries**: quay.io, Docker Hub
- **Cloud providers**: GCP, AWS, Azure for testing
- **Secret management**: HashiCorp Vault for CI credentials
- **Database**: PostgreSQL for Central data persistence

### Development Dependencies

- **Local Kubernetes**: KIND, Docker Desktop, OpenShift Local
- **Build environment**: Docker, Go toolchain, Node.js
- **Network access**: External API dependencies, image pulls

## Performance and Optimization

### Build Performance

- **Parallel builds**: Component-level parallelization
- **Docker layer caching**: Optimized Dockerfile structure
- **Incremental compilation**: Fast rebuild targets
- **Dependency caching**: Go module and Node.js dependency caching

### Test Performance

- **Unit test optimization**: Parallel execution, minimal dependencies
- **Integration test efficiency**: Focused test scopes
- **E2E test optimization**: Selective test execution based on changes

This document provides an overview of the current StackRox build, installation, and testing infrastructure, highlighting both capabilities and areas for improvement in the development experience.