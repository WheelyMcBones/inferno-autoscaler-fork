#!/usr/bin/env python3
"""
HPA Experiment Data Analyzer
Analyzes and plots HPA scaling experiment data
"""

import argparse
import json
import sys
from pathlib import Path
from datetime import datetime

import pandas as pd
import matplotlib.pyplot as plt
import matplotlib.dates as mdates
from matplotlib.patches import Rectangle


def load_experiment_data(experiment_dir):
    """Load all experiment data files"""
    exp_path = Path(experiment_dir)
    
    if not exp_path.exists():
        print(f"Error: Experiment directory not found: {experiment_dir}")
        sys.exit(1)
    
    # Load metadata
    metadata_file = exp_path / "metadata.json"
    if metadata_file.exists():
        with open(metadata_file) as f:
            metadata = json.load(f)
    else:
        metadata = {}
    
    # Load metrics CSV
    metrics_file = exp_path / "metrics.csv"
    if not metrics_file.exists():
        print(f"Error: Metrics file not found: {metrics_file}")
        sys.exit(1)
    
    df = pd.read_csv(metrics_file)
    df['timestamp'] = pd.to_datetime(df['timestamp'])
    
    # Load scaling events
    scaling_log = exp_path / "scaling.log"
    scaling_events = []
    if scaling_log.exists():
        with open(scaling_log) as f:
            content = f.read()
            # Parse scaling events (simplified)
            for block in content.split("========================================"):
                if "SCALING EVENT" in block:
                    lines = block.strip().split('\n')
                    event = {
                        'text': block.strip(),
                        'lines': lines
                    }
                    scaling_events.append(event)
    
    return metadata, df, scaling_events


def plot_experiment(df, metadata, output_file=None):
    """Create comprehensive plot of experiment data"""
    
    fig, axes = plt.subplots(4, 1, figsize=(14, 10), sharex=True)
    fig.suptitle(f"HPA Scaling Experiment: {metadata.get('experiment_name', 'Unknown')}", 
                 fontsize=14, fontweight='bold')
    
    # Convert timestamps to matplotlib dates
    timestamps = df['timestamp']
    
    # Plot 1: Replica count
    ax1 = axes[0]
    ax1.plot(timestamps, df['replicas'], marker='o', label='Current Replicas', linewidth=2, color='blue')
    ax1.plot(timestamps, df['desired_replicas'], marker='s', label='Desired Replicas', 
             linewidth=1, linestyle='--', color='orange', alpha=0.7)
    ax1.set_ylabel('Replica Count', fontweight='bold')
    ax1.legend(loc='upper left')
    ax1.grid(True, alpha=0.3)
    ax1.set_ylim(bottom=0)
    
    # Highlight scaling events
    scaling_times = df[df['replicas'] != df['replicas'].shift()]['timestamp']
    for st in scaling_times:
        ax1.axvline(x=st, color='red', linestyle=':', alpha=0.5, linewidth=1)
    
    # Plot 2: Number of waiting requests
    ax2 = axes[1]
    ax2.plot(timestamps, df['num_requests_waiting_current'], marker='o', 
             label='Current Waiting Requests', linewidth=2, color='green')
    ax2.axhline(y=df['num_requests_waiting_target'].iloc[0], color='red', 
                linestyle='--', label='Target Threshold', linewidth=2, alpha=0.7)
    ax2.set_ylabel('Waiting Requests', fontweight='bold')
    ax2.legend(loc='upper left')
    ax2.grid(True, alpha=0.3)
    ax2.set_ylim(bottom=0)
    
    # Plot 3: KV Cache Usage
    ax3 = axes[2]
    ax3.plot(timestamps, df['kv_cache_usage_current'], marker='o', 
             label='Current KV Cache Usage', linewidth=2, color='purple')
    ax3.axhline(y=df['kv_cache_usage_target'].iloc[0], color='red', 
                linestyle='--', label='Target Threshold', linewidth=2, alpha=0.7)
    ax3.set_ylabel('KV Cache Usage (%)', fontweight='bold')
    ax3.legend(loc='upper left')
    ax3.grid(True, alpha=0.3)
    ax3.set_ylim(bottom=0, top=100)
    
    # Plot 4: Active Jobs
    ax4 = axes[3]
    ax4.bar(timestamps, df['active_jobs'], label='Active Jobs', color='teal', alpha=0.7, width=0.002)
    ax4.bar(timestamps, df['completed_jobs'], label='Completed Jobs', 
            color='gray', alpha=0.5, width=0.002, bottom=df['active_jobs'])
    ax4.set_ylabel('Job Count', fontweight='bold')
    ax4.set_xlabel('Time', fontweight='bold')
    ax4.legend(loc='upper left')
    ax4.grid(True, alpha=0.3)
    ax4.set_ylim(bottom=0)
    
    # Format x-axis
    ax4.xaxis.set_major_formatter(mdates.DateFormatter('%H:%M:%S'))
    ax4.xaxis.set_major_locator(mdates.MinuteLocator(interval=2))
    plt.xticks(rotation=45, ha='right')
    
    plt.tight_layout()
    
    if output_file:
        plt.savefig(output_file, dpi=300, bbox_inches='tight')
        print(f"✓ Plot saved to: {output_file}")
    else:
        plt.show()
    
    return fig


def print_summary(df, metadata, scaling_events):
    """Print experiment summary"""
    print("=" * 60)
    print("EXPERIMENT SUMMARY")
    print("=" * 60)
    print(f"Experiment Name: {metadata.get('experiment_name', 'N/A')}")
    print(f"Start Time:      {metadata.get('start_time', 'N/A')}")
    print(f"End Time:        {metadata.get('end_time', 'N/A')}")
    print(f"Namespace:       {metadata.get('namespace', 'N/A')}")
    print(f"HPA Name:        {metadata.get('hpa_name', 'N/A')}")
    print(f"Deployment:      {metadata.get('deployment_name', 'N/A')}")
    print()
    
    print("=" * 60)
    print("SCALING STATISTICS")
    print("=" * 60)
    print(f"Initial Replicas:  {df['replicas'].iloc[0]}")
    print(f"Final Replicas:    {df['replicas'].iloc[-1]}")
    print(f"Max Replicas:      {df['replicas'].max()}")
    print(f"Min Replicas:      {df['replicas'].min()}")
    print(f"Scaling Events:    {len(scaling_events)}")
    print()
    
    print("=" * 60)
    print("METRIC STATISTICS")
    print("=" * 60)
    print("Waiting Requests:")
    print(f"  Mean:   {df['num_requests_waiting_current'].mean():.2f}")
    print(f"  Max:    {df['num_requests_waiting_current'].max():.2f}")
    print(f"  Target: {df['num_requests_waiting_target'].iloc[0]}")
    print()
    print("KV Cache Usage (%):")
    print(f"  Mean:   {df['kv_cache_usage_current'].mean():.2f}")
    print(f"  Max:    {df['kv_cache_usage_current'].max():.2f}")
    print(f"  Target: {df['kv_cache_usage_target'].iloc[0]}")
    print()
    
    print("=" * 60)
    print("JOB STATISTICS")
    print("=" * 60)
    print(f"Max Active Jobs:    {df['active_jobs'].max()}")
    print(f"Total Completed:    {df['completed_jobs'].iloc[-1]}")
    print()
    
    if scaling_events:
        print("=" * 60)
        print("SCALING EVENTS")
        print("=" * 60)
        for i, event in enumerate(scaling_events, 1):
            print(f"\nEvent {i}:")
            print(event['text'][:500])  # First 500 chars
            print()


def main():
    parser = argparse.ArgumentParser(description='Analyze HPA experiment data')
    parser.add_argument('experiment_dir', help='Path to experiment data directory')
    parser.add_argument('--plot', '-p', help='Save plot to file (e.g., plot.png)')
    parser.add_argument('--no-display', action='store_true', help='Do not display plot')
    parser.add_argument('--summary-only', action='store_true', help='Only print summary')
    
    args = parser.parse_args()
    
    # Load data
    print("Loading experiment data...")
    metadata, df, scaling_events = load_experiment_data(args.experiment_dir)
    print(f"✓ Loaded {len(df)} data points")
    print()
    
    # Print summary
    print_summary(df, metadata, scaling_events)
    
    # Create plot
    if not args.summary_only:
        if args.plot or not args.no_display:
            print("Generating plot...")
            plot_experiment(df, metadata, output_file=args.plot)
            if not args.plot and not args.no_display:
                plt.show()


if __name__ == '__main__':
    main()
