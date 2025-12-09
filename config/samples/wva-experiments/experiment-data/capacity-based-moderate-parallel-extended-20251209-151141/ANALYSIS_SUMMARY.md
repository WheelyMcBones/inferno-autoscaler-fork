# WVA Capacity-Based Experiment Analysis

**Experiment:** capacity-based-moderate-parallel-extended-20251209-151141

**Generated:** 2025-12-09 15:43:08

---

## Summary Statistics

- **KV Cache Measurements:** 31
- **Queue Measurements:** 31
- **Capacity Analyses:** 0
- **Scaling Decisions:** 26

## KV Cache Utilization

- Average: 20.3%
- Peak: 89.6%
- Min: 0.0%
- Saturation Events (>90.0%): 0/26 (0.0%)

## Queue Length

- Average: 0.0 requests
- Peak: 0 requests
- Queue Buildup Events: 0

## Scaling Behavior

- no-change: 22 (84.6%)
- scale-up: 2 (7.7%)
- scale-down: 2 (7.7%)

- Replica Range: 1 - 2

## Performance Metrics

- TTFT: 32.58 ms (avg), 82.16 ms (max)
- ITL: 18.35 ms (avg), 49.93 ms (max)
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
