# Docker Image Usage Guide

The StackRox installer is available as a Docker image that can be used as a CLI tool without needing to install Go or build the binary locally.

## Image Location

The image is published to GitHub Container Registry:
```
ghcr.io/kylape/stackrox-installer:main
```

## Basic Usage

### Using Default Configuration

The image includes a default `installer.yaml` configuration:

```bash
# Export manifests using default config
docker run --rm ghcr.io/kylape/stackrox-installer:main export securedcluster

# Show help
docker run --rm ghcr.io/kylape/stackrox-installer:main --help
```

### Using Custom Configuration

#### Method 1: Volume Mount to Standard Location

Create your custom config and mount it to the expected location:

```bash
# Create your custom config
cat > my-installer.yaml <<EOF
namespace: my-stackrox
scannerV4: true
devMode: false
images:
  central: "my-registry/stackrox:latest"
  sensor: "my-registry/stackrox:latest"
  # ... other image overrides
EOF

# Use it by mounting to /app/config/installer.yaml
docker run --rm \
  -v $(pwd)/my-installer.yaml:/app/config/installer.yaml \
  ghcr.io/kylape/stackrox-installer:main \
  -conf /app/config/installer.yaml \
  export securedcluster
```

#### Method 2: Mount to Custom Path

Mount your config to any path and specify it with `-conf`:

```bash
docker run --rm \
  -v $(pwd)/my-installer.yaml:/app/my-config.yaml \
  ghcr.io/kylape/stackrox-installer:main \
  -conf /app/my-config.yaml \
  export securedcluster
```

#### Method 3: Mount Entire Config Directory

If you have multiple config files:

```bash
# Mount entire directory
docker run --rm \
  -v $(pwd)/configs:/app/configs \
  ghcr.io/kylape/stackrox-installer:main \
  -conf /app/configs/production.yaml \
  export central
```

## Applying to Kubernetes Cluster

To apply manifests to a cluster, you need to mount your kubeconfig:

```bash
# Apply using default kubeconfig location
docker run --rm \
  -v ~/.kube:/app/.kube \
  -v $(pwd)/my-installer.yaml:/app/config/installer.yaml \
  ghcr.io/kylape/stackrox-installer:main \
  -conf /app/config/installer.yaml \
  apply securedcluster

# Apply using custom kubeconfig path
docker run --rm \
  -v $(pwd)/my-kubeconfig:/app/kubeconfig \
  -v $(pwd)/my-installer.yaml:/app/my-config.yaml \
  ghcr.io/kylape/stackrox-installer:main \
  -kubeconfig /app/kubeconfig \
  -conf /app/my-config.yaml \
  apply central
```

## Available Tags

- `main` - Latest build from main branch
- `v*` - Specific version tags (e.g., `v1.0.0`)
- `<branch>-<sha>` - Specific commit builds from branches

## Examples

### Development Workflow

```bash
# Export manifests for review
docker run --rm \
  -v $(pwd)/dev-config.yaml:/app/config/installer.yaml \
  ghcr.io/kylape/stackrox-installer:main \
  -conf /app/config/installer.yaml \
  export securedcluster > manifests.yaml

# Review the manifests
cat manifests.yaml

# Apply to development cluster
docker run --rm \
  -v ~/.kube:/app/.kube \
  -v $(pwd)/dev-config.yaml:/app/config/installer.yaml \
  ghcr.io/kylape/stackrox-installer:main \
  -conf /app/config/installer.yaml \
  apply securedcluster
```

### CI/CD Pipeline

```bash
# In a CI/CD pipeline
docker run --rm \
  -v $CI_PROJECT_DIR/configs/production.yaml:/app/config/installer.yaml \
  -v $KUBECONFIG_FILE:/app/kubeconfig \
  ghcr.io/kylape/stackrox-installer:main \
  -kubeconfig /app/kubeconfig \
  -conf /app/config/installer.yaml \
  apply central
```

### Multiple Environments

```bash
# Development
docker run --rm \
  -v $(pwd)/configs:/app/configs \
  ghcr.io/kylape/stackrox-installer:main \
  -conf /app/configs/dev.yaml \
  export securedcluster

# Staging  
docker run --rm \
  -v $(pwd)/configs:/app/configs \
  ghcr.io/kylape/stackrox-installer:main \
  -conf /app/configs/staging.yaml \
  export securedcluster

# Production
docker run --rm \
  -v $(pwd)/configs:/app/configs \
  -v ~/.kube:/app/.kube \
  ghcr.io/kylape/stackrox-installer:main \
  -conf /app/configs/production.yaml \
  apply central
```

## Tips

1. **Always use `--rm`** to clean up containers after use
2. **Mount configs as read-only** when possible: `-v $(pwd)/config.yaml:/app/config.yaml:ro`
3. **Use absolute paths** for volume mounts to avoid issues
4. **Check permissions** - ensure mounted files are readable by UID 1001 (the installer user)
5. **Use specific tags** in production rather than `main` for reproducibility