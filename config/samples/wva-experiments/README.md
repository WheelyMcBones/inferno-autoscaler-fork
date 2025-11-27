# WVA (Workload Variant Autoscaler) Experiments

This directory contains tools and notebooks for running reproducible experiments with WVA in both **Model-Based** and **Capacity-Based** modes.

## Directory Structure

```
wva-experiments/
├── README.md                           # This file
├── run-experiment.sh                   # Main experiment runner
├── monitor-wva.sh                      # WVA log collection script
├── experiment-configs/                 # Experiment configuration files
│   ├── model-based-moderate.yaml      # Model-based mode, moderate load
│   ├── model-based-high.yaml          # Model-based mode, high load
│   ├── capacity-based-moderate.yaml   # Capacity-based mode, moderate load
│   └── capacity-based-high.yaml       # Capacity-based mode, high load
├── experiment-data/                    # Collected experiment data
│   └── <experiment-name>-<timestamp>/ # Each experiment creates a timestamped directory
│       ├── experiment-config.yaml     # Config snapshot
│       ├── wva-controller-logs.jsonl  # Raw WVA controller logs
│       ├── metrics.csv                # Parsed metrics
│       └── README.md                  # Experiment summary
├── analyze-model-based.ipynb          # Notebook for model-based mode analysis
├── analyze-capacity-based.ipynb       # Notebook for capacity-based mode analysis
└── workloads/                         # Symlink to ../not_wva/workloads
```

## Quick Start

### Prerequisites

1. **Deploy your vLLM workload** (this setup doesn't handle deployment)
2. **Configure WVA** with the desired mode:
   - Model-based: Set `EXPERIMENTAL_HYBRID_OPTIMIZATION=model-only`
   - Capacity-based: Set `EXPERIMENTAL_HYBRID_OPTIMIZATION=off` (or unset)
3. **Install dependencies**:
   ```bash
   pip install pandas matplotlib jupyter
   brew install yq jq  # macOS
   ```

### Running an Experiment

```bash
# Run a moderate-load model-based experiment
./run-experiment.sh experiment-configs/model-based-moderate.yaml

# Run a high-load capacity-based experiment
./run-experiment.sh experiment-configs/capacity-based-high.yaml
```

The script will:
1. Start WVA log collection
2. Launch load generation jobs sequentially
3. Monitor metrics in real-time
4. Save all data to `experiment-data/<experiment-name>-<timestamp>/`

### Analyzing Results

Open the appropriate Jupyter notebook:

**For Model-Based Mode:**
```bash
jupyter notebook analyze-model-based.ipynb
```
This notebook extracts:
- Scaling decisions (target replicas, current replicas)
- **Predicted metrics** (ITL, TTFT from optimizer at reconciliation N)
- **Observed metrics** (actual ITL, TTFT at reconciliation N+1)
- SLO values and violations
- Model-based targets and optimization details

**For Capacity-Based Mode:**
```bash
jupyter notebook analyze-capacity-based.ipynb
```
This notebook extracts:
- Scaling decisions (scale-up, scale-down, no-change)
- Observed metrics (ITL, TTFT, replicas)
- **KV cache usage** per pod
- **Queue length** per pod
- Capacity analysis results (saturated/non-saturated replicas)

## Experiment Configurations

The `experiment-configs/` directory contains 8 pre-configured experiment files:

### Sequential Experiments (Original)
Jobs run one after another - simpler, longer duration:

- **`model-based-moderate.yaml`** - Sequential moderate load (10-15 req/min)
- **`model-based-high.yaml`** - Sequential high load (20-30 req/min)
- **`capacity-based-moderate.yaml`** - Sequential moderate load
- **`capacity-based-high.yaml`** - Sequential high load

### Parallel Experiments (Overlapping Jobs)
Jobs run concurrently with staggered starts - matches HPA experiment pattern:

- **`model-based-moderate-parallel.yaml`** - Overlapping moderate (10+15+12 req/s, peak 37 req/s)
- **`model-based-high-parallel.yaml`** - Overlapping high (20+30+15 req/s, peak 65 req/s)
- **`capacity-based-moderate-parallel.yaml`** - Overlapping moderate (capacity mode)
- **`capacity-based-high-parallel.yaml`** - Overlapping high (capacity mode)

**Parallel mode timeline example** (moderate-parallel):
```
Time:    0s        120s       240s       360s       480s       600s
Job1:    [========================================]  (10 req/s)
Job2:              [========================================]  (15 req/s)
Job3:                         [========================================]  (12 req/s)
Load:    10        25         37         27         12         0
```

This creates realistic overlapping load patterns that stress-test WVA's:
- **Model-based**: Ability to predict during dynamic load changes
- **Capacity-based**: Saturation detection under cumulative pressure

## Experiment Configuration Format

Each YAML config file defines:

```yaml
name: model-based-moderate-load
description: "Model-based WVA with moderate ShareGPT load"
mode: model-based  # or capacity-based

namespace: llm-d-autoscaler
controller_namespace: workload-variant-autoscaler-system
controller_pod_prefix: workload-variant-autoscaler-controller-manager
deployment: ms-inference-scheduling-llm-d-modelservice-decode
model_name: unsloth/Meta-Llama-3.1-8B

metrics:
  interval: 10  # seconds between log polls

workloads:
  - name: warmup
    job_manifest: ../not_wva/workloads/sharegpt-load-job-warmup.yaml
    wait_completion: true
  - name: moderate-10
    job_manifest: ../not_wva/workloads/sharegpt-load-job-moderate-10.yaml
    wait_completion: true
  - name: moderate-15
    job_manifest: ../not_wva/workloads/sharegpt-load-job-moderate-15.yaml
    wait_completion: true

output:
  base_dir: ./experiment-data
```

## Key Differences Between Modes

### Model-Based Mode
- Uses **predictive optimizer** to determine replica count
- Predicts future ITL/TTFT based on workload profile
- Logs show:
  - `Optimization solution` with predicted metrics
  - `alloc={acc=H100; numRep=X; maxBatch=512; itl=Y, ttft=Z}`
  - Model-based targets: `map[variant:X]`

### Capacity-Based Mode
- Uses **reactive capacity analysis** (KV cache + queue metrics)
- Scales based on current saturation
- Logs show:
  - `KV cache metric` per pod
  - `Queue metric` per pod
  - `Capacity analysis completed` with saturation info
  - `Metrics collected for VA` with observed ITL/TTFT

## Interpreting Results

### Model-Based Analysis
- **Prediction Accuracy**: Compare predicted ITL/TTFT (reconciliation N) with observed values (reconciliation N+1)
- **SLO Compliance**: Check if observed metrics meet SLO thresholds
- **Scaling Behavior**: Analyze when/why optimizer changes replica count

### Capacity-Based Analysis
- **Saturation Detection**: When does WVA detect saturated replicas?
- **Scale-down Safety**: How does it ensure safe scale-down?
- **KV Cache Patterns**: Correlation between cache usage and scaling
- **Queue Dynamics**: Impact of queue length on decisions

## Workload Files

Use the same workloads as HPA experiments (symlinked from `../not_wva/workloads/`):
- `sharegpt-load-job-warmup.yaml` - Initial warmup (5 minutes, low load)
- `sharegpt-load-job-moderate-10.yaml` - 10 req/min
- `sharegpt-load-job-moderate-12.yaml` - 12 req/min
- `sharegpt-load-job-moderate-15.yaml` - 15 req/min
- `sharegpt-load-job-high-20.yaml` - 20 req/min
- `sharegpt-load-job-high-30.yaml` - 30 req/min

## Troubleshooting

### WVA Not Scaling
1. Check WVA mode: `kubectl logs -n workload-variant-autoscaler-system <pod> | grep "Operating in"`
2. Verify VariantAutoscaling CRD exists: `kubectl get variantautoscaling -n llm-d-autoscaler`
3. Check metrics emission: Look for `EmitReplicaMetrics completed` in logs

### No Metrics in Logs
1. Ensure Prometheus is accessible from WVA controller
2. For model-based: Check that model profiles exist in VA spec
3. For capacity-based: Verify vLLM exposes KV cache and queue metrics

### Logs Not Collected
1. Check controller pod name: `kubectl get pods -n workload-variant-autoscaler-system`
2. Verify namespace in config matches actual deployment
3. Ensure sufficient RBAC permissions for log access

## Analysis Notebooks

### Model-Based Mode (`analyze-model-based.ipynb`)

Analyzes experiments running with **MODEL-ONLY mode** (predictive optimizer).

**Key Features:**
- Aligns predictions (reconciliation N) with observations (reconciliation N+1)
- Computes prediction accuracy for ITL and TTFT
- Detects SLO violations
- Visualizes scaling decisions and replica changes
- Generates summary statistics and error distributions

**Outputs:**
- `processed_data.csv` - Aligned prediction-observation pairs
- `plots/prediction_vs_observation.png` - Time series comparison
- `plots/scaling_behavior.png` - Replica timeline
- `plots/prediction_error_distribution.png` - Error histograms
- `ANALYSIS_SUMMARY.md` - Text summary report

**Usage:**
```bash
# Run experiment first
./run-experiment.sh experiment-configs/model-based-moderate.yaml

# Open Jupyter
jupyter notebook analyze-model-based.ipynb

# Or run all cells programmatically
jupyter execute analyze-model-based.ipynb
```

### Capacity-Based Mode (`analyze-capacity-based.ipynb`)

Analyzes experiments running with **CAPACITY-ONLY mode** (reactive scaling).

**Key Features:**
- Aggregates KV cache utilization across pods
- Tracks queue length buildup
- Detects saturation events (>90% KV cache)
- Correlates capacity metrics with scaling decisions
- Per-pod and aggregate visualizations

**Outputs:**
- `kv_cache_aggregated.csv` - KV cache metrics summary
- `queue_aggregated.csv` - Queue metrics summary
- `capacity_analysis.csv` - Saturation detection timeline
- `plots/kv_cache_utilization.png` - KV cache over time
- `plots/queue_length.png` - Queue length over time
- `plots/scaling_with_context.png` - Scaling with metrics overlay
- `plots/per_pod_kv_cache.png` - Per-pod breakdown
- `ANALYSIS_SUMMARY.md` - Text summary report

**Usage:**
```bash
# Run experiment first
./run-experiment.sh experiment-configs/capacity-based-moderate.yaml

# Open Jupyter
jupyter notebook analyze-capacity-based.ipynb

# Or run all cells programmatically
jupyter execute analyze-capacity-based.ipynb
```

**Important Notes:**
- The notebooks automatically detect the latest experiment of the matching mode
- Manually change `EXPERIMENT_DIR` if you want to analyze a specific run
- Plots are saved to `experiment-data/<experiment>/plots/`
- Both notebooks generate a `ANALYSIS_SUMMARY.md` with key findings

## References

- [HPA Experiments](../not_wva/new-experiments/) - Similar setup for HPA
- [WVA Controller Code](../../../internal/controller/variantautoscaling_controller.go)
- [Model Analyzer](../../../internal/modelanalyzer/analyzer.go)
- [Capacity Analyzer](../../../internal/capacity/analyzer.go)
