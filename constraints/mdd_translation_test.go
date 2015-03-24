package constraints

import (
	//"fmt"
	"github.com/vale1410/bule/glob"
	"github.com/vale1410/bule/mdd"
	"github.com/vale1410/bule/sat"
	"testing"
)

func TestMDDRedundant(test *testing.T) {
	//glob.Debug_flag = true
	//glob.Debug_flag = false

	glob.D("TestMDDRedundant")
	glob.MDD_max_flag = 300000
	glob.MDD_redundant_flag = false

	var t Threshold
	t.Entries = createEntries([]int64{1, 2, 1, 1, 3, 1})
	t.Typ = LE
	t.K = 5

	store := mdd.Init(len(t.Entries))
	CreateMDD(&store, t.K, t.Entries)

	if store.RemoveRedundants() != 5 {
		test.Fail()
	}
}

func TestMDDChain(test *testing.T) {
	//glob.Debug_flag = true
	//glob.Debug_flag = false

	glob.D("TestMDDChain")

	glob.MDD_max_flag = 300000
	glob.MDD_redundant_flag = false

	var t Threshold
	t.Entries = createEntries([]int64{1, 2, 1, 1, 3, 1})
	t.Typ = LE
	t.K = 5
	//t.Print10()

	store := mdd.Init(len(t.Entries))
	_, _, _, s1 := CreateMDD(&store, t.K, t.Entries)
	//store.Debug(true)
	if s1 != nil {
		test.Fail()
	}

	chain := Chain(t.Literals()) //createLiterals(i, 3)

	for i := 0; i < 4; i++ {
		//fmt.Println("\n\n Chain on index", i, i+3)

		chains := Chains{chain[i : i+3]}
		//chains[0].Print()

		store = mdd.Init(len(t.Entries))
		_, _, _, s1 := CreateMDDChain(&store, t.K, t.Entries, chains)

		if s1 != nil {
			test.Fail()
		}
		//store.Debug(true)
	}
}

func TestMDDChains1(test *testing.T) {
	//glob.Debug_flag = true

	glob.D("TestMDDChains1")

	glob.MDD_max_flag = 300000
	glob.MDD_redundant_flag = false

	var t Threshold
	t.Entries = createEntries([]int64{1, 2, 1, 1, 3, 1})
	t.Typ = LE
	t.K = 5
	//t.Print10()

	{ // check
		store := mdd.Init(len(t.Entries))
		_, _, _, s1 := CreateMDD(&store, t.K, t.Entries)
		//store.Debug(true)
		if s1 != nil {
			test.Fail()
		}
	}

	chain := Chain(t.Literals()) //createLiterals(i, 3)

	//fmt.Println("\n\n Chain on index", i, i+3)

	chains := Chains{chain[0:3], chain[3:6]}
	//chains[0].Print()

	store := mdd.Init(len(t.Entries))
	_, _, _, s1 := CreateMDDChain(&store, t.K, t.Entries, chains)

	if s1 != nil {
		test.Fail()
	}
	if len(store.Nodes) != 6 {
		store.Debug(true)
		test.Fail()
	}
	//glob.Debug_flag = false
}

func TestMDDChains2(test *testing.T) {
	//glob.Debug_flag = true

	glob.D("TestMDDChains")

	glob.MDD_max_flag = 300000
	glob.MDD_redundant_flag = false

	var t Threshold
	t.Entries = createEntries([]int64{1, 2, 1, 1, 3, 1, 3, 2, 1, 1, 1})
	t.Typ = LE
	t.K = 10
	//t.Print10()

	{ // check
		store := mdd.Init(len(t.Entries))
		_, _, _, s1 := CreateMDD(&store, t.K, t.Entries)
		//store.Debug(true)
		if s1 != nil {
			test.Fail()
		}
	}

	chain := Chain(t.Literals()) //createLiterals(i, 3)

	chains := Chains{chain[1:3], chain[5:9]}

	store := mdd.Init(len(t.Entries))
	_, _, _, s1 := CreateMDDChain(&store, t.K, t.Entries, chains)

	if s1 != nil {
		test.Fail()
	}
	if len(store.Nodes) != 31 {
		store.Debug(true)
		test.Fail()
	}
	//glob.Debug_flag = false
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
