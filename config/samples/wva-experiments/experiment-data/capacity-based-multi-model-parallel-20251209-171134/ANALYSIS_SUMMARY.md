# WVA Capacity-Based Experiment Analysis

**Experiment:** capacity-based-multi-model-parallel-20251209-171134

**Generated:** 2025-12-10 09:48:04

---

## Summary Statistics

- **KV Cache Measurements:** 45
- **Queue Measurements:** 45
- **Capacity Analyses:** 0
- **Scaling Decisions:** 42

## KV Cache Utilization

- Average: 10.9%
- Peak: 89.4%
- Min: 0.0%
- Saturation Events (>90.0%): 0/42 (0.0%)

## Queue Length

- Average: 0.0 requests
- Peak: 0 requests
- Queue Buildup Events: 0

## Scaling Behavior

- no-change: 40 (95.2%)
- scale-up: 1 (2.4%)
- scale-down: 1 (2.4%)

- Replica Range: 1 - 2

## Performance Metrics

- TTFT: 25.40 ms (avg), 75.80 ms (max)
- ITL: 13.45 ms (avg), 47.26 ms (max)
- Arrival Rate: No data collected

## Files Generated

- `kv_cache_aggregated.csv` - Processed metrics
- `queue_aggregated.csv` - Processed metrics

### Plots

- `plots/kv_cache_utilization.png` - KV cache timeline
- `plots/queue_length.png` - Queue length timeline
- `plots/latencies.png` - TTFT and ITL latencies
- `plots/arrival_rate.png` - Request arrival rate over time
- `plots/replica_scaling.png` - Replica scaling behavior
- `plots/combined_summary.png` - Combined multi-panel summary
- `plots/per_pod_kv_cache.png` - Per-pod KV cache utilization
