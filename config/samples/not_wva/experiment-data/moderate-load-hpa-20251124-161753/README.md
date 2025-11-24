# Experiment: moderate-load-hpa

**Date**: 2025-11-24 16:34:39
**Description**: HPA test with overlapping 10-15 req/s jobs for gradual scaling

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
python scripts/analyze-hpa-experiment.py experiment-data/moderate-load-hpa-20251124-161753/metrics.csv
# or
jupyter notebook analyze-hpa-experiment.ipynb
```
