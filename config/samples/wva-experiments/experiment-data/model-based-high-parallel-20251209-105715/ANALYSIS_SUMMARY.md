# WVA Model-Based Experiment Analysis

**Experiment:** model-based-high-parallel-20251209-105715

**Generated:** 2025-12-09 11:19:01

---

## Summary Statistics

- **Total Predictions:** 25
- **Total Observations:** 25
- **Aligned Pairs:** 24
- **Scaling Decisions:** 25

## SLO Configuration

- **ITL SLO:** 10 ms
- **TTFT SLO:** 1000 ms

- **SLO Violations:** 12/24 (50.0%)

## Prediction Accuracy

### ITL Predictions

- Mean Error: 6.91 ms
- Abs Mean Error: 11.46 ms
- Std Dev: 18.76 ms

### TTFT Predictions

- Mean Error: 87.69 ms
- Abs Mean Error: 95.69 ms
- Std Dev: 296.03 ms

## Scaling Behavior

- no-change: 10 (40.0%)
- scale-up: 10 (40.0%)
- scale-down: 5 (20.0%)

- Replica Range: 1 - 10

## Files Generated

- `processed_data.csv` - Aligned prediction-observation pairs
- `plots/prediction_vs_observation.png` - ITL/TTFT over time
- `plots/scaling_behavior.png` - Replica scaling timeline
- `plots/prediction_error_distribution.png` - Error histograms
