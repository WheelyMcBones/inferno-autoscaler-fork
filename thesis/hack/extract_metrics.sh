#!/bin/bash

# Extract optimization metrics from logs
# Usage: ./extract_metrics.sh logfile.log [accelerator]
# Example: ./extract_metrics.sh logfile.log H100

# Default accelerator to extract params from
TARGET_ACC="${2:-H100}"

awk -v target_acc="$TARGET_ACC" '
BEGIN {
    print "timestamp,itlAverage,ttftAverage,rate,inTk,outTk,numRep,itl,ttft,slo_itl,slo_ttft,alpha,beta,gamma,delta"
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

# Extract from "System data prepared for optimization: - { models:" line (for tuner params)
/System data prepared for optimization: - \{ models:/ {
    # Extract timestamp
    if (match($0, /"ts":"[^"]+"/) > 0) {
        ts = substr($0, RSTART+6, RLENGTH-7)
    }
    
    # Find the target accelerator section
    # Pattern: acc: TARGET_ACC, ... decodeParms: { alpha: X, beta: Y }, prefillParms: { gamma: Z, delta: W }
    acc_pattern = "acc: " target_acc
    acc_pos = index($0, acc_pattern)
    
    if (acc_pos > 0) {
        # Extract substring starting from target accelerator
        acc_section = substr($0, acc_pos)
        
        # Find the decodeParms section after this accelerator
        decode_start = index(acc_section, "decodeParms:")
        if (decode_start > 0) {
            decode_section = substr(acc_section, decode_start)
            
            # Extract alpha (first occurrence in decodeParms)
            if (match(decode_section, /alpha: [0-9.]+/) > 0) {
                alpha = substr(decode_section, RSTART+7, RLENGTH-7)
            }
            # Extract beta (first occurrence in decodeParms)
            if (match(decode_section, /beta: [0-9.]+/) > 0) {
                beta = substr(decode_section, RSTART+6, RLENGTH-6)
            }
        }
        
        # Find the prefillParms section after this accelerator
        prefill_start = index(acc_section, "prefillParms:")
        if (prefill_start > 0) {
            prefill_section = substr(acc_section, prefill_start)
            
            # Extract gamma (first occurrence in prefillParms)
            if (match(prefill_section, /gamma: [0-9.]+/) > 0) {
                gamma = substr(prefill_section, RSTART+7, RLENGTH-7)
            }
            # Extract delta (first occurrence in prefillParms)
            if (match(prefill_section, /delta: [0-9.]+/) > 0) {
                delta = substr(prefill_section, RSTART+7, RLENGTH-7)
            }
        }
    }
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
    print ts","itlAvg","ttftAvg","rate","inTk","outTk","numRep","itl","ttft","slo_itl","slo_ttft","alpha","beta","gamma","delta
}
' "$@"