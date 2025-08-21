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

## Feature Development Flow

When implementing new features, follow this standard development process:

### 1. Create Feature Branch
```bash
git checkout -b feature/your-feature-name
```

### 2. Write Design Document
Create a markdown design document that includes:
- **Problem Statement**: What issue are you solving?
- **Proposed Solution**: High-level approach and architecture
- **Implementation Details**: Key components, interfaces, and data structures
- **Testing Plan**: How you'll verify the feature works
- **Considerations**: Security, performance, backwards compatibility

Example:
```bash
# Create design doc
touch docs/feature-your-feature-name.md
# Write the design, get feedback before implementation
```

### 3. Implementation
- Follow existing code patterns and conventions
- Write clean, well-documented code
- Add appropriate error handling
- Ensure backwards compatibility where possible

### 4. Testing
- Build and test locally: `go build -o bin/installer ./installer`
- Test export functionality: `./bin/installer export <set>`
- Test apply functionality with appropriate kubeconfig
- Verify manifests are generated correctly

### 5. Commit Changes
```bash
git add .
git commit -m "feat: Add compliance container to collector DaemonSet

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

### 6. Push and Open PR
```bash
git push -u origin feature/your-feature-name
gh pr create --title "Add compliance container to collector DaemonSet" --body "$(cat <<'EOF'
## Summary
<1-3 bullet points>

## Test plan
[Checklist of TODOs for testing the pull request...]

🤖 Generated with [Claude Code](https://claude.ai/code)
EOF
)"
```

### Notes
- Always create design docs for non-trivial features
- Get design review before starting implementation
- Keep commits atomic and well-described
- Include testing instructions in PR descriptions