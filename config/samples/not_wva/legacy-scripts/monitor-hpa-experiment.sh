#!/bin/bash

# HPA Experiment Monitor
# Collects HPA metrics, events, and scaling data for analysis

set -e

# Configuration
NAMESPACE="${NAMESPACE:-llm-d-inference-scheduler}"
HPA_NAME="${HPA_NAME:-vllm-hpa-combined}"
DEPLOYMENT_NAME="${DEPLOYMENT_NAME:-ms-inference-scheduling-llm-d-modelservice-decode}"
SAMPLE_INTERVAL="${SAMPLE_INTERVAL:-5}"  # seconds between samples
OUTPUT_DIR="${OUTPUT_DIR:-./experiment-data}"
EXPERIMENT_NAME="${EXPERIMENT_NAME:-hpa-experiment-$(date +%Y%m%d-%H%M%S)}"

# Create output directory
EXPERIMENT_DIR="$OUTPUT_DIR/$EXPERIMENT_NAME"
mkdir -p "$EXPERIMENT_DIR"

# Output files
METRICS_CSV="$EXPERIMENT_DIR/metrics.csv"
EVENTS_LOG="$EXPERIMENT_DIR/events.log"
SCALING_LOG="$EXPERIMENT_DIR/scaling.log"
METADATA_FILE="$EXPERIMENT_DIR/metadata.json"
JOBS_LOG="$EXPERIMENT_DIR/jobs.log"

echo "=========================================="
echo "HPA Experiment Monitor"
echo "=========================================="
echo "Namespace:       $NAMESPACE"
echo "HPA:             $HPA_NAME"
echo "Deployment:      $DEPLOYMENT_NAME"
echo "Sample Interval: ${SAMPLE_INTERVAL}s"
echo "Output Dir:      $EXPERIMENT_DIR"
echo "=========================================="
echo ""

# Initialize CSV with headers
echo "timestamp,epoch,replicas,desired_replicas,num_requests_waiting_current,num_requests_waiting_target,kv_cache_usage_current,kv_cache_usage_target,active_jobs,completed_jobs" > "$METRICS_CSV"

# Save experiment metadata
cat > "$METADATA_FILE" <<EOF
{
  "experiment_name": "$EXPERIMENT_NAME",
  "start_time": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "namespace": "$NAMESPACE",
  "hpa_name": "$HPA_NAME",
  "deployment_name": "$DEPLOYMENT_NAME",
  "sample_interval_seconds": $SAMPLE_INTERVAL,
  "output_directory": "$EXPERIMENT_DIR"
}
EOF

echo "✓ Initialized experiment: $EXPERIMENT_NAME"
echo "✓ Output directory: $EXPERIMENT_DIR"
echo ""
echo "Starting monitoring... (Press Ctrl+C to stop)"
echo ""

# Track previous replica count to detect scaling events
PREV_REPLICAS=-1
PREV_DESIRED_REPLICAS=-1

# Cleanup function
cleanup() {
    echo ""
    echo "=========================================="
    echo "Stopping monitoring..."
    echo "=========================================="
    
    # Update metadata with end time
    END_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    TEMP_FILE=$(mktemp)
    jq --arg end_time "$END_TIME" '. + {end_time: $end_time}' "$METADATA_FILE" > "$TEMP_FILE"
    mv "$TEMP_FILE" "$METADATA_FILE"
    
    echo "✓ Data saved to: $EXPERIMENT_DIR"
    echo ""
    echo "Files created:"
    echo "  - metrics.csv      : Time-series metrics data"
    echo "  - events.log       : HPA and deployment events"
    echo "  - scaling.log      : Scaling decision log"
    echo "  - jobs.log         : Job status over time"
    echo "  - metadata.json    : Experiment metadata"
    echo ""
    echo "To analyze data:"
    echo "  cat $METRICS_CSV | column -t -s,"
    echo "  cat $SCALING_LOG"
    echo ""
    
    exit 0
}

trap cleanup SIGINT SIGTERM

# Initial event capture
echo "=== Initial State at $(date -u +%Y-%m-%dT%H:%M:%SZ) ===" >> "$EVENTS_LOG"
kubectl get events -n "$NAMESPACE" --field-selector involvedObject.name="$HPA_NAME" --sort-by='.lastTimestamp' >> "$EVENTS_LOG" 2>/dev/null || echo "No HPA events yet" >> "$EVENTS_LOG"
echo "" >> "$EVENTS_LOG"

echo "=== Initial State at $(date -u +%Y-%m-%dT%H:%M:%SZ) ===" >> "$SCALING_LOG"
kubectl describe hpa "$HPA_NAME" -n "$NAMESPACE" >> "$SCALING_LOG" 2>/dev/null || echo "HPA not found" >> "$SCALING_LOG"
echo "" >> "$SCALING_LOG"

# Main monitoring loop
while true; do
    TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    EPOCH=$(date +%s)
    
    # Get HPA status
    HPA_DATA=$(kubectl get hpa "$HPA_NAME" -n "$NAMESPACE" -o json 2>/dev/null || echo '{}')
    
    # Extract replica counts
    REPLICAS=$(echo "$HPA_DATA" | jq -r '.status.currentReplicas // 0')
    DESIRED_REPLICAS=$(echo "$HPA_DATA" | jq -r '.status.desiredReplicas // 0')
    
    # Extract metric values
    METRICS_STATUS=$(echo "$HPA_DATA" | jq -r '.status.currentMetrics // []')
    
    # Parse num_requests_waiting
    NUM_REQUESTS_CURRENT=$(echo "$METRICS_STATUS" | jq -r '.[] | select(.pods.metric.name == "num_requests_waiting") | .pods.current.averageValue // "0"' | sed 's/[^0-9.]//g')
    NUM_REQUESTS_TARGET=$(echo "$HPA_DATA" | jq -r '.spec.metrics[] | select(.pods.metric.name == "num_requests_waiting") | .pods.target.averageValue // "5"' | sed 's/[^0-9.]//g')
    
    # Parse kv_cache_usage_perc
    KV_CACHE_CURRENT=$(echo "$METRICS_STATUS" | jq -r '.[] | select(.pods.metric.name == "kv_cache_usage_perc") | .pods.current.averageValue // "0"' | sed 's/[^0-9.]//g')
    KV_CACHE_TARGET=$(echo "$HPA_DATA" | jq -r '.spec.metrics[] | select(.pods.metric.name == "kv_cache_usage_perc") | .pods.target.averageValue // "80"' | sed 's/[^0-9.]//g')
    
    # Default to 0 if empty
    NUM_REQUESTS_CURRENT=${NUM_REQUESTS_CURRENT:-0}
    NUM_REQUESTS_TARGET=${NUM_REQUESTS_TARGET:-5}
    KV_CACHE_CURRENT=${KV_CACHE_CURRENT:-0}
    KV_CACHE_TARGET=${KV_CACHE_TARGET:-80}
    
    # Count active jobs
    ACTIVE_JOBS=$(kubectl get jobs -n "$NAMESPACE" -l experiment=sharegpt-e2e -o json 2>/dev/null | jq '[.items[] | select(.status.active > 0)] | length')
    COMPLETED_JOBS=$(kubectl get jobs -n "$NAMESPACE" -l experiment=sharegpt-e2e -o json 2>/dev/null | jq '[.items[] | select(.status.succeeded > 0)] | length')
    
    # Write metrics to CSV
    echo "$TIMESTAMP,$EPOCH,$REPLICAS,$DESIRED_REPLICAS,$NUM_REQUESTS_CURRENT,$NUM_REQUESTS_TARGET,$KV_CACHE_CURRENT,$KV_CACHE_TARGET,$ACTIVE_JOBS,$COMPLETED_JOBS" >> "$METRICS_CSV"
    
    # Print status to console
    printf "[%s] Replicas: %d/%d | Waiting: %.1f/%s | KV Cache: %.1f%%/%s%% | Jobs: %d active, %d completed\n" \
        "$TIMESTAMP" "$REPLICAS" "$DESIRED_REPLICAS" \
        "$NUM_REQUESTS_CURRENT" "$NUM_REQUESTS_TARGET" \
        "$KV_CACHE_CURRENT" "$KV_CACHE_TARGET" \
        "$ACTIVE_JOBS" "$COMPLETED_JOBS"
    
    # Detect scaling events
    if [ "$REPLICAS" != "$PREV_REPLICAS" ] || [ "$DESIRED_REPLICAS" != "$PREV_DESIRED_REPLICAS" ]; then
        if [ "$PREV_REPLICAS" != "-1" ]; then
            SCALING_EVENT="SCALING EVENT at $TIMESTAMP"
            echo "========================================" >> "$SCALING_LOG"
            echo "$SCALING_EVENT" >> "$SCALING_LOG"
            echo "========================================" >> "$SCALING_LOG"
            echo "Previous:     $PREV_REPLICAS replicas (desired: $PREV_DESIRED_REPLICAS)" >> "$SCALING_LOG"
            echo "Current:      $REPLICAS replicas (desired: $DESIRED_REPLICAS)" >> "$SCALING_LOG"
            echo "Metrics:" >> "$SCALING_LOG"
            echo "  - Waiting requests: $NUM_REQUESTS_CURRENT (target: $NUM_REQUESTS_TARGET)" >> "$SCALING_LOG"
            echo "  - KV cache usage:   $KV_CACHE_CURRENT% (target: $KV_CACHE_TARGET%)" >> "$SCALING_LOG"
            echo "  - Active jobs:      $ACTIVE_JOBS" >> "$SCALING_LOG"
            echo "" >> "$SCALING_LOG"
            
            # Get recent HPA events
            echo "Recent HPA Events:" >> "$SCALING_LOG"
            kubectl get events -n "$NAMESPACE" --field-selector involvedObject.name="$HPA_NAME" --sort-by='.lastTimestamp' | tail -10 >> "$SCALING_LOG" 2>/dev/null || true
            echo "" >> "$SCALING_LOG"
            
            # Highlight in console
            echo ""
            echo "🔔 $SCALING_EVENT: $PREV_REPLICAS -> $REPLICAS replicas"
            echo ""
        fi
        
        PREV_REPLICAS=$REPLICAS
        PREV_DESIRED_REPLICAS=$DESIRED_REPLICAS
    fi
    
    # Log job status periodically (every 30 seconds)
    if [ $((EPOCH % 30)) -eq 0 ]; then
        echo "=== Job Status at $TIMESTAMP ===" >> "$JOBS_LOG"
        kubectl get jobs -n "$NAMESPACE" -l experiment=sharegpt-e2e -o wide >> "$JOBS_LOG" 2>/dev/null || echo "No jobs found" >> "$JOBS_LOG"
        echo "" >> "$JOBS_LOG"
    fi
    
    sleep "$SAMPLE_INTERVAL"
done
