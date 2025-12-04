# WVA Model-Based Experiment Analysis

**Experiment:** model-based-moderate-parallel-20251203-192044

**Generated:** 2025-12-03 22:33:50

---

## Summary Statistics

- **Total Predictions:** 11
- **Total Observations:** 12
- **Aligned Pairs:** 10
- **Scaling Decisions:** 11

## SLO Configuration

- **ITL SLO:** 10 ms
- **TTFT SLO:** 1000 ms

- **SLO Violations:** 4/10 (40.0%)

## Prediction Accuracy

### ITL Predictions

- Mean Error: -0.64 ms
- Abs Mean Error: 2.97 ms
- Std Dev: 4.57 ms

### TTFT Predictions

- Mean Error: -4.08 ms
- Abs Mean Error: 7.32 ms
- Std Dev: 9.35 ms

## Scaling Behavior

- scale-down: 5 (45.5%)
- scale-up: 4 (36.4%)
- no-change: 2 (18.2%)

- Replica Range: 1 - 10

## Files Generated

- `processed_data.csv` - Aligned prediction-observation pairs
- `plots/prediction_vs_observation.png` - ITL/TTFT over time
- `plots/scaling_behavior.png` - Replica scaling timeline
- `plots/prediction_error_distribution.png` - Error histograms
