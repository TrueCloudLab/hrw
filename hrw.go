// Package hrw implements Rendezvous hashing.
// http://en.wikipedia.org/wiki/Rendezvous_hashing.
package hrw

import (
	"encoding/binary"
	"errors"
	"math"
	"reflect"
	"sort"

	"github.com/spaolacci/murmur3"
)

type (
	// Hasher interface used by SortSliceByValue
	Hasher interface{ Hash() uint64 }

	sorter struct {
		l    int
		less func(i, j int) bool
		swap func(i, j int)
	}
)

// Boundaries of valid normalized weights
const (
	NormalizedMaxWeight = 1.0
	NormalizedMinWeight = 0.0
)

func (s *sorter) Len() int           { return s.l }
func (s *sorter) Less(i, j int) bool { return s.less(i, j) }
func (s *sorter) Swap(i, j int)      { s.swap(i, j) }

func distance(x uint64, y uint64) uint64 {
	acc := x ^ y
	// here used mmh3 64 bit finalizer
	// https://github.com/aappleby/smhasher/blob/61a0530f28277f2e850bfc39600ce61d02b518de/src/MurmurHash3.cpp#L81
	acc ^= acc >> 33
	acc = acc * 0xff51afd7ed558ccd
	acc ^= acc >> 33
	acc = acc * 0xc4ceb9fe1a85ec53
	acc ^= acc >> 33
	return acc
}

// Hash uses murmur3 hash to return uint64
func Hash(key []byte) uint64 {
	return murmur3.Sum64(key)
}

// Sort receive nodes and hash, and sort it by distance
func Sort(nodes []uint64, hash uint64) []uint64 {
	l := len(nodes)
	sorted := make([]uint64, l)
	dist := make([]uint64, l)
	for i := range nodes {
		sorted[i] = uint64(i)
		dist[i] = distance(nodes[i], hash)
	}

	sort.Slice(sorted, func(i, j int) bool {
		return dist[sorted[i]] < dist[sorted[j]]
	})
	return sorted
}

// SortByWeight receive nodes, weights and hash, and sort it by distance * weight
func SortByWeight(nodes []uint64, weights []float64, hash uint64) []uint64 {
	result := make([]uint64, len(nodes))
	copy(nodes, result)
	sortByWeight(len(nodes), false, nodes, weights, hash, reflect.Swapper(result))
	return result
}

// SortSliceByValue received []T and hash to sort by value-distance
func SortSliceByValue(slice interface{}, hash uint64) {
	rule := prepareRule(slice)
	if rule != nil {
		swap := reflect.Swapper(slice)
		sortByDistance(len(rule), false, rule, hash, swap)
	}
}

// SortHasherSliceByValue receives []Hasher and hash to sort by value-distance.
func SortHasherSliceByValue[T Hasher](slice []T, hash uint64) {
	rule := prepareHasherRule(slice)
	if rule != nil {
		swap := func(i, j int) {
			slice[i], slice[j] = slice[j], slice[i]
		}
		sortByDistance(len(rule), false, rule, hash, swap)
	}
}

// SortSliceByWeightValue received []T, weights and hash to sort by value-distance * weights
func SortSliceByWeightValue(slice interface{}, weights []float64, hash uint64) {
	rule := prepareRule(slice)
	if rule != nil {
		swap := reflect.Swapper(slice)
		sortByWeight(reflect.ValueOf(slice).Len(), false, rule, weights, hash, swap)
	}
}

// SortHasherSliceByWeightValue receives []Hasher, weights and hash to sort by value-distance * weights.
func SortHasherSliceByWeightValue[T Hasher](slice []T, weights []float64, hash uint64) {
	rule := prepareHasherRule(slice)
	if rule != nil {
		swap := func(i, j int) {
			slice[i], slice[j] = slice[j], slice[i]
		}
		sortByWeight(len(slice), false, rule, weights, hash, swap)
	}
}

// SortSliceByIndex received []T and hash to sort by index-distance
func SortSliceByIndex(slice interface{}, hash uint64) {
	length := reflect.ValueOf(slice).Len()
	swap := reflect.Swapper(slice)
	sortByDistance(length, true, nil, hash, swap)
}

// SortSliceByWeightIndex received []T, weights and hash to sort by index-distance * weights
func SortSliceByWeightIndex(slice interface{}, weights []float64, hash uint64) {
	length := reflect.ValueOf(slice).Len()
	swap := reflect.Swapper(slice)
	sortByWeight(length, true, nil, weights, hash, swap)
}

func prepareRule(slice interface{}) []uint64 {
	t := reflect.TypeOf(slice)
	if t.Kind() != reflect.Slice {
		panic("HRW sort expects slice, got " + t.Kind().String())
	}

	var (
		val    = reflect.ValueOf(slice)
		length = val.Len()
		rule   = make([]uint64, 0, length)
	)

	if length == 0 {
		return nil
	}

	switch slice := slice.(type) {
	case []int:
		var key = make([]byte, 16)
		for i := 0; i < length; i++ {
			binary.BigEndian.PutUint64(key, uint64(slice[i]))
			rule = append(rule, Hash(key))
		}
	case []uint:
		var key = make([]byte, 16)
		for i := 0; i < length; i++ {
			binary.BigEndian.PutUint64(key, uint64(slice[i]))
			rule = append(rule, Hash(key))
		}
	case []int8:
		for i := 0; i < length; i++ {
			key := byte(slice[i])
			rule = append(rule, Hash([]byte{key}))
		}
	case []uint8:
		for i := 0; i < length; i++ {
			key := slice[i]
			rule = append(rule, Hash([]byte{key}))
		}
	case []int16:
		var key = make([]byte, 8)
		for i := 0; i < length; i++ {
			binary.BigEndian.PutUint16(key, uint16(slice[i]))
			rule = append(rule, Hash(key))
		}
	case []uint16:
		var key = make([]byte, 8)
		for i := 0; i < length; i++ {
			binary.BigEndian.PutUint16(key, slice[i])
			rule = append(rule, Hash(key))
		}
	case []int32:
		var key = make([]byte, 16)
		for i := 0; i < length; i++ {
			binary.BigEndian.PutUint32(key, uint32(slice[i]))
			rule = append(rule, Hash(key))
		}
	case []uint32:
		var key = make([]byte, 16)
		for i := 0; i < length; i++ {
			binary.BigEndian.PutUint32(key, slice[i])
			rule = append(rule, Hash(key))
		}
	case []int64:
		var key = make([]byte, 32)
		for i := 0; i < length; i++ {
			binary.BigEndian.PutUint64(key, uint64(slice[i]))
			rule = append(rule, Hash(key))
		}
	case []uint64:
		var key = make([]byte, 32)
		for i := 0; i < length; i++ {
			binary.BigEndian.PutUint64(key, slice[i])
			rule = append(rule, Hash(key))
		}
	case []string:
		for i := 0; i < length; i++ {
			rule = append(rule, Hash([]byte(slice[i])))
		}

	default:
		if _, ok := val.Index(0).Interface().(Hasher); !ok {
			panic("slice elements must implement hrw.Hasher")
		}

		for i := 0; i < length; i++ {
			h := val.Index(i).Interface().(Hasher)
			rule = append(rule, h.Hash())
		}
	}
	return rule
}

func prepareHasherRule[T Hasher](hashers []T) []uint64 {
	length := len(hashers)
	if length == 0 {
		return nil
	}
	result := make([]uint64, length)
	for i := 0; i < length; i++ {
		result[i] = hashers[i].Hash()
	}
	return result
}

// ValidateWeights checks if weights are normalized between 0.0 and 1.0
func ValidateWeights(weights []float64) error {
	for i := range weights {
		if math.IsNaN(weights[i]) || weights[i] > NormalizedMaxWeight || weights[i] < NormalizedMinWeight {
			return errors.New("weights are not normalized")
		}
	}
	return nil
}

func newSorter(l int, byIndex bool, nodes []uint64, h uint64,
	swap func(i, j int)) (*sorter, []int, []uint64) {
	ind := make([]int, l)
	dist := make([]uint64, l)
	for i := 0; i < l; i++ {
		ind[i] = i
		dist[i] = getDistance(byIndex, i, nodes, h)
	}

	return &sorter{
		l: l,
		swap: func(i, j int) {
			swap(i, j)
			ind[i], ind[j] = ind[j], ind[i]
		},
	}, ind, dist
}

// sortByWeight sorts nodes by weight using provided swapper.
// nodes contains hrw hashes. If it is nil, indices are used.
func sortByWeight(l int, byIndex bool, nodes []uint64, weights []float64, hash uint64, swap func(i, j int)) {
	// if all nodes have the same distance then sort uniformly
	if allSameF64(weights) {
		sortByDistance(l, byIndex, nodes, hash, swap)
		return
	}

	s, ind, dist := newSorter(l, byIndex, nodes, hash, swap)
	s.less = func(i, j int) bool {
		ii, jj := ind[i], ind[j]
		// `maxUint64 - distance` makes the shorter distance more valuable
		// it is necessary for operation with normalized values
		wi := float64(^uint64(0)-dist[ii]) * weights[ii]
		wj := float64(^uint64(0)-dist[jj]) * weights[jj]
		return wi > wj // higher distance must be placed lower to be first
	}
	sort.Sort(s)
}

// sortByDistance sorts nodes by hrw distance using provided swapper.
// nodes contains hrw hashes. If it is nil, indices are used.
func sortByDistance(l int, byIndex bool, nodes []uint64, hash uint64, swap func(i, j int)) {
	s, ind, dist := newSorter(l, byIndex, nodes, hash, swap)
	s.less = func(i, j int) bool {
		return dist[ind[i]] < dist[ind[j]]
	}
	sort.Sort(s)
}

// getDistance return distance from nodes[i] to h.
// If byIndex is true, nodes index is used.
// Else if nodes[i] != nil, distance is calculated from this value.
// Otherwise, and hash from node index is taken.
func getDistance(byIndex bool, i int, nodes []uint64, h uint64) uint64 {
	if nodes != nil {
		return distance(nodes[i], h)
	} else if byIndex {
		return distance(uint64(i), h)
	} else {
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, uint64(i))
		return distance(Hash(buf), h)
	}
}

func allSameF64(fs []float64) bool {
	for i := range fs {
		if fs[i] != fs[0] {
			return false
		}
	}
	return true
}
