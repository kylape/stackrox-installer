# StackRox Installer Project

This is a Go-based installer for the StackRox security platform that generates and deploys Kubernetes manifests.

## Build Instructions

```bash
go build -o bin/installer ./installer
```

## Usage

The installer supports three deployment sets:
- `central` - Central components
- `securedcluster` - Secured cluster components  
- `crs` - Custom resource definitions

### Commands

```bash
# Apply manifests to cluster
./bin/installer apply <set>

# Export manifests to stdout
./bin/installer export <set>
```

### Configuration

Configuration is provided via `installer.yaml` (default) or specify with `-conf` flag.

Example:
```bash
./bin/installer -conf ./installer.yaml apply central
```

### Kubeconfig

The installer will use kubeconfig in this order:
1. `-kubeconfig` flag
2. `KUBECONFIG` environment variable  
3. In-cluster config (if running in pod)
4. `~/.kube/config`

## Project Structure

- `installer/main.go` - Main entry point
- `installer/manifest/` - Manifest generation logic
  - `manifest.go` - Core generator framework
  - `central.go`, `sensor.go`, etc. - Component-specific generators
  - `crds/` - Custom resource definitions
- `installer.yaml` - Default configuration
- `certs/` - CA certificates for TLS

## Development

- Uses Go 1.24+
- Dependencies managed with Go modules
- Integrates with StackRox main repository via replace directives
- Supports development mode with local image registries

## Testing

No specific test commands found - verify with build and basic execution.