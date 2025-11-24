#!/bin/bash

# HPA Experiment Runner
# Runs load jobs and monitors HPA scaling behavior

set -e

NAMESPACE="${NAMESPACE:-llm-d-inference-scheduler}"
EXPERIMENT_NAME="${EXPERIMENT_NAME:-hpa-experiment-$(date +%Y%m%d-%H%M%S)}"
OUTPUT_DIR="${OUTPUT_DIR:-./experiment-data}"
SAMPLE_INTERVAL="${SAMPLE_INTERVAL:-5}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKLOADS_DIR="$SCRIPT_DIR/../workloads"
SCRIPTS_SUBDIR="$SCRIPT_DIR/scripts"
MANIFESTS_SUBDIR="$SCRIPT_DIR/manifests"

echo "=========================================="
echo "HPA Scaling Experiment Runner"
echo "=========================================="
echo "Experiment:      $EXPERIMENT_NAME"
echo "Namespace:       $NAMESPACE"
echo "Output:          $OUTPUT_DIR/$EXPERIMENT_NAME"
echo "=========================================="
echo ""

# Check if monitoring script exists
MONITOR_SCRIPT="$SCRIPTS_SUBDIR/monitor-hpa-experiment.sh"
if [ ! -f "$MONITOR_SCRIPT" ]; then
    echo "❌ Error: Monitor script not found at $MONITOR_SCRIPT"
    exit 1
fi

# Check if workload files exist
for i in 1 2 3; do
    WORKLOAD="$WORKLOADS_DIR/sharegpt-load-job-$i.yaml"
    if [ ! -f "$WORKLOAD" ]; then
        echo "❌ Error: Workload file not found at $WORKLOAD"
        exit 1
    fi
done

echo "✓ All required files found"
echo ""

# Start monitoring in background
echo "Starting HPA monitor in background..."
export NAMESPACE
export EXPERIMENT_NAME
export OUTPUT_DIR
export SAMPLE_INTERVAL

# Make monitor script executable
chmod +x "$MONITOR_SCRIPT"

# Create experiment directory for monitor log
mkdir -p "$OUTPUT_DIR/$EXPERIMENT_NAME"

# Start monitor in background and capture the actual process PID
"$MONITOR_SCRIPT" > "$OUTPUT_DIR/$EXPERIMENT_NAME/monitor.log" 2>&1 &
MONITOR_PID=$!

# Give it a moment to start
sleep 2

echo "✓ Monitor started (PID: $MONITOR_PID)"
echo "  Log: $OUTPUT_DIR/$EXPERIMENT_NAME/monitor.log"
echo ""

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "=========================================="
    echo "Cleaning up..."
    echo "=========================================="
    
    # Kill monitor process and all its children
    echo "Stopping monitor process..."
    
    # First try graceful shutdown with SIGINT
    if ps -p "$MONITOR_PID" > /dev/null 2>&1; then
        echo "  Sending SIGINT to monitor (PID: $MONITOR_PID)..."
        kill -INT $MONITOR_PID 2>/dev/null || true
        
        # Wait up to 5 seconds for graceful shutdown
        for i in {1..5}; do
            if ! ps -p "$MONITOR_PID" > /dev/null 2>&1; then
                echo "  ✓ Monitor stopped gracefully"
                break
            fi
            sleep 1
        done
    fi
    
    # If still running, force kill
    if ps -p "$MONITOR_PID" > /dev/null 2>&1; then
        echo "  Force stopping monitor..."
        kill -9 $MONITOR_PID 2>/dev/null || true
        sleep 1
    fi
    
    # Also kill any remaining monitor-hpa-experiment.sh processes
    pkill -f "monitor-hpa-experiment.sh" 2>/dev/null || true
    
    echo "  Cleaning up jobs..."
    kubectl delete -f "$WORKLOADS_DIR/sharegpt-load-job-1.yaml" --ignore-not-found 2>/dev/null
    kubectl delete -f "$WORKLOADS_DIR/sharegpt-load-job-2.yaml" --ignore-not-found 2>/dev/null
    kubectl delete -f "$WORKLOADS_DIR/sharegpt-load-job-3.yaml" --ignore-not-found 2>/dev/null
    
    echo "✓ Cleanup complete"
}

trap cleanup EXIT SIGINT SIGTERM

# Give monitor time to start
sleep 3

echo "=========================================="
echo "Starting Experiment"
echo "=========================================="
echo ""

# Phase 1: Deploy first job
echo "[$(date -u +%H:%M:%S)] Phase 1: Deploying sharegpt-load-job-1..."
kubectl apply -f "$WORKLOADS_DIR/sharegpt-load-job-1.yaml"
echo "✓ Job 1 deployed"
echo "  Waiting 360 seconds..."
echo ""

sleep 360

# Phase 2: Deploy second job
echo "[$(date -u +%H:%M:%S)] Phase 2: Deploying sharegpt-load-job-2..."
kubectl apply -f "$WORKLOADS_DIR/sharegpt-load-job-2.yaml"
echo "✓ Job 2 deployed"
echo "  Waiting 360 seconds..."
echo ""

sleep 360

# Phase 3: Deploy third job
echo "[$(date -u +%H:%M:%S)] Phase 3: Deploying sharegpt-load-job-3..."
kubectl apply -f "$WORKLOADS_DIR/sharegpt-load-job-3.yaml"
echo "✓ Job 3 deployed"
echo "  Monitoring for 360 more seconds..."
echo ""

sleep 360

echo ""
echo "=========================================="
echo "Experiment Complete"
echo "=========================================="
echo ""
echo "Duration: ~18 minutes (3 phases × 6 minutes)"
echo ""
echo "Job Status:"
kubectl get jobs -n "$NAMESPACE" -l experiment=sharegpt-e2e
echo ""

echo "Current HPA Status:"
kubectl get hpa -n "$NAMESPACE"
echo ""

echo "Data collected in: $OUTPUT_DIR/$EXPERIMENT_NAME"
echo ""
echo "To view results:"
echo "  cat $OUTPUT_DIR/$EXPERIMENT_NAME/metrics.csv | column -t -s,"
echo "  cat $OUTPUT_DIR/$EXPERIMENT_NAME/scaling.log"
echo ""

# Wait a bit for monitor to finish writing
sleep 5
