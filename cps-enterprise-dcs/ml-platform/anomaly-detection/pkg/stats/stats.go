package stats

import "math"

// MeanStdDev calculates the mean and standard deviation of a slice of float64 values
func MeanStdDev(values []float64) (mean, stdDev float64) {
	n := len(values)
	if n == 0 {
		return 0, 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean = sum / float64(n)

	if n == 1 {
		return mean, 0
	}

	variance := 0.0
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(n - 1)
	stdDev = math.Sqrt(variance)

	return mean, stdDev
}

// Percentile calculates the percentile of a sorted slice
func Percentile(sortedValues []float64, percentile float64) float64 {
	n := len(sortedValues)
	if n == 0 {
		return 0
	}

	if percentile <= 0 {
		return sortedValues[0]
	}
	if percentile >= 100 {
		return sortedValues[n-1]
	}

	rank := (percentile / 100.0) * float64(n-1)
	lower := int(rank)
	upper := lower + 1

	if upper >= n {
		return sortedValues[n-1]
	}

	fraction := rank - float64(lower)
	return sortedValues[lower] + fraction*(sortedValues[upper]-sortedValues[lower])
}

// Median calculates the median of a slice
func Median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)

	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j] < sorted[i] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

// MovingAverage calculates the moving average of a slice
func MovingAverage(values []float64, windowSize int) []float64 {
	if len(values) == 0 || windowSize <= 0 {
		return nil
	}

	result := make([]float64, 0, len(values))
	sum := 0.0
	queue := make([]float64, 0, windowSize)

	for _, v := range values {
		sum += v
		queue = append(queue, v)

		if len(queue) > windowSize {
			sum -= queue[0]
			queue = queue[1:]
		}

		if len(queue) == windowSize {
			result = append(result, sum/float64(windowSize))
		}
	}

	return result
}

// ZScore calculates the z-score of a value relative to a dataset
func ZScore(value, mean, stdDev float64) float64 {
	if stdDev == 0 {
		return 0
	}
	return (value - mean) / stdDev
}
