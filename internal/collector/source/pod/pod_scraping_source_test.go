package pod

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/llm-d-incubation/workload-variant-autoscaler/internal/collector/registration"
	sourcepkg "github.com/llm-d-incubation/workload-variant-autoscaler/internal/collector/source"
	"github.com/llm-d-incubation/workload-variant-autoscaler/internal/constants"
)

var _ = Describe("PodScrapingSource", func() {
	var (
		ctx        context.Context
		fakeClient *fake.ClientBuilder
		scheme     *runtime.Scheme
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		fakeClient = fake.NewClientBuilder().WithScheme(scheme)
	})

	buildClient := func() *fake.ClientBuilder {
		return fake.NewClientBuilder().WithScheme(scheme)
	}

	Describe("NewPodScrapingSource", func() {
		It("should create source with provided service name", func() {
			config := PodScrapingSourceConfig{
				ServiceName:      "test-pool-epp",
				ServiceNamespace: "test-ns",
				MetricsPort:      9090,
			}
			source, err := NewPodScrapingSource(ctx, fakeClient.Build(), config)
			Expect(err).NotTo(HaveOccurred())
			Expect(source).NotTo(BeNil())
			Expect(source.config.ServiceName).To(Equal("test-pool-epp"))
		})

		It("should return error if ServiceName is empty", func() {
			config := PodScrapingSourceConfig{
				ServiceNamespace: "test-ns",
				MetricsPort:      9090,
			}
			_, err := NewPodScrapingSource(ctx, fakeClient.Build(), config)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("ServiceName is required"))
		})

		It("should return error if ServiceNamespace is empty", func() {
			config := PodScrapingSourceConfig{
				ServiceName: "test-pool-epp",
				MetricsPort: 9090,
			}
			_, err := NewPodScrapingSource(ctx, fakeClient.Build(), config)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("ServiceNamespace is required"))
		})

		It("should set defaults for missing config values", func() {
			config := PodScrapingSourceConfig{
				ServiceName:      "test-pool-epp",
				ServiceNamespace: "test-ns",
				MetricsPort:      9090,
			}
			source, err := NewPodScrapingSource(ctx, fakeClient.Build(), config)
			Expect(err).NotTo(HaveOccurred())
			Expect(source.config.MetricsPath).To(Equal("/metrics"))
			Expect(source.config.MetricsScheme).To(Equal("http"))
			Expect(source.config.MetricsReaderSecretName).To(BeEmpty(), "MetricsReaderSecretName should be empty by default (no EPP-specific default)")
			Expect(source.config.MetricsReaderSecretKey).To(Equal("token"))
			Expect(source.config.ScrapeTimeout).To(Equal(5 * time.Second))
			Expect(source.config.MaxConcurrentScrapes).To(Equal(10))
			Expect(source.config.DefaultTTL).To(Equal(30 * time.Second))
		})
	})

	Describe("service name validation", func() {
		It("should require service name to be provided", func() {
			config := PodScrapingSourceConfig{
				ServiceNamespace: "test-ns",
				MetricsPort:      9090,
			}
			_, err := NewPodScrapingSource(ctx, fakeClient.Build(), config)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("ServiceName is required"))
		})
	})

	Describe("isPodReady", func() {
		It("should return true for Ready pod", func() {
			pod := &corev1.Pod{
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			}
			Expect(isPodReady(pod)).To(BeTrue())
		})

		It("should return false for not Ready pod", func() {
			pod := &corev1.Pod{
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodReady,
							Status: corev1.ConditionFalse,
						},
					},
				},
			}
			Expect(isPodReady(pod)).To(BeFalse())
		})

		It("should return false for pod without Ready condition", func() {
			pod := &corev1.Pod{
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodInitialized,
							Status: corev1.ConditionTrue,
						},
					},
				},
			}
			Expect(isPodReady(pod)).To(BeFalse())
		})
	})

	Describe("discoverPods", func() {
		var (
			service *corev1.Service
			pod1    *corev1.Pod
			pod2    *corev1.Pod
			pod3    *corev1.Pod // Not ready
		)

		BeforeEach(func() {
			service = &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pool-epp",
					Namespace: "test-ns",
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"inferencepool": "test-pool-epp",
					},
				},
			}

			pod1 = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "epp-pod-1",
					Namespace: "test-ns",
					Labels: map[string]string{
						"inferencepool": "test-pool-epp",
					},
				},
				Status: corev1.PodStatus{
					PodIP: "10.0.0.1",
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			}

			pod2 = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "epp-pod-2",
					Namespace: "test-ns",
					Labels: map[string]string{
						"inferencepool": "test-pool-epp",
					},
				},
				Status: corev1.PodStatus{
					PodIP: "10.0.0.2",
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			}

			pod3 = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "epp-pod-3",
					Namespace: "test-ns",
					Labels: map[string]string{
						"inferencepool": "test-pool-epp",
					},
				},
				Status: corev1.PodStatus{
					PodIP: "10.0.0.3",
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodReady,
							Status: corev1.ConditionFalse,
						},
					},
				},
			}
		})

		It("should discover Ready pods only", func() {
			client := fakeClient.
				WithObjects(service, pod1, pod2, pod3).
				Build()

			config := PodScrapingSourceConfig{
				ServiceName:      "test-pool-epp",
				ServiceNamespace: "test-ns",
				MetricsPort:      9090,
			}
			source, err := NewPodScrapingSource(ctx, client, config)
			Expect(err).NotTo(HaveOccurred())

			pods, err := source.discoverPods(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(pods).To(HaveLen(2))
			Expect(pods[0].Name).To(BeElementOf("epp-pod-1", "epp-pod-2"))
			Expect(pods[1].Name).To(BeElementOf("epp-pod-1", "epp-pod-2"))
		})

		It("should return error if service not found", func() {
			client := fakeClient.Build()

			config := PodScrapingSourceConfig{
				ServiceName:      "nonexistent-service",
				ServiceNamespace: "test-ns",
				MetricsPort:      9090,
			}
			source, err := NewPodScrapingSource(ctx, client, config)
			Expect(err).NotTo(HaveOccurred())

			_, err = source.discoverPods(ctx)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to get service"))
		})

		It("should return empty list if no pods match selector", func() {
			client := fakeClient.
				WithObjects(service).
				Build()

			config := PodScrapingSourceConfig{
				ServiceName:      "test-pool-epp",
				ServiceNamespace: "test-ns",
				MetricsPort:      9090,
			}
			source, err := NewPodScrapingSource(ctx, client, config)
			Expect(err).NotTo(HaveOccurred())

			pods, err := source.discoverPods(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(pods).To(BeEmpty())
		})

		It("should return empty list if service has no selector (headless service)", func() {
			// Create a service without selector (headless service)
			headlessService := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "headless-service",
					Namespace: "test-ns",
				},
				Spec: corev1.ServiceSpec{
					Selector:  map[string]string{}, // Empty selector
					ClusterIP: "None",              // Headless service
				},
			}

			client := fakeClient.
				WithObjects(headlessService).
				Build()

			config := PodScrapingSourceConfig{
				ServiceName:      "headless-service",
				ServiceNamespace: "test-ns",
				MetricsPort:      9090,
			}
			source, err := NewPodScrapingSource(ctx, client, config)
			Expect(err).NotTo(HaveOccurred())

			pods, err := source.discoverPods(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(pods).To(BeEmpty(), "Should return empty list for service without selector")
		})

		It("should return empty list if service selector is nil", func() {
			// Create a service with nil selector
			serviceWithNilSelector := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nil-selector-service",
					Namespace: "test-ns",
				},
				Spec: corev1.ServiceSpec{
					Selector: nil, // Nil selector
				},
			}

			client := fakeClient.
				WithObjects(serviceWithNilSelector).
				Build()

			config := PodScrapingSourceConfig{
				ServiceName:      "nil-selector-service",
				ServiceNamespace: "test-ns",
				MetricsPort:      9090,
			}
			source, err := NewPodScrapingSource(ctx, client, config)
			Expect(err).NotTo(HaveOccurred())

			pods, err := source.discoverPods(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(pods).To(BeEmpty(), "Should return empty list for service with nil selector")
		})
	})

	Describe("getAuthToken", func() {
		var secret *corev1.Secret

		BeforeEach(func() {
			secret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "inference-gateway-sa-metrics-reader-secret",
					Namespace: "test-ns",
				},
				Data: map[string][]byte{
					"token": []byte("test-bearer-token"),
				},
			}
		})

		It("should read token from secret", func() {
			client := fakeClient.
				WithObjects(secret).
				Build()

			config := PodScrapingSourceConfig{
				ServiceName:             "test-pool-epp",
				ServiceNamespace:        "test-ns",
				MetricsPort:             9090,
				MetricsReaderSecretName: "inference-gateway-sa-metrics-reader-secret",
			}
			source, err := NewPodScrapingSource(ctx, client, config)
			Expect(err).NotTo(HaveOccurred())

			token, useAuth, err := source.getAuthToken(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(useAuth).To(BeTrue())
			Expect(token).To(Equal("test-bearer-token"))
		})

		It("should use explicit BearerToken if provided", func() {
			client := fakeClient.Build()

			config := PodScrapingSourceConfig{
				ServiceName:      "test-pool-epp",
				ServiceNamespace: "test-ns",
				BearerToken:      "explicit-token",
				MetricsPort:      9090,
			}
			source, err := NewPodScrapingSource(ctx, client, config)
			Expect(err).NotTo(HaveOccurred())

			token, useAuth, err := source.getAuthToken(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(useAuth).To(BeTrue())
			Expect(token).To(Equal("explicit-token"))
		})

		It("should skip authentication if secret not found (optional auth)", func() {
			client := fakeClient.Build()

			config := PodScrapingSourceConfig{
				ServiceName:      "test-pool-epp",
				ServiceNamespace: "test-ns",
				MetricsPort:      9090,
			}
			source, err := NewPodScrapingSource(ctx, client, config)
			Expect(err).NotTo(HaveOccurred())

			token, useAuth, err := source.getAuthToken(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(useAuth).To(BeFalse())
			Expect(token).To(BeEmpty())
		})

		It("should skip authentication if token key not found in secret (optional auth)", func() {
			secret.Data = map[string][]byte{} // Empty secret
			client := fakeClient.
				WithObjects(secret).
				Build()

			config := PodScrapingSourceConfig{
				ServiceName:             "test-pool-epp",
				ServiceNamespace:        "test-ns",
				MetricsPort:             9090,
				MetricsReaderSecretName: "inference-gateway-sa-metrics-reader-secret",
			}
			source, err := NewPodScrapingSource(ctx, client, config)
			Expect(err).NotTo(HaveOccurred())

			token, useAuth, err := source.getAuthToken(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(useAuth).To(BeFalse())
			Expect(token).To(BeEmpty())
		})

		It("should skip authentication if MetricsReaderSecretName is empty", func() {
			client := fakeClient.Build()

			config := PodScrapingSourceConfig{
				ServiceName:             "test-pool-epp",
				ServiceNamespace:        "test-ns",
				MetricsPort:             9090,
				MetricsReaderSecretName: "", // Empty - no auth
			}
			source, err := NewPodScrapingSource(ctx, client, config)
			Expect(err).NotTo(HaveOccurred())

			token, useAuth, err := source.getAuthToken(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(useAuth).To(BeFalse())
			Expect(token).To(BeEmpty())
		})
	})

	Describe("parsePrometheusMetrics", func() {
		var source *PodScrapingSource

		BeforeEach(func() {
			config := PodScrapingSourceConfig{
				ServiceName:      "test-pool-epp",
				ServiceNamespace: "test-ns",
				MetricsPort:      9090,
			}
			var err error
			source, err = NewPodScrapingSource(ctx, fakeClient.Build(), config)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should parse Prometheus text format", func() {
			metricsText := `# HELP vllm:kv_cache_usage_perc KV cache usage percentage
# TYPE vllm:kv_cache_usage_perc gauge
vllm:kv_cache_usage_perc{namespace="test-ns"} 0.75
# HELP vllm:num_requests_waiting Number of requests waiting
# TYPE vllm:num_requests_waiting gauge
vllm:num_requests_waiting{namespace="test-ns"} 5
`

			result, err := source.parsePrometheusMetrics(
				&mockReader{data: []byte(metricsText)},
				"test-pod",
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.Values).To(HaveLen(2))
			Expect(result.QueryName).To(Equal("all_metrics"))

			// Check metrics (order not guaranteed from map iteration)
			metricsByName := make(map[string]sourcepkg.MetricValue)
			for _, value := range result.Values {
				Expect(value.Labels["pod"]).To(Equal("test-pod"))
				metricsByName[value.Labels["__name__"]] = value
			}

			// Check first metric
			Expect(metricsByName).To(HaveKey(constants.VLLMKvCacheUsagePerc))
			Expect(metricsByName[constants.VLLMKvCacheUsagePerc].Value).To(Equal(0.75))

			// Check second metric
			Expect(metricsByName).To(HaveKey(constants.VLLMNumRequestsWaiting))
			Expect(metricsByName[constants.VLLMNumRequestsWaiting].Value).To(Equal(5.0))
		})

		It("should handle empty metrics", func() {
			result, err := source.parsePrometheusMetrics(
				&mockReader{data: []byte("")},
				"test-pod",
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.Values).To(BeEmpty())
		})

		It("should return error for invalid Prometheus format", func() {
			_, err := source.parsePrometheusMetrics(
				&mockReader{data: []byte("invalid prometheus format!!!")},
				"test-pod",
			)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Refresh", func() {
		var (
			service     *corev1.Service
			secret      *corev1.Secret
			readyPod1   *corev1.Pod
			readyPod2   *corev1.Pod
			mockServer1 *httptest.Server
			mockServer2 *httptest.Server
		)

		BeforeEach(func() {
			service = &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pool-epp",
					Namespace: "test-ns",
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"inferencepool": "test-pool-epp",
					},
				},
			}

			secret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "inference-gateway-sa-metrics-reader-secret",
					Namespace: "test-ns",
				},
				Data: map[string][]byte{
					"token": []byte("test-token"),
				},
			}

			// Create mock HTTP servers for pods
			mockServer1 = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify Authorization header
				auth := r.Header.Get("Authorization")
				Expect(auth).To(Equal("Bearer test-token"))
				Expect(r.URL.Path).To(Equal("/metrics"))

				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprint(w, `# HELP vllm:kv_cache_usage_perc KV cache usage
# TYPE vllm:kv_cache_usage_perc gauge
vllm:kv_cache_usage_perc{namespace="test-ns"} 0.75
# HELP vllm:num_requests_waiting Number of requests waiting
# TYPE vllm:num_requests_waiting gauge
vllm:num_requests_waiting{namespace="test-ns"} 5
`)
			}))

			mockServer2 = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				auth := r.Header.Get("Authorization")
				Expect(auth).To(Equal("Bearer test-token"))
				Expect(r.URL.Path).To(Equal("/metrics"))

				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprint(w, `# HELP vllm:kv_cache_usage_perc KV cache usage
# TYPE vllm:kv_cache_usage_perc gauge
vllm:kv_cache_usage_perc{namespace="test-ns"} 0.50
# HELP vllm:num_requests_waiting Number of requests waiting
# TYPE vllm:num_requests_waiting gauge
vllm:num_requests_waiting{namespace="test-ns"} 3
`)
			}))

			// Extract host and port from mock server URLs
			// httptest.Server URL format: "http://127.0.0.1:PORT"
			// We'll use localhost IP and extract the port
			server1URL := mockServer1.URL
			server2URL := mockServer2.URL

			// Parse URLs to get host:port
			// For testing, we'll use 127.0.0.1 as pod IP and extract port
			readyPod1 = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "epp-pod-1",
					Namespace: "test-ns",
					Labels: map[string]string{
						"inferencepool": "test-pool-epp",
					},
				},
				Status: corev1.PodStatus{
					PodIP: "127.0.0.1", // Will be overridden with actual server address
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			}

			readyPod2 = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "epp-pod-2",
					Namespace: "test-ns",
					Labels: map[string]string{
						"inferencepool": "test-pool-epp",
					},
				},
				Status: corev1.PodStatus{
					PodIP: "127.0.0.1", // Will be overridden with actual server address
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			}

			// Store server URLs for later use in tests
			_ = server1URL
			_ = server2URL
		})

		AfterEach(func() {
			if mockServer1 != nil {
				mockServer1.Close()
			}
			if mockServer2 != nil {
				mockServer2.Close()
			}
		})

		It("should scrape metrics from all Ready pods and aggregate results", func() {
			// Parse server URLs to extract ports
			server1URL := mockServer1.URL
			server2URL := mockServer2.URL

			// Extract port from URLs (format: http://127.0.0.1:PORT)
			var port1, port2 int32
			_, err := fmt.Sscanf(server1URL, "http://127.0.0.1:%d", &port1)
			Expect(err).NotTo(HaveOccurred())
			_, err = fmt.Sscanf(server2URL, "http://127.0.0.1:%d", &port2)
			Expect(err).NotTo(HaveOccurred())

			// Update pods with correct IPs
			readyPod1.Status.PodIP = "127.0.0.1"
			readyPod2.Status.PodIP = "127.0.0.1"

			// Create separate sources for each pod (since they use different ports)
			// In real scenario, all pods use same port but different IPs
			client1 := fakeClient.
				WithObjects(service, secret, readyPod1).
				Build()

			config1 := PodScrapingSourceConfig{
				ServiceName:             "test-pool-epp",
				ServiceNamespace:        "test-ns",
				MetricsPort:             port1,
				MetricsPath:             "/metrics",
				MetricsScheme:           "http",
				ScrapeTimeout:           5 * time.Second,
				MaxConcurrentScrapes:    10,
				MetricsReaderSecretName: "inference-gateway-sa-metrics-reader-secret",
			}
			source1, err := NewPodScrapingSource(ctx, client1, config1)
			Expect(err).NotTo(HaveOccurred())

			// Pre-register the metrics we expect to scrape using centralized registration
			sourceRegistry1 := sourcepkg.NewSourceRegistry()
			sourceRegistry1.MustRegister("test-pod-source", source1)
			registration.RegisterPodScrapingMetrics("test-pod-source", sourceRegistry1)

			// Test scraping from first pod - with empty RefreshSpec, it will use all registered queries
			results1, err := source1.Refresh(ctx, sourcepkg.RefreshSpec{})
			Expect(err).NotTo(HaveOccurred())

			// Should return 2 vLLM metrics that the mock server provides
			// RegisterPodScrapingMetrics also registers EPP metrics, but mock server doesn't return them
			Expect(results1).To(HaveLen(2))
			Expect(results1).To(HaveKey(constants.VLLMKvCacheUsagePerc))
			Expect(results1).To(HaveKey(constants.VLLMNumRequestsWaiting))

			// Verify KV cache metric
			kvCache := results1[constants.VLLMKvCacheUsagePerc]
			Expect(kvCache.Values).To(HaveLen(1), "KV cache metric should have 1 value from pod1")
			Expect(kvCache.Values[0].Value).To(Equal(0.75))
			Expect(kvCache.Values[0].Labels["pod"]).To(Equal("epp-pod-1"))

			// Verify queue metric
			queue := results1[constants.VLLMNumRequestsWaiting]
			Expect(queue.Values).To(HaveLen(1), "Queue metric should have 1 value from pod1")
			Expect(queue.Values[0].Value).To(Equal(5.0))
			Expect(queue.Values[0].Labels["pod"]).To(Equal("epp-pod-1"))

		})

		It("should handle unreachable pods gracefully", func() {
			// Create pod with invalid IP
			unreachablePod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "epp-pod-unreachable",
					Namespace: "test-ns",
					Labels: map[string]string{
						"inferencepool": "test-pool-epp",
					},
				},
				Status: corev1.PodStatus{
					PodIP: "192.0.2.1", // Invalid/unreachable IP (TEST-NET-1)
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			}

			client := fakeClient.
				WithObjects(service, secret, unreachablePod).
				Build()

			config := PodScrapingSourceConfig{
				ServiceName:      "test-pool-epp",
				ServiceNamespace: "test-ns",
				MetricsPort:      9090,
				ScrapeTimeout:    1 * time.Second, // Short timeout
			}
			source, err := NewPodScrapingSource(ctx, client, config)
			Expect(err).NotTo(HaveOccurred())

			// Pre-register metrics
			sourceRegistry := sourcepkg.NewSourceRegistry()
			sourceRegistry.MustRegister("test-pod-source", source)
			registration.RegisterPodScrapingMetrics("test-pod-source", sourceRegistry)

			// Should return empty results (not error) when pods are unreachable
			results, err := source.Refresh(ctx, sourcepkg.RefreshSpec{})
			Expect(err).NotTo(HaveOccurred())
			// Should have empty or no metrics due to unreachable pod
			Expect(results).To(BeEmpty())
		})

		It("should handle authentication failures", func() {
			// Create server that requires auth but we'll provide wrong token
			authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				auth := r.Header.Get("Authorization")
				if auth != "Bearer test-token" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer authServer.Close()

			var port int32
			if _, err := fmt.Sscanf(authServer.URL, "http://127.0.0.1:%d", &port); err != nil {
				Fail(fmt.Sprintf("failed to parse port from auth server URL %q: %v", authServer.URL, err))
			}

			authPod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "epp-pod-auth",
					Namespace: "test-ns",
					Labels: map[string]string{
						"inferencepool": "test-pool-epp",
					},
				},
				Status: corev1.PodStatus{
					PodIP: "127.0.0.1",
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			}

			// Use wrong token
			wrongSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "inference-gateway-sa-metrics-reader-secret",
					Namespace: "test-ns",
				},
				Data: map[string][]byte{
					"token": []byte("wrong-token"),
				},
			}

			client := fakeClient.
				WithObjects(service, wrongSecret, authPod).
				Build()

			config := PodScrapingSourceConfig{
				ServiceName:             "test-pool-epp",
				ServiceNamespace:        "test-ns",
				MetricsPort:             port,
				ScrapeTimeout:           1 * time.Second,
				MetricsReaderSecretName: "inference-gateway-sa-metrics-reader-secret",
			}
			source, err := NewPodScrapingSource(ctx, client, config)
			Expect(err).NotTo(HaveOccurred())

			// Pre-register metrics using centralized registration
			sourceRegistry := sourcepkg.NewSourceRegistry()
			sourceRegistry.MustRegister("test-pod-source", source)
			registration.RegisterPodScrapingMetrics("test-pod-source", sourceRegistry)

			// Should handle auth failure gracefully (empty results, not error)
			results, err := source.Refresh(ctx, sourcepkg.RefreshSpec{})
			Expect(err).NotTo(HaveOccurred())
			// Should be empty due to auth failure
			Expect(results).To(BeEmpty())
		})
	})

	Describe("Get", func() {
		var source *PodScrapingSource

		BeforeEach(func() {
			config := PodScrapingSourceConfig{
				ServiceName:      "test-pool-epp",
				ServiceNamespace: "test-ns",
				MetricsPort:      9090,
				DefaultTTL:       1 * time.Hour,
			}
			var err error
			source, err = NewPodScrapingSource(ctx, fakeClient.Build(), config)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return nil for uncached query", func() {
			cached := source.Get(constants.VLLMKvCacheUsagePerc, nil)
			Expect(cached).To(BeNil())
		})

		It("should return cached value if fresh", func() {
			// Manually set cache
			metricName := constants.VLLMKvCacheUsagePerc
			cacheKey := sourcepkg.BuildCacheKey(metricName, nil)
			result := sourcepkg.MetricResult{
				QueryName:   metricName,
				Values:      []sourcepkg.MetricValue{{Value: 0.75}},
				CollectedAt: time.Now(),
			}
			source.cache.Set(cacheKey, result, 1*time.Hour)

			cached := source.Get(metricName, nil)
			Expect(cached).NotTo(BeNil())
			Expect(cached.Result.QueryName).To(Equal(metricName))
			Expect(cached.Result.Values).To(HaveLen(1))
		})

		It("should return nil for expired cache", func() {
			// Manually set cache with short TTL and wait for expiration
			metricName := constants.VLLMKvCacheUsagePerc
			cacheKey := sourcepkg.BuildCacheKey(metricName, nil)
			result := sourcepkg.MetricResult{
				QueryName:   metricName,
				Values:      []sourcepkg.MetricValue{{Value: 0.75}},
				CollectedAt: time.Now(),
			}
			source.cache.Set(cacheKey, result, 100*time.Millisecond)

			// Wait for cache to expire
			time.Sleep(150 * time.Millisecond)

			cached := source.Get(metricName, nil)
			Expect(cached).To(BeNil())
		})

		Context("retrieving individual metrics after Refresh", func() {
			var (
				mockServer1 *httptest.Server
				mockServer2 *httptest.Server
				readyPod1   *corev1.Pod
				readyPod2   *corev1.Pod
				service     *corev1.Service
			)

			BeforeEach(func() {
				// Setup mock servers for two pods
				mockServer1 = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`# HELP vllm:kv_cache_usage_perc KV cache usage percentage
# TYPE vllm:kv_cache_usage_perc gauge
vllm:kv_cache_usage_perc 0.75
# HELP vllm:num_requests_waiting Number of requests waiting
# TYPE vllm:num_requests_waiting gauge
vllm:num_requests_waiting 5
`))
				}))

				mockServer2 = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`# HELP vllm:kv_cache_usage_perc KV cache usage percentage
# TYPE vllm:kv_cache_usage_perc gauge
vllm:kv_cache_usage_perc 0.85
# HELP vllm:num_requests_waiting Number of requests waiting
# TYPE vllm:num_requests_waiting gauge
vllm:num_requests_waiting 3
`))
				}))

				var port1, port2 int32
				if _, err := fmt.Sscanf(mockServer1.URL, "http://127.0.0.1:%d", &port1); err != nil {
					Fail(fmt.Sprintf("failed to parse port: %v", err))
				}
				if _, err := fmt.Sscanf(mockServer2.URL, "http://127.0.0.1:%d", &port2); err != nil {
					Fail(fmt.Sprintf("failed to parse port: %v", err))
				}

				service = &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pool-epp",
						Namespace: "test-ns",
					},
					Spec: corev1.ServiceSpec{
						Selector: map[string]string{"app": "test"},
					},
				}

				readyPod1 = &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "epp-pod-1",
						Namespace: "test-ns",
						Labels:    map[string]string{"app": "test"},
					},
					Status: corev1.PodStatus{
						PodIP: "127.0.0.1:" + fmt.Sprint(port1),
						Conditions: []corev1.PodCondition{
							{Type: corev1.PodReady, Status: corev1.ConditionTrue},
						},
					},
				}
				readyPod1.Status.PodIP = "127.0.0.1"

				readyPod2 = &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "epp-pod-2",
						Namespace: "test-ns",
						Labels:    map[string]string{"app": "test"},
					},
					Status: corev1.PodStatus{
						PodIP: "127.0.0.1",
						Conditions: []corev1.PodCondition{
							{Type: corev1.PodReady, Status: corev1.ConditionTrue},
						},
					},
				}

				client := buildClient().WithObjects(service, readyPod1, readyPod2).Build()
				config := PodScrapingSourceConfig{
					ServiceName:      "test-pool-epp",
					ServiceNamespace: "test-ns",
					MetricsPort:      port1,
					MetricsScheme:    "http",
					MetricsPath:      "/metrics",
				}
				var err error
				source, err = NewPodScrapingSource(ctx, client, config)
				Expect(err).NotTo(HaveOccurred())

				// Pre-register expected metrics using centralized registration
				sourceRegistry := sourcepkg.NewSourceRegistry()
				sourceRegistry.MustRegister("test-pod-source", source)
				registration.RegisterPodScrapingMetrics("test-pod-source", sourceRegistry)

				// Point pods to different servers
				readyPod1.Status.PodIP = "127.0.0.1"
				readyPod2.Status.PodIP = "127.0.0.1"
			})

			AfterEach(func() {
				if mockServer1 != nil {
					mockServer1.Close()
				}
				if mockServer2 != nil {
					mockServer2.Close()
				}
			})

			It("should cache each metric individually after Refresh", func() {
				// Manually inject scraped data (simulating successful scrape)
				now := time.Now()

				// Cache metrics as Refresh() would
				kvCacheMetric := sourcepkg.MetricResult{
					QueryName: constants.VLLMKvCacheUsagePerc,
					Values: []sourcepkg.MetricValue{
						{Value: 0.75, Timestamp: now, Labels: map[string]string{"pod": "epp-pod-1", "__name__": constants.VLLMKvCacheUsagePerc}},
						{Value: 0.85, Timestamp: now, Labels: map[string]string{"pod": "epp-pod-2", "__name__": constants.VLLMKvCacheUsagePerc}},
					},
					CollectedAt: now,
				}
				queueMetric := sourcepkg.MetricResult{
					QueryName: constants.VLLMNumRequestsWaiting,
					Values: []sourcepkg.MetricValue{
						{Value: 5.0, Timestamp: now, Labels: map[string]string{"pod": "epp-pod-1", "__name__": constants.VLLMNumRequestsWaiting}},
						{Value: 3.0, Timestamp: now, Labels: map[string]string{"pod": "epp-pod-2", "__name__": constants.VLLMNumRequestsWaiting}},
					},
					CollectedAt: now,
				}

				source.cache.Set(sourcepkg.BuildCacheKey(constants.VLLMKvCacheUsagePerc, nil), kvCacheMetric, 1*time.Hour)
				source.cache.Set(sourcepkg.BuildCacheKey(constants.VLLMNumRequestsWaiting, nil), queueMetric, 1*time.Hour)

				// Retrieve individual metrics
				kvCache := source.Get(constants.VLLMKvCacheUsagePerc, nil)
				Expect(kvCache).NotTo(BeNil())
				Expect(kvCache.Result.QueryName).To(Equal(constants.VLLMKvCacheUsagePerc))
				Expect(kvCache.Result.Values).To(HaveLen(2))

				queue := source.Get(constants.VLLMNumRequestsWaiting, nil)
				Expect(queue).NotTo(BeNil())
				Expect(queue.Result.QueryName).To(Equal(constants.VLLMNumRequestsWaiting))
				Expect(queue.Result.Values).To(HaveLen(2))
			})

			It("should return metrics from all pods for a specific metric name", func() {
				now := time.Now()
				metric := sourcepkg.MetricResult{
					QueryName: constants.VLLMKvCacheUsagePerc,
					Values: []sourcepkg.MetricValue{
						{Value: 0.75, Timestamp: now, Labels: map[string]string{"pod": "epp-pod-1", "__name__": constants.VLLMKvCacheUsagePerc}},
						{Value: 0.85, Timestamp: now, Labels: map[string]string{"pod": "epp-pod-2", "__name__": constants.VLLMKvCacheUsagePerc}},
					},
					CollectedAt: now,
				}
				source.cache.Set(sourcepkg.BuildCacheKey(constants.VLLMKvCacheUsagePerc, nil), metric, 1*time.Hour)
				cached := source.Get(constants.VLLMKvCacheUsagePerc, nil)
				Expect(cached).NotTo(BeNil())
				Expect(cached.Result.Values).To(HaveLen(2))

				// Verify both pods are present
				pods := make(map[string]float64)
				for _, value := range cached.Result.Values {
					pods[value.Labels["pod"]] = value.Value
				}
				Expect(pods).To(HaveKey("epp-pod-1"))
				Expect(pods).To(HaveKey("epp-pod-2"))
				Expect(pods["epp-pod-1"]).To(Equal(0.75))
				Expect(pods["epp-pod-2"]).To(Equal(0.85))
			})

			It("should return nil for non-existent metric", func() {
				cached := source.Get("nonexistent_metric", nil)
				Expect(cached).To(BeNil())
			})

			It("should preserve all labels for each metric", func() {
				now := time.Now()
				metric := sourcepkg.MetricResult{
					QueryName: constants.EPPInferencePoolAverageQueueSize,
					Values: []sourcepkg.MetricValue{
						{
							Value:     10.0,
							Timestamp: now,
							Labels: map[string]string{
								"__name__":   constants.EPPInferencePoolAverageQueueSize,
								"pod":        "epp-pod-1",
								"model_name": "llama-3-8b",
								"namespace":  "test-ns",
							},
						},
					},
					CollectedAt: now,
				}
				source.cache.Set(sourcepkg.BuildCacheKey(constants.EPPInferencePoolAverageQueueSize, nil), metric, 1*time.Hour)

				cached := source.Get(constants.EPPInferencePoolAverageQueueSize, nil)
				Expect(cached).NotTo(BeNil())
				Expect(cached.Result.Values).To(HaveLen(1))

				value := cached.Result.Values[0]
				Expect(value.Labels["pod"]).To(Equal("epp-pod-1"))
				Expect(value.Labels["model_name"]).To(Equal("llama-3-8b"))
				Expect(value.Labels["namespace"]).To(Equal("test-ns"))
			})
		})
	})

	Describe("QueryList", func() {
		It("should return empty registry initially", func() {
			config := PodScrapingSourceConfig{
				ServiceName:      "test-pool-epp",
				ServiceNamespace: "test-ns",
				MetricsPort:      9090,
			}
			source, err := NewPodScrapingSource(ctx, buildClient().Build(), config)
			Expect(err).NotTo(HaveOccurred())

			registry := source.QueryList()
			Expect(registry).NotTo(BeNil())

			// Initially empty - metrics must be pre-registered before use
			queries := registry.List()
			Expect(queries).To(BeEmpty())
		})

		It("should allow pre-registration of metrics", func() {
			config := PodScrapingSourceConfig{
				ServiceName:      "test-pool-epp",
				ServiceNamespace: "test-ns",
				MetricsPort:      9090,
			}
			source, err := NewPodScrapingSource(ctx, buildClient().Build(), config)
			Expect(err).NotTo(HaveOccurred())

			// Pre-register expected metrics using centralized registration
			sourceRegistry := sourcepkg.NewSourceRegistry()
			sourceRegistry.MustRegister("test-pod-source", source)
			registration.RegisterPodScrapingMetrics("test-pod-source", sourceRegistry)

			// Verify registration
			registry := source.QueryList()
			query := registry.Get(constants.VLLMKvCacheUsagePerc)
			Expect(query).NotTo(BeNil())
			Expect(query.Name).To(Equal(constants.VLLMKvCacheUsagePerc))
			Expect(query.Type).To(Equal(sourcepkg.QueryTypeMetricName))
		})
	})
})

// mockReader is a simple io.Reader implementation for testing
type mockReader struct {
	data []byte
	pos  int
}

func (m *mockReader) Read(p []byte) (n int, err error) {
	if m.pos >= len(m.data) {
		return 0, io.EOF
	}
	n = copy(p, m.data[m.pos:])
	m.pos += n
	return n, nil
}
