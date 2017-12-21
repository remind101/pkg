package counting

import (
	"hash/fnv"
	"math"

	"github.com/pkg/errors"
	"github.com/willf/bitset"
)

// ProbabilisticCounter implements an a linear-time counting algorithm, also known as "linear counting".
//
// From: "A Linear-Time Probabilistic Counting Algorithm for Database Applications"
// (http://dblab.kaist.ac.kr/Publication/pdf/ACM90_TODS_v15n2.pdf)
//
// "Linear counting is a two-step process. In step 1, the algorithm allocates a bit map (hash table) of size m in main
// memory. All entries are initialized to "0"s. The algorithm then scans the relation and applied a hash function to each
// data value in the column of interest. The hash function generates a bit map address and the algorithm sets this
// addressed bit to "1". In step 2, the algorithm first counts the number of empty bit map entries (equivalently,
// the number of "0" entries). It then estimates the column cardinality by dividing this count by the bit map size m
// (thus obtaining the fraction of empty bit map entries V_n) and plugging the result into the following equation:
//
//   n^ = -m * ln V_n (The symbol ^ denotes an estimator)"
//
// Therefore, ProbabilisticCounter is able to compute an approximate distinct count for a sufficiently large number of
// string values, with an error rate of less than 1.25% on average. It does so while using a very low amount of memory,
// for instance tracking up to 1 million string values would consume 128 kilobytes.
//
// ProbabilisticCounter *does not* provide a query-based API which would let callers verify whether a given string value
// has been counted before or not. There are other algorithms which are more efficient at providing that kind of
// funcionality, for instance "min-count log sketch" and "hyperloglog".
//
// Lastly, ProbabilisticCounter provides serialization/deserialization in case it is needed to use and persist this
// data structure in a database.
type ProbabilisticCounter struct {
	cardinality int
	bitset      *bitset.BitSet
}

// It creates a new ProbabilisticCounter given an *estimate* of the true cardinality of the set of values it should count.
// This estimate is needed to that an internal bit map can be initialized and sized properly.
func NewProbabilisticCounter(cardinality int) *ProbabilisticCounter {
	return &ProbabilisticCounter{
		cardinality: cardinality,
		bitset:      newBitSetForCardinality(cardinality),
	}
}

// It deserializes a new ProbabilisticCounter which has probably been serialized before with probabiliscCounter.Bytes().
// cardinality: an *estimate* of the true cardinality of the set of values it should count so that an internal bit map
// can be initialized and sized properly.
func NewProbabilisticCounterFromBytes(cardinality int, bytes []byte) (*ProbabilisticCounter, error) {
	bs := newBitSetForCardinality(cardinality)
	err := bs.UnmarshalBinary(bytes)

	if err != nil {
		return nil, errors.Wrap(err, "Unable to deserialize ProbabilisticCounter from binary data")
	}

	return &ProbabilisticCounter{
		cardinality: cardinality,
		bitset:      bs,
	}, nil
}

// It adds one or more string values to this counter.
func (pc *ProbabilisticCounter) Add(values ...string) {
	for _, value := range values {
		pc.bitset.Set(pc.bitIndex(value))
	}
}

// It provides an (approximate) distinct count of values counter so far, with an error rate less than 1.25% on average.
func (pc *ProbabilisticCounter) Count() int {
	ones := float64(pc.bitset.Count())
	length := float64(pc.bitset.Len())
	zeroes := length - ones
	estimator := -1 * length * math.Log(zeroes/length)
	return int(estimator)
}

// It serializes this ProbabilisticCounter into a slice of bytes, which can be used to persist this data structure.
func (pc *ProbabilisticCounter) Bytes() ([]byte, error) {
	bytes, err := pc.bitset.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to serialize ProbabilisticCounter into binary data")
	}
	return bytes, nil
}

func newBitSetForCardinality(cardinality int) *bitset.BitSet {
	length := lengthForCardinality(cardinality)
	return bitset.New(length)
}

var loadFactor float64 = 16

// bit map length is proportinal to the estimated cardinality, it can be between 8 and 128kb,
// those numbers were chosen in order to achieve a relatively low error rate while keeping a low memory usage.
func lengthForCardinality(cardinality int) uint {
	length := closestBase2(cardinality) * loadFactor
	var minLength = math.Exp2(16) // 8kb
	var maxLength = math.Exp2(20) // 128kb
	return uint(math.Max(minLength, math.Min(length, maxLength)))
}

func closestBase2(n int) float64 {
	count := 0
	for ; n > 0; n >>= 1 {
		count++
	}
	return math.Exp2(float64(count))
}

// It provides a bit map address given a string value using a non-crypto hashing function.
func (pc *ProbabilisticCounter) bitIndex(value string) uint {
	h := fnv.New32a()
	h.Write([]byte(value))
	return uint(h.Sum32()) % pc.bitset.Len()
}
