# Developer Guide

Guide for developers contributing to Workload-Variant-Autoscaler.

## Development Environment Setup

### Prerequisites

- Go 1.23.0+
- Docker 17.03+
- kubectl 1.32.0+
- Kind (for local testing)
- Make

### Initial Setup

1. **Clone the repository:**
   ```bash
   git clone https://github.com/llm-d-incubation/workload-variant-autoscaler.git
   cd workload-variant-autoscaler
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Install development tools:**
   ```bash
   make setup-envtest
   make controller-gen
   make kustomize
   ```

## Project Structure

```
workload-variant-autoscaler/
├── api/v1alpha1/          # CRD definitions
├── cmd/                   # Main application entry points
├── config/                # Kubernetes manifests
│   ├── crd/              # CRD manifests
│   ├── rbac/             # RBAC configurations
│   ├── manager/          # Controller deployment
│   └── samples/          # Example resources
├── deploy/                # Deployment scripts
│   ├── kubernetes/       # K8s deployment
│   ├── openshift/        # OpenShift deployment
│   └── kind/             # Local development
├── docs/                  # Documentation
├── internal/              # Private application code
│   ├── controller/       # Controller implementation
│   ├── collector/        # Metrics collection
│   ├── optimizer/        # Optimization logic
│   ├── actuator/         # Metric emission & scaling
│   └── modelanalyzer/    # Model analysis
├── pkg/                   # Public libraries
│   ├── analyzer/         # Queue theory models
│   ├── solver/           # Optimization algorithms
│   ├── core/             # Core domain models
│   └── config/           # Configuration structures
├── test/                  # Tests
│   ├── e2e/              # End-to-end tests
│   └── utils/            # Test utilities
└── tools/                 # Development tools
    └── vllm-emulator/    # Testing emulator
```

## Development Workflow

### Running Locally

**Option 1: Outside the cluster**
```bash
# Run the controller on your machine (connects to configured cluster)
make run
```

**Option 2: In a Kind cluster**
```bash
# Create a Kind cluster with emulated GPUs
make create-kind-cluster

# Deploy the controller
make deploy IMG=<your-image>

# Or deploy with llm-d infrastructure
make deploy-llm-d-wva-emulated-on-kind
```

### Making Changes

1. **Create a feature branch:**
   ```bash
   git checkout -b feature/my-feature
   ```

2. **Make your changes**

3. **Generate code if needed:**
   ```bash
   # After modifying CRDs
   make manifests generate
   ```

4. **Run tests:**
   ```bash
   make test
   ```

5. **Run linter:**
   ```bash
   make lint
   ```

## Building and Testing

### Build the Binary

```bash
make build
```

The binary will be in `bin/manager`.

### Build Docker Image

```bash
make docker-build IMG=<your-registry>/wva-controller:tag
```

### Push Docker Image

```bash
make docker-push IMG=<your-registry>/wva-controller:tag
```

### Multi-architecture Build

```bash
PLATFORMS=linux/arm64,linux/amd64 make docker-buildx IMG=<your-registry>/wva-controller:tag
```

## Testing

### Unit Tests

```bash
# Run all unit tests
make test

# Run specific package tests
go test ./internal/optimizer/...

# With coverage
go test -cover ./...
```

### E2E Tests

#### Kind E2E Tests

```bash
# Run all E2E tests
make test-e2e

# Run specific tests
make test-e2e FOCUS="single VA"

# Skip specific tests
make test-e2e SKIP="multiple VA"
```

#### OpenShift E2E Tests

```bash
# Run E2E tests on OpenShift cluster
make test-e2e-openshift

# With custom image
make test-e2e-openshift IMG=<your-registry>/wva-controller:tag

# Run specific OpenShift tests
make test-e2e-openshift FOCUS="HPA integration"
```

**Prerequisites for OpenShift E2E:**
- Access to an OpenShift cluster (OCP 4.12+)
- `oc` CLI tool configured and authenticated
- Cluster admin permissions
- Prometheus operator installed

See [Testing Guide](testing.md) for more details.

### Manual Testing

1. **Deploy to Kind cluster:**
   ```bash
   make deploy-llm-d-wva-emulated-on-kind IMG=<your-image>
   ```

2. **Create test resources:**
   ```bash
   kubectl apply -f config/samples/
   ```

3. **Monitor controller logs:**
   ```bash
   kubectl logs -n workload-variant-autoscaler-system \
     deployment/workload-variant-autoscaler-controller-manager -f
   ```

## Code Generation

### After Modifying CRDs

```bash
# Generate deepcopy, CRD manifests, and RBAC
make manifests generate
```

### Generate CRD Documentation

```bash
make crd-docs
```

Output will be in `docs/user-guide/crd-reference.md`.

## Debugging

### VSCode Launch Configuration

Create `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Controller",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/cmd/main.go",
      "env": {
        "KUBECONFIG": "${env:HOME}/.kube/config"
      },
      "args": []
    }
  ]
}
```

### Debugging in Cluster

```bash
# Build debug image
go build -gcflags="all=-N -l" -o bin/manager cmd/main.go

# Deploy and attach debugger (e.g., Delve)
```

### Viewing Controller Logs

```bash
kubectl logs -n workload-variant-autoscaler-system \
  -l control-plane=controller-manager --tail=100 -f
```

## Common Development Tasks

### Adding a New Field to CRD

1. Modify `api/v1alpha1/variantautoscaling_types.go`
2. Run `make manifests generate`
3. Update tests
4. Run `make crd-docs`
5. Update user documentation

### Adding a New Metric

1. Define metric in `internal/metrics/metrics.go`
2. Emit metric from appropriate controller location
3. Update Prometheus integration docs
4. Add to Grafana dashboards (if applicable)

### Modifying Optimization Logic

1. Update code in `pkg/solver/` or `pkg/analyzer/`
2. Add/update unit tests
3. Run `make test`
4. Update design documentation if algorithm changes

## Documentation

### Updating Documentation

After code changes, update relevant docs in:
- `docs/user-guide/` - User-facing changes
- `docs/design/` - Architecture/design changes
- `docs/integrations/` - Integration guide updates

### Testing Documentation

Verify all commands and examples in documentation work:
```bash
# Test installation steps
# Test configuration examples
# Test all code snippets
```

## Release Process

See [Releasing Guide](releasing.md) (coming soon) for the release process.

## Getting Help

- Check [CONTRIBUTING.md](../../CONTRIBUTING.md)
- Review existing code and tests
- Ask in GitHub Discussions
- Attend community meetings

## Useful Commands

```bash
# Format code
make fmt

# Vet code
make vet

# Run linter
make lint

# Fix linting issues
make lint-fix

# Clean build artifacts
rm -rf bin/ dist/

# Reset Kind cluster
make destroy-kind-cluster
make create-kind-cluster
```

## Next Steps

- Read [Testing Guide](testing.md)
- Review [Code Style Guidelines](../../CONTRIBUTING.md#coding-guidelines)
- Check out [Good First Issues](https://github.com/llm-d-incubation/workload-variant-autoscaler/labels/good%20first%20issue)

