package config

import (
	"math"
)

/**
 * Parameters
 */

// Tolerated percentile for SLOs
var SLOPercentile = 0.95

// Multiplier of average of exponential distribution to attain percentile
var SLOMargin = -float32(math.Log(1 - SLOPercentile))

// maximum number of requests in queueing system as multiples of maximum batch size
var MaxQueueToBatchRatio = 10

// accelerator transition penalty factor
var AccelPenaltyFactor = float32(0.1)

// default name of a service class
const DefaultServiceClassName string = "Free"

// default priority of a lowest service class
const DefaultLowPriority int = 100

// default priority of a highest service class
const DefaultHighPriority int = 1

// default priority of a service class (lowest)
const DefaultServiceClassPriority int = DefaultLowPriority

// default option for allocation under saturated condition
var DefaultSaturatedAllocationPolicy SaturatedAllocationPolicy = None

// Queue model type selection
type QueueModelType string

const (
	// MM1K uses M/M/1/K model (Markovian arrivals, exponential service times)
	MM1K QueueModelType = "MM1K"
	// MD1K uses M/D/1/K model (Markovian arrivals, deterministic service times)
	MD1K QueueModelType = "MD1K"
)

// Default queue model type
// MD1K is more accurate for LLM inference because service times are deterministic
// (same input/output lengths always take the same GPU computation time)
var DefaultQueueModelType QueueModelType = MD1K
