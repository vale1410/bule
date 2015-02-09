package constraints

// test class

import (
	"fmt"
	"github.com/vale1410/bule/sat"
	"testing"
)

func TestAtMostOne(t *testing.T) {

	fmt.Println("Repair Test for Pure Cardinality.")
	k := 10

	lits := make([]sat.Literal, k)

	for i, _ := range lits {
		lits[i] = sat.Literal{true, sat.NewAtomP1(sat.Pred("input"), i)}
	}
	//var clauses sat.ClauseSet

	//    fmt.Println("atMostOne size test")
	//
	//    fmt.Println()
	//    clauses := AtMostOne(Naive, "naive", lits)
	//    fmt.Println(clauses.Size())
	//
	//    fmt.Println()
	//    clauses = AtMostOne(Split, "split", lits)
	//    fmt.Println(clauses.Size())
	//
	//    fmt.Println()
	//    clauses = AtMostOne(Count, "counter", lits)
	//    fmt.Println(clauses.Size())
	//
	//    fmt.Println()
	//    clauses = AtMostOne(Sort, "sorter", lits)
	//    fmt.Println(clauses.Size())
	//
	//    clauses = AtMostOne(Heule, "heule", lits)
	//    fmt.Println(clauses.Size())
	//    //clauses.Print()

	//    clauses = AtMostOne(Log, "Log", lits)
	//    fmt.Println(clauses.Size())
	//    clauses.PrintDebug()

	//    fmt.Println("ExactlyOne size test")
	//
	//    fmt.Println()
	//    clauses = ExactlyOne(Naive, "naive", lits)
	//    fmt.Println(clauses.Size())
	//    //clauses.Print()
	//
	//    fmt.Println()
	//    clauses = ExactlyOne(Split, "split", lits)
	//    fmt.Println(clauses.Size())
	//    //clauses.Print()
	//
	//    fmt.Println()
	//    clauses = ExactlyOne(Count, "counter", lits)
	//    fmt.Println(clauses.Size())
	//    //clauses.Print()
	//
	//    fmt.Println()
	//    clauses = ExactlyOne(Sort, "sorter", lits)
	//    fmt.Println(clauses.Size())
	//    //clauses.Print()
	//
	//    clauses = ExactlyOne(Heule, "heule", lits)
	//    fmt.Println(clauses.Size())
	//    //clauses.Print()

	//clauses = ExactlyOne(Log, "Log", lits)
	//fmt.Println(clauses.Size())
	//clauses.PrintDebug()
}
