package constraints

// test class

import (
	"github.com/vale1410/bule/glob"
	"github.com/vale1410/bule/sat"
	"testing"
)

func TestAtMostOne(test *testing.T) {
	glob.D("TestTranslateAtMostOne")
	k := 6

	lits := make([]sat.Literal, k)
	atoms := make(map[string]bool)

	for i := range lits {
		lits[i] = sat.Literal{true, sat.NewAtomP1(sat.Pred("x"), i)}
		atoms[lits[i].A.Id()] = true
	}

	t := TranslateAtMostOne(Naive, "naive", lits)
	if t.Clauses.Size() == 0 {
		test.Fail()
	}
	t = TranslateAtMostOne(Split, "split", lits)
	if t.Clauses.Size() == 0 {
		test.Fail()
	}
	//	t = TranslateAtMostOne(Sort, "sorter", lits)
	//	if t.Clauses.Size() == 0 {
	//		test.Fail()
	//	}
	t = TranslateAtMostOne(Heule, "heule", lits)
	if t.Clauses.Size() == 0 {
		test.Fail()
	}
	t = TranslateAtMostOne(Log, "Log", lits)
	if t.Clauses.Size() == 0 {
		test.Fail()
	}

	//fmt.Println()
	t = TranslateAtMostOne(Count, "counter", lits)
	if t.Clauses.Size() == 0 {
		test.Fail()
	}
	//t.Clauses.PrintDebug()
	//g := sat.IdGenerator(t.Clauses.Size() * 7)
	//g.Filename = "out.cnf"
	//g.PrimaryVars = atoms
	//g.Solve(t.Clauses)
	//g.PrintSymbolTable("sym.txt")

	//fmt.Println()
	t = TranslateExactlyOne(Naive, "naive", lits)
	if t.Clauses.Size() == 0 {
		test.Fail()
	}
	t = TranslateExactlyOne(Split, "split", lits)
	if t.Clauses.Size() == 0 {
		test.Fail()
	}
	t = TranslateExactlyOne(Count, "counter", lits)
	if t.Clauses.Size() == 0 {
		test.Fail()
	}
	//t = TranslateExactlyOne(Sort, "sorter", lits)
	//if t.Clauses.Size() == 0 {
	//	test.Fail()
	//}
	t = TranslateExactlyOne(Heule, "heule", lits)
	if t.Clauses.Size() == 0 {
		test.Fail()
	}
	t = TranslateExactlyOne(Log, "Log", lits)
	if t.Clauses.Size() == 0 {
		test.Fail()
	}
}
