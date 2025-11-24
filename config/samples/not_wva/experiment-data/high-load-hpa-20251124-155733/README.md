# Experiment: high-load-hpa

**Date**: 2025-11-24 16:08:04
**Description**: HPA stress test with overlapping 20-30 req/s jobs to force scaling

## Configuration
- Namespace: llm-d-inference-scheduler
- Deployment: ms-inference-scheduling-llm-d-modelservice-decode
- Model: unsloth/Meta-Llama-3.1-8B
- HPA Enabled: true

## Files
- `metrics.csv`: Time-series metrics data (replicas, TTFT, ITL, etc.)
- `scaling-events.log`: HPA scaling events log
- `experiment-config.yaml`: Full experiment configuration

## Metrics Collected
- Replicas (current and desired)
- Num Requests Waiting
- KV Cache Usage (%)
- TTFT (Time to First Token, ms)
- ITL (Inter-Token Latency, ms)
- Request Rate (req/min)
- Active Jobs

## Analysis
Use the analysis notebook or script to visualize results:
```bash
cd ../..
python scripts/analyze-hpa-experiment.py experiment-data/high-load-hpa-20251124-155733/metrics.csv
# or
jupyter notebook analyze-hpa-experiment.ipynb
```
