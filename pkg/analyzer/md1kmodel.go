package analyzer

import (
	"bytes"
	"fmt"
	"math"
)

// M/D/1/K Finite storage single server queue with deterministic service times
// M = Markovian (Poisson) arrivals
// D = Deterministic service times
// 1 = Single server
// K = Finite capacity (queue + server)
type MD1KModel struct {
	QueueModel           // extends base class
	K          int       // limit on number in system
	p          []float64 // state probabilities (embedded Markov chain)
	sumP       float64   // sum of probabilities
	throughput float32   // effective (departure) rate
}

func NewMD1KModel(K int) *MD1KModel {
	m := &MD1KModel{
		QueueModel: QueueModel{},
		K:          K,
		p:          make([]float64, K+1),
		throughput: 0,
	}
	m.QueueModel.GetRhoMax = m.GetRhoMax
	m.QueueModel.ComputeRho = m.ComputeRho
	m.QueueModel.computeStatistics = m.computeStatistics
	return m
}

// Solve queueing model given arrival and service rates
func (m *MD1KModel) Solve(lambda float32, mu float32) {
	m.QueueModel.Solve(lambda, mu)
}

// Compute utilization of queueing model
func (m *MD1KModel) ComputeRho() float32 {
	if m.lambda == m.mu {
		return 1
	} else {
		return m.lambda / m.mu
	}
}

// Compute the maximum utilization of queueing model
func (m *MD1KModel) GetRhoMax() float32 {
	return float32(m.K)
}

// Compute state probabilities using embedded Markov chain approach
// For M/D/1/K, we observe the system at service completion epochs
// At each completion, the number of arrivals during deterministic service time S
// follows a Poisson distribution with parameter λS
func (m *MD1KModel) computeProbabilities() {
	// Initialize uniform distribution
	for i := 0; i <= m.K; i++ {
		m.p[i] = 1.0 / float64(m.K+1)
	}

	// Service time is deterministic: S = 1/μ
	serviceTime := float64(1.0 / m.mu)
	// Expected arrivals during service time
	arrivalRate := float64(m.lambda) * serviceTime

	// Iterate to find steady-state probabilities
	maxIter := 10000
	tolerance := 1e-8

	for iter := 0; iter < maxIter; iter++ {
		pNew := make([]float64, m.K+1)

		// For each state n at a service completion
		for n := 0; n <= m.K; n++ {
			if n == 0 {
				// If system is empty, it remains empty until next arrival
				pNew[0] += m.p[0]
			} else {
				// During service of one customer, j more arrive (Poisson)
				// After service completes, state goes from n to n-1+j
				for j := 0; j <= m.K; j++ {
					// Poisson probability of j arrivals during service
					poissonProb := math.Exp(-arrivalRate) *
						math.Pow(arrivalRate, float64(j)) /
						factorial(j)

					nextState := n - 1 + j
					if nextState > m.K {
						nextState = m.K // Truncate at capacity
					}

					pNew[nextState] += m.p[n] * poissonProb
				}
			}
		}

		// Check convergence
		maxDiff := 0.0
		for i := 0; i <= m.K; i++ {
			diff := math.Abs(pNew[i] - m.p[i])
			if diff > maxDiff {
				maxDiff = diff
			}
		}

		// Update probabilities
		copy(m.p, pNew)

		if maxDiff < tolerance {
			break
		}
	}

	// Normalize to ensure sum = 1
	m.sumP = 0
	for i := 0; i <= m.K; i++ {
		m.sumP += m.p[i]
	}
	for i := 0; i <= m.K; i++ {
		m.p[i] /= m.sumP
	}
	m.sumP = 1.0
}

// Evaluate performance measures of queueing model
func (m *MD1KModel) computeStatistics() {
	if !m.isValid {
		return
	}
	m.computeProbabilities()

	// Calculate average number in system
	var temp float64
	for i := 0; i <= m.K; i++ {
		temp += float64(i) * m.p[i]
	}
	m.avgNumInSystem = float32(temp)

	// Throughput (effective departure rate)
	// For M/D/1/K: throughput = λ(1 - p[K])
	m.throughput = m.lambda * (1 - float32(m.p[m.K]))

	// Average response time (by Little's Law)
	m.avgRespTime = m.avgNumInSystem / m.throughput

	// Average service time (deterministic)
	m.avgServTime = 1 / m.mu

	// Average waiting time
	m.avgWaitTime = m.avgRespTime - m.avgServTime
	if m.avgWaitTime < 0 {
		m.avgWaitTime = 0
	}

	// Average queue length
	m.avgQueueLength = m.throughput * m.avgWaitTime
}

// Helper function to compute factorial
func factorial(n int) float64 {
	if n < 0 {
		return 1
	}
	if n == 0 || n == 1 {
		return 1
	}
	// Use lookup table for small n to avoid repeated computation
	if n <= 20 {
		factorialTable := []float64{
			1, 1, 2, 6, 24, 120, 720, 5040, 40320, 362880, 3628800,
			39916800, 479001600, 6227020800, 87178291200, 1307674368000,
			20922789888000, 355687428096000, 6402373705728000,
			121645100408832000, 2432902008176640000,
		}
		return factorialTable[n]
	}
	// For larger n, use Stirling's approximation or iterative computation
	result := 1.0
	for i := 2; i <= n; i++ {
		result *= float64(i)
		if math.IsInf(result, 1) {
			// Use Stirling's approximation for very large factorials
			// n! ≈ √(2πn) * (n/e)^n
			return math.Sqrt(2*math.Pi*float64(n)) *
				math.Pow(float64(n)/math.E, float64(n))
		}
	}
	return result
}

func (m *MD1KModel) GetProbabilities() []float64 {
	return m.p
}

func (m *MD1KModel) GetThroughput() float32 {
	return m.throughput
}

func (m *MD1KModel) String() string {
	var b bytes.Buffer
	b.WriteString("MD1KModel: ")
	b.WriteString(m.QueueModel.String())
	fmt.Fprintf(&b, "K=%d; ", m.K)
	return b.String()
}
