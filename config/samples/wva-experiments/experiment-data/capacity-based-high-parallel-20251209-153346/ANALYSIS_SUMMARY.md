# WVA Capacity-Based Experiment Analysis

**Experiment:** capacity-based-high-parallel-20251209-153346

**Generated:** 2025-12-09 15:59:28

---

## Summary Statistics

- **KV Cache Measurements:** 64
- **Queue Measurements:** 64
- **Capacity Analyses:** 0
- **Scaling Decisions:** 27

## KV Cache Utilization

- Average: 41.0%
- Peak: 100.0%
- Min: 0.0%
- Saturation Events (>90.0%): 8/27 (29.6%)

## Queue Length

- Average: 126.4 requests
- Peak: 1015 requests
- Queue Buildup Events: 7

## Scaling Behavior

- no-change: 21 (77.8%)
- scale-up: 3 (11.1%)
- scale-down: 3 (11.1%)

- Replica Range: 1 - 4

## Performance Metrics

- TTFT: 1509.04 ms (avg), 18711.66 ms (max)
- ITL: 38.90 ms (avg), 72.27 ms (max)
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
