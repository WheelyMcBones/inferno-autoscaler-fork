# WVA Capacity-Based Experiment Analysis

**Experiment:** capacity-based-multi-model-parallel-20251209-141241

**Generated:** 2025-12-09 15:08:03

---

## Summary Statistics

- **KV Cache Measurements:** 37
- **Queue Measurements:** 37
- **Capacity Analyses:** 0
- **Scaling Decisions:** 33

## KV Cache Utilization

- Average: 14.8%
- Peak: 100.0%
- Min: 0.0%
- Saturation Events (>90.0%): 2/33 (6.1%)

## Queue Length

- Average: 0.6 requests
- Peak: 40 requests
- Queue Buildup Events: 1

## Scaling Behavior

- no-change: 31 (93.9%)
- scale-up: 1 (3.0%)
- scale-down: 1 (3.0%)

- Replica Range: 1 - 2

## Performance Metrics

- TTFT: 28.99 ms (avg), 118.98 ms (max)
- ITL: 14.42 ms (avg), 55.79 ms (max)
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
