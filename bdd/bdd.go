package bdd

import (
	"fmt"
	"github.com/yasushi-saito/rbtree"
	"math"
)

type Node struct {
	id    int
	level int
	wmin  int64
	wmax  int64
	right int // if variable in level is true
	left  int // if variable in level is false
}

type BddStore struct {
	nextId  int
	storage *rbtree.Tree
}

func Init(size int) (bddStore BddStore) {
	bddStore.storage = rbtree.NewTree(Compare)
	bddStore.Insert(Node{level: 0, wmin: math.MinInt64, wmax: -1})
	bddStore.Insert(Node{level: 0, wmin: 0, wmax: math.MaxInt64})
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
			if n.(Node).wmax >= math.MaxInt64 {
				fmt.Println("level", n.(Node).level, "[", n.(Node).wmin, ", +∞]")
			} else {
				fmt.Println("level", n.(Node).level, "[", n.(Node).wmin, ",", n.(Node).wmax, "]")
			}
			return false
		}

		//iter := bddStore.storage.FindGE(Node{})
		iter := bddStore.storage.Min()

		for iter.Limit() {
			fmt.Println("hello")
			anon(iter.Item())
			iter.Next()
		}
	}

}

// returns node, if exists
func (bddStore *BddStore) GetByWeight(level int, weight int64) (id int, wmin, wmax int64) {
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

func (bddStore *BddStore) Insert(n Node) (id int) {
	//debug code
	//n := Node{level: level, wmin: wmin, wmax: wmax}
	//bbdStore
	if a := bddStore.storage.Get(n); a != nil {
		fmt.Printf("%#v \n", a)
		panic("node should not exist")
	}
	//debug code end
	n.id = bddStore.nextId
	bddStore.nextId++
	bddStore.storage.Insert(n)
	return n.id
}
