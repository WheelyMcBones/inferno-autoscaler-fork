import json
import sys
import csv
import re

log_file = sys.argv[1]
csv_file = sys.argv[2]

events = []
with open(log_file, 'r') as f:
    for line in f:
        try:
            log = json.loads(line.strip())
            ts = log.get('ts', '')
            msg = log.get('msg', '')
            
            # KV cache metrics
            if 'KV cache metric' in msg:
                pod_match = re.search(r'pod=([a-z0-9-]+)', msg)
                usage_match = re.search(r'usage=([0-9.]+)', msg)
                percent_match = re.search(r'\(([0-9.]+)%\)', msg)
                
                events.append({
                    'timestamp': ts,
                    'event_type': 'kv_cache',
                    'pod': pod_match.group(1) if pod_match else None,
                    'kv_cache_usage': float(usage_match.group(1)) if usage_match else None,
                    'kv_cache_percent': float(percent_match.group(1)) if percent_match else None
                })
            
            # Queue metrics
            elif 'Queue metric' in msg:
                pod_match = re.search(r'pod=([a-z0-9-]+)', msg)
                queue_match = re.search(r'queueLength=([0-9]+)', msg)
                
                events.append({
                    'timestamp': ts,
                    'event_type': 'queue',
                    'pod': pod_match.group(1) if pod_match else None,
                    'queue_length': int(queue_match.group(1)) if queue_match else None
                })
            
            # Metrics collected
            elif 'Metrics collected for VA' in msg:
                replicas_match = re.search(r'replicas=([0-9]+)', msg)
                ttft_match = re.search(r'ttft=([0-9.]+)', msg)
                itl_match = re.search(r'itl=([0-9.]+)', msg)
                cost_match = re.search(r'cost=([0-9.]+)', msg)
                
                events.append({
                    'timestamp': ts,
                    'event_type': 'metrics',
                    'replicas': int(replicas_match.group(1)) if replicas_match else None,
                    'ttft': float(ttft_match.group(1)) if ttft_match else None,
                    'itl': float(itl_match.group(1)) if itl_match else None,
                    'cost': float(cost_match.group(1)) if cost_match else None
                })
            
            # Capacity analysis
            elif 'Capacity analysis completed' in msg:
                total_match = re.search(r'totalReplicas=([0-9]+)', msg)
                nonsaturated_match = re.search(r'nonSaturated=([0-9]+)', msg)
                scaleup_match = re.search(r'shouldScaleUp=([a-z]+)', msg)
                scaledown_match = re.search(r'scaleDownSafe=([a-z]+)', msg)
                
                events.append({
                    'timestamp': ts,
                    'event_type': 'capacity_analysis',
                    'total_replicas': int(total_match.group(1)) if total_match else None,
                    'non_saturated_replicas': int(nonsaturated_match.group(1)) if nonsaturated_match else None,
                    'should_scale_up': scaleup_match.group(1) == 'true' if scaleup_match else None,
                    'scale_down_safe': scaledown_match.group(1) == 'true' if scaledown_match else None
                })
            
            # Scaling decisions
            elif 'Processing decision' in msg:
                current_match = re.search(r'current=([0-9]+)', msg)
                target_match = re.search(r'target=([0-9]+)', msg)
                action_match = re.search(r'action=([a-z-]+)', msg)
                
                events.append({
                    'timestamp': ts,
                    'event_type': 'scaling_decision',
                    'current_replicas': int(current_match.group(1)) if current_match else None,
                    'target_replicas': int(target_match.group(1)) if target_match else None,
                    'action': action_match.group(1) if action_match else None
                })
                    
        except json.JSONDecodeError:
            continue
        except Exception as e:
            print(f"Error parsing line: {e}", file=sys.stderr)
            continue

# Write to CSV
if events:
    fieldnames = sorted(set().union(*(d.keys() for d in events)))
    with open(csv_file, 'w', newline='') as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()
        writer.writerows(events)
    print(f"Parsed {len(events)} events to {csv_file}")
else:
    print("No events found in logs", file=sys.stderr)
