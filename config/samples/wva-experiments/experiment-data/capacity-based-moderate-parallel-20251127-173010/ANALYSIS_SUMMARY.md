# WVA Capacity-Based Experiment Analysis

**Experiment:** capacity-based-moderate-parallel-20251127-173010

**Generated:** 2025-11-27 17:48:25

---

## Summary Statistics

- **KV Cache Measurements:** 160
- **Queue Measurements:** 160
- **Capacity Analyses:** 136
- **Scaling Decisions:** 4

## KV Cache Utilization

- Average: 21.3%
- Peak: 100.0%
- Min: 0.0%
- Saturation Events (>90.0%): 16/68 (23.5%)

## Queue Length

- Average: 40.4 requests
- Peak: 514 requests
- Queue Buildup Events: 14

## Capacity Analysis

- Scale-Up Recommendations: 16/136
- Scale-Down Safe: 12/136

## Scaling Behavior

- scale-up: 2 (50.0%)
- scale-down: 2 (50.0%)

- Replica Range: 1 - 2

## Performance Metrics

- TTFT: 1806.36 ms (avg), 8421.97 ms (max)
- ITL: 35.44 ms (avg), 66.51 ms (max)

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
