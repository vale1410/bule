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
    fmt.Println("atMostOne size test")

    fmt.Println()
    clauses := AtMostOne(Naive, "naive", lits)
    fmt.Println(clauses.Size())
    //clauses.Print()

    fmt.Println()
    clauses = AtMostOne(Split, "split", lits)
    fmt.Println(clauses.Size())
    //clauses.Print()

    fmt.Println()
    clauses = AtMostOne(Counter, "counter", lits)
    fmt.Println(clauses.Size())
    //clauses.Print()

    fmt.Println()
    clauses = AtMostOne(Sorter, "sorter", lits)
    fmt.Println(clauses.Size())
    //clauses.Print()

    fmt.Println("ExactlyOne size test")

    fmt.Println()
    clauses = ExactlyOne(Naive, "naive", lits)
    fmt.Println(clauses.Size())
    //clauses.Print()

    fmt.Println()
    clauses = ExactlyOne(Split, "split", lits)
    fmt.Println(clauses.Size())
    //clauses.Print()

    fmt.Println()
    clauses = ExactlyOne(Counter, "counter", lits)
    fmt.Println(clauses.Size())
    //clauses.Print()

    fmt.Println()
    clauses = ExactlyOne(Sorter, "sorter", lits)
    fmt.Println(clauses.Size())
    //clauses.Print()
}

