package bdd

import (
	"errors"
	"github.com/vale1410/bule/constraints"
)

func (b *BddStore) CreateBdd(K int64, entries []constraints.Entry) (int, int64, int64, error) {

	l := len(entries) ///level

	if b.MaxNodes < len(b.Nodes) {
		return 0, 0, 0, errors.New("Bdd max nodes reached")
	}

	//fmt.Println(l, K, entries)

	if id, wmin, wmax := b.GetByWeight(l, K); id != -1 {

		//	fmt.Println("exists", l, K, "[", wmin, wmax, "]")
		return id, wmin, wmax, nil

	} else {

		id_left, wmin_left, wmax_left, err := b.CreateBdd(K, entries[1:])
		if err == nil {
			id_right, wmin_right, wmax_right, err := b.CreateBdd(K-entries[0].Weight, entries[1:])

			if err == nil {

				var n Node

				n.level = l
				n.wmin = maxx(wmin_left, wmin_right+entries[0].Weight)
				n.wmax = min(wmax_left, wmax_right+entries[0].Weight)
				n.right = id_right
				n.left = id_left

				return b.Insert(n), n.wmin, n.wmax, nil
			} else {
				return 0, 0, 0, err
			}
		} else {
			return 0, 0, 0, err
		}
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
