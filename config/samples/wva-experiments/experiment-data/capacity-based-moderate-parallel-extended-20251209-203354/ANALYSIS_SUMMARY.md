# WVA Capacity-Based Experiment Analysis

**Experiment:** capacity-based-moderate-parallel-extended-20251209-203354

**Generated:** 2025-12-10 09:58:57

---

## Summary Statistics

- **KV Cache Measurements:** 32
- **Queue Measurements:** 32
- **Capacity Analyses:** 0
- **Scaling Decisions:** 28

## KV Cache Utilization

- Average: 16.7%
- Peak: 79.4%
- Min: 0.0%
- Saturation Events (>90.0%): 0/28 (0.0%)

## Queue Length

- Average: 0.0 requests
- Peak: 0 requests
- Queue Buildup Events: 0

## Scaling Behavior

- no-change: 24 (85.7%)
- scale-up: 2 (7.1%)
- scale-down: 2 (7.1%)

- Replica Range: 1 - 2

## Performance Metrics

- TTFT: 31.45 ms (avg), 80.78 ms (max)
- ITL: 17.06 ms (avg), 48.44 ms (max)
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
