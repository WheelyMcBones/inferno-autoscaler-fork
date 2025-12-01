# WVA Capacity-Based Experiment Analysis

**Experiment:** capacity-based-high-parallel-20251128-155627

**Generated:** 2025-11-28 19:09:25

---

## Summary Statistics

- **KV Cache Measurements:** 330
- **Queue Measurements:** 330
- **Capacity Analyses:** 288
- **Scaling Decisions:** 144

## KV Cache Utilization

- Average: 24.6%
- Peak: 97.2%
- Min: 0.0%
- Saturation Events (>90.0%): 8/146 (5.5%)

## Queue Length

- Average: 0.0 requests
- Peak: 0 requests
- Queue Buildup Events: 0

## Capacity Analysis

- Scale-Up Recommendations: 32/288
- Scale-Down Safe: 20/288

## Scaling Behavior

- no-change: 136 (94.4%)
- scale-up: 4 (2.8%)
- scale-down: 4 (2.8%)

- Replica Range: 1 - 3

## Performance Metrics

- TTFT: 40.90 ms (avg), 162.53 ms (max)
- ITL: 21.85 ms (avg), 55.03 ms (max)
- Arrival Rate: No data collected

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
