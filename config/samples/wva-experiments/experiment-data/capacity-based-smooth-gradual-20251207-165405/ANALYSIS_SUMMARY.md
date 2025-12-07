# WVA Capacity-Based Experiment Analysis

**Experiment:** capacity-based-smooth-gradual-20251207-165405

**Generated:** 2025-12-07 17:35:50

---

## Summary Statistics

- **KV Cache Measurements:** 64
- **Queue Measurements:** 64
- **Capacity Analyses:** 0
- **Scaling Decisions:** 0

## KV Cache Utilization

- Average: 13.6%
- Peak: 65.1%
- Min: 0.0%
- Saturation Events (>90.0%): 0/60 (0.0%)

## Queue Length

- Average: 0.0 requests
- Peak: 0 requests
- Queue Buildup Events: 0

## Performance Metrics

- TTFT: 28.32 ms (avg), 53.97 ms (max)
- ITL: 15.16 ms (avg), 32.03 ms (max)
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
