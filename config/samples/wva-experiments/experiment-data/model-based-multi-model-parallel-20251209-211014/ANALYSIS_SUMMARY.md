# WVA Model-Based Experiment Analysis

**Experiment:** model-based-multi-model-parallel-20251209-211014

**Generated:** 2025-12-09 21:44:22

---

## Summary Statistics

- **Total Predictions:** 39
- **Total Observations:** 39
- **Aligned Pairs:** 38
- **Scaling Decisions:** 39

## SLO Configuration

- **ITL SLO:** 10 ms
- **TTFT SLO:** 1000 ms

- **SLO Violations:** 10/38 (26.3%)

## Prediction Accuracy

### ITL Predictions

- Mean Error: -0.19 ms
- Abs Mean Error: 4.97 ms
- Std Dev: 8.20 ms

### TTFT Predictions

- Mean Error: -0.20 ms
- Abs Mean Error: 8.03 ms
- Std Dev: 13.52 ms

## Scaling Behavior

- no-change: 22 (56.4%)
- scale-up: 13 (33.3%)
- scale-down: 4 (10.3%)

- Replica Range: 1 - 4

## Files Generated

- `processed_data.csv` - Aligned prediction-observation pairs
- `plots/prediction_vs_observation.png` - ITL/TTFT over time
- `plots/scaling_behavior.png` - Replica scaling timeline
- `plots/prediction_error_distribution.png` - Error histograms
