package constraints

import (
	//  "fmt"
	"github.com/vale1410/bule/sat"
	"github.com/vale1410/bule/sorters"
)

// better rename to encoding type
type CardinalityType int

const (
	Naive CardinalityType = iota
	Sort
	Split
	Count
	Heule
	Log
)

func AtMostOne(typ CardinalityType, tag string, lits []sat.Literal) (clauses sat.ClauseSet) {

	switch typ {

	case Naive:
		for i, l := range lits {
			for j := i + 1; j < len(lits); j++ {
				clauses.AddTaggedClause(tag, sat.Neg(l), sat.Neg(lits[j]))
			}
		}

	case Sort:

		sat.SetUp(3, sorters.Pairwise)
		clauses.AddClauseSet(sat.CreateCardinality(tag, lits, 1, sorters.AtMost))

	case Split:

		// a constant that should be exposed,
		// its the cuttoff for the split method of atMostOne

		cutoff := 5

		if len(lits) <= cutoff {
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
	case Count:

		pred := sat.Pred("count")
		counterId := newId()

		first := sat.NewAtomP2(pred, counterId, 1)
		clauses.AddTaggedClause(tag+"-2", lits[0], sat.Literal{false, first})

		for i := 1; i < len(lits); i++ {
			p1 := sat.NewAtomP2(pred, counterId, i)
			clauses.AddTaggedClause(tag+"-1", sat.Neg(lits[i-1]), sat.Literal{true, p1})
			clauses.AddTaggedClause(tag+"-1", sat.Neg(lits[i]), sat.Literal{false, p1})
			if i != len(lits) {
				p2 := sat.NewAtomP2(pred, counterId, i+1)
				clauses.AddTaggedClause(tag+"-2", lits[i], sat.Literal{true, p1}, sat.Literal{false, p2})
				clauses.AddTaggedClause(tag+"-3", sat.Literal{false, p1}, sat.Literal{true, p2})
			}
		}

	case Heule:

		k := 4

		if len(lits) > k+1 {
			aux := sat.NewAtomP1(sat.Pred("heule"), newId())
			front := make([]sat.Literal, k+1)
			copy(front, lits[:k])
			front[k] = sat.Literal{true, aux}
			clauses = AtMostOne(Naive, tag, front)
			back := make([]sat.Literal, len(lits)-k+1)
			copy(back, lits[k:])
			back[len(lits)-k] = sat.Literal{false, aux}
			clauses.AddClauseSet(AtMostOne(typ, tag, back))
		} else {
			clauses = AtMostOne(Naive, tag, lits)
		}

	case Log:

		cutoff := 5 //will be a parameter of this encoding

		//this is very similar to the split encoding

		clauses = buildLogEncoding(sat.Pred("logE"), newId(), cutoff, 0, tag, lits)

	}

	return

}

func buildLogEncoding(pred sat.Pred, uId int, cutoff int, depth int, tag string, lits []sat.Literal) (clauses sat.ClauseSet) {

	//fmt.Println(depth,lits)

	if len(lits) <= cutoff {
		clauses = AtMostOne(Naive, tag, lits)
	} else {

		atom := sat.NewAtomP2(pred, uId, depth)

		first := lits[:len(lits)/2]
		for _, l := range first {
			clauses.AddTaggedClause(tag, sat.Literal{true, atom}, sat.Neg(l))
		}
		second := lits[len(lits)/2:]
		for _, l := range second {
			clauses.AddTaggedClause(tag, sat.Literal{false, atom}, sat.Neg(l))
		}

		depth++

		clauses.AddClauseSet(buildLogEncoding(pred, uId, cutoff, depth, tag, first))
		clauses.AddClauseSet(buildLogEncoding(pred, uId, cutoff, depth, tag, second))

	}

	//fmt.Println(clauses)

	return
}

func ExactlyOne(typ CardinalityType, tag string, lits []sat.Literal) (clauses sat.ClauseSet) {

	switch typ {
	case Naive:

		clauses.AddClauseSet(AtMostOne(typ, tag, lits))
		clauses.AddTaggedClause(tag, lits...)

	case Sort:

		sat.SetUp(3, sorters.Pairwise)
		clauses.AddClauseSet(sat.CreateCardinality(tag, lits, 1, sorters.Equal))

	case Split:

		clauses.AddClauseSet(AtMostOne(typ, tag, lits))
		clauses.AddTaggedClause(tag, lits...)

	case Count:
		pred := sat.Pred("count")
		counterId := newId()

		first := sat.NewAtomP2(pred, counterId, 1)
		clauses.AddTaggedClause(tag+"-2", lits[0], sat.Literal{false, first})

		for i := 1; i < len(lits); i++ {
			p1 := sat.NewAtomP2(pred, counterId, i)
			clauses.AddTaggedClause(tag+"-1", sat.Neg(lits[i-1]), sat.Literal{true, p1})
			clauses.AddTaggedClause(tag+"-1", sat.Neg(lits[i]), sat.Literal{false, p1})
			if i != len(lits) {
				p2 := sat.NewAtomP2(pred, counterId, i+1)
				clauses.AddTaggedClause(tag+"-2", lits[i], sat.Literal{true, p1}, sat.Literal{false, p2})
				clauses.AddTaggedClause(tag+"-3", sat.Literal{false, p1}, sat.Literal{true, p2})
			}
		}
		// force that last counter has to be 1, i.e. it models ExactlyOne
		final := sat.NewAtomP2(pred, counterId, len(lits))
		clauses.AddTaggedClause(tag+"-4", sat.Literal{true, final})

	case Heule:

		clauses.AddClauseSet(AtMostOne(typ, tag, lits))
		clauses.AddTaggedClause(tag, lits...)

	case Log:

		clauses.AddClauseSet(AtMostOne(typ, tag, lits))
		clauses.AddTaggedClause(tag, lits...)

	default:
		panic("CNF translation for this type not implemented yet")

	}

	return

}
