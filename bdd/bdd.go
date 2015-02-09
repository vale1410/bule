package bdd

import (
	"fmt"
	"github.com/yasushi-saito/rbtree"
	"math"
)

type Node struct {
	id    int
	level int
	wmin  int
	wmax  int
	pos   int // if variable in level is true
	neg   int // if variable in level is false
}

type BddStore struct {
	nextId  int
	storage *rbtree.Tree
}

func Init(size int) (bddStore BddStore) {
	bddStore.storage = rbtree.NewTree(Compare)
	return
}

func Compare(aa, bb rbtree.Item) int {
	a := aa.(Node)
	b := bb.(Node)
	if a.level > b.level {
		return 1
	} else if a.level < b.level {
		return -1
	} else {
		if a.wmin > b.wmax {
			return 1
		} else if b.wmin > a.wmax {
			return -1
		} else { //they intersect and are equivalent
			return 0
		}
	}
}

func (bddStore *BddStore) Debug(withTable bool) {

	fmt.Println("Bdd Nodes:")
	fmt.Println("#nodes rb-data\t:", bddStore.storage.Len())

	if withTable {
		anon := func(n rbtree.Item) bool {
			if n.(Node).wmax >= math.MaxInt32 {
				fmt.Println("level", n.(Node).level, "[", n.(Node).wmin, ", +∞]")
			} else {
				fmt.Println("level", n.(Node).level, "[", n.(Node).wmin, ",", n.(Node).wmax, "]")
			}
			return false
		}

		iter := bddStore.storage.FindGE(0)

		for iter.Limit() {
			anon(iter.Item())
			iter.Next()
		}
	}

}

// returns node, if exists
func (bddStore *BddStore) GetByWeight(level, weight int) (id, wmin, wmax int) {
	n := Node{level: level, wmin: weight, wmax: weight}
	if a := bddStore.storage.Get(n); a != nil {
		id = a.(Node).id
		wmin = a.(Node).wmin
		wmax = a.(Node).wmax
	} else {
		id = -1
	}
	return
}

func (bddStore *BddStore) Insert(level, wmin, wmax, pos, neg int) {
	//debug code
	n := Node{level: level, wmin: wmin, wmax: wmax}
	if a := bddStore.storage.Get(n); a != nil {
		fmt.Println(n)
		panic("node should not exist")
	}
	//debug code end

	bddStore.storage.Insert(Node{id: bddStore.nextId, level: level, wmin: wmin, wmax: wmax, pos: pos, neg: neg})
	bddStore.nextId++
}
