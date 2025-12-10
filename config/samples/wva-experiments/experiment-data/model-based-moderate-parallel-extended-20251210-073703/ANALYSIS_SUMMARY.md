# WVA Model-Based Experiment Analysis

**Experiment:** model-based-moderate-parallel-extended-20251210-073703

**Generated:** 2025-12-10 11:12:31

---

## Summary Statistics

- **Total Predictions:** 28
- **Total Observations:** 28
- **Aligned Pairs:** 27
- **Scaling Decisions:** 28

## SLO Configuration

- **ITL SLO:** 10 ms
- **TTFT SLO:** 1000 ms

- **SLO Violations:** 8/27 (29.6%)

## Prediction Accuracy

### ITL Predictions

- Mean Error: -0.64 ms
- Abs Mean Error: 3.61 ms
- Std Dev: 5.34 ms

### TTFT Predictions

- Mean Error: -2.52 ms
- Abs Mean Error: 6.89 ms
- Std Dev: 10.06 ms

## Scaling Behavior

- no-change: 20 (71.4%)
- scale-up: 6 (21.4%)
- scale-down: 2 (7.1%)

- Replica Range: 1 - 3

## Files Generated

- `processed_data.csv` - Aligned prediction-observation pairs
- `plots/prediction_vs_observation.png` - ITL/TTFT over time
- `plots/scaling_behavior.png` - Replica scaling timeline
- `plots/prediction_error_distribution.png` - Error histograms
