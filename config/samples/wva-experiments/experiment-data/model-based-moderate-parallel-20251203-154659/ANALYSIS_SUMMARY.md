# WVA Model-Based Experiment Analysis

**Experiment:** model-based-moderate-parallel-20251203-154659

**Generated:** 2025-12-03 16:01:58

---

## Summary Statistics

- **Total Predictions:** 15
- **Total Observations:** 17
- **Aligned Pairs:** 14
- **Scaling Decisions:** 15

## SLO Configuration

- **ITL SLO:** 10 ms
- **TTFT SLO:** 1000 ms

- **SLO Violations:** 5/14 (35.7%)

## Prediction Accuracy

### ITL Predictions

- Mean Error: -1.69 ms
- Abs Mean Error: 6.46 ms
- Std Dev: 8.12 ms

### TTFT Predictions

- Mean Error: -4.82 ms
- Abs Mean Error: 12.45 ms
- Std Dev: 14.57 ms

## Scaling Behavior

- no-change: 8 (53.3%)
- scale-up: 4 (26.7%)
- scale-down: 3 (20.0%)

- Replica Range: 1 - 10

## Files Generated

- `processed_data.csv` - Aligned prediction-observation pairs
- `plots/prediction_vs_observation.png` - ITL/TTFT over time
- `plots/scaling_behavior.png` - Replica scaling timeline
- `plots/prediction_error_distribution.png` - Error histograms
