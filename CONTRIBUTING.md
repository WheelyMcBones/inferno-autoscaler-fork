# Contributing to Workload-Variant-Autoscaler

Welcome! We're excited that you're interested in contributing to the Workload-Variant-Autoscaler (WVA) project.

## General Contributing Guidelines

For general contribution guidelines including code of conduct, commit message format, PR process, and community standards, please see the **[llm-d Contributing Guide](https://github.com/llm-d/llm-d/blob/main/CONTRIBUTING.md)**.

This document covers **WVA-specific** development setup and workflows.

## WVA-Specific Development

### Prerequisites

- Go 1.23.0+
- Docker 17.03+
- kubectl 1.31.0+ (or `oc` for OpenShift)
- Kind (for local development)
- Basic understanding of Kubernetes controllers and operators

### Setting Up Your Development Environment

1. **Fork and clone the repository:**
   ```bash
   git clone https://github.com/<your-username>/workload-variant-autoscaler.git
   cd workload-variant-autoscaler
   ```

2. **Add upstream remote:**
   ```bash
   git remote add upstream https://github.com/llm-d-incubation/workload-variant-autoscaler.git
   ```

3. **Install dependencies:**
   ```bash
   go mod download
   ```

4. **Set up a local Kind cluster with emulated GPUs:**
   ```bash
   make deploy-llm-d-wva-emulated-on-kind
   ```

5. **Run tests to verify setup:**
   ```bash
   make test
   ```

See [Developer Guide](docs/developer-guide/development.md) for detailed setup instructions.

## WVA Project Structure

```
workload-variant-autoscaler/
├── api/v1alpha1/         # CRD definitions and types
├── cmd/                  # Main application entry point
├── config/               # Kubernetes manifests
│   ├── crd/             # CRD base manifests
│   ├── rbac/            # RBAC configurations
│   ├── manager/         # Controller deployment configs
│   └── samples/         # Example VariantAutoscaling CRs
├── deploy/               # Deployment scripts
│   ├── kubernetes/      # Standard K8s deployment
│   ├── openshift/       # OpenShift-specific deployment
│   └── kind-emulator/   # Local development with Kind
├── docs/                 # Documentation
│   ├── user-guide/      # User-facing documentation
│   ├── developer-guide/ # Development and testing guides
│   ├── integrations/    # Integration guides (HPA, KEDA, Prometheus)
│   ├── tutorials/       # Step-by-step tutorials
│   └── design/          # Architecture and design docs
├── internal/             # Private application code
│   ├── controller/      # Main reconciliation logic
│   ├── collector/       # Metrics collection
│   ├── optimizer/       # Optimization engine
│   ├── actuator/        # Metric emission & actuation
│   ├── modelanalyzer/   # Model performance analysis
│   ├── metrics/         # Metrics definitions
│   └── utils/           # Utility functions
├── pkg/                  # Public libraries (inferno optimizer)
│   ├── analyzer/        # Queue theory models
│   ├── solver/          # Optimization algorithms
│   ├── core/            # Core domain models
│   ├── config/          # Configuration structures
│   └── manager/         # Optimization manager
├── test/                 # Tests
│   ├── e2e/             # End-to-end tests (Kind)
│   ├── e2e-openshift/   # OpenShift E2E tests
│   └── utils/           # Test utilities
├── tools/                # Development tools
│   └── vllm-emulator/   # vLLM emulator for testing
└── charts/               # Helm charts
    └── workload-variant-autoscaler/
```

## WVA-Specific Development Tasks

### Testing Your Changes

**Run unit tests:**
```bash
make test
```

**Run E2E tests on Kind:**
```bash
make test-e2e

# Run specific tests
make test-e2e FOCUS="single VA"
make test-e2e SKIP="multiple VA"
```

**Run E2E tests on OpenShift:**
```bash
make test-e2e-openshift
```

**Run linter:**
```bash
make lint

# Auto-fix linting issues
make lint-fix
```

### Modifying CRDs

If you modify the `VariantAutoscaling` CRD in `api/v1alpha1/`:

1. **Generate updated manifests and code:**
   ```bash
   make manifests generate
   ```

2. **Update CRD documentation:**
   ```bash
   make crd-docs
   ```

3. **Verify CRD changes:**
   ```bash
   kubectl explain variantautoscaling.spec
   ```

### Building and Deploying

**Build the controller binary:**
```bash
make build
```

**Run controller locally (connects to configured cluster):**
```bash
make run
```

**Build Docker image:**
```bash
make docker-build IMG=<your-registry>/wva-controller:tag
```

**Deploy to cluster:**
```bash
make deploy IMG=<your-registry>/wva-controller:tag
```

**Deploy with llm-d for testing:**
```bash
make deploy-llm-d-wva-emulated-on-kind IMG=<your-registry>/wva-controller:tag
```

## Documentation

### Updating Documentation

When making code changes, update relevant documentation in:
- `docs/user-guide/` - User-facing changes (CRD changes, new features)
- `docs/developer-guide/` - Development workflow changes
- `docs/integrations/` - Integration guide updates
- `docs/design/` - Architecture or design changes
- `README.md` - High-level feature changes

### Testing Documentation

Verify all commands and examples work:
```bash
# Test installation steps from docs
# Test configuration examples
# Verify all code snippets are correct
```

## WVA-Specific Code Guidelines

### Controller Development

- Use the `logger` package from `internal/logger`
- Always use backoff retries for Kubernetes API calls (see `internal/utils`)
- Update Kubernetes conditions for status visibility
- Emit metrics for observability

### Performance Modeling

When modifying queue models in `pkg/analyzer/`:
- Ensure mathematical correctness
- Add comprehensive unit tests
- Validate against real workload data when possible
- Document assumptions and limitations

### Optimization Algorithms

When modifying solvers in `pkg/solver/`:
- Consider computational complexity
- Test edge cases (zero load, overload, etc.)
- Ensure feasibility checking
- Document algorithm choices

## Getting Help

- Check [Developer Guide](docs/developer-guide/development.md)
- Review existing code and tests
- Search [GitHub Issues](https://github.com/llm-d-incubation/workload-variant-autoscaler/issues)
- Ask in GitHub Discussions
- Attend llm-d community meetings

## Common Development Tasks

### Running the Controller Locally

```bash
# Option 1: Run outside cluster (connects to KUBECONFIG cluster)
make run

# Option 2: Deploy to Kind cluster
make deploy-llm-d-wva-emulated-on-kind
```

### Debugging

```bash
# View controller logs
kubectl logs -n workload-variant-autoscaler-system \
  -l control-plane=controller-manager --tail=100 -f

# Check VariantAutoscaling status
kubectl describe variantautoscaling <name> -n <namespace>

# Check emitted metrics
kubectl get --raw "/apis/external.metrics.k8s.io/v1beta1/namespaces/<namespace>/inferno_desired_replicas" | jq
```

### Cleaning Up

```bash
# Destroy Kind cluster
make destroy-kind-cluster

# Undeploy from cluster
make undeploy

# Uninstall CRDs
make uninstall
```

## Pre-Submission Checklist

Before submitting your PR, ensure:

- [ ] `make test` passes
- [ ] `make lint` passes (fix issues with `make lint-fix`)
- [ ] `make test-e2e` passes (if controller logic changed)
- [ ] Documentation updated (if user-facing changes)
- [ ] CRD docs regenerated (if CRD changed): `make crd-docs`
- [ ] Commit messages follow [conventional commits](https://www.conventionalcommits.org/)
- [ ] PR description clearly explains the change

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.

---

Thank you for contributing to Workload-Variant-Autoscaler!

For general llm-d contribution guidelines, visit the [llm-d Contributing Guide](https://github.com/llm-d/llm-d/blob/main/CONTRIBUTING.md).
