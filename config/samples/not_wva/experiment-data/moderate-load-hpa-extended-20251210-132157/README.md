# Experiment: moderate-load-hpa-extended

**Date**: 2025-12-10 13:46:51
**Description**: Extended HPA test with overlapping 10-12 req/s jobs (20+ mins, peak ~32 req/s)

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
python scripts/analyze-hpa-experiment.py experiment-data/moderate-load-hpa-extended-20251210-132157/metrics.csv
# or
jupyter notebook analyze-hpa-experiment.ipynb
```
