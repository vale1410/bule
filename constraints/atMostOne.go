package constraints

import (
    //  "fmt"
    "github.com/vale1410/bule/sat"
    "github.com/vale1410/bule/sorters"
)

type CardinalityType int

const (
    Naive CardinalityType = iota
    Sorter
    Split
    Counter
)


func AtMostOne(typ CardinalityType, tag string, lits []sat.Literal) (clauses sat.ClauseSet) {

    switch typ {
    case Naive:
        for i, l := range lits {
            for j := i + 1; j < len(lits); j++ {
                clauses.AddTaggedClause(tag, sat.Neg(l), sat.Neg(lits[j]))
            }
        }
    case Sorter:

        sorter := sorters.CreateCardinalityNetwork(len(lits), 1, sorters.AtMost, sorters.Pairwise)
        sorter.RemoveOutput()

        which := [8]bool{false, false, false, true, true, true, false, false}
        pred := sat.Pred("sorter")
        clauses.AddClauseSet(sat.CreateEncoding(lits, which, []sat.Literal{}, "sort", pred, sorter))

    case Split:

        // a constant that should be exposed,
        // its the cuttoff for the split method of atMostOne

        cutOff := 4

        if len(lits) <= cutOff {
            return AtMostOne(Naive, tag, lits)
        } else {
            aux := sat.NewAtomP1(sat.Pred("split"), newId())
            for _, l := range lits[:len(lits)/2] {
                clauses.AddTaggedClause(tag, sat.Literal{true, aux}, sat.Neg(l))
            }
            for _, l := range lits[len(lits)/2:] {
                clauses.AddTaggedClause(tag, sat.Literal{false, aux}, sat.Neg(l))
            }

            clauses.AddClauseSet(AtMostOne(typ, tag, lits[:len(lits)/2]))
            clauses.AddClauseSet(AtMostOne(typ, tag, lits[len(lits)/2:]))

        }
    case Counter:
        pred := sat.Pred("count")
        tag := "count"

        for i := 1; i < len(lits) ; i++{
            p1 := sat.NewAtomP1(pred,i)
            p2 := sat.NewAtomP1(pred,i+1)
            clauses.AddTaggedClause(tag, sat.Literal{false, p1}, sat.Literal{true, p2})
            clauses.AddTaggedClause(tag, sat.Neg(lits[i-1]),sat.Literal{true, p1})
            clauses.AddTaggedClause(tag, sat.Literal{false, p1}, sat.Neg(lits[i]))
        }

    }

    return

}

func ExactlyOne(typ CardinalityType, tag string, lits []sat.Literal) (clauses sat.ClauseSet) {

    switch typ {
    case Naive:

        clauses.AddClauseSet(AtMostOne(typ , tag , lits ))
        clauses.AddTaggedClause(tag, lits...)

    case Sorter:

        sorter := sorters.CreateCardinalityNetwork(len(lits), 1, sorters.Equal, sorters.Pairwise)
        sorter.RemoveOutput()

        which := [8]bool{false, false, false, true, true, true, false, false}
        pred := sat.Pred("sorter")
        clauses.AddClauseSet(sat.CreateEncoding(lits, which, []sat.Literal{}, "sort", pred, sorter))

    case Split:

        clauses.AddClauseSet(AtMostOne(typ , tag , lits ))
        clauses.AddTaggedClause(tag, lits...)

    case Counter:
        pred := sat.Pred("count")
        tag := "count"
        counterId := newId()

        for i := 1; i < len(lits) ; i++{
            p1 := sat.NewAtomP2(pred,counterId,i)
            p2 := sat.NewAtomP2(pred,counterId,i+1)
            clauses.AddTaggedClause(tag, sat.Literal{false, p1}, sat.Literal{true, p2})
            clauses.AddTaggedClause(tag, sat.Neg(lits[i-1]),sat.Literal{true, p1})
            clauses.AddTaggedClause(tag, sat.Literal{false, p1}, sat.Neg(lits[i]))
        }
        // force that last counter has to be 1, i.e. it models Exactly
        final := sat.NewAtomP2(pred,counterId,len(lits)-1)
        clauses.AddTaggedClause(tag,sat.Literal{true,final})

    }

    return

}
