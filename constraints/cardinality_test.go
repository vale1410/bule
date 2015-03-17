package constraints

// test class

import (
	"fmt"
	"github.com/vale1410/bule/sat"
	"testing"
)

func TestAtMostOne(test *testing.T) {

	fmt.Println("Test for Pure Cardinality.")
	k := 6

	lits := make([]sat.Literal, k)
	atoms := make(map[string]bool)

	for i, _ := range lits {
		lits[i] = sat.Literal{true, sat.NewAtomP1(sat.Pred("x"), i)}
		atoms[lits[i].A.Id()] = true
	}

	t := AtMostOne(Naive, "naive", lits)
	t = AtMostOne(Split, "split", lits)
	t = AtMostOne(Sort, "sorter", lits)
	t = AtMostOne(Heule, "heule", lits)
	t = AtMostOne(Log, "Log", lits)

	fmt.Println()
	t = AtMostOne(Count, "counter", lits)
	t.Clauses.PrintDebug()
	//g := sat.IdGenerator(t.Clauses.Size() * 7)
	//g.Filename = "out.cnf"
	//g.PrimaryVars = atoms
	//g.Solve(t.Clauses)
	//g.PrintSymbolTable("sym.txt")

	fmt.Println()
	t = ExactlyOne(Naive, "naive", lits)
	t = ExactlyOne(Split, "split", lits)
	t = ExactlyOne(Count, "counter", lits)
	t = ExactlyOne(Sort, "sorter", lits)
	t = ExactlyOne(Heule, "heule", lits)
	t = ExactlyOne(Log, "Log", lits)
}
