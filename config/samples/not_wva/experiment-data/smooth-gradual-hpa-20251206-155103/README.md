# Experiment: smooth-gradual-hpa

**Date**: 2025-12-06 16:16:01
**Description**: HPA baseline with smooth gradual load transitions (30+ mins, 6 jobs @ 5 req/s each)

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
python scripts/analyze-hpa-experiment.py experiment-data/smooth-gradual-hpa-20251206-155103/metrics.csv
# or
jupyter notebook analyze-hpa-experiment.ipynb
```
