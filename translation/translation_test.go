package translation

import (
	"github.com/vale1410/bule/constraints"
	"github.com/vale1410/bule/sat"
	"testing"
)

func TestBDD(test *testing.T) {

	pb := createIgnasi1()

	Translate(pb, BDD)

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
