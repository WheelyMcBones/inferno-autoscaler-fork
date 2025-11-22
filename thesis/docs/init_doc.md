# Workload Variant Autoscaler: A Performance-Model-Based Approach to AI Inference Autoscaling

## Abstract

This document presents the Workload Variant Autoscaler (WVA), a Kubernetes-native autoscaling system designed for AI inference workloads. WVA employs analytical performance models derived from queueing theory to optimize resource allocation across heterogeneous accelerator types while satisfying Service Level Objectives (SLOs). The system operates in two modes: a reactive capacity-based approach for general applicability, and a predictive model-based approach for dense transformer architectures. This paper describes the theoretical foundations, architectural design, and operational mechanisms of the autoscaler.

## 1. Introduction

### 1.1 Motivation

Large Language Model (LLM) inference workloads present unique autoscaling challenges due to their heterogeneous resource requirements, variable request patterns, and strict latency constraints. Traditional Kubernetes autoscaling mechanisms, such as the Horizontal Pod Autoscaler (HPA), operate reactively on threshold-based metrics and lack awareness of model-specific performance characteristics and cost trade-offs across different accelerator types.

The Workload Variant Autoscaler addresses these limitations by introducing a global optimization framework that considers:

1. **Heterogeneous accelerator support**: Multiple GPU types with varying performance and cost characteristics
2. **Performance modeling**: Analytical models capturing prefill and decode latency as functions of batch size and request characteristics
3. **SLO-aware allocation**: Explicit time-to-first-token (TTFT) and inter-token-latency (ITL) targets
4. **Cost optimization**: Selection of accelerator types and replica counts to minimize infrastructure costs while meeting SLOs

### 1.2 System Overview

WVA is implemented as a Kubernetes controller that manages custom resources called `VariantAutoscaling`. Each variant represents a collection of inference server replicas serving a specific model on a particular accelerator configuration. The system operates as a global autoscaler, considering all variants holistically rather than making independent per-variant decisions.

The autoscaler supports two operational modes:

- **Capacity-based scaling** (default): Reactive scaling triggered by resource saturation (KV cache utilization, request queue length)
- **Model-based scaling** (optional): Predictive scaling using performance parameters to calculate optimal allocations

## 2. Core Concepts and Types

### 2.1 Fundamental Entities

#### 2.1.1 Variant

A **variant** is defined as a homogeneous collection of inference server replicas serving a specific model using a particular accelerator arrangement. Formally, a variant $V$ is characterized by:

$$V = (M, A, R)$$

where:
- $M$ denotes the model identifier
- $A$ represents the accelerator configuration (type, count, arrangement)
- $R$ is the set of replica instances

Each variant maintains both current and desired allocation states, where an allocation specifies:

```
Allocation = (accelerator, numReplicas, maxBatch, cost, performance_metrics)
```

#### 2.1.2 Accelerator

An **accelerator** $G$ represents a unit of GPU allocation, characterized by:

$$G = (name, type, multiplicity, cost, power\_profile)$$

The multiplicity parameter accounts for multi-GPU configurations required for large models. The cost parameter enables economic optimization across heterogeneous accelerator types.

#### 2.1.3 Model

A **model** $M$ encapsulates the inference characteristics of a specific LLM, including:

$$M = (name, \{P_{G_i}\}, \{N_{G_i}\})$$

where:
- $P_{G_i}$ is the performance data for accelerator $G_i$
- $N_{G_i}$ is the number of accelerator instances required to fit the model on $G_i$

Performance data includes parameters for analytical models describing prefill and decode latency.

#### 2.1.4 Service Class

A **service class** $S$ defines SLO requirements and workload priority:

$$S = (name, priority, \{T_M\})$$

where $T_M$ specifies target performance metrics for model $M$:

$$T_M = (TTFT_{target}, ITL_{target}, TPS_{target})$$

Here, $TTFT$ denotes time-to-first-token, $ITL$ denotes inter-token-latency, and $TPS$ denotes tokens-per-second throughput.

### 2.2 Performance Model Parameters

The system employs linear performance models calibrated through benchmarking:

**Decode Time (Inter-Token Latency)**:
$$ITL(b) = \alpha + \beta \cdot b$$

**Prefill Time**:
$$T_{prefill}(n_{in}, b) = \gamma + \delta \cdot n_{in} \cdot b$$

where:
- $b$ is the batch size
- $n_{in}$ is the number of input tokens
- $\alpha, \beta, \gamma, \delta$ are model-accelerator specific parameters

These linear relationships have been empirically validated across various LLM architectures and accelerator types in published research [1-5].

### 2.3 Workload Characterization

Request workload is characterized by:

$$W = (\lambda, n_{in}, n_{out})$$

where:
- $\lambda$ is the arrival rate (requests per minute)
- $n_{in}$ is the average input token count
- $n_{out}$ is the average output token count

## 3. Core Functionality

### 3.1 Architecture Overview

The WVA architecture comprises five primary components:

1. **Controller**: Kubernetes reconciliation loop managing VariantAutoscaling resources
2. **Metrics Collector**: Interface to Prometheus for gathering vLLM and system metrics
3. **Model Analyzer**: Queueing-theoretic performance analysis
4. **Optimizer**: Global allocation solver (greedy or unlimited capacity)
5. **Actuator**: Metrics emission for external autoscalers (HPA, KEDA)

The system follows a sense-analyze-decide-act loop:

```
┌─────────────┐
│ Prometheus  │ ← vLLM metrics
└──────┬──────┘
       │
       ↓
┌──────────────────┐
│ Metrics Collector│
└────────┬─────────┘
         │
         ↓
┌──────────────────────┐
│  Capacity Analyzer   │ ← Reactive approach
│  Model Analyzer      │ ← Predictive approach
└────────┬─────────────┘
         │
         ↓
┌──────────────────┐
│    Optimizer     │
└────────┬─────────┘
         │
         ↓
┌──────────────────┐
│    Actuator      │ → Emit metrics for HPA/KEDA
└──────────────────┘
```

### 3.2 Queueing Model

The model analyzer employs an M/M/1/K queueing model with state-dependent service rates to represent inference server behavior. The system models each replica as a queue with:

- **Arrival process**: Poisson with rate $\lambda$
- **Service process**: State-dependent rates $\mu(n)$ where $n$ is the batch size
- **Capacity**: Maximum occupancy $K = B_{max} + Q_{max}$
  - $B_{max}$: maximum batch size
  - $Q_{max}$: maximum queue length

#### 3.2.1 Service Rate Calculation

The state-dependent service rate for batch size $b$ is derived from the performance model:

$$\mu(b) = \frac{b}{T_{prefill}(n_{in}, b) + (n_{out} - 1) \cdot ITL(b)}$$

This formulation accounts for one prefill operation per request and $(n_{out} - 1)$ decode iterations to generate subsequent tokens.

#### 3.2.2 Performance Metrics

The queueing model computes:

1. **Average queue length**: $L_q$
2. **Average waiting time**: $W_q$
3. **Average service time**: $W_s$
4. **Average response time**: $T = W_q + W_s$
5. **Utilization**: $\rho = E[N_{service}] / B_{max}$

where $N_{service}$ denotes the number of requests in service.

The effective concurrency (average batch size) is calculated by solving:

$$T_{prefill}(n_{in}, \bar{b}) + (n_{out} - 1) \cdot ITL(\bar{b}) = W_s$$

for $\bar{b}$.

### 3.3 Optimization Formulation

#### 3.3.1 Unlimited Capacity Mode

In unlimited mode (default), the optimizer solves independent subproblems for each variant:

$$\min_{G, R} \, cost(G, R)$$

subject to:
$$TTFT(G, R, W) \leq TTFT_{target}$$
$$ITL(G, R, W) \leq ITL_{target}$$
$$R \geq R_{min}$$

where:
- $G$ is the selected accelerator type
- $R$ is the number of replicas
- $cost(G, R) = C_G \cdot N_G \cdot R$ with $C_G$ being the cost per accelerator and $N_G$ the instance count
- $TTFT(G, R, W)$ and $ITL(G, R, W)$ are estimated using the queueing model

The system evaluates all feasible $(G, R)$ pairs and selects the minimum-cost allocation satisfying SLOs.

#### 3.3.2 Limited Capacity Mode

Limited mode (future work) adds global capacity constraints:

$$\sum_{i} R_i \cdot N_{G_i} \cdot m_{G_i} \leq C_{type(G_i)} \quad \forall \text{ accelerator types}$$

where:
- $i$ indexes variants
- $m_{G_i}$ is the multiplicity of accelerator $G_i$
- $C_{type}$ is the available capacity of accelerator type $type$

A greedy algorithm allocates resources by priority, with optional saturation policies for best-effort allocation beyond SLO requirements.

### 3.4 Capacity-Based Scaling

The capacity analyzer provides a reactive alternative to model-based optimization, suitable for all model architectures. It monitors two saturation indicators:

1. **KV cache utilization**: $u_{kv} \in [0, 1]$
2. **Request queue length**: $q \in \mathbb{Z}_{\geq 0}$

#### 3.4.1 Saturation Detection

A replica is saturated if:

$$u_{kv} \geq \theta_{kv} \quad \text{or} \quad q \geq \theta_q$$

where $\theta_{kv}$ and $\theta_q$ are configurable thresholds (defaults: 0.80, 5).

#### 3.4.2 Spare Capacity Calculation

For non-saturated replicas, spare capacity is:

$$s_{kv,i} = \theta_{kv} - u_{kv,i}$$
$$s_{q,i} = \theta_q - q_i$$

Average spare capacity across $N$ non-saturated replicas:

$$\bar{s}_{kv} = \frac{1}{N}\sum_{i=1}^N s_{kv,i}$$
$$\bar{s}_q = \frac{1}{N}\sum_{i=1}^N s_{q,i}$$

#### 3.4.3 Scaling Decisions

**Scale-up** is triggered if:

$$\bar{s}_{kv} < \epsilon_{kv} \quad \text{or} \quad \bar{s}_q < \epsilon_q$$

where $\epsilon_{kv}$ and $\epsilon_q$ are spare capacity triggers (defaults: 0.10, 3).

**Scale-down** is permitted only if simulation shows adequate headroom after redistributing load across remaining replicas:

$$\frac{L_{total}}{N-1} + s_{threshold} < \theta$$

where $L_{total}$ is the total load across non-saturated replicas.

The capacity analyzer employs cost-based variant selection:
- Scale-up: Select cheapest variant
- Scale-down: Select most expensive variant

## 4. Autoscaling Operation

### 4.1 Reconciliation Loop

The controller operates with a configurable reconciliation interval (default: 60 seconds). Each reconciliation cycle:

1. **List active VariantAutoscaling resources**
2. **Group variants by model** for cross-variant optimization
3. **Collect metrics** from Prometheus
4. **Validate metrics availability** and update status conditions
5. **Execute scaling analysis** (capacity-based and/or model-based)
6. **Arbitrate final decisions** if both analyses available
7. **Update status** with desired allocations
8. **Emit metrics** for external autoscalers

### 4.2 Metrics Collection

The metrics collector queries Prometheus for vLLM inference server metrics:

**Utilization Metrics**:
- `vllm:kv_cache_usage_perc`: KV cache utilization ratio
- `vllm:num_requests_waiting`: Queue depth

**Performance Metrics**:
- `vllm:request_success_total`: Request throughput
- `vllm:time_to_first_token_seconds_{sum,count}`: TTFT statistics
- `vllm:time_per_output_token_seconds_{sum,count}`: ITL statistics
- `vllm:request_prompt_tokens_{sum,count}`: Input token statistics
- `vllm:request_generation_tokens_{sum,count}`: Output token statistics

Metrics are queried with `max_over_time[1m]` aggregation to capture peak utilization for conservative capacity analysis.

### 4.3 Hybrid Decision Architecture

When both capacity and model-based analyses are available, the system applies a hybrid decision matrix:

| Capacity Decision | Model-Based Decision | Final Decision | Rationale |
|-------------------|---------------------|----------------|-----------|
| Scale-up | Scale-down | **Capacity veto** (no change or scale-up) | Safety: capacity indicates saturation |
| No change | Scale-down | **Safety block** if unsafe | Prevent premature scale-down |
| Scale-up | No change | **Scale-up** | Capacity-driven action |
| No change | Scale-up | **Scale-up** | Model-based optimization |
| Agree | Agree | **Follow consensus** | Both analyses aligned |

This arbitration ensures capacity safety overrides dominate, preventing SLO violations.

### 4.4 Metrics Emission

The actuator emits Prometheus metrics consumed by HPA or KEDA:

- `inferno_current_replicas`: Current replica count per variant
- `inferno_desired_replicas`: Optimized target replica count
- `inferno_desired_ratio`: Ratio of desired to current replicas

External autoscalers consume these metrics to actuate scaling decisions. The system operates as a decision engine, delegating physical pod scaling to standard Kubernetes mechanisms.

### 4.5 Graceful Degradation

When metrics are unavailable (e.g., ServiceMonitor misconfiguration), the system:

1. Sets `MetricsAvailable` condition to `False` with diagnostic messages
2. Skips optimization for affected variants
3. Maintains current replica counts (no disruptive changes)
4. Emits safety-net metrics using previous desired replicas or current state
5. Continues monitoring other variants with valid metrics

Status conditions provide observability:

```yaml
status:
  conditions:
  - type: MetricsAvailable
    status: "False"
    reason: MetricsMissing
    message: "No vLLM metrics found. Ensure ServiceMonitor is configured."
  - type: OptimizationReady
    status: "False"
    reason: MetricsUnavailable
    message: "Cannot optimize without metrics."
```

### 4.6 Configuration Management

The system supports ConfigMap-based configuration with:

1. **Global defaults**: Applied to all variants
2. **Per-model overrides**: Match by `model_id` and `namespace`
3. **Automatic reload**: ConfigMap watch triggers cache updates and reconciliation

Configuration includes capacity thresholds, spare capacity triggers, and optimization intervals. Changes take effect immediately without pod restart.

## 5. Architectural Considerations

### 5.1 Model Architecture Compatibility

The performance model assumes linear batch size scaling, validated for **dense transformer architectures** (e.g., Llama, GPT variants). For specialized architectures:

- **Hybrid State Space Models (HSSM)**: State space dynamics may violate linearity assumptions
- **Mixture of Experts (MoE)**: Sparse activation and dynamic routing affect batch efficiency
- **Custom architectures**: Performance characteristics may deviate from established scaling laws

For non-standard architectures, **capacity-based scaling** (default mode) provides architecture-agnostic reactive autoscaling without requiring performance parameters.

### 5.2 Operational Modes

The system supports three operational modes:

1. **Capacity-only** (default, `EXPERIMENTAL_HYBRID_OPTIMIZATION=off`): Reactive scaling based on observed saturation
2. **Model-only** (`EXPERIMENTAL_HYBRID_OPTIMIZATION=model-only`): Predictive scaling using performance models
3. **Hybrid** (`EXPERIMENTAL_HYBRID_OPTIMIZATION=on`): Combined approach with capacity safety overrides

The default capacity-only mode ensures broad applicability while hybrid mode optimizes cost-efficiency for well-characterized workloads.

### 5.3 Integration Architecture

WVA integrates with the Kubernetes ecosystem through:

- **Custom Resource Definitions**: Native Kubernetes API for variant configuration
- **Prometheus**: Standardized metrics collection and query interface
- **HPA/KEDA**: Metric emission for autoscaler consumption
- **Controller Runtime**: Leverages established patterns for reconciliation and caching

This design promotes composability and allows WVA to operate alongside existing autoscaling infrastructure.

## 6. Related Work

The WVA approach synthesizes concepts from several research areas:

1. **Performance modeling**: Linear ITL-batch size relationships [1-5]
2. **Queueing theory**: M/M/1/K models for request processing [6]
3. **Resource allocation**: Cost-aware optimization with SLO constraints [7]
4. **Disaggregated serving**: Prefill-decode separation [1]

The system extends prior work by providing a production-ready implementation integrating analytical models with reactive capacity monitoring in a Kubernetes-native framework.

## 7. Conclusion

The Workload Variant Autoscaler provides a sophisticated autoscaling solution for AI inference workloads, combining analytical performance models with reactive capacity monitoring. The hybrid architecture ensures both cost-efficiency through predictive optimization and safety through capacity-based guardrails.

Key contributions include:

1. **Dual-mode operation**: Capacity-based (default, architecture-agnostic) and model-based (optimized for dense transformers) scaling
2. **Global optimization**: Holistic resource allocation across heterogeneous accelerator types
3. **SLO-aware allocation**: Explicit TTFT and ITL targets with queueing-theoretic validation
4. **Production-ready design**: Kubernetes-native implementation with graceful degradation and comprehensive observability

Future work includes limited capacity mode with degraded-mode handling, enhanced model architectures beyond dense transformers, and integration with inference scheduler thresholds.

## References

[1] Agrawal, A., et al. "Taming Throughput-Latency tradeoff in LLM inference with Sarathi-Serve." OSDI 2024.

[2] Griggs, T., et al. "Mélange: Cost efficient large language model serving by exploiting gpu heterogeneity." arXiv:2404.14527, 2024.

[3] Yang, Y., et al. "A queueing theoretic perspective on low-latency llm inference with variable token length." WiOpt 2024.

[4] Yuan, Z., et al. "LLM inference unveiled: Survey and roofline model insights." arXiv:2402.16363, 2024.

[5] Zhu, K., et al. "PolyServe: Efficient Multi-SLO Serving at Scale." arXiv:2507.17769, 2025.

[6] Casson, A. "Transformer FLOPs." adamcasson.com, 2023.

[7] WVA Design Documentation. "Modeling and Optimization." Internal documentation, 2025.