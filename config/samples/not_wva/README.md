# vLLM HPA with Custom Prometheus Metrics

This directory contains the configuration for setting up Horizontal Pod Autoscaler (HPA) with custom vLLM metrics from Prometheus.

## Files

- **create-prometheus-ca.sh** - Script to extract OpenShift Service CA certificate and create ConfigMap
- **prometheus-adapter-vllm-values.yaml** - Prometheus Adapter Helm values with vLLM metrics configuration
- **hpa-vllm-waiting.yaml** - HPA for num_requests_waiting metric
- **hpa-vllm-kvcache.yaml** - HPA for kv_cache_usage_perc metric  
- **hpa-vllm-combined.yaml** - HPA using both metrics with Max policy
- **setup-vllm-hpa.sh** - Automated setup script

## Setup Steps

### 1. Create the Prometheus CA ConfigMap

The Prometheus Adapter needs the OpenShift Service CA certificate to communicate with Thanos Querier over HTTPS:

```bash
cd /Users/tom/Desktop/llm-d-communitydev/tom_wva
./config/samples/not_wva/create-prometheus-ca.sh
```

This script will:
- Extract the Service CA certificate from OpenShift
- Create the `prometheus-ca` ConfigMap in the `openshift-user-workload-monitoring` namespace

### 2. Deploy/Upgrade Prometheus Adapter

```bash
helm upgrade prometheus-adapter prometheus-community/prometheus-adapter \
  -n openshift-user-workload-monitoring \
  -f config/samples/not_wva/prometheus-adapter-vllm-values.yaml
```

### 3. Verify Custom Metrics are Available

Wait about 30-60 seconds for the adapter to discover metrics, then check:

```bash
# List all custom metrics
kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1" | jq .

# Check specific vLLM metrics
kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1/namespaces/llm-d-inference-scheduler/pods/*/num_requests_waiting" | jq .
kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1/namespaces/llm-d-inference-scheduler/pods/*/kv_cache_usage_perc" | jq .
```

### 4. Deploy HPA

Choose one of the HPA configurations:

```bash
# Option 1: HPA based on waiting requests only
kubectl apply -f config/samples/not_wva/hpa-vllm-waiting.yaml

# Option 2: HPA based on KV cache usage only
kubectl apply -f config/samples/not_wva/hpa-vllm-kvcache.yaml

# Option 3: HPA using both metrics (recommended)
kubectl apply -f config/samples/not_wva/hpa-vllm-combined.yaml
```

### 5. Monitor HPA Status

```bash
# Check HPA status
kubectl describe hpa -n llm-d-inference-scheduler

# Watch HPA in real-time
kubectl get hpa -n llm-d-inference-scheduler --watch

# Check deployment replicas
kubectl get deployment ms-inference-scheduling-llm-d-modelservice-decode -n llm-d-inference-scheduler
```

## Troubleshooting

### Check Prometheus Adapter Logs

```bash
kubectl logs -n openshift-user-workload-monitoring deployment/prometheus-adapter --tail=100
```

### Verify Prometheus Connectivity

```bash
# Exec into the adapter pod
kubectl exec -n openshift-user-workload-monitoring deployment/prometheus-adapter -- \
  curl -k --header "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" \
  "https://thanos-querier.openshift-monitoring.svc.cluster.local:9091/api/v1/series?match[]=vllm:num_requests_waiting"
```

### Check vLLM Metrics in Prometheus

The metrics should be available in Prometheus with these labels:
- `namespace` - Kubernetes namespace
- `pod` - Pod name
- `engine` - vLLM engine ID
- `model_name` - Model being served

## Configuration Details

### Prometheus Adapter Setup

- **Prometheus URL**: `https://thanos-querier.openshift-monitoring.svc.cluster.local:9091`
- **Authentication**: Service account token (`/var/run/secrets/kubernetes.io/serviceaccount/token`)
- **TLS**: Uses Service CA certificate from `prometheus-ca` ConfigMap
- **Namespace**: `openshift-user-workload-monitoring`

### vLLM Metrics

1. **num_requests_waiting** - Number of requests waiting in queue
   - HPA Target: > 5 average requests
   - Scale range: 1-10 replicas

2. **kv_cache_usage_perc** - KV cache utilization (0.0-1.0)
   - HPA Target: > 80% (0.8)
   - Scale range: 1-10 replicas
   - Note: Metric is multiplied by 100 in adapter query for percentage display
