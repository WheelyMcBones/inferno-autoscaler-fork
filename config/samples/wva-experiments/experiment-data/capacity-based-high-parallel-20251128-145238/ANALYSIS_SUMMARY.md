# WVA Capacity-Based Experiment Analysis

**Experiment:** capacity-based-high-parallel-20251128-145238

**Generated:** 2025-11-28 15:11:22

---

## Summary Statistics

- **KV Cache Measurements:** 118
- **Queue Measurements:** 202
- **Capacity Analyses:** 168
- **Scaling Decisions:** 84

## KV Cache Utilization

- Average: 21.1%
- Peak: 99.6%
- Min: 0.0%
- Saturation Events (>90.0%): 6/84 (7.1%)

## Queue Length

- Average: 5.7 requests
- Peak: 122 requests
- Queue Buildup Events: 4

## Capacity Analysis

- Scale-Up Recommendations: 32/168
- Scale-Down Safe: 24/168

## Scaling Behavior

- no-change: 76 (90.5%)
- scale-up: 4 (4.8%)
- scale-down: 4 (4.8%)

- Replica Range: 1 - 3

## Performance Metrics

- TTFT: 773.36 ms (avg), 16512.52 ms (max)
- ITL: 26.26 ms (avg), 70.01 ms (max)
- Arrival Rate: 1507.20 req/s (avg), 3161.15 req/s (max)

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
