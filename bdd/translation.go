package bdd

import (
	//	"fmt"
	"github.com/vale1410/bule/constraints"
	//	"github.com/vale1410/bule/sat"
)

var totalNodes int

func Translate(pb constraints.Threshold) (b BddStore) {

	b = Init(len(pb.Entries))

	b.createBdd(pb.K, pb.Entries)

	return
}

func (b *BddStore) createBdd(K int64, entries []constraints.Entry) (id int, wmin, max int64) {

	l := len(entries) ///level
	totalNodes++

	//check if node already exists
	if id, wmin, wmax := b.GetByWeight(l, K); id != -1 {

		return id, wmin, max

	} else {

		id_left, wmin_left, wmax_left := b.createBdd(K-entries[0].Weight, entries[1:])

		id_right, wmin_right, wmax_right := b.createBdd(K, entries[1:])

		//fmt.Print(n.Left, " (", left.Lb+entries[0].w, left.Ub+entries[0].w, ")  ", n.Right, "(", right.Lb, right.Ub, ") ->")

		var n Node

		n.wmin = maxx(wmin_left+entries[0].Weight, wmin_right)
		n.wmax = min(wmax_left+entries[0].Weight, wmax_right)
		n.right = id_right
		n.left = id_left

		newNode := b.Insert(n)
		//fmt.Println(newNode, "(", n.Lb, n.Ub, ")", n.Best)
		return newNode, wmin, wmax
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
