# Golang HRW implementation

[![Build Status](https://travis-ci.org/nspcc-dev/hrw.svg?branch=master)](https://travis-ci.org/nspcc-dev/hrw)
[![codecov](https://codecov.io/gh/nspcc-dev/hrw/badge.svg)](https://codecov.io/gh/nspcc-dev/hrw)
[![Report](https://goreportcard.com/badge/github.com/nspcc-dev/hrw)](https://goreportcard.com/report/github.com/nspcc-dev/hrw)
[![GitHub release](https://img.shields.io/github/release/nspcc-dev/hrw.svg)](https://github.com/nspcc-dev/hrw)

[Rendezvous or highest random weight](https://en.wikipedia.org/wiki/Rendezvous_hashing) (HRW) hashing is an algorithm that allows clients to achieve distributed agreement on a set of k options out of a possible set of n options. A typical application is when clients need to agree on which sites (or proxies) objects are assigned to. When k is 1, it subsumes the goals of consistent hashing, using an entirely different method.

## Install

`go get github.com/nspcc-dev/hrw`

## Benchmark:

```
BenchmarkSort_fnv_10-8                                           4812801               240.9 ns/op           216 B/op          4 allocs/op
BenchmarkSort_fnv_100-8                                           434767              2600 ns/op            1848 B/op          4 allocs/op
BenchmarkSort_fnv_1000-8                                           20428             66116 ns/op           16440 B/op          4 allocs/op
BenchmarkSortByIndex_fnv_10-8                                    2505410               486.5 ns/op           352 B/op          7 allocs/op
BenchmarkSortByIndex_fnv_100-8                                    254556              4697 ns/op            1984 B/op          7 allocs/op
BenchmarkSortByIndex_fnv_1000-8                                    13581             88334 ns/op           16576 B/op          7 allocs/op
BenchmarkSortByValue_fnv_10-8                                    1761030               682.1 ns/op           592 B/op         18 allocs/op
BenchmarkSortByValue_fnv_100-8                                    258838              4675 ns/op            4480 B/op        108 allocs/op
BenchmarkSortByValue_fnv_1000-8                                    27027             44649 ns/op           40768 B/op       1008 allocs/op
BenchmarkSortHashersByValue_Reflection_fnv_10-8                  1013560              1249 ns/op             768 B/op         29 allocs/op
BenchmarkSortHashersByValue_Reflection_fnv_100-8                  106029             11414 ns/op            6096 B/op        209 allocs/op
BenchmarkSortHashersByValue_Reflection_fnv_1000-8                  10000            108977 ns/op           56784 B/op       2009 allocs/op
BenchmarkSortHashersByValue_Typed_fnv_10-8                       1577814               700.3 ns/op           584 B/op         17 allocs/op
BenchmarkSortHashersByValue_Typed_fnv_100-8                       215938              5024 ns/op            4472 B/op        107 allocs/op
BenchmarkSortHashersByValue_Typed_fnv_1000-8                       24447             46889 ns/op           40760 B/op       1007 allocs/op

BenchmarkSortByWeight_fnv_10-8                                   2924833               370.6 ns/op           448 B/op          8 allocs/op
BenchmarkSortByWeight_fnv_100-8                                   816069              1516 ns/op            2896 B/op          8 allocs/op
BenchmarkSortByWeight_fnv_1000-8                                   80391             17478 ns/op           24784 B/op          8 allocs/op
BenchmarkSortByWeightIndex_fnv_10-8                              1945612               550.3 ns/op           368 B/op          7 allocs/op
BenchmarkSortByWeightIndex_fnv_100-8                              140473              8084 ns/op            2000 B/op          7 allocs/op
BenchmarkSortByWeightIndex_fnv_1000-8                               5518            200949 ns/op           16592 B/op          7 allocs/op
BenchmarkSortByWeightValue_fnv_10-8                              1305580               909.8 ns/op           608 B/op         18 allocs/op
BenchmarkSortByWeightValue_fnv_100-8                              165410              6796 ns/op            4496 B/op        108 allocs/op
BenchmarkSortByWeightValue_fnv_1000-8                              17922             78555 ns/op           40784 B/op       1008 allocs/op
BenchmarkSortHashersByWeightValueReflection_fnv_10-8              454976              2229 ns/op             784 B/op         29 allocs/op
BenchmarkSortHashersByWeightValueReflection_fnv_100-8              76264             15332 ns/op            6112 B/op        209 allocs/op
BenchmarkSortHashersByWeightValueReflection_fnv_1000-8             80288             13192 ns/op            6112 B/op        209 allocs/op
BenchmarkSortHashersByWeightValueTyped_fnv_10-8                  1433113               901.4 ns/op           600 B/op         17 allocs/op
BenchmarkSortHashersByWeightValueTyped_fnv_100-8                  188626              5896 ns/op            4488 B/op        107 allocs/op
BenchmarkSortHashersByWeightValueTyped_fnv_1000-8                 178131              6518 ns/op            4488 B/op        107 allocs/op
```

## Example

```go
package main

import (
	"fmt"
	
	"github.com/TrueCloudLab/hrw"
)

func main() {
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
		h   = hrw.Hash(key)
	)

	hrw.SortSliceByValue(servers, h)
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
```