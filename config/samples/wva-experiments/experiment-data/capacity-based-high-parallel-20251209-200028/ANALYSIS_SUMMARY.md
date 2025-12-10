# WVA Capacity-Based Experiment Analysis

**Experiment:** capacity-based-high-parallel-20251209-200028

**Generated:** 2025-12-10 09:34:36

---

## Summary Statistics

- **KV Cache Measurements:** 63
- **Queue Measurements:** 63
- **Capacity Analyses:** 0
- **Scaling Decisions:** 35

## KV Cache Utilization

- Average: 32.0%
- Peak: 100.0%
- Min: 0.0%
- Saturation Events (>90.0%): 8/35 (22.9%)

## Queue Length

- Average: 49.4 requests
- Peak: 517 requests
- Queue Buildup Events: 8

## Scaling Behavior

- no-change: 29 (82.9%)
- scale-up: 3 (8.6%)
- scale-down: 3 (8.6%)

- Replica Range: 1 - 4

## Performance Metrics

- TTFT: 1076.72 ms (avg), 8512.82 ms (max)
- ITL: 39.18 ms (avg), 92.34 ms (max)
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
