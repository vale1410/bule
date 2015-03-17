package bdd

import (
	"github.com/vale1410/bule/constraints"
	"github.com/vale1410/bule/sat"
	"testing"
)

func TestBDD1(test *testing.T) {

	maxNodes := 300000

	var t constraints.Threshold
	t.Entries = createEntries([]int64{3, 1, 4, 3, 2, 3})
	t.Typ = constraints.AtMost
	t.K = 5

	literals := createLiterals(0, 3)

	b1 := Init(len(t.Entries), maxNodes)
	_, _, _, s1 := b1.CreateBddAMO(t.K, t.Entries, literals)

	b2 := Init(len(t.Entries), maxNodes)
	_, _, _, s2 := b2.CreateBdd(t.K, t.Entries)

	if s1 != nil || s2 != nil {
		test.Fail()
	}

	b1.Debug(true)
	b2.Debug(true)
}

func createLiterals(start int, n int) (literals []sat.Literal) {
	p := sat.Pred("x")
	literals = make([]sat.Literal, n)

	for i := 0; i < n; i++ {
		literals[i] = sat.Literal{true, sat.NewAtomP1(p, start+i)}
	}
	return
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
