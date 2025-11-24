#!/bin/bash
#
# Enhanced HPA Experiment Monitor with TTFT/ITL Collection
# Collects HPA metrics along with performance metrics (TTFT, ITL)
#

set -euo pipefail

NAMESPACE="${NAMESPACE:-llm-d-inference-scheduler}"
DEPLOYMENT="${DEPLOYMENT:-ms-inference-scheduling-llm-d-modelservice-decode}"
HPA_NAME="${HPA_NAME:-vllm-hpa-combined}"
MODEL_NAME="${MODEL_NAME:-unsloth/Meta-Llama-3.1-8B}"
INTERVAL="${INTERVAL:-5}"
OUTPUT_DIR="${OUTPUT_DIR:-experiment-data/experiment-$(date +%Y%m%d-%H%M%S)}"
PROMETHEUS_URL="${PROMETHEUS_URL:-https://thanos-querier.openshift-monitoring.svc.cluster.local:9091}"

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Output files
METRICS_CSV="$OUTPUT_DIR/metrics.csv"
EVENTS_LOG="$OUTPUT_DIR/scaling-events.log"
SUMMARY_LOG="$OUTPUT_DIR/summary.log"

echo "Starting enhanced HPA experiment monitor..."
echo "Namespace: $NAMESPACE"
echo "Deployment: $DEPLOYMENT"
echo "HPA: $HPA_NAME"
echo "Model: $MODEL_NAME"
echo "Interval: ${INTERVAL}s"
echo "Output directory: $OUTPUT_DIR"
echo ""

# Create CSV header
cat > "$METRICS_CSV" <<EOF
timestamp,replicas,desired_replicas,num_requests_waiting,num_requests_waiting_target,kv_cache_usage_perc,kv_cache_usage_target,ttft_ms,itl_ms,request_rate_per_min,active_jobs
EOF

# Track previous state for event detection
PREV_REPLICAS=0
PREV_DESIRED=0

# Get OpenShift token for Prometheus authentication
TOKEN=$(oc whoami -t 2>/dev/null || echo "")

# Function to query Prometheus
query_prometheus() {
    local query="$1"
    local metric_name="$2"
    
    if [[ -z "$TOKEN" ]]; then
        echo "0"
        return
    fi
    
    # URL encode the query
    local encoded_query=$(echo -n "$query" | jq -sRr @uri)
    
    # Query Prometheus
    local result=$(curl -s -k \
        -H "Authorization: Bearer $TOKEN" \
        "${PROMETHEUS_URL}/api/v1/query?query=${encoded_query}" 2>/dev/null || echo '{}')
    
    # Extract value
    local value=$(echo "$result" | jq -r '.data.result[0].value[1] // "0"' 2>/dev/null || echo "0")
    
    # Handle NaN and empty values
    if [[ "$value" == "NaN" ]] || [[ "$value" == "null" ]] || [[ -z "$value" ]]; then
        echo "0"
    else
        printf "%.2f" "$value" 2>/dev/null || echo "0"
    fi
}

# Function to get active job count
get_active_jobs() {
    kubectl get jobs -n "$NAMESPACE" -l experiment=sharegpt-e2e -o json 2>/dev/null | \
        jq '[.items[] | select(.status.active > 0)] | length' 2>/dev/null || echo "0"
}

# Log scaling event
log_event() {
    local message="$1"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] $message" | tee -a "$EVENTS_LOG"
}

# Cleanup function
cleanup() {
    echo ""
    echo "Stopping monitor..."
    log_event "Monitor stopped"
    exit 0
}

trap cleanup SIGINT SIGTERM

# Initial message
log_event "Monitor started - collecting metrics every ${INTERVAL}s"

# Main monitoring loop
while true; do
    TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
    
    # Get HPA status
    HPA_DATA=$(kubectl get hpa "$HPA_NAME" -n "$NAMESPACE" -o json 2>/dev/null || echo '{}')
    
    # Extract replica counts
    CURRENT_REPLICAS=$(echo "$HPA_DATA" | jq -r '.status.currentReplicas // 0')
    DESIRED_REPLICAS=$(echo "$HPA_DATA" | jq -r '.status.desiredReplicas // 0')
    
    # Extract metrics from HPA status
    NUM_WAITING=$(echo "$HPA_DATA" | jq -r '.status.currentMetrics[] | select(.pods.metric.name=="num_requests_waiting") | .pods.current.averageValue // "0"' | sed 's/[^0-9.]//g')
    NUM_WAITING_TARGET=$(echo "$HPA_DATA" | jq -r '.spec.metrics[] | select(.pods.metric.name=="num_requests_waiting") | .pods.target.averageValue // "0"' | sed 's/[^0-9.]//g')
    
    KV_CACHE=$(echo "$HPA_DATA" | jq -r '.status.currentMetrics[] | select(.pods.metric.name=="kv_cache_usage_perc") | .pods.current.averageValue // "0"' | sed 's/m$//' | awk '{print $1/1000}')
    KV_CACHE_TARGET=$(echo "$HPA_DATA" | jq -r '.spec.metrics[] | select(.pods.metric.name=="kv_cache_usage_perc") | .pods.target.averageValue // "0"' | sed 's/m$//' | awk '{print $1/1000}')
    
    # Default values if empty
    NUM_WAITING=${NUM_WAITING:-0}
    NUM_WAITING_TARGET=${NUM_WAITING_TARGET:-0}
    KV_CACHE=${KV_CACHE:-0}
    KV_CACHE_TARGET=${KV_CACHE_TARGET:-0}
    
    # Query Prometheus for performance metrics
    # TTFT: Time to First Token (ms)
    TTFT_QUERY='sum(rate(vllm:time_to_first_token_seconds_sum{model_name="'$MODEL_NAME'",namespace="'$NAMESPACE'"}[1m]))/sum(rate(vllm:time_to_first_token_seconds_count{model_name="'$MODEL_NAME'",namespace="'$NAMESPACE'"}[1m])) * 1000'
    TTFT=$(query_prometheus "$TTFT_QUERY" "ttft")
    
    # ITL: Inter-Token Latency (ms)
    ITL_QUERY='sum(rate(vllm:time_per_output_token_seconds_sum{model_name="'$MODEL_NAME'",namespace="'$NAMESPACE'"}[1m]))/sum(rate(vllm:time_per_output_token_seconds_count{model_name="'$MODEL_NAME'",namespace="'$NAMESPACE'"}[1m])) * 1000'
    ITL=$(query_prometheus "$ITL_QUERY" "itl")
    
    # Request rate (req/min)
    REQ_RATE_QUERY='sum(rate(vllm:request_success_total{model_name="'$MODEL_NAME'",namespace="'$NAMESPACE'"}[1m])) * 60'
    REQ_RATE=$(query_prometheus "$REQ_RATE_QUERY" "request_rate")
    
    # Get active job count
    ACTIVE_JOBS=$(get_active_jobs)
    
    # Write to CSV
    echo "$TIMESTAMP,$CURRENT_REPLICAS,$DESIRED_REPLICAS,$NUM_WAITING,$NUM_WAITING_TARGET,$KV_CACHE,$KV_CACHE_TARGET,$TTFT,$ITL,$REQ_RATE,$ACTIVE_JOBS" >> "$METRICS_CSV"
    
    # Detect scaling events
    if [[ "$CURRENT_REPLICAS" != "$PREV_REPLICAS" ]] || [[ "$DESIRED_REPLICAS" != "$PREV_DESIRED" ]]; then
        if [[ "$DESIRED_REPLICAS" -gt "$CURRENT_REPLICAS" ]]; then
            log_event "SCALE UP TRIGGERED: $CURRENT_REPLICAS -> $DESIRED_REPLICAS (waiting: $NUM_WAITING, cache: $KV_CACHE)"
        elif [[ "$DESIRED_REPLICAS" -lt "$CURRENT_REPLICAS" ]]; then
            log_event "SCALE DOWN TRIGGERED: $CURRENT_REPLICAS -> $DESIRED_REPLICAS (waiting: $NUM_WAITING, cache: $KV_CACHE)"
        fi
        
        if [[ "$CURRENT_REPLICAS" != "$PREV_REPLICAS" ]]; then
            log_event "REPLICAS CHANGED: $PREV_REPLICAS -> $CURRENT_REPLICAS"
        fi
    fi
    
    # Print status line
    printf "\r[%s] Replicas: %d/%d | Waiting: %.0f/%.0f | Cache: %.1f%%/%.1f%% | TTFT: %.1fms | ITL: %.1fms | Rate: %.1f/min | Jobs: %d" \
        "$TIMESTAMP" "$CURRENT_REPLICAS" "$DESIRED_REPLICAS" "$NUM_WAITING" "$NUM_WAITING_TARGET" \
        "$(echo "$KV_CACHE * 100" | bc -l)" "$(echo "$KV_CACHE_TARGET * 100" | bc -l)" \
        "$TTFT" "$ITL" "$REQ_RATE" "$ACTIVE_JOBS"
    
    # Update previous state
    PREV_REPLICAS=$CURRENT_REPLICAS
    PREV_DESIRED=$DESIRED_REPLICAS
    
    sleep "$INTERVAL"
done
