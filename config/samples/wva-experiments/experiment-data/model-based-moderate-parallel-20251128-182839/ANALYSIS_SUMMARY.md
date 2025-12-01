# WVA Model-Based Experiment Analysis

**Experiment:** model-based-moderate-parallel-20251128-182839

**Generated:** 2025-11-28 19:08:17

---

## Summary Statistics

- **Total Predictions:** 58
- **Total Observations:** 58
- **Aligned Pairs:** 57
- **Scaling Decisions:** 58

## SLO Configuration

- **ITL SLO:** 10 ms
- **TTFT SLO:** 1000 ms

- **SLO Violations:** 38/57 (66.7%)

## Prediction Accuracy

### ITL Predictions

- Mean Error: 6.26 ms
- Abs Mean Error: 11.28 ms
- Std Dev: 17.53 ms

### TTFT Predictions

- Mean Error: 163.24 ms
- Abs Mean Error: 173.55 ms
- Std Dev: 833.05 ms

## Scaling Behavior

- no-change: 44 (75.9%)
- scale-down: 8 (13.8%)
- scale-up: 6 (10.3%)

- Replica Range: 1 - 3

## Files Generated

- `processed_data.csv` - Aligned prediction-observation pairs
- `plots/prediction_vs_observation.png` - ITL/TTFT over time
- `plots/scaling_behavior.png` - Replica scaling timeline
- `plots/prediction_error_distribution.png` - Error histograms
