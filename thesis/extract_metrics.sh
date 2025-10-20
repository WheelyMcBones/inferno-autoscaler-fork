#!/bin/bash

# Extract optimization metrics from logs
# Usage: ./extract_metrics.sh logfile.log

awk '
BEGIN {
    print "timestamp,itlAverage,ttftAverage,rate,inTk,outTk,numRep,itl,ttft,slo_itl,slo_ttft"
}

# Extract from "System data prepared for optimization: - { servers:" line
/System data prepared for optimization: - \{ servers:/ {
    # Extract timestamp
    if (match($0, /"ts":"[^"]+"/) > 0) {
        ts = substr($0, RSTART+6, RLENGTH-7)
    }
    # Extract itlAverage
    if (match($0, /itlAverage: [0-9.]+/) > 0) {
        itlAvg = substr($0, RSTART+12, RLENGTH-12)
    }
    # Extract ttftAverage
    if (match($0, /ttftAverage: [0-9.]+/) > 0) {
        ttftAvg = substr($0, RSTART+13, RLENGTH-13)
    }
    # # Extract arrivalRate
    # if (match($0, /arrivalRate: [0-9.]+/) > 0) {
    #     arrRate = substr($0, RSTART+13, RLENGTH-13)
    # }
}

# Extract from "Optimization solution - system:" line
/Optimization solution - system:/ {
    # Extract rate
    if (match($0, /rate=[0-9.]+/) > 0) {
        rate = substr($0, RSTART+5, RLENGTH-5)
    }
    # Extract inTk
    if (match($0, /inTk=[0-9.]+/) > 0) {
        inTk = substr($0, RSTART+5, RLENGTH-5)
    }
    # Extract outTk
    if (match($0, /outTk=[0-9.]+/) > 0) {
        outTk = substr($0, RSTART+6, RLENGTH-6)
    }
    # Extract numRep
    if (match($0, /numRep=[0-9.]+/) > 0) {
        numRep = substr($0, RSTART+7, RLENGTH-7)
    }
    # Extract itl (from alloc)
    if (match($0, /itl=[0-9.]+/) > 0) {
        itl = substr($0, RSTART+4, RLENGTH-4)
    }
    # Extract ttft (from alloc)
    if (match($0, /ttft=[0-9.]+/) > 0) {
        ttft = substr($0, RSTART+5, RLENGTH-5)
    }
    # Extract slo-itl
    if (match($0, /slo-itl=[0-9.]+/) > 0) {
        slo_itl = substr($0, RSTART+8, RLENGTH-8)
    }
    # Extract slo-ttft
    if (match($0, /slo-ttft=[0-9.]+/) > 0) {
        slo_ttft = substr($0, RSTART+9, RLENGTH-9)
    }
    
    # Print the combined record
    print ts","itlAvg","ttftAvg","rate","inTk","outTk","numRep","itl","ttft","slo_itl","slo_ttft
}
' "$@"