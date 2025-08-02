# GitHub Actions CI/CD Workflow Design

## Problem Statement

The StackRox installer project needs automated CI/CD to:
- Run unit tests on pull requests and main branch
- Build and validate the installer binary
- Create a containerized CLI image that can be used as a tool
- Ensure code quality and prevent regressions

## Proposed Solution

Implement a GitHub Actions workflow that provides comprehensive CI/CD pipeline with:
1. **Testing Pipeline**: Run Go unit tests and build validation
2. **Container Build**: Create a Docker image with the installer binary as a CLI tool
3. **Multi-platform Support**: Build for linux/amd64 and linux/arm64
4. **Artifact Management**: Publish images to a container registry

## Implementation Details

### GitHub Actions Workflow Structure
- **Trigger**: On push to main, pull requests, and tags
- **Jobs**:
  1. `test` - Run unit tests and linting
  2. `build` - Build installer binary for multiple architectures  
  3. `docker` - Build and push container image

### Container Image Design
- **Base Image**: Use minimal alpine image with ca-certificates
- **Binary**: Include compiled installer binary at `/usr/local/bin/installer`
- **Entrypoint**: Configure to run installer directly
- **Configuration**: Support both default config and custom mounted configs
- **Usage Examples**:
  ```bash
  # Use default config
  docker run <image> export securedcluster
  
  # Use custom config via volume mount
  docker run -v /path/to/my-config.yaml:/app/config/installer.yaml <image> -conf /app/config/installer.yaml export securedcluster
  
  # Use custom config via bind mount
  docker run -v $(pwd)/my-installer.yaml:/app/my-config.yaml <image> -conf /app/my-config.yaml export securedcluster
  ```

### Key Components

#### 1. Testing Job
```yaml
- Go setup (1.24+)
- Dependency caching
- Run `go test ./...`
- Run `go build` validation
- Code quality checks
```

#### 2. Build Job
```yaml
- Cross-compilation for linux/amd64, linux/arm64
- Binary artifact upload
- Version tagging from git
```

#### 3. Docker Job  
```yaml
- Multi-platform buildx setup
- Docker build with installer binary
- Push to GitHub Container Registry (ghcr.io)
- Tag with git sha and semver if tagged
```

## Testing Plan

1. **Local Testing**: Validate Dockerfile builds correctly
2. **Workflow Testing**: Push to feature branch and verify workflow runs
3. **Image Testing**: Pull built image and test CLI functionality
4. **Integration Testing**: Verify `docker run <image> export securedcluster` works

## Considerations

### Security
- Use GitHub's built-in GITHUB_TOKEN for registry authentication
- No external secrets required for basic functionality
- Container runs as non-root user when possible

### Performance  
- Cache Go modules and Docker layers
- Use GitHub's actions cache for dependencies
- Parallel job execution where possible

### Backwards Compatibility
- No breaking changes to existing functionality
- Container image provides same CLI interface
- Maintains existing build process

### Registry Strategy
- Use GitHub Container Registry (ghcr.io) for simplicity
- Public images for open source project
- Automated cleanup of old images