package bdd

import (
	"fmt"
	"github.com/vale1410/bule/constraints"
	"github.com/vale1410/bule/sat"
	"testing"
)

func Test1(test *testing.T) {
	//
	//	fmt.Println("test stuff")
	//
	//	b := Init(100)
	//
	//	b.Insert(Node{level: 5})
	//	b.Insert(Node{level: 7})
	//	b.Insert(Node{level: 2})
	//	b.Insert(Node{level: 3})
	//
	//	if id, wmin, wmax := b.GetByWeight(0, -1); id != -1 {
	//		fmt.Println(id, wmin, wmax)
	//	} else {
	//		fmt.Println(id, wmin, wmax)
	//		fmt.Println("shit")
	//	}
	//
	//	b.Debug(true)

}

func TestBDD(test *testing.T) {

	pb := createIgnasi2()

	b := Init(len(pb.Entries))
	_, _, _ = b.CreateBdd(pb.K, pb.Entries)

	b.Debug(true)

	fmt.Println("are you happy now?")

}

func createEntries(weights []int64) (entries []constraints.Entry) {
	p := sat.Pred("x")
	entries = make([]constraints.Entry, len(weights))

	for i := 0; i < len(weights); i++ {
		l := sat.Literal{true, sat.NewAtomP1(p, i)}
		entries[i] = constraints.Entry{l, weights[i]}
	}
	return
}

func createIgnasi1() (t constraints.Threshold) {

	weights := []int64{4, 3, 1, 1, 1, 1}
	t.K = 5

	t.Desc = "Ignasi 1"
	t.Typ = constraints.AtMost
	t.Entries = createEntries(weights)
	return
}

func createIgnasi2() (t constraints.Threshold) {

	weights := []int64{7, 6, 2, 2, 2, 2, 1, 1, 1, 1, 1}
	t.K = 12

	t.Desc = "Ignasi 2"
	t.Typ = constraints.AtMost
	t.Entries = createEntries(weights)
	return
}
