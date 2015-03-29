package constraints

import (
	//  "fmt"
	"github.com/vale1410/bule/sat"
)

type OneTranslationType int

const (
	Naive OneTranslationType = iota
	Sort
	Split
	Count
	Heule
	Log
)

type CardTranslation struct {
	PB      *Threshold
	Typ     OneTranslationType
	Aux     []sat.Literal
	Clauses sat.ClauseSet
}

func TranslateAtMostOne(typ OneTranslationType, tag string, lits []sat.Literal) (trans CardTranslation) {

	var clauses sat.ClauseSet

	switch typ {

	case Naive:
		for i, l := range lits {
			for j := i + 1; j < len(lits); j++ {
				clauses.AddTaggedClause(tag, sat.Neg(l), sat.Neg(lits[j]))
			}
		}

	case Split:

		// a constant that should be exposed,
		// its the cuttoff for the split method of atMostOne

		cutoff := 5

		if len(lits) <= cutoff {
			return TranslateAtMostOne(Naive, tag, lits)
		} else {
			aux := sat.NewAtomP1(sat.Pred("split"), newId())
			trans.Aux = append(trans.Aux, sat.Literal{true, aux})

			for _, l := range lits[:len(lits)/2] {
				clauses.AddTaggedClause(tag, sat.Literal{true, aux}, sat.Neg(l))
			}
			for _, l := range lits[len(lits)/2:] {
				clauses.AddTaggedClause(tag, sat.Literal{false, aux}, sat.Neg(l))
			}

			clauses.AddClauseSet(TranslateAtMostOne(typ, tag, lits[:len(lits)/2]).Clauses)
			clauses.AddClauseSet(TranslateAtMostOne(typ, tag, lits[len(lits)/2:]).Clauses)

		}
	case Count:

		pred := sat.Pred("c")
		counterId := newId()

		auxs := make([]sat.Literal, len(lits))
		for i, _ := range auxs {
			auxs[i] = sat.Literal{true, sat.NewAtomP2(pred, counterId, i)}
		}
		trans.Aux = auxs

		// S_i -> S_{i-1}
		for i := 1; i < len(lits); i++ {
			clauses.AddTaggedClause(tag, auxs[i-1], sat.Neg(auxs[i]))
		}

		// X_i -> S_i
		for i := 0; i < len(lits); i++ {
			clauses.AddTaggedClause(tag, auxs[i], sat.Neg(lits[i]))
		}

		// X_i-1 -> -S_i
		for i := 1; i < len(lits); i++ {
			clauses.AddTaggedClause(tag, sat.Neg(auxs[i]), sat.Neg(lits[i-1]))
		}

		// (S_i-1 /\ -S_i) -> X_i-1
		for i := 1; i <= len(lits); i++ {
			if i != len(lits) {
				clauses.AddTaggedClause(tag, sat.Neg(auxs[i-1]), auxs[i], lits[i-1])
			} else {
				clauses.AddTaggedClause(tag, sat.Neg(auxs[i-1]), lits[i-1])
			}
		}

	case Heule:

		k := 4 // fixed size for the heule encoding

		if len(lits) > k+1 {
			aux := sat.NewAtomP1(sat.Pred("heule"), newId())
			trans.Aux = append(trans.Aux, sat.Literal{true, aux})

			front := make([]sat.Literal, k+1)
			copy(front, lits[:k])
			front[k] = sat.Literal{true, aux}

			trans2 := TranslateAtMostOne(Naive, tag, front)
			clauses.AddClauseSet(trans2.Clauses)

			back := make([]sat.Literal, len(lits)-k+1)
			copy(back, lits[k:])
			back[len(lits)-k] = sat.Literal{false, aux}

			trans2 = TranslateAtMostOne(typ, tag, back)
			trans.Aux = append(trans.Aux, trans2.Aux...)
			clauses.AddClauseSet(trans2.Clauses)

		} else {
			trans2 := TranslateAtMostOne(Naive, tag, lits)
			clauses.AddClauseSet(trans2.Clauses)
		}

	case Log:

		cutoff := 5 //will be a parameter of this encoding
		clauses = buildLogEncoding(sat.Pred("logE"), newId(), cutoff, 0, tag, lits)
	case Sort:
		panic("CNF translation for this type not implemented yet")
	default:
		panic("CNF translation for this type not implemented yet")

	}

	trans.Typ = typ
	trans.Clauses = clauses

	return

}

func TranslateExactlyOne(typ OneTranslationType, tag string, lits []sat.Literal) (trans CardTranslation) {

	var clauses sat.ClauseSet

	switch typ {
	case Heule, Log, Naive, Split:

		trans2 := TranslateAtMostOne(typ, tag, lits)
		trans.Aux = append(trans.Aux, trans2.Aux...)
		clauses.AddClauseSet(trans2.Clauses)
		clauses.AddTaggedClause(tag, lits...)

	case Count:

		pred := sat.Pred("count")
		counterId := newId()

		auxs := make([]sat.Literal, len(lits))
		for i, _ := range auxs {
			auxs[i] = sat.Literal{true, sat.NewAtomP2(pred, counterId, i)}
		}
		trans.Aux = auxs

		// S_i -> S_{i-1}
		for i := 1; i < len(lits); i++ {
			clauses.AddTaggedClause(tag, auxs[i-1], sat.Neg(auxs[i]))
		}

		// X_i -> S_i
		for i := 0; i < len(lits); i++ {
			clauses.AddTaggedClause(tag, auxs[i], sat.Neg(lits[i]))
		}

		// X_i-1 -> -S_i
		for i := 1; i < len(lits); i++ {
			clauses.AddTaggedClause(tag, sat.Neg(auxs[i]), sat.Neg(lits[i-1]))
		}

		// (S_i-1 /\ -S_i) -> X_i-1
		for i := 1; i <= len(lits); i++ {
			if i != len(lits) {
				clauses.AddTaggedClause(tag, sat.Neg(auxs[i-1]), auxs[i], lits[i-1])
			} else {
				clauses.AddTaggedClause(tag, sat.Neg(auxs[i-1]), lits[i-1])
			}
		}

		clauses.AddTaggedClause(tag+"Ex1", auxs[0])

	case Sort:
		panic("CNF translation for this type not implemented yet")
	default:
		panic("CNF translation for this type not implemented yet")
	}

	trans.Typ = typ
	trans.Clauses = clauses

	return

}

func buildLogEncoding(pred sat.Pred, uId int, cutoff int, depth int, tag string, lits []sat.Literal) (clauses sat.ClauseSet) {
	if len(lits) <= cutoff {
		trans2 := TranslateAtMostOne(Naive, tag, lits)
		clauses.AddClauseSet(trans2.Clauses)
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
	return
}
