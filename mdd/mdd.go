package mdd

import (
	"fmt"
	"github.com/vale1410/bule/glob"
	//"math"
)

// per default 0,1 are ids given

type Node struct {
	Id       int
	Level    int
	Children []int
	Parents  []int
}

func (n Node) str() string {
	return fmt.Sprintf("%d-%v", n.Level, n.Children)
}

type MddStore struct {
	NextId   int
	Nodes    []*Node
	Levels   [][]*Node
	MaxNodes int
	Top      int
	store    map[string]bool
}

func (node Node) IsZero() bool {
	return node.Id == 0
}

func (node Node) IsOne() bool {
	return node.Id == 1
}

func (node Node) Print() {
	fmt.Print(node.Id, "\t", node.Level, node.Children)
}

func Init(size int) (b MddStore) {
	b.Nodes = make([]*Node, 2)
	b.MaxNodes = glob.MDD_max_flag
	b.Nodes[0] = &Node{Id: 0, Level: 0} // id 0
	b.Nodes[1] = &Node{Id: 1, Level: 0} // id 1
	b.store[b.Nodes[0].str()] = true
	b.store[b.Nodes[1].str()] = true
	b.NextId = 2
	return
}

func (mddStore *MddStore) Insert(n Node) (id int) {

	if mddStore.store[n.str()] {
		n.Print()
		panic("node should not exist")
	}

	n.Id = mddStore.NextId
	mddStore.Nodes = append(mddStore.Nodes, &n)
	mddStore.NextId++
	mddStore.store[n.str()] = true

	glob.A(mddStore.NextId == len(mddStore.Nodes), "nextId calculation and length of Nodes list is wrong")

	return n.Id
}
