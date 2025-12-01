# WVA Capacity-Based Experiment Analysis

**Experiment:** capacity-based-moderate-parallel-20251127-181520

**Generated:** 2025-11-27 18:39:41

---

## Summary Statistics

- **KV Cache Measurements:** 72
- **Queue Measurements:** 126
- **Capacity Analyses:** 108
- **Scaling Decisions:** 54

## KV Cache Utilization

- Average: 16.3%
- Peak: 55.9%
- Min: 0.0%
- Saturation Events (>90.0%): 0/54 (0.0%)

## Queue Length

- Average: 0.0 requests
- Peak: 0 requests
- Queue Buildup Events: 0

## Capacity Analysis

- Scale-Up Recommendations: 24/108
- Scale-Down Safe: 12/108

## Scaling Behavior

- no-change: 46 (85.2%)
- scale-up: 4 (7.4%)
- scale-down: 4 (7.4%)

- Replica Range: 1 - 3

## Performance Metrics

- TTFT: 34.69 ms (avg), 59.91 ms (max)
- ITL: 19.32 ms (avg), 35.36 ms (max)

## Files Generated

- `kv_cache_aggregated.csv` - Processed metrics
- `queue_aggregated.csv` - Processed metrics
- `capacity_analysis.csv` - Processed metrics

### Plots

- `plots/kv_cache_utilization.png` - KV cache timeline
- `plots/queue_length.png` - Queue length timeline
- `plots/latencies.png` - TTFT and ITL latencies
- `plots/replica_scaling.png` - Replica scaling behavior
- `plots/combined_summary.png` - Combined multi-panel summary
- `plots/per_pod_kv_cache.png` - Per-pod KV cache utilization
