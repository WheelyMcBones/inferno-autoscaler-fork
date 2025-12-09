# WVA Model-Based Experiment Analysis

**Experiment:** model-based-multi-model-parallel-20251208-220210

**Generated:** 2025-12-09 09:54:55

---

## Summary Statistics

- **Total Predictions:** 36
- **Total Observations:** 37
- **Aligned Pairs:** 35
- **Scaling Decisions:** 36

## SLO Configuration

- **ITL SLO:** 10 ms
- **TTFT SLO:** 1000 ms

- **SLO Violations:** 4/35 (11.4%)

## Prediction Accuracy

### ITL Predictions

- Mean Error: -1.23 ms
- Abs Mean Error: 1.62 ms
- Std Dev: 3.32 ms

### TTFT Predictions

- Mean Error: -3.73 ms
- Abs Mean Error: 3.89 ms
- Std Dev: 6.45 ms

## Scaling Behavior

- no-change: 27 (75.0%)
- scale-down: 5 (13.9%)
- scale-up: 4 (11.1%)

- Replica Range: 1 - 10

## Files Generated

- `processed_data.csv` - Aligned prediction-observation pairs
- `plots/prediction_vs_observation.png` - ITL/TTFT over time
- `plots/scaling_behavior.png` - Replica scaling timeline
- `plots/prediction_error_distribution.png` - Error histograms
