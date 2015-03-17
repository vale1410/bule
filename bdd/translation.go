package bdd

import (
	"errors"
	//	"fmt"
	"github.com/vale1410/bule/constraints"
	"github.com/vale1410/bule/sat"
	"math"
)

// TODO: current assumption: amo is in order with entries, starts somewhere
// TODO: amo is same polarity as entries, and coefficients in entries are ascending for amo
func (b *BddStore) CreateBddAMO(K int64, entries []constraints.Entry, amo []sat.Literal) (int, int64, int64, error) {

	l := len(entries) ///level

	if b.MaxNodes < len(b.Nodes) {
		return 0, 0, 0, errors.New("Bdd max nodes reached")
	}

	//fmt.Println(l, K, entries)

	if id, wmin_cache, wmax_cache := b.GetByWeight(l, K); id != -1 {

		//	fmt.Println("exists", l, K, "[", wmin, wmax, "]")

		return id, wmin_cache, wmax_cache, nil

	} else {
		//domain of variable [0,1], extend to [0..n] soon (MDDs)
		// entry of variable domain, atom: Dom: 2

		var n Node
		var err error

		if len(amo) > 0 && amo[0] == entries[0].Literal { //amo mode
			// iterate over the amo
			n.level = l
			n.children = make([]int, len(amo)+1)
			n.children[0], n.wmin, n.wmax, err = b.CreateBddAMO(K, entries[len(amo):], []sat.Literal{})

			if err != nil {
				return 0, 0, 0, err
			}

			for i, _ := range amo {

				if amo[i] != entries[i].Literal {
					panic("amo and PB are not aligned!!!! ")
				}

				var wmin2, wmax2 int64
				n.children[i+1], wmin2, wmax2, err = b.CreateBddAMO(K-entries[i].Weight, entries[len(amo):], amo)
				n.wmin = maxx(n.wmin, wmin2+entries[i].Weight)
				n.wmax = min(n.wmax, wmax2+entries[i].Weight)

				if err != nil {
					return 0, 0, 0, err
				}

			}

		} else { //usual mode
			dom := 2
			n.level = l
			n.children = make([]int, dom)
			n.wmin = math.MinInt64
			n.wmax = math.MaxInt64

			var err error
			for i := int64(0); i < int64(dom); i++ {
				var wmin2, wmax2 int64

				n.children[i], wmin2, wmax2, err = b.CreateBddAMO(K-i*entries[0].Weight, entries[1:], amo)

				n.wmin = maxx(n.wmin, wmin2+i*entries[0].Weight)
				n.wmax = min(n.wmax, wmax2+i*entries[0].Weight)

				if err != nil {
					return 0, 0, 0, err
				}
			}
		}

		return b.Insert(n), n.wmin, n.wmax, nil
	}
}

func (b *BddStore) CreateBdd(K int64, entries []constraints.Entry) (int, int64, int64, error) {

	l := len(entries) ///level

	if b.MaxNodes < len(b.Nodes) {
		return 0, 0, 0, errors.New("Bdd max nodes reached")
	}

	//fmt.Println(l, K, entries)

	if id, wmin_cache, wmax_cache := b.GetByWeight(l, K); id != -1 {

		//	fmt.Println("exists", l, K, "[", wmin, wmax, "]")
		return id, wmin_cache, wmax_cache, nil

	} else {
		//domain of variable [0,1], extend to [0..n] soon (MDDs)
		// entry of variable domain, atom: Dom: 2

		dom := 2

		var n Node
		n.level = l
		n.children = make([]int, dom)
		n.wmin = math.MinInt64
		n.wmax = math.MaxInt64

		var err error
		for i := int64(0); i < int64(dom); i++ {
			var wmin2, wmax2 int64
			n.children[i], wmin2, wmax2, err = b.CreateBdd(K-i*entries[0].Weight, entries[1:])
			n.wmin = maxx(n.wmin, wmin2+i*entries[0].Weight)
			n.wmax = min(n.wmax, wmax2+i*entries[0].Weight)

			if err != nil {
				return 0, 0, 0, err
			}

			//			}
		}

		return b.Insert(n), n.wmin, n.wmax, nil
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
