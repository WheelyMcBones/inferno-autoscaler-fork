# WVA Model-Based Experiment Analysis

**Experiment:** model-based-multi-model-parallel-20251209-232619

**Generated:** 2025-12-10 00:00:56

---

## Summary Statistics

- **Total Predictions:** 34
- **Total Observations:** 34
- **Aligned Pairs:** 33
- **Scaling Decisions:** 34

## SLO Configuration

- **ITL SLO:** 10 ms
- **TTFT SLO:** 1000 ms

- **SLO Violations:** 9/33 (27.3%)

## Prediction Accuracy

### ITL Predictions

- Mean Error: 2.71 ms
- Abs Mean Error: 4.01 ms
- Std Dev: 8.83 ms

### TTFT Predictions

- Mean Error: 3.86 ms
- Abs Mean Error: 6.67 ms
- Std Dev: 14.10 ms

## Scaling Behavior

- no-change: 23 (67.6%)
- scale-up: 8 (23.5%)
- scale-down: 3 (8.8%)

- Replica Range: 1 - 3

## Files Generated

- `processed_data.csv` - Aligned prediction-observation pairs
- `plots/prediction_vs_observation.png` - ITL/TTFT over time
- `plots/scaling_behavior.png` - Replica scaling timeline
- `plots/prediction_error_distribution.png` - Error histograms
