# Quick Start Guide - WVA Analysis Notebook

## Prerequisites

```bash
# Install required Python packages
pip install matplotlib numpy pandas jupyter

# Ensure bash script is executable
chmod +x extract_metrics.sh
```

## Running Your First Experiment

### Step 1: Prepare Your Log File
Place your WVA controller log file in the `thesis/` directory:
```bash
cp /path/to/your/wva-logs.txt thesis/ttft_contextaware_scorers_logs.txt
```

### Step 2: Open the Notebook
```bash
cd thesis/
jupyter notebook wva_analysis.ipynb
```

### Step 3: Configure and Run
1. In Cell 1, update `LOG_FILE` to your log filename
2. Click "Cell" → "Run All" (or press Shift+Enter repeatedly)
3. Wait for processing to complete (~1-2 minutes)

### Step 4: View Results
Check the newly created directory:
```bash
ls -la experiments/$(ls -t experiments/ | head -1)/
```

## Outputs You'll Get

### 6 Publication-Quality Plots
- ITL performance with warmup gaps
- TTFT performance with warmup gaps  
- Load pattern evolution
- ITL vs replica scaling timeline
- TTFT vs replica scaling timeline
- Combined ITL & TTFT analysis

### 3 Data Files
- Extracted metrics CSV
- Processed metrics CSV
- Original log backup

### 3 Analysis Files
- Warmup gap statistics
- Scaling event log
- Performance summary

## Customizing Your Analysis

### Change SLO Thresholds
In Cell 4 (detect_scaling_events function):
```python
# Modify the SLO threshold
scaling_events, warmup_gaps = detect_scaling_events(df, slo_threshold=10.0)  # Custom ITL SLO
```

### Adjust Plot Styling
In Cell 1:
```python
# Change plot style
plt.style.use('ggplot')  # or 'bmh', 'classic', 'dark_background'
```

### Filter Data by Time Range
In Cell 3 (after loading data):
```python
# Analyze only first 10 minutes
df = df[df['time_minutes'] <= 10]
```

## Troubleshooting

### Error: "No such file or directory"
- Check that `LOG_FILE` path is correct
- Ensure `extract_metrics.sh` is in the same directory
- Verify log file exists and has content

### Error: "Empty DataFrame"
- Check that your log file contains optimization data
- Look for lines with "System data prepared for optimization"
- Verify bash script is extracting data correctly:
  ```bash
  ./extract_metrics.sh your_log.txt | head
  ```

### Plots Not Showing
- Ensure `%matplotlib inline` is in Cell 1
- Try restarting kernel: Kernel → Restart & Run All
- Check that matplotlib is installed: `pip install matplotlib`

### Missing Columns in CSV
- Your logs might be in a different format
- Check the bash script output manually
- Update the AWK patterns in `extract_metrics.sh` if needed

## Advanced Usage

### Batch Processing Multiple Logs
```python
import glob

log_files = glob.glob('logs/*.txt')
for log_file in log_files:
    LOG_FILE = log_file
    # Run analysis...
```

### Export to LaTeX/Papers
```python
# Save plots with specific DPI for papers
plt.savefig(plot_path, dpi=600, format='pdf')  # High-res PDF
```

### Compare Multiple Experiments
```python
import pandas as pd

exp1 = pd.read_csv('experiments/20251020_143022/data/processed_metrics.csv')
exp2 = pd.read_csv('experiments/20251020_150315/data/processed_metrics.csv')

# Plot comparison
plt.plot(exp1['time_minutes'], exp1['itlAverage'], label='Experiment 1')
plt.plot(exp2['time_minutes'], exp2['itlAverage'], label='Experiment 2')
plt.legend()
```

## Best Practices

1. **Always run full notebook** - Don't skip cells, they depend on each other
2. **One log file per run** - Creates clean, separate experiments
3. **Descriptive log filenames** - e.g., `baseline_1000rpm_3nodes.txt`
4. **Review manifest.txt** - Confirms experiment configuration
5. **Keep experiments directory** - Historical data for comparisons

## Directory Naming

Experiments are automatically named with timestamps:
```
experiments/20251020_143022/  ← Oct 20, 2025 at 14:30:22
```

You can create symbolic links for easier reference:
```bash
ln -s experiments/20251020_143022 experiments/baseline_test
ln -s experiments/20251020_150315 experiments/peak_load_test
```

## Getting Help

- Check the notebook markdown cells for explanations
- Review the `experiments/README.md` for output descriptions
- Examine example outputs in existing experiment directories
- Read the bash script comments in `extract_metrics.sh`
