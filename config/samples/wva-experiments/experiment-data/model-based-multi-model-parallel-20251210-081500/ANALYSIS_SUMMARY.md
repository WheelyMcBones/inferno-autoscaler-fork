# WVA Model-Based Experiment Analysis

**Experiment:** model-based-multi-model-parallel-20251210-081500

**Generated:** 2025-12-10 10:08:35

---

## Summary Statistics

- **Total Predictions:** 38
- **Total Observations:** 38
- **Aligned Pairs:** 37
- **Scaling Decisions:** 38

## SLO Configuration

- **ITL SLO:** 10 ms
- **TTFT SLO:** 1000 ms

- **SLO Violations:** 12/37 (32.4%)

## Prediction Accuracy

### ITL Predictions

- Mean Error: 1.68 ms
- Abs Mean Error: 4.24 ms
- Std Dev: 8.61 ms

### TTFT Predictions

- Mean Error: 1.81 ms
- Abs Mean Error: 7.67 ms
- Std Dev: 14.13 ms

## Scaling Behavior

- no-change: 27 (71.1%)
- scale-up: 9 (23.7%)
- scale-down: 2 (5.3%)

- Replica Range: 1 - 2

## Files Generated

- `processed_data.csv` - Aligned prediction-observation pairs
- `plots/prediction_vs_observation.png` - ITL/TTFT over time
- `plots/scaling_behavior.png` - Replica scaling timeline
- `plots/prediction_error_distribution.png` - Error histograms
