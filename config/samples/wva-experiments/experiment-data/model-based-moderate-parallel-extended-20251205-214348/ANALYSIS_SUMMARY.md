# WVA Model-Based Experiment Analysis

**Experiment:** model-based-moderate-parallel-extended-20251205-214348

**Generated:** 2025-12-05 22:43:17

---

## Summary Statistics

- **Total Predictions:** 28
- **Total Observations:** 27
- **Aligned Pairs:** 26
- **Scaling Decisions:** 28

## SLO Configuration

- **ITL SLO:** 10 ms
- **TTFT SLO:** 1000 ms

- **SLO Violations:** 8/26 (30.8%)

## Prediction Accuracy

### ITL Predictions

- Mean Error: -0.22 ms
- Abs Mean Error: 3.67 ms
- Std Dev: 5.03 ms

### TTFT Predictions

- Mean Error: -2.18 ms
- Abs Mean Error: 7.16 ms
- Std Dev: 9.70 ms

## Scaling Behavior

- no-change: 20 (71.4%)
- scale-up: 4 (14.3%)
- scale-down: 4 (14.3%)

- Replica Range: 1 - 6

## Files Generated

- `processed_data.csv` - Aligned prediction-observation pairs
- `plots/prediction_vs_observation.png` - ITL/TTFT over time
- `plots/scaling_behavior.png` - Replica scaling timeline
- `plots/prediction_error_distribution.png` - Error histograms
