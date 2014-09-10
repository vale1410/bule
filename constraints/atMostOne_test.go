package constraints

// test class

import (
	"fmt"
	"github.com/vale1410/bule/sat"
	"testing"
)

func TestAtMostOne(t *testing.T) {

	k := 200

	lits := make([]sat.Literal, k)

	for i, _ := range lits {
		lits[i] = sat.Literal{true, sat.NewAtomP1(sat.Pred("input"), i)}
	}

	fmt.Println()
	clauses := atMostOne(Naive, "naive", lits)
    fmt.Println(clauses.Size())
	//clauses.Print()

	fmt.Println()
	clauses = atMostOne(Split, "split", lits)
    fmt.Println(clauses.Size())
	//clauses.Print()

	fmt.Println()
	clauses = atMostOne(Counter, "counter", lits)
    fmt.Println(clauses.Size())
	//clauses.Print()

	fmt.Println()
	clauses = atMostOne(Sorter, "sorter", lits)
    fmt.Println(clauses.Size())
	//clauses.Print()
}
