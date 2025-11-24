#!/bin/bash

set -e

MONITORING_NAMESPACE="openshift-user-workload-monitoring"
PROMETHEUS_SECRET_NS="openshift-monitoring"
PROM_CA_CERT_PATH="/tmp/prometheus-ca.crt"

echo "Extracting OpenShift Service CA certificate for Prometheus Adapter..."

# Method 1: Extract Service CA from openshift-service-ca.crt ConfigMap (preferred)
if kubectl get configmap openshift-service-ca.crt -n $PROMETHEUS_SECRET_NS &> /dev/null; then
    echo "Extracting Service CA from openshift-service-ca.crt ConfigMap"
    kubectl get configmap openshift-service-ca.crt -n $PROMETHEUS_SECRET_NS -o jsonpath='{.data.service-ca\.crt}' > $PROM_CA_CERT_PATH 2>/dev/null || true
    if [ -s "$PROM_CA_CERT_PATH" ]; then
        echo "✓ Extracted Service CA from openshift-service-ca.crt ConfigMap"
    fi
fi

# Method 2: Extract Service CA from openshift-config namespace
if [ ! -s "$PROM_CA_CERT_PATH" ]; then
    echo "Trying to extract Service CA from openshift-config namespace"
    kubectl get configmap openshift-service-ca -n openshift-config -o jsonpath='{.data.service-ca\.crt}' > $PROM_CA_CERT_PATH 2>/dev/null || true
    if [ -s "$PROM_CA_CERT_PATH" ]; then
        echo "✓ Extracted Service CA from openshift-config namespace"
    fi
fi

# Method 3: Fallback to thanos-querier-tls secret
if [ ! -s "$PROM_CA_CERT_PATH" ]; then
    echo "Service CA not found, falling back to server certificate from thanos-querier-tls"
    if kubectl get secret thanos-querier-tls -n $PROMETHEUS_SECRET_NS &> /dev/null; then
        echo "Extracting certificate from thanos-querier-tls secret"
        kubectl get secret thanos-querier-tls -n $PROMETHEUS_SECRET_NS -o jsonpath='{.data.tls\.crt}' | base64 -d > $PROM_CA_CERT_PATH
        if [ -s "$PROM_CA_CERT_PATH" ]; then
            echo "✓ Extracted certificate from thanos-querier-tls secret"
        fi
    fi
fi

# Verify we have a valid certificate
if [ ! -s "$PROM_CA_CERT_PATH" ]; then
    echo "✗ Failed to extract OpenShift Service CA certificate"
    echo "Tried: openshift-service-ca.crt ConfigMap, openshift-config ConfigMap, and thanos-querier-tls secret"
    exit 1
fi

echo "Creating prometheus-ca ConfigMap in $MONITORING_NAMESPACE namespace..."

# Create or update the prometheus-ca ConfigMap
kubectl create configmap prometheus-ca \
    --from-file=ca.crt=$PROM_CA_CERT_PATH \
    -n $MONITORING_NAMESPACE \
    --dry-run=client -o yaml | kubectl apply -f -

echo "✓ prometheus-ca ConfigMap created/updated successfully"

# Cleanup temp file
rm -f $PROM_CA_CERT_PATH

echo ""
echo "Now you can upgrade the Prometheus Adapter:"
echo "  helm upgrade prometheus-adapter prometheus-community/prometheus-adapter \\"
echo "    -n $MONITORING_NAMESPACE \\"
echo "    -f config/samples/not_wva/prometheus-adapter-vllm-values.yaml"
