package mdd

import (
	"fmt"
	"github.com/vale1410/bule/glob"
	"github.com/yasushi-saito/rbtree"
	"math"
)

type Node struct {
	Id       int
	Level    int
	Wmin     int64
	Wmax     int64
	Children []int
	//right int // if variable in level is true
	//left  int // if variable in level is false
}

func (node Node) IsZero() bool {
	return node.Wmin <= math.MinInt64+10000000
}

func (node Node) IsOne() bool {
	return node.Wmax >= math.MaxInt64-10000000
}

func printNode(node Node) {
	if node.IsZero() {
		fmt.Print(node.Id, "\t", node.Level, "\t[ -∞,", node.Wmax, "]")
	} else if node.IsOne() {
		fmt.Print(node.Id, "\t", node.Level, "\t[", node.Wmin, ", +∞]")
	} else {
		fmt.Print(node.Id, "\t", node.Level, "\t[", node.Wmin, ",", node.Wmax, "]")
	}
	fmt.Println(" c: ", node.Children)
}

type MddStore struct {
	NextId   int
	Nodes    []*Node
	MaxNodes int
	Top      int
	storage  *rbtree.Tree
}

func Init(size int) (b MddStore) {
	b.storage = rbtree.NewTree(Compare)
	b.Nodes = make([]*Node, 2)
	b.MaxNodes = glob.MDD_max_flag

	b.Nodes[0] = &Node{Id: 0, Level: 0, Wmin: math.MinInt64 + 100000, Wmax: -1} // id 0
	b.Nodes[1] = &Node{Id: 1, Level: 0, Wmin: 0, Wmax: math.MaxInt64 - 100000}  // id 1
	b.storage.Insert(*b.Nodes[0])
	b.storage.Insert(*b.Nodes[1])
	b.NextId = 2
	return
}

func Compare(aa, bb rbtree.Item) int {
	a := aa.(Node)
	b := bb.(Node)
	if a.Level > b.Level {
		return 1
	} else if a.Level < b.Level {
		return -1
	} else {
		if a.Wmin > b.Wmax {
			return 1
		} else if b.Wmin > a.Wmax {
			return -1
		} else { //they intersect and are equivalent
			return 0
		}
	}
}

// cleans the bdd from redundant nodes
func (store *MddStore) RemoveRedundants() (removed int) {

	rep := make(map[int]int, len(store.Nodes))
	for i, node := range store.Nodes {
		if i > 1 {
			equal := true
			id := (*node).Children[0]
			for j, child := range (*node).Children {
				equal = equal && (id == child)

				if child_new, b := rep[child]; b {
					(*node).Children[j] = child_new
				}
			}
			if equal {
				if id_deep, b := rep[id]; b {
					id = id_deep
				}
				rep[node.Id] = id
				*node = Node{}
				removed++
			}
		}
	}
	if id, b := rep[store.Top]; b {
		store.Top = id
	}

	return
}

//preparation for MDDs, gives out ids of decendancts
func (b *MddStore) ClauseIds(n Node) (v int, level int, des []int) {
	//children = []int{b.checkId(n.left), b.checkId(n.right)}
	return n.Id, n.Level, n.Children
}

func (b *MddStore) checkId(id int) int {
	if b.Nodes[id].IsZero() {
		return 0
	} else if b.Nodes[id].IsOne() {
		return 1
	} else {
		return id
	}
}

func (store *MddStore) Debug(withTable bool) {

	fmt.Println("Mdd Nodes:")
	count := 0
	if withTable {
		fmt.Println("id\tlevel\tinterval")
		for i, node := range store.Nodes {
			if i == 0 || (*node).Id > 1 {
				count++
				printNode(*node)
			}
		}

		//iter := mddStore.storage.Min()
		//for !iter.Limit() {
		//	printNode(iter.Item().(Node))
		//	iter = iter.Next()
		//}
	}
	fmt.Println("#nodes rb-data\t:", count)

}

// returns node, if exists
func (mddStore *MddStore) GetByWeight(level int, weight int64) (id int, wmin, wmax int64) {
	n := Node{Level: level, Wmin: weight, Wmax: weight}
	if a := mddStore.storage.Get(n); a != nil {
		id = a.(Node).Id
		wmin = a.(Node).Wmin
		wmax = a.(Node).Wmax
	} else {
		id = -1
	}
	return
}

func (mddStore *MddStore) Insert(n Node) (id int) {

	//check code start TODO: remove for performance
	if a := mddStore.storage.Get(n); a != nil {
		fmt.Println("FAIL")
		printNode(a.(Node))
		panic("node should not exist")
	}
	//check code end

	n.Id = mddStore.NextId
	mddStore.Nodes = append(mddStore.Nodes, &n)
	mddStore.NextId++

	glob.A(mddStore.NextId == len(mddStore.Nodes), "nextId calculation and length of Nodes list is wrong")

	mddStore.storage.Insert(n)

	return n.Id
}
