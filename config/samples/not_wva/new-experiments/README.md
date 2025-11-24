# New Experiment System

Configuration-driven HPA experiments with TTFT/ITL metrics collection.

## Quick Start

```bash
./start-experiments.sh
```

## Components

- **start-experiments.sh**: Interactive menu to run experiments and view results
- **run-experiment.sh**: Executes experiments from YAML configs
- **monitor-hpa-enhanced.sh**: Collects HPA metrics + TTFT/ITL from Prometheus
- **experiment-configs/**: YAML configuration files defining experiment setups
- **analyze-hpa-experiment.ipynb**: Jupyter notebook for data analysis

## Configuration Format

```yaml
experiment:
  name: baseline-hpa
  hpa_enabled: true
  phases:
    - name: warmup
      job_manifest: sharegpt-load-job-warmup.yaml
      duration_seconds: 60
    - name: load-phase-1
      job_manifest: sharegpt-load-job-1.yaml
      duration_seconds: 300
```

## Metrics Collected

- **TTFT**: Time to First Token (ms) - identical to WVA collector
- **ITL**: Inter-Token Latency (ms) - identical to WVA collector  
- **HPA Metrics**: replicas, kv_cache_usage, num_requests_waiting, request_rate

Output: CSV files in `../experiment-data/`

## Creating New Experiments

1. Copy `experiment-configs/baseline-hpa.yaml`
2. Modify phases and parameters
3. Run via `./start-experiments.sh`
