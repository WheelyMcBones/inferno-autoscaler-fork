# WVA Capacity-Based Experiment Analysis

**Experiment:** capacity-based-moderate-parallel-20251126-182520

**Generated:** 2025-11-27 16:07:42

---

## Summary Statistics

- **KV Cache Measurements:** 36
- **Queue Measurements:** 36
- **Capacity Analyses:** 32
- **Scaling Decisions:** 4

## KV Cache Utilization

- Average: 27.8%
- Peak: 99.8%
- Min: 0.0%
- Saturation Events (>90.0%): 6/32 (18.8%)

## Queue Length

- Average: 61.2 requests
- Peak: 600 requests
- Queue Buildup Events: 6

## Capacity Analysis

- Scale-Up Recommendations: 6/32
- Scale-Down Safe: 2/32

## Scaling Behavior

- scale-up: 2 (50.0%)
- scale-down: 2 (50.0%)

- Replica Range: 1 - 2

## Performance Metrics

- TTFT: 3056.02 ms (avg), 16345.01 ms (max)
- ITL: 38.13 ms (avg), 90.55 ms (max)

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
