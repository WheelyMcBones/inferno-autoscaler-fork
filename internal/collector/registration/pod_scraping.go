package registration

import (
	"github.com/llm-d-incubation/workload-variant-autoscaler/internal/collector/source"
	"github.com/llm-d-incubation/workload-variant-autoscaler/internal/constants"
)

// RegisterPodScrapingMetrics registers expected metrics for Pod sources
// including vLLM reported metrics and EndPointPicker reported metrics.
func RegisterPodScrapingMetrics(sourceName string, sourceRegistry *source.SourceRegistry) {
	podSource := sourceRegistry.Get(sourceName)
	if podSource == nil {
		// Source not registered yet, skip
		return
	}

	registry := podSource.QueryList()

	// Register vLLM KV cache usage metric
	registry.MustRegister(source.QueryTemplate{
		Name:        constants.VLLMKvCacheUsagePerc,
		Type:        source.QueryTypeMetricName,
		Template:    constants.VLLMKvCacheUsagePerc,
		Params:      []string{},
		Description: "vLLM KV cache utilization percentage (0.0-1.0) scraped directly from pod /metrics endpoint",
	})

	// Register vLLM queue length metric
	registry.MustRegister(source.QueryTemplate{
		Name:        constants.VLLMNumRequestsWaiting,
		Type:        source.QueryTypeMetricName,
		Template:    constants.VLLMNumRequestsWaiting,
		Params:      []string{},
		Description: "Number of requests waiting in vLLM queue, scraped directly from pod /metrics endpoint",
	})

	// Register EPP average KV cache utilization metric
	registry.MustRegister(source.QueryTemplate{
		Name:        constants.EPPInferencePoolAverageKvCacheUtilization,
		Type:        source.QueryTypeMetricName,
		Template:    constants.EPPInferencePoolAverageKvCacheUtilization,
		Params:      []string{},
		Description: "Average KV cache utilization reported by EndPointPicker, scraped directly from pod /metrics endpoint",
	})

	// Register EPP average queue size metric
	registry.MustRegister(source.QueryTemplate{
		Name:        constants.EPPInferencePoolAverageQueueSize,
		Type:        source.QueryTypeMetricName,
		Template:    constants.EPPInferencePoolAverageQueueSize,
		Params:      []string{},
		Description: "Average queue size reported by EndPointPicker, scraped directly from pod /metrics endpoint",
	})
}
