package sorters

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
)

func TestBubbleSortSize(t *testing.T) {

	sorter := CreateSortingNetwork(5, -1, Bubble)
	fmt.Println(sorter)

	PrintSorterTikZ(sorter, "test.tex")

}

//func TestRemoveOutput(t *testing.T) {
//
//	size := 5
//	k := 2
//	typ := OddEven
//
//	sorter := CreateCardinalityNetwork(size, k, AtMost, typ)
//	fmt.Println(sorter)
//	sorter.RemoveOutput()
//	fmt.Println(sorter)
//
//	if len(sorter.Comparators) == 0 {
//		panic("removeOutput or propagateBackwards is broken!")
//	}
//
//}

//func TestSizes(t *testing.T) {
//
//	fmt.Println("Comparing sizes: Bitonic/OddEven/Bubble/Pairwise")
//	var typs [3]SortingNetworkType
//	typs[0] = OddEven
//	typs[1] = Pairwise
//	typs[2] = Bitonic
//	//typs[3] = Bubble
//
//	N := 10
//
//	fmt.Println("Merging Networks")
//
//	for n := 2; n <= N; n++ {
//		for k := 1; k < n; k++ {
//			fmt.Print(n, "/", k, ";")
//			for _, typ := range typs {
//				sorter := CreateSortingNetwork(n, k, typ)
//				fmt.Print(len(sorter.Comparators), ";")
//			}
//			fmt.Println()
//		}
//	}
//
//	fmt.Println("Cardinality Networks")
//
//	for n := 2; n <= N; n++ {
//		for k := 1; k < n; k++ {
//			fmt.Print(n, "/", k, ";")
//			for _, typ := range typs {
//				sorter := CreateCardinalityNetwork(n, k, AtMost, typ)
//				fmt.Print(len(sorter.Comparators), ";")
//			}
//			fmt.Println()
//		}
//	}
//
//}

//func TestNormalize(t *testing.T) {
//
//	fmt.Println("Test: Normalization")
//
//	sorter := CreateSortingNetwork(8, -1, Pairwise)
//
//	ids1 := make(map[int]bool, len(sorter.Comparators)*4)
//
//	for _, comp := range sorter.Comparators {
//		ids1[comp.A] = true
//		ids1[comp.B] = true
//		ids1[comp.C] = true
//		ids1[comp.D] = true
//	}
//
//	offset := 100
//
//	var in []int
//
//	max := sorter.Normalize(offset, in)
//
//	if max-offset != len(ids1) {
//		t.Error("Normalize failed")
//	}
//
//}

//func TestStuff(t *testing.T) {
//
//	fmt.Println("Print to TeX shit")
//	sorter := CreateSortingNetwork(16,-1,Pairwise)
//	PrintSorterTikZ(sorter,"pairwise16.tex")
//}

// TestCardinality check constraint sum n <= k
// TestAtLeast check constraint sum n >= k
//func TestCardinality(t *testing.T) {
//
//	fmt.Println("Test: Bitonic/OddEven/Bubble/Pairwise")
//	var typs [4]SortingNetworkType
//	typs[0] = OddEven
//	typs[1] = Bitonic
//	typs[2] = Bubble
//	typs[3] = Pairwise
//
//	for _, typ := range typs {
//
//		sizes := []int{3, 4, 6, 9, 9, 9, 33, 68, 123, 250}
//		ks := []int{2, 2, 3, 2, 6, 7, 29, 8, 8, 100}
//
//		for i, size := range sizes {
//			cardinalityAtMost(size, ks[i], t, typ)
//			cardinalityAtLeast(size, ks[i], t, typ)
//			cutSorting(size, ks[i], t, typ)
//			normalSorting(size, t, typ)
//		}
//
//		for x := 5; x < 100; x = x + 20 {
//			for y := 1; y < x; y = y + 6 {
//				sizes = []int{x}
//				ks = []int{y}
//
//				for i, size := range sizes {
//					cardinalityAtMost(size, ks[i], t, typ)
//					cardinalityAtLeast(size, ks[i], t, typ)
//					cutSorting(size, ks[i], t, typ)
//					normalSorting(size, t, typ)
//				}
//			}
//		}
//	}
//}

func cardinalityAtLeast(size int, k int, t *testing.T, typ SortingNetworkType) {

	array1 := rand.Perm(size)
	array2 := make([]int, size)

	copy(array2, array1)
	sort.Ints(array2)

	mapping := make(map[int]int, size)

	sorter := CreateSortingNetwork(size, -1, typ)

	for i := 0; i < k; i++ {
		mapping[sorter.Out[i]] = 1
		sorter.Out[i] = 1
		array2[i] = 1
	}

	sorter.PropagateBackwards(mapping)
	sortAndCompareArrays(sorter, array1, array2, t)
}

func cardinalityAtMost(size int, k int, t *testing.T, typ SortingNetworkType) {

	array1 := rand.Perm(size)
	array2 := make([]int, size)

	copy(array2, array1)
	sort.Ints(array2)

	mapping := make(map[int]int, size)

	sorter := CreateSortingNetwork(size, -1, typ)

	for i := size - k; i < size; i++ {
		mapping[sorter.Out[i]] = 0
		sorter.Out[i] = 0
		array2[i] = 0
	}
	sorter.PropagateBackwards(mapping)
	sortAndCompareArrays(sorter, array1, array2, t)
}

func cutSorting(size int, cut int, t *testing.T, typ SortingNetworkType) {

	array1 := rand.Perm(cut)
	array2 := rand.Perm(size - cut)
	sort.Ints(array1)
	sort.Ints(array2)
	result := make([]int, size)

	array3 := append(array1, array2...)
	copy(result, array3)

	//fmt.Println("array1",array1)
	//fmt.Println("array2",array2)
	//fmt.Println("array3",array3)
	sort.Ints(result)
	//fmt.Println("result",result)

	sorter := CreateSortingNetwork(size, cut, typ)
	//PrintSorterTikZ(sorter,"tmp/cutSorter.tex")
	sortAndCompareArrays(sorter, array3, result, t)
}

func normalSorting(size int, t *testing.T, typ SortingNetworkType) {

	array1 := rand.Perm(size)
	array2 := make([]int, size)
	copy(array2, array1)
	sort.Ints(array2)
	sorter := CreateSortingNetwork(len(array1), -1, typ)
	sortAndCompareArrays(sorter, array1, array2, t)
}

func sortAndCompareArrays(sorter Sorter, array1, array2 []int, t *testing.T) {

	mapping := make(map[int]int, len(sorter.Comparators))

	for i, x := range sorter.In {
		mapping[x] = array1[i]
	}

	for _, comp := range sorter.Comparators {

		b, bok := mapping[comp.B]
		a, aok := mapping[comp.A]

		if !aok {
			t.Error("not in mapping", comp.A)
		}

		if !bok {
			t.Error("not in mapping", comp.B)
		}

		if comp.D > 1 { // 0,1, specific meaning
			mapping[comp.D] = max(a, b)
		}
		if comp.C > 1 { // 0,1, specific meaning
			mapping[comp.C] = min(a, b)
		}

	}

	output := make([]int, len(array1))

	e := false

	for i, x := range sorter.Out {
		if x <= 1 {
			output[i] = x
		} else {
			output[i] = mapping[x]
		}
		if output[i] != array2[i] {
			t.Error("Output array does not coincide in position", i)
			e = true
		}
	}

	if e {
		t.Error("ideal", len(array2), array2)
		t.Error("output", len(output), output)
		t.Error("sorter", sorter)
		t.Error("mapping", mapping)
		if len(sorter.Comparators) < 100 {
			printSorterDot(sorter, "sorter")
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func min(a, b int) int {
	if a > b {
		return b
	} else {
		return a
	}
}
