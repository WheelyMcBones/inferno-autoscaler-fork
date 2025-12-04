# WVA Model-Based Experiment Analysis

**Experiment:** model-based-moderate-parallel-20251203-144456

**Generated:** 2025-12-03 15:04:17

---

## Summary Statistics

- **Total Predictions:** 17
- **Total Observations:** 17
- **Aligned Pairs:** 16
- **Scaling Decisions:** 17

## SLO Configuration

- **ITL SLO:** 10 ms
- **TTFT SLO:** 1000 ms

- **SLO Violations:** 6/16 (37.5%)

## Prediction Accuracy

### ITL Predictions

- Mean Error: -0.28 ms
- Abs Mean Error: 5.83 ms
- Std Dev: 8.34 ms

### TTFT Predictions

- Mean Error: -0.64 ms
- Abs Mean Error: 11.13 ms
- Std Dev: 14.76 ms

## Scaling Behavior

- no-change: 9 (52.9%)
- scale-up: 4 (23.5%)
- scale-down: 4 (23.5%)

- Replica Range: 1 - 7

## Files Generated

- `processed_data.csv` - Aligned prediction-observation pairs
- `plots/prediction_vs_observation.png` - ITL/TTFT over time
- `plots/scaling_behavior.png` - Replica scaling timeline
- `plots/prediction_error_distribution.png` - Error histograms
