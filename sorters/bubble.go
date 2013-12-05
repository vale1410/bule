package sorters

func bubbleSort(newId *int, array []int, comparators *[]Comparator) {

	n := len(array)
	if n > 1 {
		for i := 0; i < n-1; i++ {
			compareAndSwap(newId, array, comparators, i, i+1)
		}
		bubbleSort(newId, array[:n-1], comparators)
	}
}
