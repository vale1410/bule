package bdd

import (
	"fmt"
	"github.com/yasushi-saito/rbtree"
	"math"
)

type Node struct {
	id       int
	level    int
	wmin     int64
	wmax     int64
	children []int
	//right int // if variable in level is true
	//left  int // if variable in level is false
}

func (node Node) IsZero() bool {
	return node.wmin <= math.MinInt64+10000000
}

func (node Node) IsOne() bool {
	return node.wmax >= math.MaxInt64-10000000
}

func printNode(node Node) bool {
	if node.IsZero() {
		fmt.Print(node.id, "\t", node.level, "\t[ -∞,", node.wmax, "]")
	} else if node.IsOne() {
		fmt.Print(node.id, "\t", node.level, "\t[", node.wmin, ", +∞]")
	} else {
		fmt.Print(node.id, "\t", node.level, "\t[", node.wmin, ",", node.wmax, "]")
	}
	fmt.Println(" children ", node.children)
	return true
}

type BddStore struct {
	NextId   int
	Nodes    []*Node
	MaxNodes int
	storage  *rbtree.Tree
}

func Init(size int, maxNodes int) (b BddStore) {
	b.storage = rbtree.NewTree(Compare)
	b.Nodes = make([]*Node, 2)
	b.MaxNodes = maxNodes

	b.Nodes[0] = &Node{id: 0, level: 0, wmin: math.MinInt64 + 100000, wmax: -1} // id 0
	b.Nodes[1] = &Node{id: 1, level: 0, wmin: 0, wmax: math.MaxInt64 - 100000}  // id 1
	b.storage.Insert(*b.Nodes[0])
	b.storage.Insert(*b.Nodes[1])
	b.NextId = 2
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

//preparation for MDDs, gives out ids of decendancts
func (b *BddStore) ClauseIds(n Node) (v int, level int, des []int) {
	//children = []int{b.checkId(n.left), b.checkId(n.right)}
	return n.id, n.level, n.children
}

func (b *BddStore) checkId(id int) int {
	if b.Nodes[id].IsZero() {
		return 0
	} else if b.Nodes[id].IsOne() {
		return 1
	} else {
		return id
	}
}

func (bddStore *BddStore) Debug(withTable bool) {

	fmt.Println("Bdd Nodes:")
	fmt.Println("#nodes rb-data\t:", bddStore.storage.Len())

	if withTable {
		fmt.Println("id\tlevel\tinterval")

		iter := bddStore.storage.Min()
		for !iter.Limit() {
			printNode(iter.Item().(Node))
			iter = iter.Next()
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

	//check code start
	if a := bddStore.storage.Get(n); a != nil {
		fmt.Println("FAIL")
		printNode(a.(Node))
		panic("node should not exist")
	}
	//check code end
	n.id = bddStore.NextId
	bddStore.Nodes = append(bddStore.Nodes, &n)
	bddStore.NextId++
	if bddStore.NextId != len(bddStore.Nodes) {
		panic("nextId calculation and length of Nodes list is wrong")
	}
	bddStore.storage.Insert(n)
	return n.id
}
