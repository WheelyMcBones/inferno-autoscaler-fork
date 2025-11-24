# New Experiment System

Configuration-driven HPA experiments with TTFT/ITL metrics collection.

## Quick Start

```bash
./start-experiments.sh
```

## Components

- **start-experiments.sh**: Interactive menu to run experiments and view results
- **run-experiment.sh**: Executes experiments from YAML configs (supports parallel jobs)
- **monitor-hpa-enhanced.sh**: Collects HPA metrics + TTFT/ITL from Prometheus
- **experiment-configs/**: YAML configuration files defining experiment setups
- **analyze-hpa-experiment.ipynb**: Jupyter notebook for data analysis

## Configuration Format

### Sequential Mode (default)

Jobs run one at a time with 30s gap between phases:

```yaml
jobs:
  - name: "phase-1-low"
    manifest: "../../workloads/sharegpt-load-job-1.yaml"
    duration: 360
  - name: "phase-2-high"
    manifest: "../../workloads/sharegpt-load-job-3.yaml"
    duration: 360
```

### Parallel Mode (with overlapping jobs)

Add `start_delay` to run jobs in parallel with staggered start times:

```yaml
jobs:
  - name: "job-1"
    manifest: "../../workloads/sharegpt-load-job-high-20.yaml"  # 20 req/s
    duration: 360
    start_delay: 0  # Starts immediately
    
  - name: "job-2"
    manifest: "../../workloads/sharegpt-load-job-high-30.yaml"  # 30 req/s
    duration: 360
    start_delay: 120  # Starts 2min after experiment begins
    
  - name: "job-3"
    manifest: "../../workloads/sharegpt-load-job-high-20.yaml"  # 20 req/s
    duration: 360
    start_delay: 240  # Starts 4min after experiment begins
```

**Timeline for above config:**

```bash
Time:    0s        120s       240s       360s       480s       600s
Job1:    [========================================]  (20 req/s)
Job2:              [========================================]  (30 req/s)
Job3:                         [========================================]  (20 req/s)
Total:   20        50         50         30         20         0
```

This creates sustained 50 req/s periods (120-480s) to force HPA scale-up.

### Scale-Down Observation (Cooldown)

To capture HPA scale-down behavior after jobs complete, add a cooldown period:

```yaml
cooldown:
  enabled: true
  duration: 300  # Monitor for 5 minutes after jobs complete
```

**What happens:**

1. Jobs complete (or reach their duration)
2. Jobs are deleted immediately (triggers scale-down)
3. Monitoring continues for `duration` seconds
4. Captures full scale-down: 10 replicas → 1 replica
5. Metrics CSV includes the complete scaling cycle

**Without cooldown:** Experiment ends immediately after jobs finish, missing scale-down data.

## Metrics Collected

**Performance Metrics (identical to WVA collector):**

- **TTFT**: Time to First Token (ms) - `sum(rate(metric_sum[1m]))/sum(rate(metric_count[1m])) * 1000`
- **ITL**: Inter-Token Latency (ms) - same pattern as TTFT
- Uses rate() with [1m] window for current average, not cumulative totals

**HPA Metrics:**

- Replicas: desired, current, ready
- Custom metrics: kv_cache_usage_perc, num_requests_waiting
- Request rate (req/min)

**Output:** CSV files in `../experiment-data/`

## Available Experiment Configs

- **baseline-hpa.yaml**: Sequential 7→7→14 req/s (moderate load, no overlap) + 5min cooldown
- **moderate-load-hpa.yaml**: Parallel 10+15+12 req/s with overlaps (gradual scaling: 10→25→27 req/s) + 5min cooldown
- **high-load-hpa.yaml**: Parallel 20+30+20 req/s with overlaps (aggressive stress test: 20→50→50 req/s) + 5min cooldown

### Experiment Comparison

| Config | Mode | Jobs | Peak Load | Use Case |
|--------|------|------|-----------|----------|
| baseline | Sequential | 7→7→14 req/s | 14 req/s | Basic HPA testing, no overlap |
| moderate-load | Parallel | 10→15→12 req/s | 27 req/s | Gradual scale-up, realistic load |
| high-load | Parallel | 20→30→20 req/s | 50 req/s | Stress test, aggressive scaling |

## Creating New Experiments

1. Copy one of the configs: `baseline-hpa.yaml`, `moderate-load-hpa.yaml`, or `high-load-hpa.yaml`
2. Modify jobs, durations, start_delay values
3. Add/remove cooldown period as needed
4. Run via `./start-experiments.sh`

**Tips:**

- Use `start_delay` when you need cumulative load (multiple jobs running simultaneously)
- Omit `start_delay` (or set to 0) for sequential execution
- Parallel mode auto-detected when any job has `start_delay > 0`
- Enable `cooldown` to capture complete HPA scaling cycle (up AND down)
- **moderate-load** is recommended for observing gradual HPA behavior
- **high-load** is for stress testing maximum scaling capacity
