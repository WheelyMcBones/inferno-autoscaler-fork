package collector

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	llmdVariantAutoscalingV1alpha1 "github.com/llm-d-incubation/workload-variant-autoscaler/api/v1alpha1"
	"github.com/llm-d-incubation/workload-variant-autoscaler/internal/constants"
	"github.com/llm-d-incubation/workload-variant-autoscaler/internal/interfaces"
	"github.com/llm-d-incubation/workload-variant-autoscaler/internal/logger"
	"github.com/llm-d-incubation/workload-variant-autoscaler/internal/utils"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	appsv1 "k8s.io/api/apps/v1"
)

// SaturationMetricsCollector collects vLLM metrics from Prometheus
type SaturationMetricsCollector struct {
	promAPI promv1.API
}

// NewSaturationMetricsCollector creates a new metrics collector
func NewSaturationMetricsCollector(promAPI promv1.API) *SaturationMetricsCollector {
	return &SaturationMetricsCollector{
		promAPI: promAPI,
	}
}

// escapePrometheusLabelValue escapes a label value for safe use in Prometheus queries.
// This prevents PromQL injection attacks by escaping quotes and backslashes.
// Prometheus label values can contain any characters, including forward slashes,
// but must be properly escaped when embedded in query strings.
func escapePrometheusLabelValue(value string) string {
	// Escape backslashes first (must be done before escaping quotes)
	value = regexp.MustCompile(`\\`).ReplaceAllString(value, `\\`)
	// Escape double quotes
	value = regexp.MustCompile(`"`).ReplaceAllString(value, `\"`)
	return value
}

// contextWithRespectedDeadline creates a timeout context that respects the parent context deadline.
// If the parent has a deadline shorter than the desired timeout, uses the parent's remaining time minus a buffer.
// Returns the context and cancel function.
func contextWithRespectedDeadline(parent context.Context, desiredTimeout time.Duration) (context.Context, context.CancelFunc) {
	deadline, hasDeadline := parent.Deadline()
	if !hasDeadline {
		// No parent deadline, use desired timeout
		return context.WithTimeout(parent, desiredTimeout)
	}

	// Calculate remaining time from parent deadline
	remaining := time.Until(deadline)
	if remaining <= 0 {
		// Parent already expired, use minimal timeout
		return context.WithTimeout(parent, time.Millisecond)
	}

	// If remaining time is less than desired, use remaining minus buffer
	const deadlineBuffer = 100 * time.Millisecond
	if remaining < desiredTimeout {
		timeout := remaining - deadlineBuffer
		if timeout < time.Millisecond {
			timeout = time.Millisecond
		}
		return context.WithTimeout(parent, timeout)
	}

	// Parent deadline is generous, use desired timeout
	return context.WithTimeout(parent, desiredTimeout)
}

// CollectReplicaMetrics collects KV cache and queue metrics for all replicas of a model.
// It queries Prometheus for:
// - constants.VLLMKvCacheUsagePerc (KV cache utilization 0.0-1.0)
// - constants.VLLMNumRequestsWaiting (queue length)
//
// Uses max_over_time[1m] to capture peak values in the last minute for safety-first
// guardrails. This prevents missing saturation events that could occur between
// instant queries and provides more conservative analysis.
//
// Uses deployment-to-pod mapping for accurate attribution.
// Each deployment corresponds to a VA, and we get
// the actual pods for each deployment using the pod lists.
func (cmc *SaturationMetricsCollector) CollectReplicaMetrics(
	ctx context.Context,
	modelID string,
	namespace string,
	deployments map[string]*appsv1.Deployment,
	variantAutoscalings map[string]*llmdVariantAutoscalingV1alpha1.VariantAutoscaling,
	variantCosts map[string]float64,
) ([]interfaces.ReplicaMetrics, error) {

	// Validate input to prevent injection and ensure valid queries
	// if err := validatePrometheusLabel(namespace, "namespace"); err != nil {
	// 	return nil, err
	// }
	// if err := validatePrometheusLabel(modelID, "modelID"); err != nil {
	// 	return nil, err
	// }

	// Query KV cache and queue metrics in parallel for better performance
	// Use result struct to avoid race conditions on error variables
	type queryResult struct {
		kvMetrics    map[string]float64
		queueMetrics map[string]int
		kvErr        error
		queueErr     error
	}
	result := &queryResult{}
	var resultMutex sync.Mutex
	var wg sync.WaitGroup

	wg.Add(2)

	// Query KV cache metrics in parallel
	go func() {
		defer wg.Done()
		kv, err := cmc.queryKvCacheMetrics(ctx, modelID, namespace)
		resultMutex.Lock()
		result.kvMetrics = kv
		result.kvErr = err
		resultMutex.Unlock()
	}()

	// Query queue metrics in parallel
	go func() {
		defer wg.Done()
		queue, err := cmc.queryQueueMetrics(ctx, modelID, namespace)
		resultMutex.Lock()
		result.queueMetrics = queue
		result.queueErr = err
		resultMutex.Unlock()
	}()

	wg.Wait()

	// Check for errors after both queries complete
	if result.kvErr != nil {
		return nil, fmt.Errorf("failed to query KV cache metrics: %w", result.kvErr)
	}
	if result.queueErr != nil {
		return nil, fmt.Errorf("failed to query queue metrics: %w", result.queueErr)
	}

	// Use results from struct
	kvMetricsMap := result.kvMetrics
	queueMetricsMap := result.queueMetrics

	// Merge metrics by pod and assign to variants using deployment-to-pod mapping
	replicaMetrics := cmc.mergeMetrics(ctx, kvMetricsMap, queueMetricsMap, modelID, namespace, deployments, variantAutoscalings, variantCosts)

	logger.Log.Debugf("Collected replica metrics: modelID=%s, namespace=%s, replicaCount=%d",
		modelID, namespace, len(replicaMetrics))

	return replicaMetrics, nil
}

// queryKvCacheMetrics queries constants.VLLMKvCacheUsagePerc metric with max_over_time[1m]
// to capture peak KV cache usage in the last minute for conservative analysis.
func (cmc *SaturationMetricsCollector) queryKvCacheMetrics(
	ctx context.Context,
	modelID string,
	namespace string,
) (map[string]float64, error) {

	// Query for peak KV cache usage over last minute across all pods of this model (all variants)
	// Using max_over_time ensures we don't miss saturation events between queries
	// The outer 'max by (pod)' aggregates multiple scrape samples per pod into one value
	// vLLM uses 'model_name' label for the model identifier
	// Escape label values to prevent PromQL injection
	query := fmt.Sprintf(`max by (pod) (max_over_time(%s{namespace="%s",model_name="%s"}[1m]))`,
		constants.VLLMKvCacheUsagePerc, escapePrometheusLabelValue(namespace), escapePrometheusLabelValue(modelID))

	// Add timeout to prevent hanging on Prometheus issues (respects parent deadline)
	queryCtx, cancel := contextWithRespectedDeadline(ctx, 5*time.Second)
	defer cancel()

	result, warnings, err := utils.QueryPrometheusWithBackoff(queryCtx, cmc.promAPI, query)
	if err != nil {
		return nil, fmt.Errorf("prometheus query failed: %w", err)
	}

	if len(warnings) > 0 {
		logger.Log.Warnf("Prometheus query returned warnings: query=%s, warnings=%v",
			query, warnings)
	}

	metricsMap := make(map[string]float64)

	if result.Type() == model.ValVector {
		vector := result.(model.Vector)
		for _, sample := range vector {
			podName := string(sample.Metric["pod"])
			if podName == "" {
				// Try alternative label names
				podName = string(sample.Metric["pod_name"])
			}

			if podName != "" {
				kvValue := float64(sample.Value)
				metricsMap[podName] = kvValue
				logger.Log.Infof("KV cache metric: pod=%s, usage=%.3f (%.1f%%)",
					podName, kvValue, kvValue*100)
			}
		}
	}

	logger.Log.Debugf("KV cache metrics collected (max over 1m): modelID=%s, namespace=%s, podCount=%d",
		modelID, namespace, len(metricsMap))

	return metricsMap, nil
}

// queryQueueMetrics queries constants.VLLMNumRequestsWaiting metric with max_over_time[1m]
// to capture peak queue length in the last minute for conservative saturation analysis.
func (cmc *SaturationMetricsCollector) queryQueueMetrics(
	ctx context.Context,
	modelID string,
	namespace string,
) (map[string]int, error) {

	// Query for peak queue length over last minute
	// Using max_over_time ensures we catch burst traffic that could saturate the system
	// The outer 'max by (pod)' aggregates multiple scrape samples per pod into one value
	// vLLM uses 'model_name' label for the model identifier
	// Escape label values to prevent PromQL injection
	query := fmt.Sprintf(`max by (pod) (max_over_time(%s{namespace="%s",model_name="%s"}[1m]))`,
		constants.VLLMNumRequestsWaiting, escapePrometheusLabelValue(namespace), escapePrometheusLabelValue(modelID))

	// Add timeout to prevent hanging on Prometheus issues (respects parent deadline)
	queryCtx, cancel := contextWithRespectedDeadline(ctx, 5*time.Second)
	defer cancel()

	result, warnings, err := utils.QueryPrometheusWithBackoff(queryCtx, cmc.promAPI, query)
	if err != nil {
		return nil, fmt.Errorf("prometheus query failed: %w", err)
	}

	if len(warnings) > 0 {
		logger.Log.Warnf("Prometheus query returned warnings: query=%s, warnings=%v",
			query, warnings)
	}

	metricsMap := make(map[string]int)

	if result.Type() == model.ValVector {
		vector := result.(model.Vector)
		for _, sample := range vector {
			podName := string(sample.Metric["pod"])
			if podName == "" {
				podName = string(sample.Metric["pod_name"])
			}

			if podName != "" {
				queueLen := int(sample.Value)
				metricsMap[podName] = queueLen
				logger.Log.Infof("Queue metric: pod=%s, queueLength=%d", podName, queueLen)
			}
		}
	}

	logger.Log.Debugf("Queue metrics collected (max over 1m): modelID=%s, namespace=%s, podCount=%d",
		modelID, namespace, len(metricsMap))

	return metricsMap, nil
}

// mergeMetrics combines KV cache and queue metrics into ReplicaMetrics structs.
// It uses deployment-to-pod mapping to assign pods to variants.
func (cmc *SaturationMetricsCollector) mergeMetrics(
	ctx context.Context,
	kvMetrics map[string]float64,
	queueMetrics map[string]int,
	modelID string,
	namespace string,
	deployments map[string]*appsv1.Deployment,
	variantAutoscalings map[string]*llmdVariantAutoscalingV1alpha1.VariantAutoscaling,
	variantCosts map[string]float64,
) []interfaces.ReplicaMetrics {

	// Use union of pod names from both metric sets
	podSet := make(map[string]bool)
	for pod := range kvMetrics {
		podSet[pod] = true
	}
	for pod := range queueMetrics {
		podSet[pod] = true
	}

	// Query kube_pod_info to get existing pods with their deployment owners.
	// This returns a map of the existing pods to their deployment names via a single Prometheus query.
	podToDeployment := cmc.getExistingPodsForDeploymentsFromPrometheus(ctx, namespace, deployments, podSet)

	// Filter out deleted pods (pods with stale reported vLLM metrics)
	for podName := range podSet {
		if _, exists := podToDeployment[podName]; !exists {
			logger.Log.Debugf("Filtering deleted pod: pod=%s, namespace=%s, model=%s",
				podName, namespace, modelID)
			delete(podSet, podName)
		}
	}

	replicaMetrics := make([]interfaces.ReplicaMetrics, 0, len(podSet))

	for podName := range podSet {
		// Get VariantAutoscaling name as Deployment name
		variantName := podToDeployment[podName]

		// Get metrics (defaults to 0 if not present)
		kvUsage := kvMetrics[podName]
		queueLen := queueMetrics[podName]

		// Get accelerator name from VariantAutoscaling label
		acceleratorName := ""
		if va, ok := variantAutoscalings[variantName]; ok && va != nil {
			if va.Labels != nil {
				if accName, exists := va.Labels["inference.optimization/acceleratorName"]; exists {
					acceleratorName = accName
				}
			}
		}

		if acceleratorName == "" {
			logger.Log.Warnf("Missing acceleratorName label on VariantAutoscaling: variant=%s, pod=%s", variantName, podName)
		}

		// Look up cost by variant name, default to DefaultVariantCost if not found
		cost := 10.0
		if variantCosts != nil {
			if c, ok := variantCosts[variantName]; ok {
				cost = c
			}
		}

		metric := interfaces.ReplicaMetrics{
			PodName:         podName,
			ModelID:         modelID,
			Namespace:       namespace,
			VariantName:     variantName,
			AcceleratorName: acceleratorName,
			KvCacheUsage:    kvUsage,
			QueueLength:     queueLen,
			Cost:            cost,
		}

		replicaMetrics = append(replicaMetrics, metric)
	}

	return replicaMetrics
}

// getExistingPodsForDeploymentsFromPrometheus queries kube_pod_info to get existing pods and their Deployment owners.
// This combines deleted Pods filtering with Deployment matching in a single Prometheus query.
//
// Returns a map of podName → deploymentName for Pods that:
// 1. Currently exist (non-deleted Pods with stale metrics)
// 2. Are in the candidate set (from vLLM metrics)
// 3. Can be matched to a Deployment corresponding to an existing VariantAutoscaling
//
// Matching strategy:
// - For ReplicaSet-managed pods: matches ReplicaSet name to Deployment name
//
// TODO(note): this approach may still be subject to staleness, as the scrape interval (typically 15-30s)
// adds latency between pod termination and metric removal
func (cmc *SaturationMetricsCollector) getExistingPodsForDeploymentsFromPrometheus(
	ctx context.Context,
	namespace string,
	deployments map[string]*appsv1.Deployment,
	candidatePods map[string]bool,
) map[string]string {
	podToDeployment := make(map[string]string)

	// Build Pod name regex filter from Deployment names
	var podQueryFilter string
	if len(deployments) > 0 {
		deploymentNames := make([]string, 0, len(deployments))
		for deploymentName := range deployments {
			escapedName := escapePrometheusLabelValue(deploymentName)
			deploymentNames = append(deploymentNames, escapedName+"-.*")
		}
		podQueryFilter = fmt.Sprintf(`,pod=~"%s"`, strings.Join(deploymentNames, "|"))
	}

	// Query kube_pod_info for current pods in namespace with deployment name filtering
	// kube_pod_info is a gauge metric from kube-state-metrics that reflects current pod state
	// Note: this may still be subject to staleness due to scrape intervals - the observed lag is typically ~30s.
	query := fmt.Sprintf(`kube_pod_info{namespace="%s"%s}`, escapePrometheusLabelValue(namespace), podQueryFilter)

	result, warnings, err := utils.QueryPrometheusWithBackoff(ctx, cmc.promAPI, query)
	if err != nil {
		logger.Log.Errorf("Failed to query kube_pod_info: namespace=%s, error=%v", namespace, err)
		// On error, fall back to name-based matching for all candidates
		for podName := range candidatePods {
			if deploymentName := matchToDeploymentByName(podName, deployments); deploymentName != "" {
				podToDeployment[podName] = deploymentName
			}
		}
		return podToDeployment
	}

	if len(warnings) > 0 {
		logger.Log.Warnf("Prometheus query warnings: query=%s, warnings=%v", query, warnings)
	}

	// Extract Pod and Deployment from kube_pod_info labels
	if result.Type() == model.ValVector {
		vector := result.(model.Vector)
		for _, sample := range vector {
			podName := string(sample.Metric["pod"])
			if podName == "" {
				logger.Log.Warnf("Empty pod name in kube_pod_info metric: namespace=%s, metric=%v", namespace, sample.Metric)
				continue
			}
			// Validate pod name is present in the candidate list
			if !candidatePods[podName] {
				continue
			}

			// Extract owner information from kube_pod_info labels
			createdByKind := string(sample.Metric["created_by_kind"])
			createdByName := string(sample.Metric["created_by_name"])

			if createdByName == "" {
				logger.Log.Debugf("Pod has no owner: pod=%s", podName)
				continue
			}

			// Try to match owner name to known Deployments
			deploymentName := matchToDeploymentByName(createdByName, deployments)
			if deploymentName == "" {
				logger.Log.Warnf("Could not match pod to owner: pod=%s, owner_kind=%s, owner_name=%s",
					podName, createdByKind, createdByName)
				continue
			}

			podToDeployment[podName] = deploymentName
		}
	}

	logger.Log.Debugf("Pod to owner mapping: namespace=%s, candidatePods=%d, matched=%d",
		namespace, len(candidatePods), len(podToDeployment))

	return podToDeployment
}

// matchToDeploymentByName matches a name (of a Pod, ReplicaSet) to a Deployment.
// Returns the Deployment name with the longest matching prefix, or exact match, or empty string if no match is found.
func matchToDeploymentByName(name string, deployments map[string]*appsv1.Deployment) string {
	// Check for exact match
	if _, exists := deployments[name]; exists {
		return name
	}

	// Check for prefix match (for ReplicaSets and Pods)
	var matchedDeployment string
	maxPrefixLen := 0

	for deploymentName := range deployments {
		prefix := deploymentName + "-"
		if strings.HasPrefix(name, prefix) && len(prefix) > maxPrefixLen {
			matchedDeployment = deploymentName
			maxPrefixLen = len(prefix)
		}
	}

	return matchedDeployment
}
