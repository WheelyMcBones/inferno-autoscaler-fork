# HPA Experiments with TTFT/ITL Metrics

Tools for running HPA scaling experiments with vLLM, collecting performance metrics (TTFT, ITL) for comparison with WVA.

## Quick Start

```bash
# Interactive launcher - choose between new or legacy system
./launch.sh

# Or go directly to new system (recommended)
cd new-experiments
./start-experiments.sh
```

## Directory Structure

```bash
not_wva/
├── launch.sh                    # Main launcher
├── new-experiments/             # ✅ Recommended: Config-driven system with TTFT/ITL
├── legacy-scripts/              # Original manual experiment system
├── scripts/                     # Shared setup scripts
├── manifests/                   # Kubernetes configs (HPA, Prometheus)
├── workloads/                   # Load generation jobs
└── experiment-data/             # Results (auto-created)
```

## Prerequisites

Required:

- `kubectl`, `oc` (OpenShift CLI)
- `yq` (install: `brew install yq`)
- `jq`

For analysis:

- `python3` with pandas, matplotlib
- `jupyter` (optional, for notebooks)

## One-Time Setup

### 1. Deploy vLLM Service

```bash
kubectl apply -f manifests/vllm-service.yaml
```

Verify:

```bash
kubectl get svc vllm-service -n llm-d-inference-scheduler
kubectl get endpoints vllm-service -n llm-d-inference-scheduler
```

### 2. Setup Prometheus

```bash
# Create CA certificate
./scripts/create-prometheus-ca.sh

# Deploy Prometheus Adapter
helm upgrade prometheus-adapter prometheus-community/prometheus-adapter \
  -n openshift-user-workload-monitoring \
  -f manifests/prometheus-adapter-vllm-values.yaml

# Verify custom metrics
kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1" | jq .
```

### 3. Deploy HPA

```bash
kubectl apply -f manifests/hpa-vllm-combined.yaml

# Verify
kubectl get hpa vllm-hpa-combined -n llm-d-inference-scheduler
```

## Running Experiments

See `new-experiments/README.md` for detailed usage.

Quick example:

```bash
cd new-experiments
./start-experiments.sh
# Select option 1, then choose baseline-hpa
```

## Troubleshooting

**No metrics data:**

```bash
# Check if vLLM is up
kubectl get pods -n llm-d-inference-scheduler -l llm-d.ai/model=ms-inference-scheduling-llm-d-modelservice

# Check if receiving requests
kubectl logs -n llm-d-inference-scheduler <pod-name> -c vllm --tail=20
```

**HPA not scaling:**

```bash
# Check current metrics vs targets
kubectl get hpa vllm-hpa-combined -n llm-d-inference-scheduler
```

**Jobs not starting:**

```bash
# Check service exists
kubectl get svc vllm-service -n llm-d-inference-scheduler
```
