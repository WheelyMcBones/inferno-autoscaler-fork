# WVA Model-Based Experiment Analysis

**Experiment:** model-based-high-parallel-20251209-224705

**Generated:** 2025-12-10 13:02:37

---

## Summary Statistics

- **Total Predictions:** 22
- **Total Observations:** 22
- **Aligned Pairs:** 21
- **Scaling Decisions:** 22

## SLO Configuration

- **ITL SLO:** 10 ms
- **TTFT SLO:** 1000 ms

- **SLO Violations:** 12/21 (57.1%)

## Prediction Accuracy

### ITL Predictions

- Mean Error: 11.29 ms
- Abs Mean Error: 12.60 ms
- Std Dev: 19.64 ms

### TTFT Predictions

- Mean Error: 335.89 ms
- Abs Mean Error: 338.08 ms
- Std Dev: 1247.83 ms

## Scaling Behavior

- scale-up: 12 (54.5%)
- no-change: 6 (27.3%)
- scale-down: 4 (18.2%)

- Replica Range: 1 - 6

## Files Generated

- `processed_data.csv` - Aligned prediction-observation pairs
- `plots/prediction_vs_observation.png` - ITL/TTFT over time
- `plots/scaling_behavior.png` - Replica scaling timeline
- `plots/prediction_error_distribution.png` - Error histograms
