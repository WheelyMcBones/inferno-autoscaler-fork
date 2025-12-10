# WVA Model-Based Experiment Analysis

**Experiment:** model-based-multi-model-parallel-20251209-215450

**Generated:** 2025-12-09 22:31:12

---

## Summary Statistics

- **Total Predictions:** 34
- **Total Observations:** 34
- **Aligned Pairs:** 33
- **Scaling Decisions:** 34

## SLO Configuration

- **ITL SLO:** 10 ms
- **TTFT SLO:** 1000 ms

- **SLO Violations:** 14/33 (42.4%)

## Prediction Accuracy

### ITL Predictions

- Mean Error: 3.26 ms
- Abs Mean Error: 4.63 ms
- Std Dev: 10.02 ms

### TTFT Predictions

- Mean Error: 4.74 ms
- Abs Mean Error: 7.48 ms
- Std Dev: 15.36 ms

## Scaling Behavior

- no-change: 19 (55.9%)
- scale-up: 11 (32.4%)
- scale-down: 4 (11.8%)

- Replica Range: 1 - 3

## Files Generated

- `processed_data.csv` - Aligned prediction-observation pairs
- `plots/prediction_vs_observation.png` - ITL/TTFT over time
- `plots/scaling_behavior.png` - Replica scaling timeline
- `plots/prediction_error_distribution.png` - Error histograms
