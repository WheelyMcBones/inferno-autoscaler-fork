# WVA Capacity-Based Experiment Analysis

**Experiment:** capacity-based-moderate-parallel-20251127-205831

**Generated:** 2025-11-28 15:45:13

---

## Summary Statistics

- **KV Cache Measurements:** 68
- **Queue Measurements:** 121
- **Capacity Analyses:** 104
- **Scaling Decisions:** 52

## KV Cache Utilization

- Average: 21.3%
- Peak: 89.7%
- Min: 0.0%
- Saturation Events (>90.0%): 0/52 (0.0%)

## Queue Length

- Average: 0.0 requests
- Peak: 0 requests
- Queue Buildup Events: 0

## Capacity Analysis

- Scale-Up Recommendations: 20/104
- Scale-Down Safe: 12/104

## Scaling Behavior

- no-change: 48 (92.3%)
- scale-up: 2 (3.8%)
- scale-down: 2 (3.8%)

- Replica Range: 1 - 2

## Performance Metrics

- TTFT: 39.32 ms (avg), 87.37 ms (max)
- ITL: 21.76 ms (avg), 54.07 ms (max)
- Arrival Rate: 1395.06 req/s (avg), 2512.00 req/s (max)

## Files Generated

- `kv_cache_aggregated.csv` - Processed metrics
- `queue_aggregated.csv` - Processed metrics
- `capacity_analysis.csv` - Processed metrics

### Plots

- `plots/kv_cache_utilization.png` - KV cache timeline
- `plots/queue_length.png` - Queue length timeline
- `plots/latencies.png` - TTFT and ITL latencies
- `plots/arrival_rate.png` - Request arrival rate over time
- `plots/replica_scaling.png` - Replica scaling behavior
- `plots/combined_summary.png` - Combined multi-panel summary
- `plots/per_pod_kv_cache.png` - Per-pod KV cache utilization
