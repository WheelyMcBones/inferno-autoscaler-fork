# WVA Model-Based Experiment Analysis

**Experiment:** model-based-moderate-parallel-extended-20251205-181954

**Generated:** 2025-12-05 19:09:47

---

## Summary Statistics

- **Total Predictions:** 23
- **Total Observations:** 23
- **Aligned Pairs:** 22
- **Scaling Decisions:** 23

## SLO Configuration

- **ITL SLO:** 10 ms
- **TTFT SLO:** 1000 ms

- **SLO Violations:** 9/22 (40.9%)

## Prediction Accuracy

### ITL Predictions

- Mean Error: 1.61 ms
- Abs Mean Error: 3.45 ms
- Std Dev: 4.76 ms

### TTFT Predictions

- Mean Error: 0.99 ms
- Abs Mean Error: 5.97 ms
- Std Dev: 8.31 ms

## Scaling Behavior

- no-change: 17 (73.9%)
- scale-down: 4 (17.4%)
- scale-up: 2 (8.7%)

- Replica Range: 1 - 4

## Files Generated

- `processed_data.csv` - Aligned prediction-observation pairs
- `plots/prediction_vs_observation.png` - ITL/TTFT over time
- `plots/scaling_behavior.png` - Replica scaling timeline
- `plots/prediction_error_distribution.png` - Error histograms
