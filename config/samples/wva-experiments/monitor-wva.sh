#!/bin/bash
#
# WVA Log Collector
# Continuously collects WVA controller logs and saves to JSONL file
#

set -euo pipefail

# Ensure common binary paths are available
export PATH="/usr/local/bin:/opt/homebrew/bin:$PATH"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Parse arguments
if [[ $# -lt 2 ]]; then
    print_error "Usage: $0 <controller-namespace> <controller-pod-prefix> <output-file> [interval]"
    echo ""
    echo "Example:"
    echo "  $0 workload-variant-autoscaler-system workload-variant-autoscaler-controller-manager ./logs.jsonl 10"
    exit 1
fi

CONTROLLER_NS="$1"
CONTROLLER_PREFIX="$2"
OUTPUT_FILE="$3"
INTERVAL="${4:-10}"  # Default 10 seconds

print_info "WVA Log Collector Started"
print_info "Namespace: $CONTROLLER_NS"
print_info "Pod prefix: $CONTROLLER_PREFIX"
print_info "Output: $OUTPUT_FILE"
print_info "Poll interval: ${INTERVAL}s"
echo ""

# Find controller pod
print_info "Finding WVA controller pod..."

# WVA runs in HA mode with leader election
# We need to find the leader pod by checking the lease
LEASE_NAME="72dd1cf1.llm-d.ai"  # WVA controller manager lease name

# Get the leader pod from the lease
LEADER_POD=$(kubectl get lease "$LEASE_NAME" -n "$CONTROLLER_NS" -o jsonpath='{.spec.holderIdentity}' 2>/dev/null | cut -d'_' -f1)

if [[ -n "$LEADER_POD" ]]; then
    POD="$LEADER_POD"
    print_info "Found leader pod from lease: $POD"
else
    print_warn "Could not determine leader from lease, falling back to first pod..."
    POD=$(kubectl get pods -n "$CONTROLLER_NS" -o name | grep "$CONTROLLER_PREFIX" | head -n1 | sed 's/pod\///')
    print_info "Using pod: $POD (may not be leader)"
fi

if [[ -z "$POD" ]]; then
    print_error "No controller pod found with prefix: $CONTROLLER_PREFIX"
    exit 1
fi
echo ""

# Initialize output file with header if it doesn't exist
if [[ ! -f "$OUTPUT_FILE" ]]; then
    print_info "Creating new log file: $OUTPUT_FILE"
    touch "$OUTPUT_FILE"
fi

# Track seen log lines to avoid duplicates
# Use a temporary file since associative arrays require bash 4+
SEEN_LINES_FILE=$(mktemp)
trap 'rm -f "$SEEN_LINES_FILE"; print_info "Stopping log collection. Collected $LINE_COUNT log lines."; exit 0' INT TERM EXIT

LINE_COUNT=0

print_info "Starting log collection (press Ctrl+C to stop)..."
echo ""

while true; do
    # Check if we're still monitoring the leader pod
    CURRENT_LEADER=$(kubectl get lease "$LEASE_NAME" -n "$CONTROLLER_NS" -o jsonpath='{.spec.holderIdentity}' 2>/dev/null | cut -d'_' -f1)
    
    if [[ -n "$CURRENT_LEADER" ]] && [[ "$CURRENT_LEADER" != "$POD" ]]; then
        print_warn "Leader election change detected: $POD -> $CURRENT_LEADER"
        POD="$CURRENT_LEADER"
        print_info "Switched to new leader pod, continuing log collection..."
    fi
    
    # Check if pod still exists and is running
    POD_STATUS=$(kubectl get pod "$POD" -n "$CONTROLLER_NS" -o jsonpath='{.status.phase}' 2>/dev/null || echo "NotFound")
    
    if [[ "$POD_STATUS" != "Running" ]]; then
        print_warn "Pod $POD is no longer running (status: $POD_STATUS). Waiting for new leader..."
        sleep 5
        
        # Try to get new leader from lease
        CURRENT_LEADER=$(kubectl get lease "$LEASE_NAME" -n "$CONTROLLER_NS" -o jsonpath='{.spec.holderIdentity}' 2>/dev/null | cut -d'_' -f1)
        
        if [[ -n "$CURRENT_LEADER" ]]; then
            print_info "Detected new leader pod: $CURRENT_LEADER"
            POD="$CURRENT_LEADER"
            print_info "Switched to new leader pod, continuing log collection..."
        else
            print_error "No leader pod found. Exiting."
            exit 1
        fi
    fi
    
    # Get logs from the controller pod, filtering out lines containing "Found"
    if [[ $LINE_COUNT -eq 0 ]]; then
        # First run: get recent logs (last 5 minutes to capture context)
        LOGS=$(kubectl logs -n "$CONTROLLER_NS" "$POD" --since=5m 2>/dev/null | grep -v "Found" || grep -v "Token" || true)
    else
        # Subsequent runs: get logs since last poll interval + buffer
        LOGS=$(kubectl logs -n "$CONTROLLER_NS" "$POD" --since="$((INTERVAL + 5))s" 2>/dev/null | grep -v "Found" || grep -v "Token" || true)
    fi
    
    if [[ -n "$LOGS" ]]; then
        # Process each line
        while IFS= read -r line; do
            # Skip empty lines
            [[ -z "$line" ]] && continue
            
            # Create a hash of the line to detect duplicates
            # macOS md5 outputs "MD5 (stdin) = <hash>", so extract just the hash
            if command -v md5 >/dev/null 2>&1; then
                LINE_HASH=$(echo "$line" | md5 | awk '{print $NF}')
            else
                LINE_HASH=$(echo "$line" | md5sum | cut -d' ' -f1)
            fi
            
            # Skip if we've already seen this exact line
            if grep -q "^$LINE_HASH$" "$SEEN_LINES_FILE" 2>/dev/null; then
                continue
            fi
            
            # Mark this line as seen
            echo "$LINE_HASH" >> "$SEEN_LINES_FILE"
            
            # Append to output file
            echo "$line" >> "$OUTPUT_FILE"
            ((LINE_COUNT++))
            
            # Print progress every 10 lines
            if (( LINE_COUNT % 10 == 0 )); then
                if echo "$line" | jq -e '.msg' >/dev/null 2>&1; then
                    MSG=$(echo "$line" | jq -r '.msg // empty' 2>/dev/null || echo "")
                    if [[ -n "$MSG" ]]; then
                        print_info "Collected $LINE_COUNT lines | Latest: ${MSG:0:80}..."
                    fi
                fi
            fi
        done <<< "$LOGS"
    fi
    
    # Sleep before next poll
    sleep "$INTERVAL"
done
