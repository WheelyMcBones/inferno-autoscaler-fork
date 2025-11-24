# Legacy Scripts

Original manual HPA experiment runner. **Recommendation: Use `../new-experiments/` instead.**

## Scripts

- **run-hpa-experiment.sh**: Manual experiment runner with hardcoded parameters
- **monitor-hpa-experiment.sh**: Basic monitoring (no TTFT/ITL collection)
- **analyze-hpa-experiment.py**: CLI analysis tool
- **cleanup-experiments.sh**: Clean up experiment jobs and data
- **view-experiment.sh**: View experiment results

## Why Legacy?

These scripts require manual editing to change experiment parameters and don't collect TTFT/ITL metrics. The new system offers:

- YAML-based configuration
- TTFT/ITL metrics matching WVA's collector
- Interactive menus
- Better analysis tools
