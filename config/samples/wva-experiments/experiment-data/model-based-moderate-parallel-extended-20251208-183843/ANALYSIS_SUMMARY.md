# WVA Model-Based Experiment Analysis

**Experiment:** model-based-moderate-parallel-extended-20251208-183843

**Generated:** 2025-12-08 19:02:18

---

## Summary Statistics

- **Total Predictions:** 21
- **Total Observations:** 22
- **Aligned Pairs:** 20
- **Scaling Decisions:** 21

## SLO Configuration

- **ITL SLO:** 10 ms
- **TTFT SLO:** 1000 ms

- **SLO Violations:** 5/20 (25.0%)

## Prediction Accuracy

### ITL Predictions

- Mean Error: -0.07 ms
- Abs Mean Error: 1.28 ms
- Std Dev: 2.49 ms

### TTFT Predictions

- Mean Error: -1.11 ms
- Abs Mean Error: 2.64 ms
- Std Dev: 4.96 ms

## Scaling Behavior

- scale-down: 9 (42.9%)
- no-change: 7 (33.3%)
- scale-up: 5 (23.8%)

- Replica Range: 1 - 9

## Files Generated

- `processed_data.csv` - Aligned prediction-observation pairs
- `plots/prediction_vs_observation.png` - ITL/TTFT over time
- `plots/scaling_behavior.png` - Replica scaling timeline
- `plots/prediction_error_distribution.png` - Error histograms
