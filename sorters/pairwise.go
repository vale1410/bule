package sorters

import (
//    "fmt"
)

func pairwiseMerge(newId *int, array []int, comparators *[]Comparator, lo int, hi int, r int) {
	step := r * 2
	if step < hi-lo {
		pairwiseMerge(newId, array, comparators, lo, hi, step)
		pairwiseMerge(newId, array, comparators, lo+r, hi-r, step)
		for i := lo + r; i <= hi-r; i += step {
			compareAndSwap(newId, array, comparators, i, i+r)
		}
	} else {
		//compareAndSwap(newId, array, comparators, lo, lo+r)
	}
}

func pairwiseSplit(newId *int, array []int, comparators *[]Comparator, lo int, hi int) {
    //fmt.Println("pairwiseSplit",lo,hi)
	mid := lo + ((hi - lo) / 2)
	for i := 0; i <= mid-lo; i++ {
        //fmt.Println("compareAndSwap Split",lo+i,mid+i+1)
		compareAndSwap(newId, array, comparators, lo+i, mid+i+1)
	}
}

func pairwiseSort(newId *int, array []int, comparators *[]Comparator, lo int, hi int) {
	if (hi - lo) >= 1 {
		mid := lo + ((hi - lo) / 2)
		pairwiseSplit(newId, array, comparators, lo, hi)
		pairwiseSort(newId, array, comparators, lo, mid)
		pairwiseSort(newId, array, comparators, mid+1, hi)
		pairwiseMerge(newId, array, comparators, lo, hi, 1)
	}
}
