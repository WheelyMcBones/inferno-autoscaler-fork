# WVA Model-Based Experiment Analysis

**Experiment:** model-based-moderate-parallel-extended-20251209-095600

**Generated:** 2025-12-09 10:18:05

---

## Summary Statistics

- **Total Predictions:** 26
- **Total Observations:** 26
- **Aligned Pairs:** 25
- **Scaling Decisions:** 26

## SLO Configuration

- **ITL SLO:** 10 ms
- **TTFT SLO:** 1000 ms

- **SLO Violations:** 8/25 (32.0%)

## Prediction Accuracy

### ITL Predictions

- Mean Error: 0.06 ms
- Abs Mean Error: 3.21 ms
- Std Dev: 4.82 ms

### TTFT Predictions

- Mean Error: -2.60 ms
- Abs Mean Error: 6.47 ms
- Std Dev: 8.66 ms

## Scaling Behavior

- no-change: 18 (69.2%)
- scale-up: 6 (23.1%)
- scale-down: 2 (7.7%)

- Replica Range: 1 - 3

## Files Generated

- `processed_data.csv` - Aligned prediction-observation pairs
- `plots/prediction_vs_observation.png` - ITL/TTFT over time
- `plots/scaling_behavior.png` - Replica scaling timeline
- `plots/prediction_error_distribution.png` - Error histograms
