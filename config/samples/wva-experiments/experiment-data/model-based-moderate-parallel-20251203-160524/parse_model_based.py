import json
import sys
import csv
import re

def parse_optimization_solution(solution_str):
    """Extract predicted ITL and TTFT from optimization solution."""
    itl_match = re.search(r'itl=([0-9.]+)', solution_str)
    ttft_match = re.search(r'ttft=([0-9.]+)', solution_str)
    replicas_match = re.search(r'numRep=([0-9]+)', solution_str)
    
    return {
        'predicted_itl': float(itl_match.group(1)) if itl_match else None,
        'predicted_ttft': float(ttft_match.group(1)) if ttft_match else None,
        'predicted_replicas': int(replicas_match.group(1)) if replicas_match else None
    }

log_file = sys.argv[1]
csv_file = sys.argv[2]

events = []
with open(log_file, 'r') as f:
    for line in f:
        try:
            log = json.loads(line.strip())
            ts = log.get('ts', '')
            msg = log.get('msg', '')
            level = log.get('level', '')
            
            # Extract different event types
            if 'Optimization solution' in msg:
                prediction = parse_optimization_solution(msg)
                events.append({
                    'timestamp': ts,
                    'event_type': 'prediction',
                    **prediction
                })
            elif 'Processing decision' in msg:
                # Extract scaling decision
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
            elif 'Found SLO for model' in msg:
                # Extract SLO values
                slo_itl_match = re.search(r'slo-tpot=([0-9]+)', msg)
                slo_ttft_match = re.search(r'slo-ttft=([0-9]+)', msg)
                
                events.append({
                    'timestamp': ts,
                    'event_type': 'slo',
                    'slo_itl': int(slo_itl_match.group(1)) if slo_itl_match else None,
                    'slo_ttft': int(slo_ttft_match.group(1)) if slo_ttft_match else None
                })
            elif 'Metrics collected for VA' in msg or 'observed' in msg.lower():
                # Try to extract observed metrics (these come from Prometheus)
                replicas_match = re.search(r'replicas=([0-9]+)', msg)
                ttft_match = re.search(r'ttft=([0-9.]+)', msg)
                itl_match = re.search(r'itl=([0-9.]+)', msg)
                
                if any([replicas_match, ttft_match, itl_match]):
                    events.append({
                        'timestamp': ts,
                        'event_type': 'observed_metrics',
                        'observed_replicas': int(replicas_match.group(1)) if replicas_match else None,
                        'observed_ttft': float(ttft_match.group(1)) if ttft_match else None,
                        'observed_itl': float(itl_match.group(1)) if itl_match else None
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
