package constraints

import (
	//	"fmt"
	"github.com/vale1410/bule/config"
	"github.com/vale1410/bule/mdd"
	"github.com/vale1410/bule/sat"
	"testing"
)

//func TestSlice(test *testing.T) {
//
//	b := []int{0, 1, 2, 3, 4}
//	c := []int{0, 1, 2, 3, 4}
//
//	fmt.Println(b[len(c)])
//
//}

func TestMDD1(test *testing.T) {

	config.MaxMDD_flag = 300000

	var t Threshold
	t.Entries = createEntries([]int64{1, 2, 1, 1, 3, 1})
	t.Typ = AtMost
	t.K = 5
	t.Print10()

	b1 := mdd.Init(len(t.Entries))
	_, _, _, s1 := CreateMDD(&b1, t.K, t.Entries)
	//b1.Debug(true)
	if s1 != nil {
		test.Fail()
	}

	for i := 0; i < 4; i++ {
		//	fmt.Println("\n\n Chain on index", i, i+3)
		literals := createLiterals(i, 3)

		b1 = mdd.Init(len(t.Entries))
		_, _, _, s1 := CreateMDDChain(&b1, t.K, t.Entries, literals)

		if s1 != nil {
			test.Fail()
		}

		//	b1.Debug(true)
	}
}

func createLiterals(start int, n int) (literals []sat.Literal) {
	p := sat.Pred("x")
	literals = make([]sat.Literal, n)

	for i := 0; i < n; i++ {
		literals[i] = sat.Literal{true, sat.NewAtomP1(p, start+i)}
	}
	return
}

func createEntries(weights []int64) (entries []Entry) {
	p := sat.Pred("x")
	entries = make([]Entry, len(weights))

	for i := 0; i < len(weights); i++ {
		l := sat.Literal{true, sat.NewAtomP1(p, i)}
		entries[i] = Entry{l, weights[i]}
	}
	return
}
