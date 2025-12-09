# WVA Model-Based Experiment Analysis

**Experiment:** model-based-moderate-parallel-extended-20251209-082407

**Generated:** 2025-12-09 08:54:33

---

## Summary Statistics

- **Total Predictions:** 22
- **Total Observations:** 22
- **Aligned Pairs:** 21
- **Scaling Decisions:** 22

## SLO Configuration

- **ITL SLO:** 10 ms
- **TTFT SLO:** 1000 ms

- **SLO Violations:** 7/21 (33.3%)

## Prediction Accuracy

### ITL Predictions

- Mean Error: 0.17 ms
- Abs Mean Error: 2.12 ms
- Std Dev: 3.33 ms

### TTFT Predictions

- Mean Error: -4.42 ms
- Abs Mean Error: 5.47 ms
- Std Dev: 6.14 ms

## Scaling Behavior

- no-change: 14 (63.6%)
- scale-up: 6 (27.3%)
- scale-down: 2 (9.1%)

- Replica Range: 1 - 3

## Files Generated

- `processed_data.csv` - Aligned prediction-observation pairs
- `plots/prediction_vs_observation.png` - ITL/TTFT over time
- `plots/scaling_behavior.png` - Replica scaling timeline
- `plots/prediction_error_distribution.png` - Error histograms
