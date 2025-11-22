# Dynamic Parameter Estimation for LLM Inference Server Autoscaling: A Kalman Filter Approach

## Executive Summary

This document provides a comprehensive theoretical and practical foundation for a system that integrates Extended Kalman Filtering with queueing-theoretic models to enable dynamic autoscaling of Large Language Model (LLM) inference servers. The system addresses the challenge of allocating GPU resources to meet Service Level Objectives (SLOs) while minimizing cost, in the face of uncertain and time-varying system parameters.

---

## 1. Introduction and Motivation

### 1.1 The LLM Inference Challenge

Large Language Model inference presents unique operational challenges:

1. **Two-phase processing**: LLM inference consists of two distinct phases:
   - **Prefill phase**: Processing input tokens and generating the first output token
   - **Decode phase**: Iteratively generating subsequent output tokens

2. **Resource intensity**: Each phase has different computational characteristics and resource requirements

3. **Batching dynamics**: Modern inference engines (like vLLM) use continuous batching with state-dependent service rates

4. **Performance metrics**: Users care about:
   - **TTFT** (Time to First Token): Latency from request submission to first token
   - **ITL** (Inter-Token Latency): Time between consecutive output tokens
   - **TPS** (Tokens Per Second): Throughput measure

### 1.2 The Autoscaling Problem

The autoscaling system must:

- Allocate accelerators (GPUs) to inference servers
- Meet SLO targets for TTFT, ITL, and TPS
- Minimize operational cost
- Handle multiple service classes with different priorities
- Adapt to changing workload patterns

The key challenge: **Queueing model parameters are unknown and time-varying**.

---

## 2. Queueing Theory Foundations

### 2.1 Classical M/M/1/K Queue

The system uses an **M/M/1/K** queueing model as its foundation:

**M/M/1/K Characteristics:**

- **M** (Markovian): Exponential inter-arrival times with rate λ
- **M** (Markovian): Exponential service times with rate μ  
- **1**: Single server
- **K**: Finite system capacity (K customers maximum)

**State probabilities** for M/M/1/K (from `mm1kmodel.go`):

For ρ = λ/μ ≠ 1:
$$p_0 = \frac{1-\rho}{1-\rho^{K+1}}$$

For ρ = 1:
$$p_0 = \frac{1}{K+1}$$

Then for i = 1, 2, ..., K:
$$p_i = p_0 \rho^i$$

**Key performance metrics:**

1. **Average number in system:**
$$L = \sum_{i=0}^{K} i \cdot p_i$$

2. **Effective throughput** (accounts for blocking):
$$\lambda_{eff} = \lambda(1 - p_K)$$

3. **Average response time** (Little's Law):
$$W = \frac{L}{\lambda_{eff}}$$

4. **Average waiting time:**
$$W_q = W - \frac{1}{\mu}$$

5. **Average queue length:**
$$L_q = \lambda_{eff} \cdot W_q$$

### 2.2 State-Dependent Service Rates

LLM inference servers exhibit **state-dependent service rates** due to batching. The service rate depends on the number of requests being processed concurrently.

From `mm1modelstatedependent.go`, the model uses:

- Service rate vector: μ(n) for n = 1, 2, ..., N (where N is max batch size)
- State n represents having n requests in the system

**State probability recursion:**
$$p_{n+1} = p_n \cdot \frac{\lambda}{\mu(n+1)}$$

With normalization:
$$\sum_{n=0}^{K} p_n = 1$$

**Utilization** (proportion of time serving):
$$\rho = 1 - p_0$$

### 2.3 LLM-Specific Service Rate Model

For LLM inference, the service rate for batch size n is modeled as:

$$\mu(n) = \frac{n}{T_{prefill}(n) + (m-1) \cdot T_{decode}(n)}$$

Where:
- n: current batch size
- m: average number of output tokens per request
- $T_{prefill}(n)$: Prefill time for batch size n
- $T_{decode}(n)$: Per-token decode time for batch size n

**Prefill time model** (from `queueanalyzer.go`):
$$T_{prefill}(n) = \gamma + \delta \cdot k \cdot n$$

Where:
- γ (gamma): Base prefill overhead
- δ (delta): Per-token-per-batch scaling factor
- k: average number of input tokens
- n: batch size

**Decode time model:**
$$T_{decode}(n) = \alpha + \beta \cdot n$$

Where:
- α (alpha): Base decode overhead per token
- β (beta): Per-request scaling factor
- n: batch size

**These parameters (α, β, γ, δ) are what the Kalman Filter estimates**.

### 2.4 Performance Metric Calculations

From the queueing model, we derive SLO-relevant metrics:

**TTFT (Time to First Token):**
$$TTFT = W_q + T_{prefill}(n_{eff})$$

Where $n_{eff}$ is the effective concurrency level.

**ITL (Inter-Token Latency):**
$$ITL = T_{decode}(n_{eff})$$

**TPS (Tokens Per Second):**
$$TPS = \lambda_{eff} \cdot m$$

Where m is the average number of output tokens per request.

**Effective concurrency calculation** (from `queueanalyzer.go`):

Given average service time $\bar{S}$, find n such that:
$$T_{prefill}(n) + (m-1) \cdot T_{decode}(n) = \bar{S}$$

Solving:
$$n = \frac{\bar{S} - (\gamma + \alpha(m-1))}{\delta k + \beta(m-1)}$$

Bounded by: $n \in [0, N_{max}]$

---

## 3. Kalman Filtering Theory

### 3.1 The Standard Kalman Filter

The Kalman Filter provides optimal state estimation for linear systems with Gaussian noise.

**Note**: The implementation uses the `github.com/llm-inferno/kalman-filter` Go package, which provides the Extended Kalman Filter implementation with numerical Jacobian computation.

**System model:**
$$\mathbf{x}_{k} = \mathbf{F}_{k-1}\mathbf{x}_{k-1} + \mathbf{w}_{k-1}$$
$$\mathbf{z}_{k} = \mathbf{H}_{k}\mathbf{x}_{k} + \mathbf{v}_{k}$$

Where:
- $\mathbf{x}_k \in \mathbb{R}^n$: State vector at time k
- $\mathbf{F}_{k-1}$: State transition matrix
- $\mathbf{w}_{k-1} \sim \mathcal{N}(0, \mathbf{Q}_{k-1})$: Process noise
- $\mathbf{z}_k \in \mathbb{R}^m$: Measurement vector
- $\mathbf{H}_k$: Measurement matrix
- $\mathbf{v}_k \sim \mathcal{N}(0, \mathbf{R}_k)$: Measurement noise

**Kalman Filter Algorithm** (from `basickf.go`):

**Prediction Step:**
$$\hat{\mathbf{x}}_{k|k-1} = \mathbf{F}_{k-1}\hat{\mathbf{x}}_{k-1|k-1}$$
$$\mathbf{P}_{k|k-1} = \mathbf{F}_{k-1}\mathbf{P}_{k-1|k-1}\mathbf{F}_{k-1}^T + \mathbf{Q}_{k-1}$$

**Update Step:**

1. Innovation:
$$\mathbf{y}_k = \mathbf{z}_k - \mathbf{H}_k\hat{\mathbf{x}}_{k|k-1}$$

2. Innovation covariance:
$$\mathbf{S}_k = \mathbf{H}_k\mathbf{P}_{k|k-1}\mathbf{H}_k^T + \mathbf{R}_k$$

3. Kalman gain:
$$\mathbf{K}_k = \mathbf{P}_{k|k-1}\mathbf{H}_k^T\mathbf{S}_k^{-1}$$

4. State update:
$$\hat{\mathbf{x}}_{k|k} = \hat{\mathbf{x}}_{k|k-1} + \mathbf{K}_k\mathbf{y}_k$$

5. Covariance update:
$$\mathbf{P}_{k|k} = (\mathbf{I} - \mathbf{K}_k\mathbf{H}_k)\mathbf{P}_{k|k-1}$$

### 3.2 Extended Kalman Filter (EKF)

The EKF extends Kalman filtering to **nonlinear systems**, which is necessary for our queueing model.

**Nonlinear system model:**
$$\mathbf{x}_{k} = f(\mathbf{x}_{k-1}) + \mathbf{w}_{k-1}$$
$$\mathbf{z}_{k} = h(\mathbf{x}_{k}) + \mathbf{v}_{k}$$

Where f and h are nonlinear functions.

**EKF Algorithm** (from `extendedkf.go`):

**Prediction Step:**
$$\hat{\mathbf{x}}_{k|k-1} = f(\hat{\mathbf{x}}_{k-1|k-1})$$
$$\mathbf{F}_k = \left.\frac{\partial f}{\partial \mathbf{x}}\right|_{\mathbf{x}=\hat{\mathbf{x}}_{k-1|k-1}}$$
$$\mathbf{P}_{k|k-1} = \mathbf{F}_k\mathbf{P}_{k-1|k-1}\mathbf{F}_k^T + \mathbf{Q}_{k-1}$$

**Update Step:**
$$\mathbf{y}_k = \mathbf{z}_k - h(\hat{\mathbf{x}}_{k|k-1})$$
$$\mathbf{H}_k = \left.\frac{\partial h}{\partial \mathbf{x}}\right|_{\mathbf{x}=\hat{\mathbf{x}}_{k|k-1}}$$
$$\mathbf{S}_k = \mathbf{H}_k\mathbf{P}_{k|k-1}\mathbf{H}_k^T + \mathbf{R}_k$$
$$\mathbf{K}_k = \mathbf{P}_{k|k-1}\mathbf{H}_k^T\mathbf{S}_k^{-1}$$
$$\hat{\mathbf{x}}_{k|k} = \hat{\mathbf{x}}_{k|k-1} + \mathbf{K}_k\mathbf{y}_k$$
$$\mathbf{P}_{k|k} = (\mathbf{I} - \mathbf{K}_k\mathbf{H}_k)\mathbf{P}_{k|k-1}$$

### 3.3 Jacobian Computation

For the EKF, Jacobians are computed numerically using **symmetric finite differences** (from `numerical_jacobian.go`):

$$\frac{\partial f_i}{\partial x_j} \approx \frac{f_i(x + \epsilon e_j) - f_i(x - \epsilon e_j)}{2\epsilon}$$

Where:
- $\epsilon = \delta \cdot |x_j|$ if $x_j \neq 0$
- $\epsilon = \delta$ if $x_j = 0$
- $\delta$ is a small constant (default 0.01)
- $e_j$ is the j-th unit vector

This provides second-order accuracy: $O(\epsilon^2)$.

---

## 4. Model Tuner: Kalman Filter Application

### 4.0 Tuner Architecture

The **Tuner** (from `tuner.go`) encapsulates the complete parameter estimation workflow:

**Key Components:**

- **Configurator**: Manages filter initialization, noise covariances, and constraints
- **Filter**: Extended Kalman Filter instance from the `kalman-filter` package
- **Environment**: Current workload and observation data
- **Stasher**: State backup/restore mechanism for validation rollback

**Tuner Workflow:**

1. **Run()**: Executes one tuning cycle:
   - Stash current state
   - Predict step (time update)
   - Update step (measurement update)
   - Extract tuned parameters
   - Validate results
   - If validation fails, unstash (rollback)

2. **TunedResults**: Output structure containing:
   - ServiceParms: Updated (α, β, γ, δ) parameters
   - Innovation: Measurement residual vector
   - Covariance: Posterior estimate uncertainty

3. **UpdateEnvironment()**: Allows updating workload conditions between tuning cycles

### 4.1 State Vector Definition

The **state vector** represents the queueing model parameters:

$$\mathbf{x} = \begin{bmatrix} \alpha \\ \beta \\ \delta \\ \gamma \end{bmatrix}$$

Where:
- α (index 0): Base decode time per token (milliseconds)
- β (index 1): Decode time scaling factor with batch size (milliseconds)
- δ (index 2): Prefill per-token-per-batch scaling factor (milliseconds)
- γ (index 3): Base prefill overhead (milliseconds)

Note: The current implementation estimates all four parameters (α, β, γ, δ) simultaneously, enabling comprehensive model adaptation.

### 4.2 State Transition Model

The model assumes **constant parameters** (random walk):

$$f(\mathbf{x}_k) = \mathbf{x}_k$$

In matrix form:
$$\mathbf{F} = \mathbf{I}$$

This reflects that we expect α and β to change slowly over time.

**Process noise covariance** (from `configurator.go`):

$$\mathbf{Q} = \text{diag}\left[(c_0 \cdot \alpha)^2, (c_1 \cdot \beta)^2, (c_2 \cdot \delta)^2, (c_3 \cdot \gamma)^2\right]$$

Where $c_i$ are **percent change** parameters (typically 5%, from `tuner_defaults.go`), representing expected parameter variation.

### 4.3 Observation Model

The **observation vector** consists of measured performance metrics:

$$\mathbf{z} = \begin{bmatrix} TTFT^{obs} \\ ITL^{obs} \end{bmatrix}$$

Where:
- $TTFT^{obs}$: Observed Time to First Token (average queueing time + prefill time, in milliseconds)
- $ITL^{obs}$: Observed Inter-Token Latency (average token generation time, in milliseconds)

**Observation function** $h(\mathbf{x})$ (from `tuner.go`):

Given current state $\mathbf{x} = [\alpha, \beta, \delta, \gamma]^T$ and environment conditions:

- λ: Arrival rate (requests/minute)
- k: Average input tokens per request
- m: Average output tokens per request  
- N: Max batch size

1. Construct queueing model configuration using current state estimates:

$$\text{PrefillParms} = \{\gamma = x_3, \delta = x_2\}$$
$$\text{DecodeParms} = \{\alpha = x_0, \beta = x_1\}$$

2. Compute state-dependent service rates:

$$\mu(n) = \frac{n}{\gamma + \delta \cdot k \cdot n + (m-1)(\alpha + \beta \cdot n)}$$

for n = 1, 2, ..., N

3. Solve M/M/1/K state-dependent queue:
   - Convert arrival rate: $\lambda_{sec} = \lambda / 60$ (requests/second)
   - Compute state probabilities $p_n$
   - Calculate average waiting time $W_q$
   - Calculate average prefill time $T_{prefill}$
   - Calculate average token generation time $T_{token}$

4. Return predictions:

$$h(\mathbf{x}) = \begin{bmatrix} W_q + T_{prefill} \\ T_{token} \end{bmatrix} = \begin{bmatrix} TTFT \\ ITL \end{bmatrix}$$

This is a **nonlinear function** of all four parameters (α, β, γ, δ), requiring the EKF.

### 4.4 Measurement Noise Covariance

The **measurement noise covariance** is estimated from expected measurement accuracy (from `configurator.go`):

$$\mathbf{R} = \text{diag}\left[\sigma_{TTFT}^2, \sigma_{ITL}^2\right]$$

Where:
$$\sigma_i^2 = \left(\frac{\epsilon}{t_p}\right)^2 \frac{z_i^2}{\gamma_{factor}}$$

Parameters (from `tuner_defaults.go`):

- $\epsilon$: Error level (0.05 = 5%)
- $t_p$: Student-t percentile value (1.96 for 95% confidence)
- $\gamma_{factor}$: Gamma factor (1.0)
- $z_i$: Expected observation value

This models measurement uncertainty that scales with the magnitude of the observation.

### 4.5 State Constraints

The filter supports **bounded state estimation** to ensure physical validity:

$$\alpha_{min} \leq \alpha \leq \alpha_{max}$$
$$\beta_{min} \leq \beta \leq \beta_{max}$$
$$\gamma_{min} \leq \gamma \leq \gamma_{max}$$
$$\delta_{min} \leq \delta \leq \delta_{max}$$

After each update, states are clamped:

$$\mathbf{x}_{k|k} = \max(\mathbf{x}_{min}, \min(\mathbf{x}_{k|k}, \mathbf{x}_{max}))$$

This prevents the filter from estimating physically impossible parameter values (e.g., negative service times).

### 4.6 Tuner Validation and Stashing Mechanism

The tuner implements a **validation and rollback** mechanism to ensure update quality (from `tuner.go`):

1. **Stashing**: Before applying updates, the filter state (X and P) is saved using a `Stasher` object.

2. **Parameter Positivity Check**: All parameters (α, β, γ, δ) must be strictly positive:
   $$\alpha > 0, \quad \beta > 0, \quad \gamma > 0, \quad \delta > 0$$

3. **Normalized Innovation Squared (NIS) Test**: The innovation is validated using the NIS statistic:

$$NIS = \mathbf{y}_k^T \mathbf{S}_k^{-1} \mathbf{y}_k$$

Where $\mathbf{y}_k$ is the innovation vector and $\mathbf{S}_k$ is the innovation covariance matrix.

Under the assumption that the filter is operating correctly, NIS follows a chi-squared distribution with degrees of freedom equal to the measurement dimension (2 for [TTFT, ITL]). For a 95% confidence interval with 2 degrees of freedom, the chi-squared critical value is:

$$\chi^2_{0.95, 2} \approx 5.99$$

However, the implementation uses a more conservative threshold of **7.378**, corresponding to approximately 97.5% confidence (from `tuner_defaults.go`):

$$NIS < 7.378$$

If NIS exceeds this threshold, the update is rejected as an outlier, and the filter state is **unstashed** (rolled back) to the previous values.

4. **Rollback on Failure**: If any validation check fails, the tuner calls `UnStash()` to restore the previous filter state, ensuring robustness against bad measurements or model mismatches.

---

## 5. Autoscaler Optimization Framework

### 5.1 Resource Allocation Problem

**Decision variables:**
- For each server s: Choose accelerator type and number of replicas

**Objectives:**
- Minimize total cost
- Meet all SLO constraints
- Respect accelerator capacity limits

**Constraints:**

1. **SLO constraints** (per server):
For TTFT:

$$TTFT_s \leq TTFT_s^{target}$$

For ITL:

$$ITL_s \leq ITL_s^{target}$$

For TPS:

$$TPS_s \geq TPS_s^{target}$$

2. **Capacity constraints** (per accelerator type):

$$\sum_s n_s^{(a)} \leq C^{(a)}$$

Where $n_s^{(a)}$ is the number of accelerator type a assigned to server s.

3. **Priority constraints**: Higher priority servers must be satisfied before lower priority servers.

### 5.2 Greedy Allocation Algorithm

The system uses a **greedy algorithm** with priority ordering (from `greedy.go`):

**Algorithm Overview:**

1. **Initialization:**
   - For each server s, compute candidate allocations for each accelerator type
   - Sort allocations by value (transition penalty + cost)
   - Create server entry with priority and allocation list

2. **Priority-based greedy selection:**
   ```
   While servers remain:
     - Select highest priority server with lowest delta
     - Delta = value difference to next allocation option
     - If resources available for current allocation:
       * Assign allocation
       * Update available resources
     - Else:
       * Move to next allocation option
       * Reinsert into priority queue
   ```

3. **Best-effort allocation:**
   - For unallocated servers, apply saturation policy:
     * **None**: No additional allocation
     * **PriorityExhaustive**: Allocate greedily by priority
     * **PriorityRoundRobin**: Round-robin within priority groups
     * **RoundRobin**: Round-robin across all servers

**Delta calculation:**
$$\Delta_s = V(a_{next}) - V(a_{current})$$

Where V(a) is the allocation value (cost + transition penalty).

Servers with large Δ are deprioritized because moving to their next option has high cost.

### 5.3 Allocation Value and Transition Penalty

**Allocation value** combines cost and transition penalty:

$$V(a_{new}) = C(a_{new}) + P(a_{current} \rightarrow a_{new})$$

**Transition penalty** (from `allocation.go`):

If same accelerator type:
$$P = |C(a_{new}) - C(a_{current})|$$

If different accelerator type:
$$P = \phi \cdot (C(a_{current}) + C(a_{new})) + |C(a_{new}) - C(a_{current})|$$

Where φ is the accelerator penalty factor (default 0.1).

This penalizes switching accelerator types to promote stability.

### 5.4 Sizing: Target Rate Computation

For a given SLO target, the system uses **binary search** to find the maximum sustainable arrival rate (from `queueanalyzer.go`):

**Binary search algorithm:**

1. Define search range: $[\lambda_{min}, \lambda_{max}]$

2. For TTFT target:
   ```
   Find λ* such that: TTFT(λ*) = TTFT_target
   ```

3. For ITL target:
   ```
   Find λ* such that: ITL(λ*) = ITL_target
   ```

4. Final allocation rate:
   ```
   λ_alloc = min(λ*_TTFT, λ*_ITL, λ*_TPS)
   ```

**Binary search implementation** (from `utils.go`):

Given function $f(\lambda)$ and target y:

```
Evaluate f(λ_min) and f(λ_max)
Determine if f is increasing or decreasing

While not converged:
  λ_mid = (λ_min + λ_max) / 2
  y_mid = f(λ_mid)
  
  If |y_mid - y| < tolerance:
    Return λ_mid
  
  Update search bounds based on monotonicity
```

Convergence tolerance: $|y_{mid} - y| / y \leq 10^{-6}$

---

## 6. System Integration and Data Flow

### 6.1 Observer Pattern and Environment Structure

The system uses an **Environment structure** to encapsulate workload observations (from `environment.go`):

**Environment structure:**

- Lambda: Request arrival rate (requests/minute)
- AvgInputToks: Average input tokens per request
- AvgOutputToks: Average output tokens per request
- MaxBatchSize: Maximum batch size
- AvgTTFT: Average Time to First Token (milliseconds)
- AvgITL: Average Inter-Token Latency (milliseconds)

The Environment provides a `GetObservations()` method that returns the measurement vector:

$$\mathbf{z} = \begin{bmatrix} \text{AvgTTFT} \\ \text{AvgITL} \end{bmatrix}$$

**Environment validation**: The `Valid()` method ensures all fields are positive and non-zero before tuning.

### 6.2 Tuner Service Architecture

**TunerService** maintains tuners for multiple servers (from `tunerservice/`):

**Components:**

1. **TunerServer** (HTTP API):
   ```
   GET /getparams?server_name=<name>
   → Returns {server_name, alpha, beta}
   ```

2. **TunerService** (core logic):
   - Manages tuner instances per server
   - Periodically updates all tuners
   - Coordinates with collector for observations

**Update cycle:**

```
Every tuner_period seconds:
  1. Fetch environments from collector
  2. For each server:
     - Update tuner environment
     - Run prediction step
     - Run update step
     - Extract parameters (α, β)
  3. Expose updated parameters via API
```

### 6.3 Integration with Autoscaler

**Control loop:**

```
1. Collector gathers metrics from inference servers
   → Provides to TunerService

2. TunerService runs EKF for each server
   → Estimates α, β parameters
   → Exposes via /getparams API

3. Autoscaler queries /getparams
   → Gets current (α, β) for each server
   → Updates queueing models

4. Autoscaler runs optimization
   → Computes allocations using updated models
   → Generates desired state

5. Orchestrator applies changes
   → Scales servers to meet desired state
```

This creates a **closed-loop adaptive system** where model parameters track system behavior.

---

## 7. Mathematical Properties and Convergence

### 7.1 Kalman Filter Optimality

The Kalman Filter is **optimal** in the minimum mean squared error (MMSE) sense for linear Gaussian systems:

$$\hat{\mathbf{x}}_{k|k} = \arg\min_{\hat{\mathbf{x}}} E[||\mathbf{x}_k - \hat{\mathbf{x}}||^2 | \mathbf{z}_{1:k}]$$

For the **Extended Kalman Filter**, optimality is not guaranteed due to linearization, but it provides:
- **Consistency**: Estimation error covariance converges
- **Bounded error**: Under observability and controllability conditions

### 7.2 Observability Analysis

The system is **observable** if the observation model provides sufficient information to estimate all state components.

**Observability condition:** The observability matrix must have full rank:

$$\mathcal{O} = \begin{bmatrix} \mathbf{H}_k \\ \mathbf{H}_k\mathbf{F}_{k-1} \\ \vdots \\ \mathbf{H}_k\mathbf{F}_{k-1}^{n-1} \end{bmatrix}$$

For our system:

- State: [α, β, δ, γ] (4 parameters)
- Observations: [TTFT, ITL] (2 measurements)

Both observations depend on all four parameters through the queueing model, providing observability when:

1. Workload is non-zero (λ > 0)
2. System is not saturated (ρ < 1)
3. Batch size varies over time
4. Both prefill and decode phases are active (input and output tokens > 0)

The observation function couples all parameters through the service rate formulation, ensuring that changes in any parameter affect the observed metrics, making the system theoretically observable despite having more parameters (4) than measurements (2).

### 7.3 Stability and Convergence

**Filter stability** requires:
1. Process noise Q > 0 (system is controllable)
2. Measurement noise R > 0 (measurements are informative)
3. Initial covariance P₀ > 0

Under these conditions, the **covariance matrix** P converges to a steady-state value:

$$\lim_{k \rightarrow \infty} \mathbf{P}_k = \mathbf{P}_{\infty}$$

For time-invariant systems, $\mathbf{P}_{\infty}$ satisfies the **Discrete Algebraic Riccati Equation**:

$$\mathbf{P}_{\infty} = \mathbf{F}\mathbf{P}_{\infty}\mathbf{F}^T - \mathbf{F}\mathbf{P}_{\infty}\mathbf{H}^T(\mathbf{H}\mathbf{P}_{\infty}\mathbf{H}^T + \mathbf{R})^{-1}\mathbf{H}\mathbf{P}_{\infty}\mathbf{F}^T + \mathbf{Q}$$

### 7.4 Queueing Model Accuracy

The queueing model's accuracy depends on several assumptions:

**Valid assumptions:**
- Markovian arrivals (reasonable for aggregated traffic)
- State-dependent service (matches batched inference)
- Finite capacity (matches real systems)

**Limitations:**
- Assumes exponential service time distribution
- Ignores request size variability
- Does not model pre-emption or context switching

**Model validation:** Comparing predictions to observations through innovation analysis:

$$\mathbf{y}_k = \mathbf{z}_k - h(\hat{\mathbf{x}}_{k|k-1})$$

If the model is accurate, innovations should be:

- Zero-mean: $E[\mathbf{y}_k] = 0$
- White: $E[\mathbf{y}_k\mathbf{y}_j^T] = 0$ for $k \neq j$
- Consistent with S: $E[\mathbf{y}_k\mathbf{y}_k^T] = \mathbf{S}_k$

---

## 8. Computational Complexity

### 8.1 Kalman Filter Complexity

**Per iteration complexity:**

1. **Prediction**: $O(n^3)$ for matrix multiplication
   - $\mathbf{F}\mathbf{P}\mathbf{F}^T$: $O(n^3)$

2. **Update**: $O(n^2m + m^3)$
   - $\mathbf{H}\mathbf{P}\mathbf{H}^T$: $O(nm^2)$
   - $\mathbf{S}^{-1}$: $O(m^3)$
   - $\mathbf{K}$ computation: $O(n^2m)$

For our system: n = 2, m = 2, so complexity is $O(1)$ (constant).

**Numerical Jacobian computation:**
- For each state dimension: 2 function evaluations
- Total: 2n function evaluations = 4 queueing model solves

### 8.2 Queueing Model Solution Complexity

**M/M/1/K state-dependent model:**

1. **State probability computation**: $O(K)$
   - Recursive computation of $p_n$
   - Normalization: $O(K)$

2. **Performance metrics**: $O(K)$
   - Expectation calculations: $\sum_{n=0}^K f(n) p_n$

For typical values (K ≈ 100), this is efficient: ~100 operations.

### 8.3 Optimization Complexity

**Greedy algorithm:**

1. **Allocation computation** (per server): $O(A)$
   - A = number of accelerator types
   - Each allocation: solve queueing model

2. **Greedy selection**: $O(S \log S)$
   - S = number of servers
   - Priority queue operations

3. **Total**: $O(S \cdot A + S \log S)$

For typical deployments (S ~ 100, A ~ 5), this completes in milliseconds.

---

## 9. Practical Considerations

### 9.1 Initialization Strategy

**Initial state estimate** $\mathbf{x}_0$:
- Option 1: Use offline profiling data
- Option 2: Conservative defaults (high α, β for stability)
- Option 3: Rapid adaptation phase with high process noise

**Initial covariance** $\mathbf{P}_0$:
- Should reflect uncertainty in initial estimate
- Typical: $\mathbf{P}_0 = \text{diag}[(0.1 \cdot \alpha_0)^2, (0.1 \cdot \beta_0)^2]$

### 9.2 Noise Parameter Tuning

**Process noise Q:**
- Reflects expected parameter variation
- Too small: slow adaptation to changes
- Too large: noisy estimates, instability
- Guideline: 1-10% expected change per step

**Measurement noise R:**
- Reflects observation accuracy
- Should account for:
  * Metric aggregation window
  * Sampling variability
  * System noise

### 9.3 Handling Edge Cases

**Zero or low load:**
- Service time observations become unreliable
- Solution: Increase R (measurement uncertainty)
- Alternative: Pause updates until load increases

**Saturated system:**
- Queueing model assumptions violated
- Large queue times, high variance
- Solution: Detect saturation, flag for scaling

**Transient behavior:**
- Sudden load changes
- Large innovations
- Solution: Adaptive R based on innovation magnitude

### 9.4 Multi-Server Considerations

**Independent tuners** per server:
- Each server has separate state estimates
- Allows heterogeneous accelerator types
- No information sharing between servers

**Potential improvements:**
- Hierarchical filtering: shared hyperparameters
- Information sharing: similar workloads
- Multi-agent coordination

---

## 10. Evaluation Metrics

### 10.1 Filter Performance Metrics

**Estimation accuracy:**
$$RMSE = \sqrt{\frac{1}{T}\sum_{t=1}^T ||\mathbf{x}_t - \hat{\mathbf{x}}_t||^2}$$

**Innovation analysis:**
- Innovation mean: Should be ≈ 0
- Innovation covariance: Should match $\mathbf{S}_k$
- Normalized Innovation Squared (NIS):
$$\epsilon_k = \mathbf{y}_k^T \mathbf{S}_k^{-1} \mathbf{y}_k$$

Should follow $\chi^2$ distribution with m degrees of freedom.

### 10.2 Autoscaler Performance Metrics

**SLO compliance:**
- Fraction of time SLOs are met
- SLO violation severity

**Resource efficiency:**
- Total GPU-hours used
- Cost per request served
- Utilization distribution

**Adaptation speed:**
- Time to detect workload change
- Time to rescale

### 10.3 End-to-End Metrics

**Prediction vs. Reality:**
Compare predicted performance (from queueing model) to observed:
- TTFT prediction error
- ITL prediction error
- TPS prediction error

**Economic metrics:**
- Total operational cost
- Cost per SLO violation
- Opportunity cost of over-provisioning

---

## 11. Kubernetes Controller Architecture

### 11.1 Controller Pattern Fundamentals

The Workload Variant Autoscaler implements the **Kubernetes Operator Pattern**, which extends Kubernetes functionality through custom controllers. This section describes the architectural foundations and reconciliation loop implementation.

**Reference**: [Kubernetes Controllers](https://kubernetes.io/docs/concepts/architecture/controller/), [Operator Pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

#### 11.1.1 Control Loop Theory

A **controller** in Kubernetes is a control loop that continuously drives the cluster toward a desired state. Borrowing from control theory, the controller:

1. **Observes** the current state of the system
2. **Compares** current state to desired state  
3. **Acts** to reconcile any differences
4. **Repeats** continuously

This mirrors the thermostat analogy: the desired temperature (spec) vs. actual room temperature (status), with the heating/cooling system as the actuator.

**Mathematical Representation:**

Let $s(t)$ represent system state at time $t$, $s^*$ the desired state, and $u(t)$ the control action:

$$u(t) = K(s^* - s(t))$$

where $K$ is the controller gain (reconciliation logic). The system evolves:

$$\frac{ds}{dt} = f(s, u(t))$$

The controller aims to drive $s(t) \rightarrow s^*$ as $t \rightarrow \infty$.

#### 11.1.2 Kubernetes API Server Integration

Controllers interact with the cluster through the **Kubernetes API Server**, which provides:

- **Declarative API**: Resources defined by specifications (desired state)
- **Watch mechanism**: Efficient change notifications
- **Optimistic concurrency**: Resource versioning for conflict resolution
- **RBAC**: Fine-grained access control

**API Operations:**
- `GET`: Read current state
- `LIST`: Enumerate resources
- `WATCH`: Stream updates
- `UPDATE`: Modify resources
- `PATCH`: Partial updates

### 11.2 VariantAutoscaling Custom Resource Definition (CRD)

The WVA defines a **VariantAutoscaling** custom resource (from `variantautoscaling_types.go`):

```go
type VariantAutoscalingSpec struct {
    ModelID      string        // Model identifier
    SLOClassRef  ConfigMapKeyRef // SLO configuration reference
    ModelProfile ModelProfile  // Performance characteristics
}

type VariantAutoscalingStatus struct {
    CurrentAlloc          Allocation      // Current resource allocation
    DesiredOptimizedAlloc OptimizedAlloc // Target allocation
    Actuation            ActuationStatus // Actuation state
    Conditions           []Condition     // Status conditions
}
```

**Key Design Principles:**

1. **Spec defines intent**: User declares desired model serving configuration
2. **Status reflects reality**: Controller updates status with current and desired allocations
3. **Conditions track progress**: Standard Kubernetes condition types for observability

**Condition Types:**
- `OptimizationReady`: Indicates successful optimization computation
- `ActuationComplete`: Indicates scaling actions completed
- `TuningActive`: Indicates parameter estimation is running

### 11.3 Reconciliation Loop Implementation

The **Reconcile** function is the heart of the controller (from `variantautoscaling_controller.go`):

```go
func (r *VariantAutoscalingReconciler) Reconcile(
    ctx context.Context, 
    req ctrl.Request,
) (ctrl.Result, error)
```

**Reconciliation Flow:**

```
┌─────────────────────────────────────────────────────────┐
│                   Reconciliation Loop                    │
├─────────────────────────────────────────────────────────┤
│                                                           │
│  1. Read Configuration                                    │
│     - Optimization interval                               │
│     - Accelerator costs                                   │
│     - Service class definitions                           │
│                                                           │
│  2. List Active VariantAutoscalings                       │
│     - Filter by status (active vs suspended)              │
│                                                           │
│  3. Create System Model                                   │
│     - Build inferno.System with accelerators              │
│     - Configure optimizer spec (unlimited mode)           │
│                                                           │
│  4. Tune Model Parameters (if enabled)                    │
│     - TunerManager.TuneModelPerfParams()                  │
│     - Update α, β, γ, δ for each server                  │
│                                                           │
│  5. Analyze Models                                        │
│     - For each variant:                                   │
│       * ModelAnalyzer.AnalyzeModel()                      │
│       * Compute potential allocations                     │
│       * Estimate performance for each allocation          │
│                                                           │
│  6. Optimize Allocations                                  │
│     - VariantAutoscalingsEngine.Optimize()                │
│     - Greedy algorithm with priority ordering             │
│     - Compute desired state for all variants              │
│                                                           │
│  7. Emit Metrics                                          │
│     - Prometheus metrics for current state                │
│     - Optimization decisions                              │
│                                                           │
│  8. Update Status                                         │
│     - Write DesiredOptimizedAlloc to status               │
│     - Update conditions                                   │
│                                                           │
│  9. Actuate Changes                                       │
│     - Actuator.UpdateDeploymentReplicas()                 │
│     - Scale inference server deployments                  │
│                                                           │
│  10. Requeue                                              │
│     - Schedule next reconciliation                        │
│     - Default: 60 seconds (configurable)                  │
│                                                           │
└─────────────────────────────────────────────────────────┘
```

**Key Implementation Details:**

1. **Idempotency**: Reconcile can be called multiple times; must produce same result
2. **Error handling**: Return error to trigger exponential backoff retry
3. **Requeue control**: Return `ctrl.Result{RequeueAfter: duration}` for periodic execution
4. **Status updates**: Separate from spec updates to avoid conflicts

**Code Structure:**

```go
// Read configuration
interval, err := r.readOptimizationConfig(ctx)
requeueDuration, _ := time.ParseDuration(interval)

// List active VAs
var vaList VariantAutoscalingList
r.List(ctx, &vaList)
activeVAs := filterActiveVariantAutoscalings(vaList.Items)

// Build system model
systemData := utils.CreateSystemData(acceleratorCm, serviceClassCm)
system := inferno.NewSystem()
system.SetFromSpec(&systemData.Spec)

// Tune parameters
if r.TunerMgr.IsEnabled() {
    r.TunerMgr.TuneModelPerfParams(systemData)
}

// Analyze and optimize
modelAnalyzer := analyzer.NewModelAnalyzer(system)
// ... analysis logic ...

engine := variantAutoscalingOptimizer.NewVariantAutoscalingsEngine(manager, system)
optimizedAllocation, err := engine.Optimize(ctx, *updateList, allAnalyzerResponses)

// Update status and actuate
// ... status update logic ...

return ctrl.Result{RequeueAfter: requeueDuration}, nil
```

### 11.4 Controller-Runtime Framework

WVA uses **controller-runtime**, a Go library providing abstractions for building Kubernetes controllers:

**Key Components:**

1. **Manager**: Coordinates multiple controllers, handles leader election
2. **Client**: Cached, type-safe Kubernetes API client
3. **Scheme**: Type registry for custom resources
4. **Cache**: Informer-based caching layer
5. **Recorder**: Event recorder for user-facing events

**Setup Code:**

```go
mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
    Scheme: scheme,
    Metrics: metricsserver.Options{BindAddress: ":8080"},
    LeaderElection: true,
    LeaderElectionID: "workload-variant-autoscaler-lock",
})

reconciler := &VariantAutoscalingReconciler{
    Client:   mgr.GetClient(),
    Scheme:   mgr.GetScheme(),
    PromAPI:  promAPI,
    TunerMgr: tunerMgr,
}

err = ctrl.NewControllerManagedBy(mgr).
    For(&v1alpha1.VariantAutoscaling{}).
    Watches(&appsv1.Deployment{}, handler.EnqueueRequestsFromMapFunc(...)).
    WithOptions(controller.Options{MaxConcurrentReconciles: 1}).
    Complete(reconciler)
```

**Reference**: [controller-runtime documentation](https://pkg.go.dev/sigs.k8s.io/controller-runtime)

### 11.5 Watch and Event Mechanisms

Controllers use **watches** to receive notifications of resource changes, avoiding polling:

**Primary Watch**: `VariantAutoscaling` resources
- Triggers reconciliation when spec or status changes
- Uses informers with local cache

**Secondary Watches**: `Deployment` resources
- Monitors inference server deployments
- Maps deployment changes back to owning VariantAutoscaling
- Enables reactive reconciliation when external changes occur

**Event Types:**
- `Create`: New resource created
- `Update`: Resource modified
- `Delete`: Resource deleted

**Filtering with Predicates:**

```go
predicate.Funcs{
    UpdateFunc: func(e event.UpdateEvent) bool {
        // Only reconcile on generation changes (spec updates)
        return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
    },
    DeleteFunc: func(e event.DeleteEvent) bool {
        // Reconcile deletions to clean up
        return true
    },
}
```

This reduces unnecessary reconciliations, improving efficiency.

### 11.6 Leader Election and High Availability

For production deployments, multiple controller replicas run with **leader election**:

**Algorithm**: Kubernetes lease-based leader election
- One replica acquires a `Lease` resource
- Leader performs reconciliation
- Other replicas wait on standby
- Automatic failover on leader failure

**Configuration:**

```go
ctrl.Options{
    LeaderElection: true,
    LeaderElectionID: "workload-variant-autoscaler-lock",
    LeaderElectionNamespace: "workload-variant-autoscaler-system",
    LeaseDuration: 15 * time.Second,
    RenewDeadline: 10 * time.Second,
    RetryPeriod: 2 * time.Second,
}
```

**Benefits:**
- High availability: Automatic failover
- Consistency: Single controller modifying resources
- Scalability: Standby replicas ready to take over

### 11.7 RBAC and Security

The controller requires specific permissions defined via **RBAC** (Role-Based Access Control):

**Kubebuilder RBAC Markers:**

```go
// +kubebuilder:rbac:groups=llmd.ai,resources=variantautoscalings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=llmd.ai,resources=variantautoscalings/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch
```

**Generated RBAC Resources:**
- `ClusterRole`: Defines permissions
- `ClusterRoleBinding`: Binds role to ServiceAccount
- `ServiceAccount`: Identity for controller pods

**Security Principles:**
- **Least privilege**: Only necessary permissions granted
- **Separation of duties**: Status updates separate from spec
- **Audit trail**: All API operations logged

### 11.8 Integration with Prometheus

The controller exposes **Prometheus metrics** for observability and integration with external autoscalers (like Kubernetes HPA).

**Inferno Output Metrics** (from `internal/constants/metrics.go`):

The WVA emits the following metrics to Prometheus, defined in `InfernoReplicaScalingTotal`, `InfernoDesiredReplicas`, `InfernoCurrentReplicas`, and `InfernoDesiredRatio` constants:

1. **`inferno_replica_scaling_total`** (Counter):
   - Tracks the total number of scaling operations performed
   - Labels: `variant_name`, `namespace`, `direction` (up/down), `reason`, `accelerator_type`
   - Purpose: Audit trail of all scaling decisions

2. **`inferno_desired_replicas`** (Gauge):
   - Desired number of replicas determined by WVA optimization
   - Labels: `variant_name`, `namespace`, `accelerator_type`
   - Purpose: Target state for external autoscalers to consume

3. **`inferno_current_replicas`** (Gauge):
   - Current actual number of replicas from Deployment status
   - Labels: `variant_name`, `namespace`, `accelerator_type`
   - Purpose: Real-time state tracking

4. **`inferno_desired_ratio`** (Gauge):
   - Ratio of desired to current replicas
   - Labels: `variant_name`, `namespace`, `accelerator_type`
   - Purpose: Convergence monitoring and alerting

**Metrics Emission** (from `internal/actuator/actuator.go`):

```go
func (a *Actuator) EmitMetrics(
    ctx context.Context, 
    va *VariantAutoscaling,
) error {
    currentReplicas, err := a.getCurrentDeploymentReplicas(ctx, va)
    if err != nil {
        // Fallback to VariantAutoscaling status
        currentReplicas = int32(va.Status.CurrentAlloc.NumReplicas)
    }
    
    return a.MetricsEmitter.EmitReplicaMetrics(
        ctx,
        va,
        currentReplicas,  // Real current from Deployment
        int32(va.Status.DesiredOptimizedAlloc.NumReplicas),  // WVA target
        va.Status.DesiredOptimizedAlloc.Accelerator,
    )
}
```

**Scraping Configuration:**

```yaml
apiVersion: v1
kind: Service
metadata:
  name: controller-metrics
  labels:
    control-plane: controller-manager
spec:
  selector:
    control-plane: controller-manager
  ports:
  - name: metrics
    port: 8080
    targetPort: 8080
```

Prometheus scrapes `http://<service>:8080/metrics` for time-series data. These metrics enable:
- **HPA Integration**: External autoscalers consume `inferno_desired_replicas`
- **Observability**: Grafana dashboards track scaling behavior
- **Alerting**: Prometheus alerts on convergence failures (ratio ≠ 1)

### 11.9 ConfigMap-Based Configuration

The controller uses **ConfigMaps** for external configuration without requiring recompilation or redeployment.

**1. Main Controller Configuration** (`variantautoscaling-config`):

From `config/manager/configmap.yaml`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: variantautoscaling-config
  namespace: workload-variant-autoscaler-system
data:
  # Prometheus configuration
  PROMETHEUS_BASE_URL: "https://kube-prometheus-stack-prometheus.workload-variant-autoscaler-monitoring.svc.cluster.local:9090"
  PROMETHEUS_TLS_INSECURE_SKIP_VERIFY: "true"
  
  # Optimization interval - controls reconciliation frequency
  GLOBAL_OPT_INTERVAL: "60s"
  
  # Scale-to-zero configuration
  WVA_SCALE_TO_ZERO: "false"
```

**2. Accelerator Configuration** (`accelerator-unit-costs`):

```yaml
data:
  L40S: |
    {
      "name": "L40S",
      "count": 1,
      "cost": 1.5,
      "memory": 48
    }
  H100: |
    {
      "name": "H100", 
      "count": 1,
      "cost": 3.0,
      "memory": 80
    }
```

**3. Service Class Configuration** (`service-classes-config`):

```yaml
data:
  premium: |
    {
      "name": "premium",
      "priority": 1,
      "ttft_ms": 500,
      "itl_ms": 50,
      "tps": 100
    }
```

**Reading ConfigMaps** (from `internal/controller/variantautoscaling_controller.go`):

```go
func (r *VariantAutoscalingReconciler) readOptimizationConfig(
    ctx context.Context,
) (string, error) {
    cm := &corev1.ConfigMap{}
    err := r.Get(ctx, types.NamespacedName{
        Name: configMapName,
        Namespace: configMapNamespace,
    }, cm)
    
    // Returns value of GLOBAL_OPT_INTERVAL key
    return cm.Data["GLOBAL_OPT_INTERVAL"], nil
}
```

This enables **runtime reconfiguration** of optimization intervals, Prometheus endpoints, and autoscaling behavior without controller restarts.

### 11.10 Actuation Strategy

The **Actuator** component serves as the metrics emission layer, exposing WVA's optimization decisions to external systems. Unlike traditional autoscalers that directly modify Kubernetes resources, WVA delegates actual scaling to external autoscalers (like HPA) through Prometheus metrics.

**Architecture Design:**

WVA follows a **separation of concerns** pattern:
1. **WVA Controller**: Performs optimization, determines desired allocations
2. **Actuator**: Emits optimization results as Prometheus metrics
3. **External Autoscaler** (HPA/KEDA): Consumes metrics and scales Deployments

**Actuator Implementation** (from `internal/actuator/actuator.go`):

```go
type Actuator struct {
    Client         client.Client
    MetricsEmitter *metrics.MetricsEmitter
}

func (a *Actuator) EmitMetrics(
    ctx context.Context,
    va *llmdOptv1alpha1.VariantAutoscaling,
) error {
    // Get real-time replica count from Deployment
    currentReplicas, err := a.getCurrentDeploymentReplicas(ctx, va)
    if err != nil {
        // Fallback to VariantAutoscaling status
        currentReplicas = int32(va.Status.CurrentAlloc.NumReplicas)
    }
    
    // Emit Prometheus metrics
    return a.MetricsEmitter.EmitReplicaMetrics(
        ctx,
        va,
        currentReplicas,  // Current state
        int32(va.Status.DesiredOptimizedAlloc.NumReplicas),  // WVA target
        va.Status.DesiredOptimizedAlloc.Accelerator,
    )
}

func (a *Actuator) getCurrentDeploymentReplicas(
    ctx context.Context,
    va *llmdOptv1alpha1.VariantAutoscaling,
) (int32, error) {
    var deploy appsv1.Deployment
    err := utils.GetDeploymentWithBackoff(
        ctx, a.Client, va.Name, va.Namespace, &deploy,
    )
    if err != nil {
        return 0, err
    }
    
    // Prefer status.replicas (actual running pods)
    if deploy.Status.Replicas >= 0 {
        return deploy.Status.Replicas, nil
    }
    
    // Fallback to spec.replicas
    if deploy.Spec.Replicas != nil {
        return *deploy.Spec.Replicas, nil
    }
    
    return 1, nil  // Default
}
```

**Actuation Flow:**

1. **Reconciliation**: Controller optimizes allocations every `GLOBAL_OPT_INTERVAL`
2. **Status Update**: Write desired allocation to `VariantAutoscaling.Status.DesiredOptimizedAlloc`
3. **Metrics Emission**: Actuator calls `EmitMetrics()` to publish to Prometheus
4. **External Consumption**: HPA/KEDA scrapes `inferno_desired_replicas` metric
5. **Scaling**: External autoscaler modifies Deployment.Spec.Replicas
6. **Convergence**: Process repeats until current == desired

**Benefits of Metrics-Based Actuation:**

- **Decoupling**: WVA focuses on optimization, not Kubernetes resource management
- **Flexibility**: Any autoscaler can consume WVA's signals
- **Observability**: All decisions visible in Prometheus/Grafana
- **Safety**: External autoscalers can apply their own rate limiting and safeguards
- **Gradualism**: HPA can implement gradual scale-up/scale-down policies

**Integration with HPA:**

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: llama-70b-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: llama-70b-deployment
  minReplicas: 1
  maxReplicas: 10
  metrics:
  - type: External
    external:
      metric:
        name: inferno_desired_replicas
        selector:
          matchLabels:
            variant_name: llama-70b
      target:
        type: Value
        value: "1"  # HPA scales to match this metric value
```

This architecture enables WVA to provide intelligent optimization while leveraging battle-tested Kubernetes autoscaling infrastructure.

## 12. Related Work and Context

### 12.1 Queueing Theory for Computer Systems

Classical queueing theory has been applied to computer systems since the 1960s:

**Foundational Works:**
- **Kleinrock, L. (1975)**: *Queueing Systems, Volume 1: Theory*. Wiley-Interscience. 
  - Comprehensive treatment of M/M/1, M/M/c, M/G/1 queues
  - Birth-death processes and limiting distributions
  - Little's Law and operational analysis
  
- **Kleinrock, L. (1976)**: *Queueing Systems, Volume 2: Computer Applications*. Wiley-Interscience.
  - Application to computer networks and time-sharing systems
  - Priority queueing and scheduling disciplines
  
- **Bolch, G., Greiner, S., de Meer, H., & Trivedi, K. S. (2006)**: *Queueing Networks and Markov Chains* (2nd ed.). Wiley.
  - Queueing networks and product-form solutions
  - Numerical methods for non-product-form networks
  - Jackson networks and BCMP theorem

**M/M/1/K Queues with State-Dependent Service:**
- **Gross, D., Shortle, J. F., Thompson, J. M., & Harris, C. M. (2008)**: *Fundamentals of Queueing Theory* (4th ed.). Wiley.
  - Finite capacity queues and blocking probabilities
  - State-dependent service rates: $\mu_n = \mu(n)$
  - Erlang loss formula and trunk reservation

**Modern Applications:**
- **Autoscaling**: Predictive models for cloud resource allocation (Amazon, Google, Microsoft)
- **Load balancing**: Power-of-two-choices, join-shortest-queue policies
- **Microservices**: Service mesh performance modeling

**LLM-Specific Queueing Models:**
- **Agrawal et al. (2024)**: "Taming Throughput-Latency Tradeoff in LLM Inference with Sarathi-Serve"
  - Chunked-prefill scheduling
  - Stall-free batched inference
  
- **Griggs et al. (2024)**: "Shepherd: Optimizing LLM Serving with Efficient Co-location"
  - Interference modeling between co-located models
  - Performance isolation techniques

### 12.2 Kalman Filtering: Theory and Applications

**Original Development:**
- **Kalman, R. E. (1960)**: "A New Approach to Linear Filtering and Prediction Problems". *Journal of Basic Engineering*, 82(1), 35-45.
  - Recursive optimal estimation for linear systems
  - Prediction-correction framework
  - Riccati equation for error covariance
  
- **Kalman, R. E., & Bucy, R. S. (1961)**: "New Results in Linear Filtering and Prediction Theory". *Journal of Basic Engineering*, 83(1), 95-108.
  - Continuous-time Kalman-Bucy filter
  - Optimality proofs

**Extended Kalman Filter:**
- **Jazwinski, A. H. (1970)**: *Stochastic Processes and Filtering Theory*. Academic Press.
  - EKF for nonlinear systems via linearization
  - Taylor series expansion and Jacobian computation
  - Stability and convergence analysis
  
- **Anderson, B. D. O., & Moore, J. B. (1979)**: *Optimal Filtering*. Prentice-Hall.
  - Theoretical foundations of optimal estimation
  - Observability and controllability
  - Discrete-time formulations

**Numerical Methods:**
- **Simon, D. (2006)**: *Optimal State Estimation: Kalman, H∞, and Nonlinear Approaches*. Wiley.
  - Numerical Jacobian computation techniques
  - Alternative nonlinear filters (UKF, particle filters)
  - Practical implementation considerations
  
- **Einicke, G. A. (2019)**: *Smoothing, Filtering and Prediction* (2nd ed.). Amazon.
  - Modern numerical algorithms
  - Square-root filtering for numerical stability
  - Adaptive filtering techniques

**Applications in Systems and Control:**
- **Maybeck, P. S. (1979)**: *Stochastic Models, Estimation, and Control, Volume 1*. Academic Press.
  - Navigation and guidance systems (GPS, inertial navigation)
  - Target tracking and sensor fusion
  
- **Bar-Shalom, Y., Li, X. R., & Kirubarajan, T. (2001)**: *Estimation with Applications to Tracking and Navigation*. Wiley.
  - Multi-target tracking
  - Data association and gating

**System Identification:**
- **Ljung, L. (1999)**: *System Identification: Theory for the User* (2nd ed.). Prentice-Hall.
  - Parameter estimation for dynamical systems
  - Adaptive control and self-tuning regulators
  - Convergence analysis and persistent excitation

**Innovation Analysis and Validation:**
- **Mehra, R. K. (1970)**: "On the Identification of Variances and Adaptive Kalman Filtering". *IEEE Transactions on Automatic Control*, 15(2), 175-184.
  - Adaptive noise covariance estimation
  - Chi-squared tests for innovation sequences
  - Filter consistency checks

### 12.3 LLM Inference Optimization

**Serving Systems:**
- **Kwon, W., Li, Z., Zhuang, S., Sheng, Y., Zheng, L., Yu, C. H., ... & Stoica, I. (2023)**: "Efficient Memory Management for Large Language Model Serving with PagedAttention". *SOSP 2023*.
  - **vLLM**: Virtual memory-style KV cache management
  - Continuous batching with dynamic request arrival
  - 2-24× throughput improvement
  
- **Yu, G. I., Jeong, J. S., Kim, G. W., Kim, S., & Chun, B. G. (2022)**: "Orca: A Distributed Serving System for Transformer-Based Generative Models". *OSDI 2022*.
  - Iteration-level scheduling
  - Selective batching to reduce queueing delays
  
- **Wu, X., Duan, Q., Zheng, Z., & Yang, Z. (2023)**: "FastServe: Fast Distributed Inference Serving for Large Language Models". *arXiv:2305.05920*.
  - Preemptive scheduling with skip-join MLFQ
  - Job size prediction and priority queues

**Performance Modeling:**
- **Yang, L., et al. (2024)**: "Hybrid Prefill-Decode Architecture for LLM Inference"
  - Analytical models for prefill/decode separation
  - Resource allocation strategies
  
- **Yuan, Z., et al. (2024)**: "LLM Inference Throughput Optimization"
  - Batch size optimization using queueing theory
  - Trade-offs between latency and throughput
  
- **Zhu, Q., et al. (2025)**: "Predictive Autoscaling for LLM Workloads"
  - Time-series forecasting for arrival rates
  - Proactive scaling strategies

**Benchmarking and Characterization:**
- **Casson, A. (2023)**: "Understanding Transformer Inference Performance"
  - FLOPS-based performance modeling
  - Memory bandwidth bottlenecks
  - Roofline analysis for LLMs

---

## 12. Future Research Directions

### 12.1 Improved Observability

**Hierarchical State Estimation:**
- Multi-level filtering: separate fast and slow parameter dynamics
- Hyperparameter learning: automatic Q and R tuning
- Cross-server information sharing for similar workloads

**Enhanced Metrics:**
- Per-request latency distributions (not just means)
- Request size heterogeneity
- Context-aware metrics (prompt length, output length)

### 12.2 Advanced Control Strategies

**Model Predictive Control (MPC):**
- Optimize over finite horizon
- Anticipate future workload changes
- Constraint handling (GPU limits, SLO bounds)

**Reinforcement Learning:**
- Learn scaling policies from experience
- Handle complex multi-objective optimization
- Adaptive exploration-exploitation

### 12.3 Enhanced Queueing Models

**Beyond Markovian Assumptions:**
- General service time distributions (M/G/1)
- Correlated arrivals (MMPP, MAP)
- Heavy-tailed distributions for request sizes

**Advanced Scheduling:**
- Priority queues with preemption
- Size-based scheduling (SRPT, PSJF)
- Context-aware routing

### 12.4 Multi-Tenant Considerations

**Fairness and Isolation:**
- Per-tenant SLOs
- Fair queueing algorithms
- Resource reservation vs. sharing

**Interference Modeling:**
- Contention for KV cache
- Memory bandwidth interference
- Network I/O impact

### 12.5 Hardware Heterogeneity

**Mixed Accelerator Types:**
- GPUs: H100, A100, L40S, L4
- NPUs: Trainium, Inferentia
- Specialized AI chips: TPU, Gaudi

**Optimization:**
- Model-accelerator matching
- Performance portability
- Cost-performance Pareto frontiers

---

## 13. Mathematical Foundations from Literature

### 13.1 M/M/1 Queue: Detailed Theory

The **M/M/1 queue** is the simplest queueing model, yet fundamental to understanding more complex systems. This section provides comprehensive mathematical foundations based on Kleinrock (1975) and Wikipedia.

**Reference**: [M/M/1 Queue - Wikipedia](https://en.wikipedia.org/wiki/M/M/1_queue)

#### 13.1.1 Model Definition

An M/M/1 queue is a stochastic process with state space {0, 1, 2, 3, ...} representing the number of customers in the system.

**Assumptions:**
- **Arrivals**: Poisson process with rate λ (exponential inter-arrival times)
- **Service**: Exponential distribution with rate μ (mean service time 1/μ)
- **Independence**: All arrival and service times are independent
- **Single server**: FCFS (First-Come, First-Served) discipline
- **Infinite buffer**: No limit on queue length

**Continuous-Time Markov Chain (CTMC):**

The system evolves as a birth-death process with transition rate matrix:

$$\mathbf{Q} = \begin{bmatrix}
-\lambda & \lambda & 0 & 0 & \cdots \\
\mu & -(\lambda+\mu) & \lambda & 0 & \cdots \\
0 & \mu & -(\lambda+\mu) & \lambda & \cdots \\
0 & 0 & \mu & -(\lambda+\mu) & \cdots \\
\vdots & \vdots & \vdots & \vdots & \ddots
\end{bmatrix}$$

**State Transition Diagram:**
```
     λ       λ       λ       λ
  0 ←→ 1 ←→ 2 ←→ 3 ←→ 4 ←→ ...
     μ       μ       μ       μ
```

#### 13.1.2 Stationary Distribution

For stability, we require **ρ = λ/μ < 1**, where ρ is the **utilization** or **traffic intensity**.

The **stationary probability** of being in state n is:

$$\pi_n = (1 - \rho) \rho^n, \quad n = 0, 1, 2, \ldots$$

This is a **geometric distribution** with parameter (1 - ρ).

**Verification:**
$$\sum_{n=0}^{\infty} \pi_n = (1-\rho) \sum_{n=0}^{\infty} \rho^n = (1-\rho) \cdot \frac{1}{1-\rho} = 1$$

**Idle probability**: $\pi_0 = 1 - \rho$ (probability server is idle)

**Busy probability**: $1 - \pi_0 = \rho$ (fraction of time server is busy)

#### 13.1.3 Performance Metrics

**Average Number in System (L):**

$$L = E[N] = \sum_{n=0}^{\infty} n \pi_n = \sum_{n=0}^{\infty} n(1-\rho)\rho^n = \frac{\rho}{1-\rho}$$

**Variance:**

$$\text{Var}(N) = \frac{\rho}{(1-\rho)^2}$$

**Average Number in Queue (Lq):**

$$L_q = L - \rho = \frac{\rho^2}{1-\rho}$$

(Customers in queue = total - those in service)

**Average Response Time (W):**

By **Little's Law**: $L = \lambda W$

$$W = \frac{L}{\lambda} = \frac{\rho}{\lambda(1-\rho)} = \frac{1}{\mu - \lambda}$$

**Average Waiting Time (Wq):**

$$W_q = W - \frac{1}{\mu} = \frac{\rho}{\mu(1-\rho)} = \frac{\lambda}{\mu(\mu - \lambda)}$$

**Response Time Distribution:**

For FCFS discipline, the response time has probability density function:

$$f_W(t) = (\mu - \lambda)e^{-(\mu-\lambda)t}, \quad t > 0$$

This is an **exponential distribution** with rate (μ - λ).

#### 13.1.4 Little's Law

**Little's Law** is a fundamental result relating average quantities:

$$L = \lambda W$$

In words: *Average number in system = Arrival rate × Average time in system*

**Proof Sketch:**
- Consider time interval [0, T]
- Let A(T) = number of arrivals, D(T) = number of departures
- Let $\int_0^T N(t) dt$ = total customer-time in system
- As T → ∞: $L = \lim_{T \to \infty} \frac{1}{T} \int_0^T N(t) dt$ and $\lambda = \lim_{T \to \infty} \frac{A(T)}{T}$
- Total time = $\int_0^T N(t) dt \approx A(T) \cdot W$
- Therefore: $L = \lambda W$

**Extensions:**
- Applies to **any stable queueing system** (not just M/M/1)
- Applies to sub-systems: $L_q = \lambda W_q$
- Operational form: requires only time-average quantities

#### 13.1.5 Busy Period Analysis

The **busy period** is the duration from when a customer arrives to an empty system until the system becomes empty again.

**Laplace Transform** of busy period B:

$$E[e^{-sB}] = \frac{1}{2\lambda}\left(\lambda + \mu + s - \sqrt{(\lambda + \mu + s)^2 - 4\lambda\mu}\right)$$

**Mean Busy Period:**

$$E[B] = \frac{1}{\mu - \lambda}$$

This equals the mean response time W!

**Variance:**

$$\text{Var}(B) = \frac{1}{\mu^2(1-\rho)^2}$$

### 13.2 Extended Kalman Filter: Detailed Foundations

The Extended Kalman Filter extends optimal linear estimation to nonlinear systems through linearization. This section provides comprehensive mathematical foundations based on Jazwinski (1970) and Wikipedia.

**Reference**: [Extended Kalman Filter - Wikipedia](https://en.wikipedia.org/wiki/Extended_Kalman_filter)

#### 13.2.1 Nonlinear System Model

**State Space Representation:**

$$\mathbf{x}_k = f(\mathbf{x}_{k-1}, \mathbf{u}_{k-1}) + \mathbf{w}_{k-1}$$
$$\mathbf{z}_k = h(\mathbf{x}_k) + \mathbf{v}_k$$

Where:
- $\mathbf{x}_k \in \mathbb{R}^n$: State vector at time k
- $\mathbf{u}_k \in \mathbb{R}^l$: Control input (optional)
- $f: \mathbb{R}^n \times \mathbb{R}^l \to \mathbb{R}^n$: Nonlinear state transition function
- $\mathbf{w}_k \sim \mathcal{N}(\mathbf{0}, \mathbf{Q}_k)$: Process noise (Gaussian, zero-mean)
- $\mathbf{z}_k \in \mathbb{R}^m$: Measurement vector
- $h: \mathbb{R}^n \to \mathbb{R}^m$: Nonlinear observation function
- $\mathbf{v}_k \sim \mathcal{N}(\mathbf{0}, \mathbf{R}_k)$: Measurement noise (Gaussian, zero-mean)

**Key Assumptions:**
- Process and measurement noise are independent
- Noise processes are white (uncorrelated in time)
- Functions f and h are **differentiable**

#### 13.2.2 Linearization via Taylor Series

The EKF linearizes the nonlinear functions f and h around the current estimate using **first-order Taylor expansions**.

**State Transition Linearization:**

$$f(\mathbf{x}_{k-1}) \approx f(\hat{\mathbf{x}}_{k-1|k-1}) + \mathbf{F}_{k-1}(\mathbf{x}_{k-1} - \hat{\mathbf{x}}_{k-1|k-1})$$

Where the **Jacobian** $\mathbf{F}_{k-1}$ is:

$$\mathbf{F}_{k-1} = \left.\frac{\partial f}{\partial \mathbf{x}}\right|_{\mathbf{x} = \hat{\mathbf{x}}_{k-1|k-1}}$$

**Element-wise:**

$$[\mathbf{F}_{k-1}]_{ij} = \frac{\partial f_i}{\partial x_j}\bigg|_{\mathbf{x} = \hat{\mathbf{x}}_{k-1|k-1}}$$

**Observation Linearization:**

$$h(\mathbf{x}_k) \approx h(\hat{\mathbf{x}}_{k|k-1}) + \mathbf{H}_k(\mathbf{x}_k - \hat{\mathbf{x}}_{k|k-1})$$

Where the **observation Jacobian** $\mathbf{H}_k$ is:

$$\mathbf{H}_k = \left.\frac{\partial h}{\partial \mathbf{x}}\right|_{\mathbf{x} = \hat{\mathbf{x}}_{k|k-1}}$$

#### 13.2.3 EKF Algorithm (Discrete-Time)

**Initialization:**
$$\hat{\mathbf{x}}_{0|0} = E[\mathbf{x}_0]$$
$$\mathbf{P}_{0|0} = E[(\mathbf{x}_0 - \hat{\mathbf{x}}_{0|0})(\mathbf{x}_0 - \hat{\mathbf{x}}_{0|0})^T]$$

**Prediction Step:**

1. **State prediction:**
$$\hat{\mathbf{x}}_{k|k-1} = f(\hat{\mathbf{x}}_{k-1|k-1}, \mathbf{u}_{k-1})$$

2. **Jacobian computation:**
$$\mathbf{F}_{k-1} = \left.\frac{\partial f}{\partial \mathbf{x}}\right|_{\hat{\mathbf{x}}_{k-1|k-1}, \mathbf{u}_{k-1}}$$

3. **Covariance prediction:**
$$\mathbf{P}_{k|k-1} = \mathbf{F}_{k-1} \mathbf{P}_{k-1|k-1} \mathbf{F}_{k-1}^T + \mathbf{Q}_{k-1}$$

**Update Step:**

1. **Observation Jacobian:**
$$\mathbf{H}_k = \left.\frac{\partial h}{\partial \mathbf{x}}\right|_{\hat{\mathbf{x}}_{k|k-1}}$$

2. **Innovation (measurement residual):**
$$\mathbf{y}_k = \mathbf{z}_k - h(\hat{\mathbf{x}}_{k|k-1})$$

3. **Innovation covariance:**
$$\mathbf{S}_k = \mathbf{H}_k \mathbf{P}_{k|k-1} \mathbf{H}_k^T + \mathbf{R}_k$$

4. **Kalman gain:**
$$\mathbf{K}_k = \mathbf{P}_{k|k-1} \mathbf{H}_k^T \mathbf{S}_k^{-1}$$

5. **State update:**
$$\hat{\mathbf{x}}_{k|k} = \hat{\mathbf{x}}_{k|k-1} + \mathbf{K}_k \mathbf{y}_k$$

6. **Covariance update (Joseph form for numerical stability):**
$$\mathbf{P}_{k|k} = (\mathbf{I} - \mathbf{K}_k \mathbf{H}_k) \mathbf{P}_{k|k-1} (\mathbf{I} - \mathbf{K}_k \mathbf{H}_k)^T + \mathbf{K}_k \mathbf{R}_k \mathbf{K}_k^T$$

Or simplified form:
$$\mathbf{P}_{k|k} = (\mathbf{I} - \mathbf{K}_k \mathbf{H}_k) \mathbf{P}_{k|k-1}$$

#### 13.2.4 Optimality and Limitations

**Optimality:**
- EKF is **optimal for linear systems** (reduces to standard Kalman filter)
- For nonlinear systems, EKF is **suboptimal** due to linearization errors
- Provides **minimum mean squared error (MMSE) estimate** only at linearization point

**Limitations:**
1. **Linearization errors**: Taylor expansion truncates higher-order terms
   - Accuracy depends on degree of nonlinearity
   - May diverge if linearization is poor
   
2. **Covariance underestimation**: Tends to be **overconfident**
   - True uncertainty often larger than $\mathbf{P}_k$
   - Can cause filter divergence
   - Mitigation: add "stabilizing noise" to Q
   
3. **Computational cost**: Requires Jacobian computation at each step
   - Analytical derivatives: error-prone for complex f, h
   - Numerical derivatives: computationally expensive
   
4. **Local linearization**: Assumes small deviations from nominal trajectory
   - Poor performance for highly nonlinear systems
   - May miss multiple modes in posterior distribution

**Alternatives:**
- **Unscented Kalman Filter (UKF)**: Uses deterministic sampling (sigma points) instead of linearization
- **Particle Filter**: Monte Carlo approach for arbitrary distributions
- **Iterated EKF**: Iteratively refines linearization point in update step

#### 13.2.5 Chi-Squared Test for Innovation

The **innovation sequence** $\{\mathbf{y}_k\}$ provides a statistical test for filter consistency.

**Theory:** Under the assumption that:
- The model is correct (f and h accurately represent the system)
- The noise statistics (Q and R) are correct
- The filter has converged

The **Normalized Innovation Squared (NIS)** statistic:

$$\epsilon_k = \mathbf{y}_k^T \mathbf{S}_k^{-1} \mathbf{y}_k$$

follows a **chi-squared distribution** with m degrees of freedom (measurement dimension):

$$\epsilon_k \sim \chi^2_m$$

**Chi-Squared Distribution:**

The probability density function is:

$$f(\chi^2; m) = \frac{1}{2^{m/2}\Gamma(m/2)} (\chi^2)^{m/2-1} e^{-\chi^2/2}$$

**Properties:**
- Mean: $E[\chi^2_m] = m$
- Variance: $\text{Var}(\chi^2_m) = 2m$

**Confidence Intervals:**

For m = 2 (as in WVA with [TTFT, ITL] observations):

| Confidence Level | Critical Value $\chi^2_{m, \alpha}$ |
|-----------------|-------------------------------------|
| 90%             | 4.605                              |
| 95%             | 5.991                              |
| 97.5%           | 7.378                              |
| 99%             | 9.210                              |

**Reference**: [Chi-Squared Distribution - Wikipedia](https://en.wikipedia.org/wiki/Chi-squared_distribution)

**Validation Test:**

If $\epsilon_k > \chi^2_{m, \alpha}$, reject the measurement as an **outlier** with confidence level α.

WVA uses the threshold **7.378** (97.5% confidence), meaning:
- If innovation is consistent with model: Accept with 97.5% probability
- If innovation suggests model mismatch: Reject with 2.5% false alarm rate

**Average NIS Test:**

Over N observations, the average should satisfy:

$$\bar{\epsilon} = \frac{1}{N}\sum_{k=1}^N \epsilon_k \approx m$$

If $\bar{\epsilon} \ll m$: Filter is **overconfident** (P too small, R too large)

If $\bar{\epsilon} \gg m$: Filter is **underconfident** (P too large, R too small) or model mismatch

### 13.3 Little's Law: Generalization and Proofs

**Little's Law** is one of the most fundamental results in queueing theory, applicable far beyond M/M/1 queues.

#### 13.3.1 General Statement

For any **stable queueing system**:

$$L = \lambda W$$

Where:
- $L$: Time-average number of customers in system
- $\lambda$: Time-average arrival rate
- $W$: Average time a customer spends in system

**Operational Form** (no probabilistic assumptions):

$$\bar{N} = \bar{\Lambda} \bar{T}$$

Where:
- $\bar{N} = \lim_{t \to \infty} \frac{1}{t} \int_0^t N(s) ds$: Time-average number
- $\bar{\Lambda} = \lim_{t \to \infty} \frac{A(t)}{t}$: Time-average arrival rate
- $\bar{T}$: Average time per customer

**Generality:**
- Works for **any queueing discipline** (FCFS, LIFO, PS, SRPT, etc.)
- Works for **any arrival process** (Poisson, deterministic, bursty, etc.)
- Works for **any service distribution**
- Works for **open or closed systems**
- Only requires **stability** (arrival rate < service rate)

#### 13.3.2 Applications to Sub-Systems

Little's Law applies to **any part of the system**:

**Queue only:**
$$L_q = \lambda W_q$$

**Server only:**
$$L_s = \lambda W_s$$

**Per-class in multi-class systems:**
$$L_i = \lambda_i W_i$$

**Example for WVA:**

If TTFT = $W_q + T_{prefill}$ and ITL = $T_{decode}$:

$$L_{queue} = \lambda \cdot W_q$$
$$L_{prefill} = \lambda \cdot T_{prefill}$$
$$L_{decode} = \lambda \cdot (m-1) \cdot T_{decode}$$

Where m is average output tokens.

### 13.4 Birth-Death Processes and Balance Equations

The M/M/1/K queue with state-dependent service rates is a **birth-death process**.

**General Birth-Death Process:**

States: $\{0, 1, 2, \ldots, N\}$ (finite or infinite)

Birth rates: $\lambda_n$ (transition from n to n+1)

Death rates: $\mu_n$ (transition from n to n-1)

**Global Balance Equations:**

For each state n, the flow in equals flow out:

$$\pi_{n-1} \lambda_{n-1} + \pi_{n+1} \mu_{n+1} = \pi_n (\lambda_n + \mu_n)$$

**Detailed Balance (Reversibility):**

For birth-death processes:

$$\pi_n \lambda_n = \pi_{n+1} \mu_{n+1}$$

**Solution (Recursive):**

$$\pi_n = \pi_0 \prod_{i=1}^{n} \frac{\lambda_{i-1}}{\mu_i}$$

**Normalization:**

$$\pi_0 = \frac{1}{1 + \sum_{n=1}^{N} \prod_{i=1}^{n} \frac{\lambda_{i-1}}{\mu_i}}$$

**Application to M/M/1/K with State-Dependent Service:**

For WVA's queueing model:
- $\lambda_n = \lambda$ for $n < K$ (constant arrival rate)
- $\lambda_n = 0$ for $n \geq K$ (finite buffer blocks arrivals)
- $\mu_n = \mu(n)$ as defined by the service rate formula

This produces the steady-state probabilities used in performance metric calculations.

---

### 12.1 Advanced Filtering Techniques

**Unscented Kalman Filter (UKF):**
- Better handling of nonlinearity
- Sigma point sampling instead of linearization
- Potentially higher accuracy for queueing model

**Adaptive filtering:**
- Online estimation of Q and R
- Detecting and adapting to model changes
- Outlier rejection

### 12.2 Enhanced Queueing Models

**More realistic assumptions:**
- Non-exponential service time distributions (phase-type, etc.)
- Request size variability
- Priority queueing for different service classes

**Multi-server models:**
- M/M/c/K for multi-GPU setups
- Load balancing effects

### 12.3 Optimization Improvements

**Exact methods:**
- Mixed-integer programming formulation
- Branch-and-bound or cutting plane methods
- Stronger optimality guarantees

**Online learning:**
- Reinforcement learning for allocation decisions
- Learning from past allocation outcomes
- Exploration vs. exploitation trade-offs

### 12.4 System-Level Extensions

**Hierarchical control:**
- Cluster-level optimization
- Regional resource allocation
- Multi-tenant considerations

**Fault tolerance:**
- Handling server failures
- Graceful degradation
- Recovery strategies

**Energy optimization:**
- Power-aware scheduling
- GPU power capping
- Carbon-aware allocation

---

## 13. Conclusion

This document has presented the theoretical and practical foundations of a system that integrates Extended Kalman Filtering with queueing-theoretic models for adaptive autoscaling of LLM inference servers. The key contributions of this approach are:

1. **Dynamic parameter estimation**: The EKF continuously adapts queueing model parameters (α, β) to match real system behavior, overcoming the challenge of unknown and time-varying service characteristics.

2. **Principled uncertainty quantification**: The Kalman framework provides not just parameter estimates but also confidence bounds (covariance matrix P), enabling risk-aware decision making.

3. **Integration of estimation and optimization**: The tuner provides updated models to the autoscaler, creating a closed-loop adaptive system that responds to workload changes.

4. **Practical implementation**: The system is modular, extensible, and suitable for production deployment with real inference servers.

The mathematical foundation spans:
- **Queueing theory**: M/M/1/K models with state-dependent service rates
- **Kalman filtering**: Extended Kalman Filter for nonlinear state estimation  
- **Optimization**: Greedy algorithms with priority-based allocation
- **Numerical methods**: Finite difference Jacobians, binary search, matrix operations

This work bridges the gap between classical queueing theory and modern machine learning serving infrastructure, providing a systematic approach to resource management under uncertainty. The combination of model-based prediction (queueing theory) with adaptive estimation (Kalman filtering) represents a powerful paradigm for managing complex, dynamic systems.

Future work can extend this foundation in multiple directions: more sophisticated filtering techniques, richer queueing models, advanced optimization methods, and system-level enhancements for fault tolerance and energy efficiency. The framework presented here provides a solid basis for such extensions while remaining practical and deployable in real-world settings.

---

## 14. References

### 14.1 Queueing Theory

1. **Kleinrock, L. (1975)**. *Queueing Systems, Volume 1: Theory*. New York: Wiley-Interscience.
   - Comprehensive foundation of queueing theory
   - M/M/1, M/M/c, M/G/1 queue analysis
   - Birth-death processes and limiting distributions

2. **Kleinrock, L. (1976)**. *Queueing Systems, Volume 2: Computer Applications*. New York: Wiley-Interscience.
   - Application to computer networks
   - Time-sharing systems and scheduling disciplines
   - Priority queueing

3. **Bolch, G., Greiner, S., de Meer, H., & Trivedi, K. S. (2006)**. *Queueing Networks and Markov Chains* (2nd ed.). Hoboken, NJ: Wiley.
   - Queueing networks and product-form solutions
   - BCMP theorem and Jackson networks
   - Numerical methods for complex networks

4. **Gross, D., Shortle, J. F., Thompson, J. M., & Harris, C. M. (2008)**. *Fundamentals of Queueing Theory* (4th ed.). Hoboken, NJ: Wiley.
   - Finite capacity queues (M/M/1/K)
   - State-dependent service rates
   - Erlang formulas and trunk reservation

5. **Harrison, P., & Patel, N. M. (1992)**. *Performance Modelling of Communication Networks and Computer Architectures*. Boston, MA: Addison-Wesley.
   - Response time distributions in queueing networks
   - Processor sharing disciplines
   - Spectral expansion methods

6. **Little, J. D. C. (1961)**. "A Proof for the Queuing Formula: L = λW". *Operations Research*, 9(3), 383-387.
   - Original proof of Little's Law
   - Operational analysis framework

### 14.2 Kalman Filtering and State Estimation

7. **Kalman, R. E. (1960)**. "A New Approach to Linear Filtering and Prediction Problems". *Journal of Basic Engineering*, 82(1), 35-45.
   - Original Kalman filter paper
   - Recursive optimal estimation
   - Discrete-time formulation

8. **Kalman, R. E., & Bucy, R. S. (1961)**. "New Results in Linear Filtering and Prediction Theory". *Journal of Basic Engineering*, 83(1), 95-108.
   - Continuous-time Kalman-Bucy filter
   - Riccati equation
   - Optimality proofs

9. **Jazwinski, A. H. (1970)**. *Stochastic Processes and Filtering Theory*. New York: Academic Press.
   - Extended Kalman Filter foundations
   - Nonlinear filtering theory
   - Convergence and stability analysis

10. **Anderson, B. D. O., & Moore, J. B. (1979)**. *Optimal Filtering*. Englewood Cliffs, NJ: Prentice-Hall.
    - Theoretical foundations of optimal estimation
    - Observability and controllability
    - Discrete Algebraic Riccati Equation

11. **Maybeck, P. S. (1979)**. *Stochastic Models, Estimation, and Control, Volume 1*. Mathematics in Science and Engineering, Vol. 141. New York: Academic Press.
    - Practical EKF implementation
    - Navigation and guidance applications
    - Numerical considerations

12. **Bar-Shalom, Y., Li, X. R., & Kirubarajan, T. (2001)**. *Estimation with Applications to Tracking and Navigation*. New York: Wiley.
    - Multi-target tracking
    - Data association algorithms
    - Innovation analysis and validation

13. **Simon, D. (2006)**. *Optimal State Estimation: Kalman, H∞, and Nonlinear Approaches*. Hoboken, NJ: Wiley.
    - Comprehensive treatment of nonlinear filters
    - Unscented Kalman Filter (UKF)
    - Particle filters
    - Numerical Jacobian computation

14. **Einicke, G. A. (2019)**. *Smoothing, Filtering and Prediction* (2nd ed.). Amazon Prime Publishing.
    - Modern numerical algorithms
    - Square-root filtering
    - Adaptive filtering techniques

15. **Mehra, R. K. (1970)**. "On the Identification of Variances and Adaptive Kalman Filtering". *IEEE Transactions on Automatic Control*, 15(2), 175-184.
    - Adaptive noise covariance estimation
    - Chi-squared innovation tests
    - Filter consistency checks

### 14.3 System Identification and Adaptive Control

16. **Ljung, L. (1999)**. *System Identification: Theory for the User* (2nd ed.). Upper Saddle River, NJ: Prentice-Hall.
    - Parameter estimation for dynamical systems
    - Persistent excitation and identifiability
    - Convergence analysis

17. **Åström, K. J., & Wittenmark, B. (1995)**. *Adaptive Control* (2nd ed.). Reading, MA: Addison-Wesley.
    - Self-tuning regulators
    - Model reference adaptive control
    - Stability and robustness

### 14.4 LLM Inference Systems

18. **Kwon, W., Li, Z., Zhuang, S., Sheng, Y., Zheng, L., Yu, C. H., Gonzalez, J., Zhang, H., & Stoica, I. (2023)**. "Efficient Memory Management for Large Language Model Serving with PagedAttention". In *Proceedings of the 29th Symposium on Operating Systems Principles (SOSP '23)*. ACM.
    - vLLM system architecture
    - PagedAttention for KV cache management
    - Continuous batching implementation

19. **Yu, G. I., Jeong, J. S., Kim, G. W., Kim, S., & Chun, B. G. (2022)**. "Orca: A Distributed Serving System for Transformer-Based Generative Models". In *Proceedings of the 16th USENIX Symposium on Operating Systems Design and Implementation (OSDI '22)*. USENIX Association.
    - Iteration-level scheduling
    - Selective batching for latency optimization

20. **Wu, X., Duan, Q., Zheng, Z., & Yang, Z. (2023)**. "FastServe: Fast Distributed Inference Serving for Large Language Models". *arXiv preprint arXiv:2305.05920*.
    - Preemptive scheduling with skip-join MLFQ
    - Job size prediction
    - Priority queue management

21. **Agrawal, A., Kedia, N., Panwar, A., Mohan, J., Kwatra, N., Gulavani, B., Tumanov, A., & Ramjee, R. (2024)**. "Taming Throughput-Latency Tradeoff in LLM Inference with Sarathi-Serve". In *Proceedings of the 18th USENIX Symposium on Operating Systems Design and Implementation (OSDI '24)*.
    - Chunked-prefill scheduling
    - Stall-free batched inference
    - Throughput-latency optimization

22. **Griggs, A., et al. (2024)**. "Shepherd: Optimizing LLM Serving with Efficient Co-location".
    - Interference modeling for co-located models
    - Performance isolation techniques
    - Resource sharing strategies

### 14.5 Performance Modeling and Benchmarking

23. **Casson, A. (2023)**. "Understanding Transformer Inference Performance: A FLOPS-Based Analysis".
    - FLOPS compute formulas for transformers
    - Memory bandwidth bottlenecks
    - Roofline model for LLMs

24. **Yang, L., et al. (2024)**. "Hybrid Prefill-Decode Architecture for LLM Inference".
    - Analytical models for P/D separation
    - Resource allocation strategies
    - Performance prediction

25. **Yuan, Z., et al. (2024)**. "LLM Inference Throughput Optimization via Queueing Theory".
    - Batch size optimization models
    - Latency-throughput trade-offs
    - Stability analysis

26. **Zhu, Q., et al. (2025)**. "Predictive Autoscaling for LLM Workloads".
    - Time-series forecasting methods
    - Proactive vs. reactive scaling
    - Workload prediction accuracy

### 14.6 Kubernetes and Cloud Native Systems

27. **Burns, B., Grant, B., Oppenheimer, D., Brewer, E., & Wilkes, J. (2016)**. "Borg, Omega, and Kubernetes". *ACM Queue*, 14(1), 70-93.
    - Evolution from Borg to Kubernetes
    - Control loops and reconciliation
    - Resource management at scale

28. **Dobies, J., & Wood, J. (2020)**. *Kubernetes Operators: Automating the Container Orchestration Platform*. Sebastopol, CA: O'Reilly Media.
    - Operator pattern fundamentals
    - Custom Resource Definitions (CRDs)
    - Controller implementation best practices

29. **CNCF Operator White Paper (2021)**. "Kubernetes Operators". Cloud Native Computing Foundation.
    - Operator pattern standardization
    - Lifecycle management
    - Best practices and anti-patterns

30. **Kubernetes Documentation**. "Controllers". Retrieved from https://kubernetes.io/docs/concepts/architecture/controller/
    - Official Kubernetes controller documentation
    - Control loop theory
    - API server integration

31. **Kubernetes Documentation**. "Operator Pattern". Retrieved from https://kubernetes.io/docs/concepts/extend-kubernetes/operator/
    - Official operator pattern guide
    - Custom resource management
    - Extension patterns

### 14.7 Optimization and Resource Allocation

32. **Papadimitriou, C. H., & Steiglitz, K. (1998)**. *Combinatorial Optimization: Algorithms and Complexity*. Mineola, NY: Dover Publications.
    - Greedy algorithms and approximation ratios
    - NP-completeness theory
    - Bin packing and scheduling problems

33. **Vazirani, V. V. (2001)**. *Approximation Algorithms*. Berlin: Springer-Verlag.
    - Approximation algorithm design
    - Priority-based heuristics
    - Performance guarantees

34. **Delimitrou, C., & Kozyrakis, C. (2014)**. "Quasar: Resource-Efficient and QoS-Aware Cluster Management". In *Proceedings of the 19th International Conference on Architectural Support for Programming Languages and Operating Systems (ASPLOS '14)*. ACM.
    - ML-based resource allocation
    - QoS prediction and interference
    - Collaborative filtering for performance modeling

### 14.8 Online References

35. **Wikipedia Contributors**. "M/M/1 Queue". *Wikipedia, The Free Encyclopedia*. Retrieved from https://en.wikipedia.org/wiki/M/M/1_queue
    - M/M/1 queue mathematical foundations
    - Stationary distribution and performance metrics
    - Busy period analysis

36. **Wikipedia Contributors**. "Extended Kalman Filter". *Wikipedia, The Free Encyclopedia*. Retrieved from https://en.wikipedia.org/wiki/Extended_Kalman_filter
    - EKF algorithm and formulation
    - Linearization techniques
    - Limitations and alternatives

37. **Wikipedia Contributors**. "Chi-Squared Distribution". *Wikipedia, The Free Encyclopedia*. Retrieved from https://en.wikipedia.org/wiki/Chi-squared_distribution
    - Chi-squared probability distribution
    - Critical values and confidence intervals
    - Applications in hypothesis testing

38. **Wikipedia Contributors**. "Little's Law". *Wikipedia, The Free Encyclopedia*. Retrieved from https://en.wikipedia.org/wiki/Little%27s_law
    - Little's Law statement and proof
    - Operational analysis
    - Applications to queueing systems

### 14.9 Software and Tools

39. **llm-inferno/kalman-filter** (v0.1.0). Go package for Extended Kalman Filtering. Retrieved from https://github.com/llm-inferno/kalman-filter
    - Implementation used in WVA tuner
    - Numerical Jacobian computation
    - Discrete-time EKF algorithm

40. **Kubernetes controller-runtime**. Go library for building Kubernetes controllers. Retrieved from https://pkg.go.dev/sigs.k8s.io/controller-runtime
    - Manager and client abstractions
    - Watch and event mechanisms
    - Leader election implementation

41. **Kubebuilder**. SDK for building Kubernetes APIs using CRDs. Retrieved from https://book.kubebuilder.io/
    - Scaffolding tools for operators
    - RBAC marker generation
    - Webhook integration

### 14.10 Historical and Foundational Works

42. **Erlang, A. K. (1909)**. "The Theory of Probabilities and Telephone Conversations". *Nyt Tidsskrift for Matematik B*, 20, 33-39.
    - Foundation of queueing theory
    - Erlang formulas for telephony

43. **Kendall, D. G. (1953)**. "Stochastic Processes Occurring in the Theory of Queues and their Analysis by the Method of the Imbedded Markov Chain". *The Annals of Mathematical Statistics*, 24(3), 338-354.
    - Kendall's notation (A/B/c)
    - Embedded Markov chain analysis

44. **Kingman, J. F. C. (1961)**. "The Single Server Queue in Heavy Traffic". *Mathematical Proceedings of the Cambridge Philosophical Society*, 57(4), 902-904.
    - Heavy traffic approximations
    - Reflected Brownian motion
    - Diffusion limits

45. **Wiener, N. (1949)**. *Extrapolation, Interpolation, and Smoothing of Stationary Time Series*. Cambridge, MA: MIT Press.
    - Foundation of optimal filtering
    - Wiener filter for stationary processes

46. **Brown, R. G., & Hwang, P. Y. C. (1997)**. *Introduction to Random Signals and Applied Kalman Filtering* (3rd ed.). New York: John Wiley & Sons.
    - Practical Kalman filtering guide
    - Continuous-time formulations
    - Applications to navigation systems

---

## 15. Appendices

### Appendix A: Notation and Symbols

**Queueing Theory:**
- λ: Arrival rate (requests per second or per minute)
- μ: Service rate (requests per second)
- μ(n): State-dependent service rate for batch size n
- ρ: Utilization (λ/μ)
- N: Number of customers in system (random variable)
- K: Maximum queue capacity
- L: Average number in system (E[N])
- W: Average response time (time in system)
- Wq: Average waiting time (time in queue)
- TTFT: Time to First Token (milliseconds)
- ITL: Inter-Token Latency (milliseconds)
- TPS: Tokens Per Second
- k: Average input tokens per request
- m: Average output tokens per request

**Service Time Parameters:**
- α (alpha): Base decode time per token (ms)
- β (beta): Decode time scaling factor with batch size (ms)
- γ (gamma): Base prefill overhead (ms)
- δ (delta): Prefill per-token-per-batch scaling factor (ms)

**Kalman Filter:**
- x: State vector
- z: Measurement (observation) vector
- u: Control input vector
- f(·): Nonlinear state transition function
- h(·): Nonlinear observation function
- F: State transition Jacobian matrix
- H: Observation Jacobian matrix
- P: State error covariance matrix
- Q: Process noise covariance matrix
- R: Measurement noise covariance matrix
- K: Kalman gain matrix
- y: Innovation (measurement residual)
- S: Innovation covariance matrix
- w: Process noise (Gaussian)
- v: Measurement noise (Gaussian)

**Subscripts:**
- k: Time index
- k|k: Estimate at time k given measurements up to k (posterior)
- k|k-1: Prediction at time k given measurements up to k-1 (prior)

**Mathematical Notation:**
- E[·]: Expected value (mean)
- Var(·): Variance
- Cov(·,·): Covariance
- ||·||: Euclidean norm
- ⊗: Kronecker product
- ∇: Gradient operator
- ∂f/∂x: Partial derivative of f with respect to x
- ~: Distributed as (e.g., x ~ N(0, 1))
- N(μ, Σ): Gaussian (normal) distribution with mean μ and covariance Σ
- χ²_m: Chi-squared distribution with m degrees of freedom

### Appendix B: Algorithm Pseudocode

**Algorithm 1: Extended Kalman Filter**

```
Initialize:
  x_0|0 ← initial state estimate
  P_0|0 ← initial covariance estimate
  
For k = 1, 2, 3, ...:
  
  // Prediction Step
  x_k|k-1 ← f(x_k-1|k-1, u_k-1)
  F_k-1 ← ∂f/∂x | x_k-1|k-1
  P_k|k-1 ← F_k-1 * P_k-1|k-1 * F_k-1^T + Q_k-1
  
  // Update Step (when measurement z_k arrives)
  H_k ← ∂h/∂x | x_k|k-1
  y_k ← z_k - h(x_k|k-1)
  S_k ← H_k * P_k|k-1 * H_k^T + R_k
  K_k ← P_k|k-1 * H_k^T * S_k^(-1)
  x_k|k ← x_k|k-1 + K_k * y_k
  P_k|k ← (I - K_k * H_k) * P_k|k-1
  
  // Validation (optional)
  NIS ← y_k^T * S_k^(-1) * y_k
  If NIS > threshold:
    Reject update, keep x_k|k-1 and P_k|k-1
```

**Algorithm 2: Queueing Model Solution (M/M/1/K State-Dependent)**

```
Input: λ (arrival rate), μ(n) (service rates), K (capacity)
Output: Performance metrics (L, W, TTFT, ITL)

// Compute stationary probabilities
Initialize: p[0] ← 1.0
For n = 1 to K:
  If n ≤ K:
    p[n] ← p[n-1] * (λ / μ(n))
  Else:
    p[n] ← 0
    
// Normalization
sum ← Σ p[n] for n = 0 to K
For n = 0 to K:
  p[n] ← p[n] / sum
  
// Performance metrics
L ← Σ n * p[n] for n = 0 to K
W ← L / λ_effective
W_q ← W - (1 / average_service_rate)

// LLM-specific metrics
TTFT ← W_q + average_prefill_time
ITL ← average_decode_time

Return {L, W, W_q, TTFT, ITL}
```

**Algorithm 3: Greedy Allocation with Priorities**

```
Input: Servers S with priorities, Accelerator inventory A
Output: Allocation map

// Initialize
For each server s in S:
  Compute candidate allocations for each accelerator type
  Sort by value (cost + transition penalty)
  Add to priority queue Q with (priority, delta)
  
// Greedy selection
While Q not empty:
  s ← Pop server with highest priority, lowest delta
  a ← Current allocation option for s
  
  If resources available for a:
    Assign allocation a to server s
    Update available resources
    Remove s from Q
  Else:
    Move to next allocation option for s
    Recompute delta
    Re-insert s into Q with updated delta
    
// Best-effort for unallocated
For each unallocated server s (by priority):
  Find any available accelerators
  Assign if possible
  
Return allocation map
```

### Appendix C: Configuration Examples

**Example 1: VariantAutoscaling Custom Resource**

```yaml
apiVersion: llmd.ai/v1alpha1
kind: VariantAutoscaling
metadata:
  name: llama3-8b-variant
  namespace: default
spec:
  modelID: "meta-llama/Llama-3-8B"
  sloClassRef:
    name: service-classes-config
    key: premium
  modelProfile:
    accelerators:
    - acc: "L40S"
      accCount: 1
      perfParms:
        decodeParms:
          alpha: "5.2"
          beta: "0.8"
        prefillParms:
          gamma: "10.5"
          delta: "0.05"
      maxBatchSize: 128
    - acc: "H100"
      accCount: 1
      perfParms:
        decodeParms:
          alpha: "3.5"
          beta: "0.6"
        prefillParms:
          gamma: "7.2"
          delta: "0.03"
      maxBatchSize: 256
```

**Example 2: Service Class Configuration**

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: service-classes-config
  namespace: workload-variant-autoscaler-system
data:
  premium: |
    {
      "name": "premium",
      "priority": 1,
      "ttft_ms": 500,
      "itl_ms": 50,
      "tps": 100
    }
  standard: |
    {
      "name": "standard",
      "priority": 2,
      "ttft_ms": 1000,
      "itl_ms": 100,
      "tps": 50
    }
  best-effort: |
    {
      "name": "best-effort",
      "priority": 3,
      "ttft_ms": 2000,
      "itl_ms": 200,
      "tps": 25
    }
```

**Example 3: Tuner Configuration**

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: tuner-config
  namespace: workload-variant-autoscaler-system
data:
  tuning-enabled: "true"
  tuning-period: "30s"
  max-nis-threshold: "7.378"
  percent-change: "0.05"
  error-level: "0.05"
  state-constraints: |
    {
      "alpha": {"min": 0.1, "max": 100.0},
      "beta": {"min": 0.01, "max": 10.0},
      "gamma": {"min": 0.1, "max": 100.0},
      "delta": {"min": 0.001, "max": 1.0}
    }
```

---

**Document Version**: 2.0  
**Last Updated**: November 2, 2025  
**Authors**: Workload Variant Autoscaler Project Contributors  
**License**: Apache 2.0

---