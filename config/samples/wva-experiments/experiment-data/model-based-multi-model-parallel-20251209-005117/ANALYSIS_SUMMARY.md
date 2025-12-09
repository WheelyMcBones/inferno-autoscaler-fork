# WVA Model-Based Experiment Analysis

**Experiment:** model-based-multi-model-parallel-20251209-005117

**Generated:** 2025-12-09 09:45:54

---

## Summary Statistics

- **Total Predictions:** 33
- **Total Observations:** 34
- **Aligned Pairs:** 32
- **Scaling Decisions:** 33

## SLO Configuration

- **ITL SLO:** 10 ms
- **TTFT SLO:** 1000 ms

- **SLO Violations:** 7/32 (21.9%)

## Prediction Accuracy

### ITL Predictions

- Mean Error: 0.69 ms
- Abs Mean Error: 2.05 ms
- Std Dev: 4.66 ms

### TTFT Predictions

- Mean Error: 0.82 ms
- Abs Mean Error: 3.94 ms
- Std Dev: 7.95 ms

## Scaling Behavior

- no-change: 24 (72.7%)
- scale-up: 6 (18.2%)
- scale-down: 3 (9.1%)

- Replica Range: 1 - 3

## Files Generated

- `processed_data.csv` - Aligned prediction-observation pairs
- `plots/prediction_vs_observation.png` - ITL/TTFT over time
- `plots/scaling_behavior.png` - Replica scaling timeline
- `plots/prediction_error_distribution.png` - Error histograms
