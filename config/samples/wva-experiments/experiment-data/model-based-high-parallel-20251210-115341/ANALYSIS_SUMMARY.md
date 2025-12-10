# WVA Model-Based Experiment Analysis

**Experiment:** model-based-high-parallel-20251210-115341

**Generated:** 2025-12-10 12:49:45

---

## Summary Statistics

- **Total Predictions:** 25
- **Total Observations:** 25
- **Aligned Pairs:** 24
- **Scaling Decisions:** 25

## SLO Configuration

- **ITL SLO:** 10 ms
- **TTFT SLO:** 1000 ms

- **SLO Violations:** 14/24 (58.3%)

## Prediction Accuracy

### ITL Predictions

- Mean Error: 13.57 ms
- Abs Mean Error: 17.09 ms
- Std Dev: 23.19 ms

### TTFT Predictions

- Mean Error: 171.92 ms
- Abs Mean Error: 176.30 ms
- Std Dev: 495.10 ms

## Scaling Behavior

- scale-up: 16 (64.0%)
- no-change: 6 (24.0%)
- scale-down: 3 (12.0%)

- Replica Range: 1 - 10

## Files Generated

- `processed_data.csv` - Aligned prediction-observation pairs
- `plots/prediction_vs_observation.png` - ITL/TTFT over time
- `plots/scaling_behavior.png` - Replica scaling timeline
- `plots/prediction_error_distribution.png` - Error histograms
