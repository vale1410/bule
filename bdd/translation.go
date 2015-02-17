package bdd

import (
	//	"fmt"
	"github.com/vale1410/bule/constraints"
	//	"github.com/vale1410/bule/sat"
	//	"strconv"
)

var totalNodes int

func (b *BddStore) CreateBdd(K int64, entries []constraints.Entry) (int, int64, int64) {

	l := len(entries) ///level
	totalNodes++

	//fmt.Println(l, K, entries)

	//check if node already exists
	if id, wmin, wmax := b.GetByWeight(l, K); id != -1 {

		//	fmt.Println("exists", l, K, "[", wmin, wmax, "]")
		return id, wmin, wmax

	} else {

		id_left, wmin_left, wmax_left := b.CreateBdd(K-entries[0].Weight, entries[1:])
		id_right, wmin_right, wmax_right := b.CreateBdd(K, entries[1:])

		var n Node

		n.level = l
		n.wmin = maxx(wmin_left+entries[0].Weight, wmin_right)
		n.wmax = min(wmax_left+entries[0].Weight, wmax_right)
		n.right = id_right
		n.left = id_left

		return b.Insert(n), n.wmin, n.wmax
	}
}

func min(a, b int64) int64 {
	if a <= b {
		return a
	} else {
		return b
	}
}

func maxx(a, b int64) int64 {
	if a >= b {
		return a
	} else {
		return b
	}
}
