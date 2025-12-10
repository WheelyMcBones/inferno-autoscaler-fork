# WVA Model-Based Experiment Analysis

**Experiment:** model-based-moderate-parallel-extended-20251210-071015

**Generated:** 2025-12-10 11:15:11

---

## Summary Statistics

- **Total Predictions:** 27
- **Total Observations:** 27
- **Aligned Pairs:** 26
- **Scaling Decisions:** 27

## SLO Configuration

- **ITL SLO:** 10 ms
- **TTFT SLO:** 1000 ms

- **SLO Violations:** 10/26 (38.5%)

## Prediction Accuracy

### ITL Predictions

- Mean Error: 4.99 ms
- Abs Mean Error: 8.21 ms
- Std Dev: 14.05 ms

### TTFT Predictions

- Mean Error: 126.77 ms
- Abs Mean Error: 133.23 ms
- Std Dev: 622.98 ms

## Scaling Behavior

- no-change: 22 (81.5%)
- scale-up: 4 (14.8%)
- scale-down: 1 (3.7%)

- Replica Range: 1 - 2

## Files Generated

- `processed_data.csv` - Aligned prediction-observation pairs
- `plots/prediction_vs_observation.png` - ITL/TTFT over time
- `plots/scaling_behavior.png` - Replica scaling timeline
- `plots/prediction_error_distribution.png` - Error histograms
