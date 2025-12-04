# WVA Model-Based Experiment Analysis

**Experiment:** model-based-moderate-parallel-20251203-162542

**Generated:** 2025-12-03 16:40:08

---

## Summary Statistics

- **Total Predictions:** 13
- **Total Observations:** 16
- **Aligned Pairs:** 12
- **Scaling Decisions:** 13

## SLO Configuration

- **ITL SLO:** 10 ms
- **TTFT SLO:** 1000 ms

- **SLO Violations:** 4/12 (33.3%)

## Prediction Accuracy

### ITL Predictions

- Mean Error: -3.03 ms
- Abs Mean Error: 4.06 ms
- Std Dev: 4.78 ms

### TTFT Predictions

- Mean Error: -6.31 ms
- Abs Mean Error: 8.42 ms
- Std Dev: 9.98 ms

## Scaling Behavior

- scale-down: 5 (38.5%)
- no-change: 4 (30.8%)
- scale-up: 4 (30.8%)

- Replica Range: 1 - 10

## Files Generated

- `processed_data.csv` - Aligned prediction-observation pairs
- `plots/prediction_vs_observation.png` - ITL/TTFT over time
- `plots/scaling_behavior.png` - Replica scaling timeline
- `plots/prediction_error_distribution.png` - Error histograms
