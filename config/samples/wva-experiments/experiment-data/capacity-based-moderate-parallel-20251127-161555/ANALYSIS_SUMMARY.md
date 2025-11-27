# WVA Capacity-Based Experiment Analysis

**Experiment:** capacity-based-moderate-parallel-20251127-161555

**Generated:** 2025-11-27 16:35:59

---

## Summary Statistics

- **KV Cache Measurements:** 40
- **Queue Measurements:** 40
- **Capacity Analyses:** 34
- **Scaling Decisions:** 6

## KV Cache Utilization

- Average: 26.4%
- Peak: 100.0%
- Min: 0.0%
- Saturation Events (>90.0%): 8/34 (23.5%)

## Queue Length

- Average: 78.5 requests
- Peak: 810 requests
- Queue Buildup Events: 6

## Capacity Analysis

- Scale-Up Recommendations: 6/34
- Scale-Down Safe: 4/34

## Scaling Behavior

- scale-up: 4 (66.7%)
- scale-down: 2 (33.3%)

- Replica Range: 1 - 2

## Performance Metrics

- TTFT: 3942.35 ms (avg), 18962.01 ms (max)
- ITL: 39.35 ms (avg), 96.18 ms (max)

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
