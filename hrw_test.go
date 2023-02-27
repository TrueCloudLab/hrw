package hrw

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

type (
	hashString string
	unknown    byte
	slices     struct {
		actual interface{}
		expect interface{}
	}

	Uint32Slice []uint32
)

var testKey = []byte("0xff51afd7ed558ccd")

func (p Uint32Slice) Len() int           { return len(p) }
func (p Uint32Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Uint32Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func Example() {
	// given a set of servers
	servers := []string{
		"one.example.com",
		"two.example.com",
		"three.example.com",
		"four.example.com",
		"five.example.com",
		"six.example.com",
	}

	// HRW can consistently select a uniformly-distributed set of servers for
	// any given key
	var (
		key = []byte("/examples/object-key")
		h   = Hash(key)
	)

	SortSliceByValue(servers, h)
	for id := range servers {
		fmt.Printf("trying GET %s%s\n", servers[id], key)
	}

	// Output:
	// trying GET three.example.com/examples/object-key
	// trying GET two.example.com/examples/object-key
	// trying GET five.example.com/examples/object-key
	// trying GET six.example.com/examples/object-key
	// trying GET one.example.com/examples/object-key
	// trying GET four.example.com/examples/object-key
}

func (h hashString) Hash() uint64 {
	return Hash([]byte(h))
}

func TestSortSliceByIndex(t *testing.T) {
	actual := []string{"a", "b", "c", "d", "e", "f"}
	expect := []string{"e", "a", "c", "f", "d", "b"}
	hash := Hash(testKey)
	SortSliceByIndex(actual, hash)
	require.Equal(t, expect, actual)
}

func TestValidateWeights(t *testing.T) {
	weights := []float64{10, 10, 10, 2, 2, 2}
	err := ValidateWeights(weights)
	require.Error(t, err)
	weights = []float64{math.NaN(), 1, 1, 0.2, 0.2, 0.2}
	err = ValidateWeights(weights)
	require.Error(t, err)
	weights = []float64{1, 1, 1, 0.2, 0.2, 0.2}
	err = ValidateWeights(weights)
	require.NoError(t, err)
}

func TestSortSliceByWeightIndex(t *testing.T) {
	actual := []string{"a", "b", "c", "d", "e", "f"}
	weights := []float64{1, 1, 1, 0.2, 0.2, 0.2}
	expect := []string{"a", "c", "b", "e", "f", "d"}
	hash := Hash(testKey)
	SortSliceByWeightIndex(actual, weights, hash)
	require.Equal(t, expect, actual)
}

func TestSortSliceByValue(t *testing.T) {
	actual := []string{"a", "b", "c", "d", "e", "f"}
	expect := []string{"d", "f", "c", "b", "a", "e"}
	hash := Hash(testKey)
	SortSliceByValue(actual, hash)
	require.Equal(t, expect, actual)
}

func TestSortSliceByValueFail(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		var (
			actual []int
			hash   = Hash(testKey)
		)
		require.NotPanics(t, func() { SortSliceByValue(actual, hash) })
	})

	t.Run("must be slice", func(t *testing.T) {
		actual := 10
		hash := Hash(testKey)
		require.Panics(t, func() { SortSliceByValue(actual, hash) })
	})

	t.Run("must 'fail' for unknown type", func(t *testing.T) {
		actual := []unknown{1, 2, 3, 4, 5}
		hash := Hash(testKey)
		require.Panics(t, func() { SortSliceByValue(actual, hash) })
	})
}

func TestSortSliceByValueHasher(t *testing.T) {
	actual := []hashString{"a", "b", "c", "d", "e", "f"}
	expect := []hashString{"d", "f", "c", "b", "a", "e"}
	hash := Hash(testKey)
	SortSliceByValue(actual, hash)
	require.Equal(t, expect, actual)
}

func TestSortHasherSliceByValue(t *testing.T) {
	actual := []hashString{"a", "b", "c", "d", "e", "f"}
	expect := []hashString{"d", "f", "c", "b", "a", "e"}
	hash := Hash(testKey)
	SortHasherSliceByValue(actual, hash)
	require.EqualValues(t, expect, actual)
}

func TestSortHasherSliceByWeightValue(t *testing.T) {
	actual := []hashString{"a", "b", "c", "d", "e", "f"}
	weights := []float64{1.0, 1.0, 1.0, 1.0, 1.0, 1.0}
	expect := []hashString{"d", "f", "c", "b", "a", "e"}
	hash := Hash(testKey)
	SortHasherSliceByWeightValue(actual, weights, hash)
	require.EqualValues(t, expect, actual)
}

func TestSortSliceByValueIntSlice(t *testing.T) {
	cases := []slices{
		{
			actual: []int{0, 1, 2, 3, 4, 5},
			expect: []int{2, 0, 5, 3, 1, 4},
		},

		{
			actual: []uint{0, 1, 2, 3, 4, 5},
			expect: []uint{2, 0, 5, 3, 1, 4},
		},

		{
			actual: []int8{0, 1, 2, 3, 4, 5},
			expect: []int8{5, 2, 1, 4, 0, 3},
		},

		{
			actual: []uint8{0, 1, 2, 3, 4, 5},
			expect: []uint8{5, 2, 1, 4, 0, 3},
		},

		{
			actual: []int16{0, 1, 2, 3, 4, 5},
			expect: []int16{1, 0, 3, 2, 4, 5},
		},

		{
			actual: []uint16{0, 1, 2, 3, 4, 5},
			expect: []uint16{1, 0, 3, 2, 4, 5},
		},

		{
			actual: []int32{0, 1, 2, 3, 4, 5},
			expect: []int32{5, 1, 2, 0, 3, 4},
		},

		{
			actual: []uint32{0, 1, 2, 3, 4, 5},
			expect: []uint32{5, 1, 2, 0, 3, 4},
		},

		{
			actual: []int64{0, 1, 2, 3, 4, 5},
			expect: []int64{5, 3, 0, 1, 4, 2},
		},

		{
			actual: []uint64{0, 1, 2, 3, 4, 5},
			expect: []uint64{5, 3, 0, 1, 4, 2},
		},
	}
	hash := Hash(testKey)

	for _, tc := range cases {
		SortSliceByValue(tc.actual, hash)
		require.Equal(t, tc.expect, tc.actual)
	}
}

func TestSort(t *testing.T) {
	nodes := []uint64{1, 2, 3, 4, 5}
	hash := Hash(testKey)
	actual := Sort(nodes, hash)
	expected := []uint64{3, 1, 4, 2, 0}
	require.Equal(t, expected, actual)
}

func TestDistribution(t *testing.T) {
	const (
		size    = 10
		keys    = 100000
		percent = 0.03
	)
	// We use χ2 method to determine similarity of distribution with uniform distribution.
	// χ2 = Σ((n-N)**2/N)
	// https://www.medcalc.org/manual/chi-square-table.php p=0.1
	var chiTable = map[int]float64{9: 14.68, 99: 117.407}

	t.Run("sort", func(t *testing.T) {
		var (
			i      uint64
			nodes  [size]uint64
			counts = make(map[uint64]uint64, size)
			key    = make([]byte, 16)
		)

		for i = 0; i < size; i++ {
			nodes[i] = i
		}

		for i = 0; i < keys; i++ {
			binary.BigEndian.PutUint64(key, i+size)
			hash := Hash(key)
			counts[Sort(nodes[:], hash)[0]]++
		}

		var chi2 float64
		mean := float64(keys) / float64(size)
		delta := mean * percent
		for node, count := range counts {
			d := mean - float64(count)
			chi2 += math.Pow(float64(count)-mean, 2) / mean
			require.True(t, d < delta && (0-d) < delta,
				"Node %d received %d keys, expected %.0f (+/- %.2f)", node, count, mean, delta)
		}
		require.True(t, chi2 < chiTable[size-1],
			"Chi2 condition for .9 is not met (expected %.2f <= %.2f)", chi2, chiTable[size-1])
	})

	t.Run("sortByIndex", func(t *testing.T) {
		var (
			i      uint64
			a, b   [size]uint64
			counts = make(map[uint64]int, size)
			key    = make([]byte, 16)
		)

		for i = 0; i < size; i++ {
			a[i] = i
		}

		for i = 0; i < keys; i++ {
			copy(b[:], a[:])

			binary.BigEndian.PutUint64(key, i+size)
			hash := Hash(key)
			SortSliceByIndex(b[:], hash)
			counts[b[0]]++
		}

		var chi2 float64
		mean := float64(keys) / float64(size)
		delta := mean * percent
		for node, count := range counts {
			d := mean - float64(count)
			chi2 += math.Pow(float64(count)-mean, 2) / mean
			require.True(t, d < delta && (0-d) < delta,
				"Node %d received %d keys, expected %.0f (+/- %.2f)", node, count, mean, delta)
		}
		require.True(t, chi2 < chiTable[size-1],
			"Chi2 condition for .9 is not met (expected %.2f <= %.2f)", chi2, chiTable[size-1])
	})

	t.Run("sortByValue", func(t *testing.T) {
		var (
			i      uint64
			a, b   [size]int
			counts = make(map[int]int, size)
			key    = make([]byte, 16)
		)

		for i = 0; i < size; i++ {
			a[i] = int(i)
		}

		for i = 0; i < keys; i++ {
			copy(b[:], a[:])
			binary.BigEndian.PutUint64(key, i+size)
			hash := Hash(key)
			SortSliceByValue(b[:], hash)
			counts[b[0]]++
		}

		var chi2 float64
		mean := float64(keys) / float64(size)
		delta := mean * percent
		for node, count := range counts {
			d := mean - float64(count)
			chi2 += math.Pow(float64(count)-mean, 2) / mean
			require.True(t, d < delta && (0-d) < delta,
				"Node %d received %d keys, expected %.0f (+/- %.2f)", node, count, mean, delta)
		}
		require.True(t, chi2 < chiTable[size-1],
			"Chi2 condition for .9 is not met (expected %.2f <= %.2f)", chi2, chiTable[size-1])
	})

	t.Run("sortByStringValue", func(t *testing.T) {
		var (
			i      uint64
			a, b   [size]string
			counts = make(map[string]int, size)
			key    = make([]byte, 16)
		)

		for i = 0; i < size; i++ {
			a[i] = strconv.FormatUint(i, 10)
		}

		for i = 0; i < keys; i++ {
			copy(b[:], a[:])
			binary.BigEndian.PutUint64(key, i+size)
			hash := Hash(key)
			SortSliceByValue(b[:], hash)
			counts[b[0]]++
		}

		var chi2 float64
		mean := float64(keys) / float64(size)
		delta := mean * percent
		for node, count := range counts {
			d := mean - float64(count)
			chi2 += math.Pow(float64(count)-mean, 2) / mean
			require.True(t, d < delta && (0-d) < delta,
				"Node %d received %d keys, expected %.0f (+/- %.2f)", node, count, mean, delta)
		}
		require.True(t, chi2 < chiTable[size-1],
			"Chi2 condition for .9 is not met (expected %.2f <= %.2f)", chi2, chiTable[size-1])
	})

	t.Run("sortByInt32Value", func(t *testing.T) {
		var (
			i      uint64
			a, b   [size]int32
			counts = make(map[int32]int, size)
			key    = make([]byte, 16)
		)

		for i = 0; i < size; i++ {
			a[i] = int32(i)
		}

		for i = 0; i < keys; i++ {
			copy(b[:], a[:])
			binary.BigEndian.PutUint64(key, i+size)
			hash := Hash(key)
			SortSliceByValue(b[:], hash)
			counts[b[0]]++
		}

		var chi2 float64
		mean := float64(keys) / float64(size)
		delta := mean * percent
		for node, count := range counts {
			d := mean - float64(count)
			chi2 += math.Pow(float64(count)-mean, 2) / mean
			require.True(t, d < delta && (0-d) < delta,
				"Node %d received %d keys, expected %.0f (+/- %.2f)", node, count, mean, delta)
		}
		require.True(t, chi2 < chiTable[size-1],
			"Chi2 condition for .9 is not met (expected %.2f <= %.2f)", chi2, chiTable[size-1])
	})

	t.Run("sortByWeightValue", func(t *testing.T) {
		var (
			i            uint64
			a, b, result [size]int
			w            [size]float64
			key          = make([]byte, 16)
		)

		for i = 0; i < size; i++ {
			a[i] = int(i)
			w[i] = float64(size-i) / float64(size)
		}
		for i = 0; i < keys; i++ {
			copy(b[:], a[:])
			binary.BigEndian.PutUint64(key, i+size)
			hash := Hash(key)
			SortSliceByWeightValue(b[:], w[:], hash)
			result[b[0]]++
		}

		for i := 0; i < size-1; i++ {
			require.True(t, bool(w[i] > w[i+1]) == bool(result[i] > result[i+1]),
				"result array %v must be corresponded to weights %v", result, w)
		}
	})

	t.Run("sortByWeightValueShuffledWeight", func(t *testing.T) {
		var (
			i            uint64
			a, b, result [size]int
			w            [size]float64
			key          = make([]byte, 16)
		)

		for i = 0; i < size; i++ {
			a[i] = int(i)
			w[i] = float64(size-i) / float64(size)
		}

		rand.Shuffle(size, func(i, j int) {
			w[i], w[j] = w[j], w[i]
		})
		for i = 0; i < keys; i++ {
			copy(b[:], a[:])
			binary.BigEndian.PutUint64(key, i+size)
			hash := Hash(key)
			SortSliceByWeightValue(b[:], w[:], hash)
			result[b[0]]++
		}
		for i := 0; i < size-1; i++ {
			require.True(t, bool(w[i] > w[i+1]) == bool(result[i] > result[i+1]),
				"result array %v must be corresponded to weights %v", result, w)
		}
	})

	t.Run("sortByWeightValueEmptyWeight", func(t *testing.T) {
		var (
			i      uint64
			a, b   [size]int
			w      [size]float64
			counts = make(map[int]int, size)
			key    = make([]byte, 16)
		)

		for i = 0; i < size; i++ {
			a[i] = int(i)
		}

		for i = 0; i < keys; i++ {
			copy(b[:], a[:])
			binary.BigEndian.PutUint64(key, i+size)
			hash := Hash(key)
			SortSliceByWeightValue(b[:], w[:], hash)
			counts[b[0]]++
		}

		var chi2 float64
		mean := float64(keys) / float64(size)
		delta := mean * percent
		for node, count := range counts {
			d := mean - float64(count)
			chi2 += math.Pow(float64(count)-mean, 2) / mean
			require.True(t, d < delta && (0-d) < delta,
				"Node %d received %d keys, expected %.0f (+/- %.2f)", node, count, mean, delta)
		}
		require.True(t, chi2 < chiTable[size-1],
			"Chi2 condition for .9 is not met (expected %.2f <= %.2f)", chi2, chiTable[size-1])
	})

	t.Run("sortByWeightValueUniformWeight", func(t *testing.T) {
		var (
			i      uint64
			a, b   [size]int
			w      [size]float64
			counts = make(map[int]int, size)
			key    = make([]byte, 16)
		)

		for i = 0; i < size; i++ {
			a[i] = int(i)
			w[i] = 0.5
		}

		for i = 0; i < keys; i++ {
			copy(b[:], a[:])
			binary.BigEndian.PutUint64(key, i+size)
			hash := Hash(key)
			SortSliceByWeightValue(b[:], w[:], hash)
			counts[b[0]]++
		}

		var chi2 float64
		mean := float64(keys) / float64(size)
		delta := mean * percent
		for node, count := range counts {
			d := mean - float64(count)
			chi2 += math.Pow(float64(count)-mean, 2) / mean
			require.True(t, d < delta && (0-d) < delta,
				"Node %d received %d keys, expected %.0f (+/- %.2f)", node, count, mean, delta)
		}
		require.True(t, chi2 < chiTable[size-1],
			"Chi2 condition for .9 is not met (expected %.2f <= %.2f)", chi2, chiTable[size-1])
	})

	t.Run("sortByWeightValueAbsoluteW", func(t *testing.T) {
		const keys = 1
		var (
			i    uint64
			a, b [size]int
			w    [size]float64
			key  = make([]byte, 16)
		)

		for i = 0; i < size; i++ {
			a[i] = int(i)
		}
		w[size-1] = 1

		for i = 0; i < keys; i++ {
			copy(b[:], a[:])
			binary.BigEndian.PutUint64(key, i+size)
			hash := Hash(key)
			SortSliceByWeightValue(b[:], w[:], hash)
			require.True(t, b[0] == a[size-1],
				"expected last value of %v to be the first with highest distance", a)
		}

	})

	t.Run("sortByWeightValueNormalizedWeight", func(t *testing.T) {
		var (
			i              uint64
			a, b, result   [size]uint64
			w, normalizedW [size]float64
			key            = make([]byte, 16)
		)

		for i = 0; i < size; i++ {
			a[i] = i
			w[int(i)] = 10
		}
		w[0] = 100

		// Here let's use logarithm normalization
		for i = 0; i < size; i++ {
			normalizedW[i] = math.Log2(w[i]) / math.Log2(w[0])
		}

		for i = 0; i < keys; i++ {
			copy(b[:], a[:])
			binary.BigEndian.PutUint64(key, i+size)
			hash := Hash(key)
			SortSliceByWeightValue(b[:], normalizedW[:], hash)
			for j := range b {
				result[b[j]] += uint64(len(b) - j)
			}
		}
		cutResult := result[1:]
		var total uint64
		for i := range cutResult {
			total += cutResult[i]
		}

		var chi2 float64
		mean := float64(total) / float64(len(cutResult))
		delta := mean * percent
		for node, count := range cutResult {
			d := mean - float64(count)
			chi2 += math.Pow(float64(count)-mean, 2) / mean
			require.True(t, d < delta && (0-d) < delta,
				"Node %d received %d keys, expected %.0f (+/- %.2f)", node, count, mean, delta)
		}
		require.True(t, chi2 < chiTable[size-1],
			"Chi2 condition for .9 is not met (expected %.2f <= %.2f)", chi2, chiTable[size-1])
	})

	t.Run("hash collision", func(t *testing.T) {
		var (
			i      uint64
			counts = make(map[uint64]uint64)
			key    = make([]byte, 16)
		)

		for i = 0; i < keys; i++ {
			binary.BigEndian.PutUint64(key, i+size)
			hash := Hash(key)
			counts[hash]++
		}

		for node, count := range counts {
			if count > 1 {
				t.Errorf("Node %d received %d keys", node, count)
			}
		}
	})
}

func BenchmarkSort_fnv_10(b *testing.B) {
	hash := Hash(testKey)
	_ = benchmarkSort(b, 10, hash)
}

func BenchmarkSort_fnv_100(b *testing.B) {
	hash := Hash(testKey)
	_ = benchmarkSort(b, 100, hash)
}

func BenchmarkSort_fnv_1000(b *testing.B) {
	hash := Hash(testKey)
	_ = benchmarkSort(b, 1000, hash)
}

func BenchmarkSortByIndex_fnv_10(b *testing.B) {
	hash := Hash(testKey)
	benchmarkSortByIndex(b, 10, hash)
}

func BenchmarkSortByIndex_fnv_100(b *testing.B) {
	hash := Hash(testKey)
	benchmarkSortByIndex(b, 100, hash)
}

func BenchmarkSortByIndex_fnv_1000(b *testing.B) {
	hash := Hash(testKey)
	benchmarkSortByIndex(b, 1000, hash)
}

func BenchmarkSortByValue_fnv_10(b *testing.B) {
	hash := Hash(testKey)
	benchmarkSortByValue(b, 10, hash)
}

func BenchmarkSortByValue_fnv_100(b *testing.B) {
	hash := Hash(testKey)
	benchmarkSortByValue(b, 100, hash)
}

func BenchmarkSortByValue_fnv_1000(b *testing.B) {
	hash := Hash(testKey)
	benchmarkSortByValue(b, 1000, hash)
}

func BenchmarkSortHashersByValue_Reflection_fnv_10(b *testing.B) {
	hash := Hash(testKey)
	benchmarkSortHashersByValueReflection(b, 10, hash)
}

func BenchmarkSortHashersByValue_Reflection_fnv_100(b *testing.B) {
	hash := Hash(testKey)
	benchmarkSortHashersByValueReflection(b, 100, hash)
}

func BenchmarkSortHashersByValue_Reflection_fnv_1000(b *testing.B) {
	hash := Hash(testKey)
	benchmarkSortHashersByValueReflection(b, 1000, hash)
}

func BenchmarkSortHashersByValue_Typed_fnv_10(b *testing.B) {
	hash := Hash(testKey)
	benchmarkSortHashersByValueTyped(b, 10, hash)
}

func BenchmarkSortHashersByValue_Typed_fnv_100(b *testing.B) {
	hash := Hash(testKey)
	benchmarkSortHashersByValueTyped(b, 100, hash)
}

func BenchmarkSortHashersByValue_Typed_fnv_1000(b *testing.B) {
	hash := Hash(testKey)
	benchmarkSortHashersByValueTyped(b, 1000, hash)
}

func BenchmarkSortByWeight_fnv_10(b *testing.B) {
	hash := Hash(testKey)
	_ = benchmarkSortByWeight(b, 10, hash)
}

func BenchmarkSortByWeight_fnv_100(b *testing.B) {
	hash := Hash(testKey)
	_ = benchmarkSortByWeight(b, 100, hash)
}

func BenchmarkSortByWeight_fnv_1000(b *testing.B) {
	hash := Hash(testKey)
	_ = benchmarkSortByWeight(b, 1000, hash)
}

func BenchmarkSortByWeightIndex_fnv_10(b *testing.B) {
	hash := Hash(testKey)
	benchmarkSortByWeightIndex(b, 10, hash)
}

func BenchmarkSortByWeightIndex_fnv_100(b *testing.B) {
	hash := Hash(testKey)
	benchmarkSortByWeightIndex(b, 100, hash)
}

func BenchmarkSortByWeightIndex_fnv_1000(b *testing.B) {
	hash := Hash(testKey)
	benchmarkSortByWeightIndex(b, 1000, hash)
}

func BenchmarkSortByWeightValue_fnv_10(b *testing.B) {
	hash := Hash(testKey)
	benchmarkSortByWeightValue(b, 10, hash)
}

func BenchmarkSortByWeightValue_fnv_100(b *testing.B) {
	hash := Hash(testKey)
	benchmarkSortByWeightValue(b, 100, hash)
}

func BenchmarkSortByWeightValue_fnv_1000(b *testing.B) {
	hash := Hash(testKey)
	benchmarkSortByWeightValue(b, 1000, hash)
}

func BenchmarkSortHashersByWeightValueReflection_fnv_10(b *testing.B) {
	benchmarkSortHashersByWeightValueRelection(b, 10, Hash(testKey))
}

func BenchmarkSortHashersByWeightValueReflection_fnv_100(b *testing.B) {
	benchmarkSortHashersByWeightValueRelection(b, 100, Hash(testKey))
}

func BenchmarkSortHashersByWeightValueReflection_fnv_1000(b *testing.B) {
	benchmarkSortHashersByWeightValueRelection(b, 100, Hash(testKey))
}

func BenchmarkSortHashersByWeightValueTyped_fnv_10(b *testing.B) {
	benchmarkSortHashersByWeightValueTyped(b, 10, Hash(testKey))
}

func BenchmarkSortHashersByWeightValueTyped_fnv_100(b *testing.B) {
	benchmarkSortHashersByWeightValueTyped(b, 100, Hash(testKey))
}

func BenchmarkSortHashersByWeightValueTyped_fnv_1000(b *testing.B) {
	benchmarkSortHashersByWeightValueTyped(b, 100, Hash(testKey))
}

func benchmarkSort(b *testing.B, n int, hash uint64) uint64 {
	servers := make([]uint64, n)
	for i := uint64(0); i < uint64(len(servers)); i++ {
		servers[i] = i
	}

	b.ResetTimer()
	b.ReportAllocs()

	var x uint64
	for i := 0; i < b.N; i++ {
		x += Sort(servers, hash)[0]
	}
	return x
}

func benchmarkSortByIndex(b *testing.B, n int, hash uint64) {
	servers := make([]uint64, n)
	for i := uint64(0); i < uint64(len(servers)); i++ {
		servers[i] = i
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		SortSliceByIndex(servers, hash)
	}
}

func benchmarkSortByValue(b *testing.B, n int, hash uint64) {
	servers := make([]string, n)
	for i := uint64(0); i < uint64(len(servers)); i++ {
		servers[i] = "localhost:" + strconv.FormatUint(60000-i, 10)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		SortSliceByValue(servers, hash)
	}
}

func benchmarkSortByWeight(b *testing.B, n int, hash uint64) uint64 {
	servers := make([]uint64, n)
	weights := make([]float64, n)
	for i := uint64(0); i < uint64(len(servers)); i++ {
		weights[i] = float64(uint64(n)-i) / float64(n)
		servers[i] = i
	}

	b.ResetTimer()
	b.ReportAllocs()

	var x uint64
	for i := 0; i < b.N; i++ {
		x += SortByWeight(servers, weights, hash)[0]
	}
	return x
}

func benchmarkSortByWeightIndex(b *testing.B, n int, hash uint64) {
	servers := make([]uint64, n)
	weights := make([]float64, n)
	for i := uint64(0); i < uint64(len(servers)); i++ {
		weights[i] = float64(uint64(n)-i) / float64(n)
		servers[i] = i
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		SortSliceByWeightIndex(servers, weights, hash)
	}
}

func benchmarkSortByWeightValue(b *testing.B, n int, hash uint64) {
	servers := make([]string, n)
	weights := make([]float64, n)
	for i := uint64(0); i < uint64(len(servers)); i++ {
		weights[i] = float64(uint64(n)-i) / float64(n)
		servers[i] = "localhost:" + strconv.FormatUint(60000-i, 10)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		SortSliceByWeightValue(servers, weights, hash)
	}
}

func benchmarkSortHashersByWeightValueRelection(b *testing.B, n int, hash uint64) {
	servers := make([]hashString, n)
	weights := make([]float64, n)
	for i := uint64(0); i < uint64(len(servers)); i++ {
		weights[i] = float64(uint64(n)-i) / float64(n)
		servers[i] = hashString("localhost:" + strconv.FormatUint(60000-i, 10))
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		SortSliceByWeightValue(servers, weights, hash)
	}
}

func benchmarkSortHashersByWeightValueTyped(b *testing.B, n int, hash uint64) {
	servers := make([]hashString, n)
	weights := make([]float64, n)
	for i := uint64(0); i < uint64(len(servers)); i++ {
		weights[i] = float64(uint64(n)-i) / float64(n)
		servers[i] = hashString("localhost:" + strconv.FormatUint(60000-i, 10))
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		SortHasherSliceByWeightValue(servers, weights, hash)
	}
}

func benchmarkSortHashersByValueReflection(b *testing.B, n int, hash uint64) {
	servers := make([]hashString, n)
	for i := uint64(0); i < uint64(len(servers)); i++ {
		servers[i] = hashString("localhost:" + strconv.FormatUint(60000-i, 10))
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		SortSliceByValue(servers, hash)
	}
}

func benchmarkSortHashersByValueTyped(b *testing.B, n int, hash uint64) {
	servers := make([]hashString, n)
	for i := uint64(0); i < uint64(len(servers)); i++ {
		servers[i] = hashString("localhost:" + strconv.FormatUint(60000-i, 10))
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		SortHasherSliceByValue(servers, hash)
	}
}
