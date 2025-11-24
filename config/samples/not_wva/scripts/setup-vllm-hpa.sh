#!/bin/bash
set -e

NAMESPACE=${NAMESPACE:-"llm-d-inference-scheduler"}
ADAPTER_NAMESPACE=${ADAPTER_NAMESPACE:-"openshift-user-workload-monitoring"}

echo "=== vLLM HPA Experiment Setup ==="
echo "Target namespace: $NAMESPACE"
echo "Adapter namespace: $ADAPTER_NAMESPACE"
echo ""

# Step 1: Install Prometheus Adapter
echo "Step 1: Installing Prometheus Adapter..."
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

helm upgrade --install prometheus-adapter prometheus-community/prometheus-adapter \
  --namespace $ADAPTER_NAMESPACE \
  --create-namespace \
  -f config/samples/prometheus-adapter-vllm-values.yaml

echo "Waiting for prometheus-adapter to be ready..."
kubectl wait --for=condition=available --timeout=300s \
  deployment/prometheus-adapter -n $ADAPTER_NAMESPACE

# Step 2: Verify custom metrics API
echo ""
echo "Step 2: Verifying custom metrics API..."
sleep 10
kubectl get apiservices v1beta1.custom.metrics.k8s.io -o yaml

# Step 3: Check available metrics
echo ""
echo "Step 3: Checking available custom metrics..."
kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1" | jq .

# Step 4: Test metric queries (if deployment exists)
echo ""
echo "Step 4: Testing metric queries..."
if kubectl get deployment vllm-deployment -n $NAMESPACE &>/dev/null; then
  echo "Testing num_requests_waiting metric:"
  kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1/namespaces/$NAMESPACE/deployments/vllm-deployment/num_requests_waiting" | jq .
  
  echo ""
  echo "Testing kv_cache_usage_perc metric:"
  kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1/namespaces/$NAMESPACE/deployments/vllm-deployment/kv_cache_usage_perc" | jq .
else
  echo "Deployment vllm-deployment not found in namespace $NAMESPACE, skipping metric query test"
fi

echo ""
echo "=== Setup Complete ==="
echo ""
echo "Available HPA configurations:"
echo "  1. hpa-vllm-waiting.yaml     - Scale on num_requests_waiting (target: 5)"
echo "  2. hpa-vllm-kvcache.yaml     - Scale on kv_cache_usage_perc (target: 80%)"
echo "  3. hpa-vllm-combined.yaml    - Scale on both metrics (uses Max policy)"
echo ""
echo "To deploy an HPA:"
echo "  kubectl apply -f config/samples/hpa-vllm-waiting.yaml"
echo "  kubectl apply -f config/samples/hpa-vllm-kvcache.yaml"
echo "  kubectl apply -f config/samples/hpa-vllm-combined.yaml"
