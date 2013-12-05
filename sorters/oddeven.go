package sorters

func oddevenMerge(newId *int, array []int, comparators *[]Comparator, lo int, hi int, r int) {
	step := r * 2
	if step < hi-lo {
		oddevenMerge(newId, array, comparators, lo, hi, step)
		oddevenMerge(newId, array, comparators, lo+r, hi-r, step)
		for i := lo + r; i <= hi-r; i += step {
			compareAndSwap(newId, array, comparators, i, i+r)
		}
	} else {
		compareAndSwap(newId, array, comparators, lo, lo+r)
	}
}

func oddevenSort(newId *int, array []int, comparators *[]Comparator, lo int, hi int) {
	if (hi - lo) >= 1 {
		mid := lo + ((hi - lo) / 2)
		oddevenSort(newId, array, comparators, lo, mid)
		oddevenSort(newId, array, comparators, mid+1, hi)
		oddevenMerge(newId, array, comparators, lo, hi, 1)
	}
}
