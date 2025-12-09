# WVA Capacity-Based Experiment Analysis

**Experiment:** capacity-based-multi-model-parallel-20251209-160253

**Generated:** 2025-12-09 17:08:13

---

## Summary Statistics

- **KV Cache Measurements:** 38
- **Queue Measurements:** 38
- **Capacity Analyses:** 0
- **Scaling Decisions:** 35

## KV Cache Utilization

- Average: 14.6%
- Peak: 98.9%
- Min: 0.0%
- Saturation Events (>90.0%): 2/35 (5.7%)

## Queue Length

- Average: 0.8 requests
- Peak: 55 requests
- Queue Buildup Events: 1

## Scaling Behavior

- no-change: 33 (94.3%)
- scale-up: 1 (2.9%)
- scale-down: 1 (2.9%)

- Replica Range: 1 - 2

## Performance Metrics

- TTFT: 28.68 ms (avg), 122.13 ms (max)
- ITL: 14.50 ms (avg), 53.92 ms (max)
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
