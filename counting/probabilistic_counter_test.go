package counting

import (
	"fmt"
	"math"

	"testing"
)

func TestProbabilisticCounterErrorRate(t *testing.T) {
	var maxError = 0.0125
	cardinalities := []int{250, 1000, 2500, 10000, 25000, 100000, 250000, 1000000}
	for _, card := range cardinalities {
		assertErrorRate(t, card, maxError)
	}
}

func assertErrorRate(t *testing.T, cardinality int, maxError float64) {
	counter := NewProbabilisticCounter(cardinality)
	fillCounter(counter, cardinality)
	e := errorRate(counter, cardinality)
	if e > maxError {
		t.Fatalf("when cardinality=%v, expected error rate (%v) to be less than %v", cardinality, e, maxError)
	}
}

func fillCounter(counter *ProbabilisticCounter, n int) {
	for i := 0; i < n; i++ {
		counter.Add(fmt.Sprintf("test-%d", i))
	}
}

func errorRate(counter *ProbabilisticCounter, cardinality int) float64 {
	cardinality64 := float64(cardinality)
	count := float64(counter.Count())
	return math.Abs(cardinality64-count) / cardinality64
}
