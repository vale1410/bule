package sorters

func buildTriangleBitonic(newId *int, array []int, comparators *[]Comparator, lo int, hi int) {
	if (hi - lo) >= 1 {
		//fmt.Println("compare", lo, hi)
		compareAndSwap(newId, array, comparators, lo, hi)
		buildTriangleBitonic(newId, array, comparators, lo+1, hi-1)
	}
}

func waterfallBitonic(newId *int, array []int, comparators *[]Comparator, lo int, hi int) {
	if (hi - lo) >= 1 {
		//fmt.Println("waterfall", lo, hi)
		mid := lo + ((hi - lo) / 2)
		for i := 0; i <= mid-lo; i++ {
			//fmt.Println("compare", lo+i, mid+i+1)
			compareAndSwap(newId, array, comparators, lo+i, mid+i+1)
		}
		waterfallBitonic(newId, array, comparators, lo, mid)
		waterfallBitonic(newId, array, comparators, mid+1, hi)
	}
}

func triangleBitonic(newId *int, array []int, comparators *[]Comparator, lo int, hi int) {
	if (hi - lo) >= 1 {
		//fmt.Println("triangle", lo, hi)
		mid := lo + ((hi - lo) / 2)
		triangleBitonic(newId, array, comparators, lo, mid)
		triangleBitonic(newId, array, comparators, mid+1, hi)
		buildTriangleBitonic(newId, array, comparators, lo, hi)
		waterfallBitonic(newId, array, comparators, lo, mid)
		waterfallBitonic(newId, array, comparators, mid+1, hi)
	}
}
