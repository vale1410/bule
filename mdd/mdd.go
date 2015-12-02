package mdd

import (
	"fmt"
	"github.com/vale1410/bule/glob"
	//"math"
)

// 0,1 are ids given to FALSE and TRUE terminal

type Node struct {
	Id       int
	Level    int
	Children []int
	Parents  []int
}

func (node Node) IsZero() bool {
	return node.Id == 0
}

func (node Node) IsOne() bool {
	return node.Id == 1
}

func str(level int, children []int) string {
	return fmt.Sprintf("%d-%v", level, children)
}

func (n Node) str() string {
	return str(n.Level, n.Children)
}

// TODO: implement
func (mdd *MddStore) isConsistent() {

	// check if all levels nodes have same number of children
	// no loops (ids of children are smaller)

}

func (mdd *MddStore) NewNode(l int, children []int) int {
	if id, b := mdd.store[str(l, children)]; b {
		return id
	} else {
		id := mdd.NextId
		mdd.NextId++
		n := Node{Id: id, Level: l, Children: children}
		mdd.Nodes = append(mdd.Nodes, &n)
		mdd.store[n.str()] = n.Id
		return id
	}
}

type MddStore struct {
	NextId int
	Nodes  []*Node
	//	Levels   [][]*Node
	MaxNodes int
	Top      int
	store    map[string]int
}

func (node Node) Print() {
	fmt.Print(node.Id, "\t", node.Level, node.Children)
}

func Init() (b MddStore) {
	b.Nodes = make([]*Node, 2)
	b.store = make(map[string]int, 2)
	b.MaxNodes = *glob.MDD_max_flag
	b.Nodes[0] = &Node{Id: 0, Level: 0} // id 0
	b.Nodes[1] = &Node{Id: 1, Level: 0} // id 1
	b.store[b.Nodes[0].str()] = 0
	b.store[b.Nodes[1].str()] = 1
	b.NextId = 2
	return
}

func (mdd *MddStore) PrintDOT() {

	fmt.Println("digraph G {")
	for _, node := range mdd.Nodes {
		for i, x := range node.Children {
			fmt.Printf("  %v -> %v [ label=\"x_%v=%v \"];\n", node.Id, x, node.Level, i)
		}
	}
	fmt.Println("}")
}
