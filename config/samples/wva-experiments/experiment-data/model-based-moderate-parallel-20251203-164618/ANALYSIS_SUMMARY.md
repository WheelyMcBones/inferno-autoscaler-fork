# WVA Model-Based Experiment Analysis

**Experiment:** model-based-moderate-parallel-20251203-164618

**Generated:** 2025-12-03 17:18:41

---

## Summary Statistics

- **Total Predictions:** 16
- **Total Observations:** 17
- **Aligned Pairs:** 15
- **Scaling Decisions:** 16

## SLO Configuration

- **ITL SLO:** 10 ms
- **TTFT SLO:** 1000 ms

- **SLO Violations:** 4/15 (26.7%)

## Prediction Accuracy

### ITL Predictions

- Mean Error: -2.17 ms
- Abs Mean Error: 5.52 ms
- Std Dev: 7.10 ms

### TTFT Predictions

- Mean Error: -5.71 ms
- Abs Mean Error: 11.05 ms
- Std Dev: 13.28 ms

## Scaling Behavior

- no-change: 8 (50.0%)
- scale-up: 4 (25.0%)
- scale-down: 4 (25.0%)

- Replica Range: 1 - 8

## Files Generated

- `processed_data.csv` - Aligned prediction-observation pairs
- `plots/prediction_vs_observation.png` - ITL/TTFT over time
- `plots/scaling_behavior.png` - Replica scaling timeline
- `plots/prediction_error_distribution.png` - Error histograms
