# WVA Model-Based Experiment Analysis

**Experiment:** model-based-smooth-gradual-20251206-144853

**Generated:** 2025-12-06 15:16:41

---

## Summary Statistics

- **Total Predictions:** 27
- **Total Observations:** 27
- **Aligned Pairs:** 26
- **Scaling Decisions:** 27

## SLO Configuration

- **ITL SLO:** 10 ms
- **TTFT SLO:** 1000 ms

- **SLO Violations:** 9/26 (34.6%)

## Prediction Accuracy

### ITL Predictions

- Mean Error: -0.22 ms
- Abs Mean Error: 1.35 ms
- Std Dev: 2.28 ms

### TTFT Predictions

- Mean Error: -0.65 ms
- Abs Mean Error: 2.61 ms
- Std Dev: 4.92 ms

## Scaling Behavior

- no-change: 17 (63.0%)
- scale-up: 5 (18.5%)
- scale-down: 5 (18.5%)

- Replica Range: 1 - 5

## Files Generated

- `processed_data.csv` - Aligned prediction-observation pairs
- `plots/prediction_vs_observation.png` - ITL/TTFT over time
- `plots/scaling_behavior.png` - Replica scaling timeline
- `plots/prediction_error_distribution.png` - Error histograms
