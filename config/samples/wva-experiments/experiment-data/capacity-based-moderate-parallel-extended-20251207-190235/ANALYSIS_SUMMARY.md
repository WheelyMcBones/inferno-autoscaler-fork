# WVA Capacity-Based Experiment Analysis

**Experiment:** capacity-based-moderate-parallel-extended-20251207-190235

**Generated:** 2025-12-07 19:26:10

---

## Summary Statistics

- **KV Cache Measurements:** 58
- **Queue Measurements:** 58
- **Capacity Analyses:** 0
- **Scaling Decisions:** 0

## KV Cache Utilization

- Average: 21.2%
- Peak: 82.9%
- Min: 0.0%
- Saturation Events (>90.0%): 0/50 (0.0%)

## Queue Length

- Average: 0.0 requests
- Peak: 0 requests
- Queue Buildup Events: 0

## Performance Metrics

- TTFT: 33.14 ms (avg), 77.83 ms (max)
- ITL: 18.41 ms (avg), 48.07 ms (max)
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
