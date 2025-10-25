package analyzer

import (
	"bytes"

	"github.com/llm-d-incubation/workload-variant-autoscaler/internal/logger"
)

// M/D/1 model with state-dependent service rate
// This models batching where service rate depends on the number of requests being processed
type MD1ModelStateDependent struct {
	MD1KModel                 // extends base class
	servRate        []float32 // state-dependent service rate
	avgNumInServers float32
}

func NewMD1ModelStateDependent(K int, servRate []float32) *MD1ModelStateDependent {
	m := MD1ModelStateDependent{
		MD1KModel:       *NewMD1KModel(K),
		servRate:        servRate,
		avgNumInServers: 0,
	}

	m.QueueModel.ComputeRho = m.ComputeRho
	m.QueueModel.computeStatistics = m.computeStatistics
	return &m
}

// Solve queueing model given arrival and service rates
func (m *MD1ModelStateDependent) Solve(lambda float32, mu float32) {
	m.MD1KModel.Solve(lambda, mu)
}

// Compute utilization of queueing model
func (m *MD1ModelStateDependent) ComputeRho() float32 {
	return 1 - float32(m.p[0])
}

// Evaluate performance measures of queueing model
func (m *MD1ModelStateDependent) computeStatistics() {
	if !m.isValid {
		return
	}

	// M/D/1/K Model with state-dependent service rates
	//
	// Approach:
	// 1. Use balance equations for state probabilities (same as M/M/1/K)
	// 2. Apply Pollaczek-Khintchine correction for waiting time
	//
	// Key insight: For Poisson arrivals, state probabilities depend only on
	// arrival rate and service rate balance (λ vs μ), not service distribution.
	// The service distribution (exponential vs deterministic) affects waiting
	// time through the coefficient of variation.
	//
	// Pollaczek-Khintchine formula: W_q = (ρ/(1-ρ)) × ((C_a² + C_s²)/2) × E[S]
	// - M/M/1: C_s = 1 → W_M = ρτ/(1-ρ)
	// - M/D/1: C_s = 0 → W_D = ρτ/(2(1-ρ)) = 0.5 × W_M
	//
	// References:
	// - Gross & Harris (1998): Section 2.5 on M/G/1 queues
	// - Kleinrock (1975): Vol 1, Section 5.3, Pollaczek-Khintchine formula

	logger.Log.Debugw("[MD1K_COMPUTE] Computing statistics",
		"lambda", m.lambda,
		"mu", m.mu)

	// Step 1: Compute state probabilities using balance equations
	m.computeProbabilitiesApprox()

	// Step 2: Calculate avgNumInServers and avgNumInSystem
	num := len(m.servRate)
	var avgNumInServers float64
	var avgNumInSystem float64
	sumP := m.p[0]
	for i := 1; i <= m.K; i++ {
		avgNumInSystem += float64(i) * m.p[i]
		sumP += m.p[i]
		if i == num {
			avgNumInServers = avgNumInSystem + (1-sumP)*float64(num)
		}
	}
	m.avgNumInServers = float32(avgNumInServers)
	m.avgNumInSystem = float32(avgNumInSystem)

	// Step 3: Throughput
	m.throughput = m.lambda * (1 - float32(m.p[m.K]))

	// Step 4: Calculate response and service times using Little's Law
	m.avgRespTime = m.avgNumInSystem / m.throughput
	m.avgServTime = m.avgNumInServers / m.throughput

	// Step 5: Calculate waiting time as if M/M/1/K
	waitTimeAsIfMM1K := m.avgRespTime - m.avgServTime
	logger.Log.Debugw("[MD1K_COMPUTE] Before correction",
		"waitTimeAsIfMM1K_ms", waitTimeAsIfMM1K*1000)

	// Step 6: Apply Pollaczek-Khintchine correction for deterministic service
	// M/D/1 has half the waiting time of M/M/1 (C_s=0 vs C_s=1)
	m.avgWaitTime = 0.5 * waitTimeAsIfMM1K
	if m.avgWaitTime < 0 {
		m.avgWaitTime = 0
	}

	// Log the correction with proper ratio handling
	if waitTimeAsIfMM1K > 0.0001 { // threshold to avoid floating point noise
		ratio := m.avgWaitTime / waitTimeAsIfMM1K
		logger.Log.Debugw("[MD1K_COMPUTE] After 0.5x correction",
			"avgWaitTime_ms", m.avgWaitTime*1000,
			"ratio", ratio)
	} else {
		// No waiting time at low load - this is expected
		logger.Log.Debugw("[MD1K_COMPUTE] After 0.5x correction",
			"avgWaitTime_ms", m.avgWaitTime*1000,
			"ratio", "N/A (no queueing)")
	}

	// Step 7: Recalculate response time to maintain R = W + S relationship
	m.avgRespTime = m.avgWaitTime + m.avgServTime

	// Step 8: Recalculate avgNumInSystem using Little's Law: N = λR
	m.avgNumInSystem = m.throughput * m.avgRespTime

	// Step 9: Queue length by Little's Law: L_q = λW
	m.avgQueueLength = m.throughput * m.avgWaitTime
	logger.Log.Debugw("[MD1K_COMPUTE] Final metrics",
		"avgRespTime_ms", m.avgRespTime*1000,
		"avgServTime_ms", m.avgServTime*1000,
		"avgQueueLength", m.avgQueueLength)
} // Probability computation for M/D/1/K with state-dependent service rates
// Theory: We solve the balance equations directly for the continuous-time chain.
// For M/G/1/K with state-dependent rates, the balance equations are:
//
//	λπ[n] = μ[n]π[n+1] for 0 ≤ n < K
//	Σπ[n] = 1
//
// This gives: π[n] = π[0] * (λ/μ[0]) * (λ/μ[1]) * ... * (λ/μ[n-1])
//
// The deterministic service time affects WAITING time (via Pollaczek-Khintchine),
// but the STATE PROBABILITIES follow the same balance equations as M/M/1/K.
//
// References:
// - Gross & Harris (1998): Fundamentals of Queueing Theory, 4th ed., Ch. 2
// - Kleinrock (1975): Queueing Systems Vol 1, Section 3.3
func (m *MD1ModelStateDependent) computeProbabilitiesApprox() {
	num := len(m.servRate)

	// Build probabilities using balance equations: π[n+1] = (λ/μ[n]) * π[n]
	m.p[0] = 1.0
	for i := 1; i <= m.K; i++ {
		idx := min(i-1, num-1)
		serviceRate := float64(m.servRate[idx])
		m.p[i] = m.p[i-1] * float64(m.lambda) / serviceRate
	}

	// Normalize
	m.sumP = 0
	for i := 0; i <= m.K; i++ {
		m.sumP += m.p[i]
	}
	for i := 0; i <= m.K; i++ {
		m.p[i] /= m.sumP
	}
	m.sumP = 1.0
	m.rho = m.ComputeRho()
}

func (m *MD1ModelStateDependent) GetAvgNumInServers() float32 {
	return m.avgNumInServers
}

func (m *MD1ModelStateDependent) String() string {
	var b bytes.Buffer
	b.WriteString("MD1ModelStateDependent: ")
	b.WriteString(m.MD1KModel.String())
	return b.String()
}
