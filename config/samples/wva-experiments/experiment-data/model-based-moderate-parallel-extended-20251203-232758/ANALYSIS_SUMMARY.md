# WVA Model-Based Experiment Analysis

**Experiment:** model-based-moderate-parallel-extended-20251203-232758

**Generated:** 2025-12-03 23:58:24

---

## Summary Statistics

- **Total Predictions:** 28
- **Total Observations:** 28
- **Aligned Pairs:** 27
- **Scaling Decisions:** 28

## SLO Configuration

- **ITL SLO:** 10 ms
- **TTFT SLO:** 1000 ms

- **SLO Violations:** 4/27 (14.8%)

## Prediction Accuracy

### ITL Predictions

- Mean Error: -1.92 ms
- Abs Mean Error: 2.46 ms
- Std Dev: 3.71 ms

### TTFT Predictions

- Mean Error: -6.87 ms
- Abs Mean Error: 7.13 ms
- Std Dev: 6.06 ms

## Scaling Behavior

- no-change: 20 (71.4%)
- scale-up: 4 (14.3%)
- scale-down: 4 (14.3%)

- Replica Range: 1 - 5

## Files Generated

- `processed_data.csv` - Aligned prediction-observation pairs
- `plots/prediction_vs_observation.png` - ITL/TTFT over time
- `plots/scaling_behavior.png` - Replica scaling timeline
- `plots/prediction_error_distribution.png` - Error histograms
