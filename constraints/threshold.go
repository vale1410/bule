package constraints

import (
	"github.com/vale1410/bule/sat"
)

type EquationType int

const (
	AtMost EquationType = iota
	AtLeast
	Equal
	Optimization
)

type Entry struct {
	Literal sat.Literal
	Weight  int64
}

type Threshold struct {
	Id      int // unique id to reference Threshold in encodings
	Desc    string
	Entries []Entry
	K       int64
	Typ     EquationType
	Pred    sat.Pred
	Clauses sat.ClauseSet
	LitIn   []sat.Literal //Bags flattened, input to Sorter
}

func (t *Threshold) OnlyFacts() (is bool, cs sat.ClauseSet) {

	t.NormalizeAtMost()
	is = false

	if t.K <= 0 {
		is = true
		cs = sat.NewClauseSet(len(t.Entries))
		for _, x := range t.Entries {
			cs.AddTaggedClause("OF", sat.Neg(x.Literal))
		}
	}

	return
}

func (t *Threshold) SingleClause() (is bool, literals []sat.Literal) {

	t.NormalizeAtLeast(false)

	var clause sat.Clause

	entries := make([]Entry, len(t.Entries))
	copy(entries, t.Entries)
	K := t.K

	// normalize to coefficients 1
	allOne := true
	literals = make([]sat.Literal, len(entries))

	for i, x := range entries {
		if x.Weight*x.Weight != 1 {
			allOne = false
			break
		}
		literals[i] = x.Literal

		if x.Weight == -1 {
			K += -x.Weight
			literals[i] = sat.Neg(clause.Literals[i])
		}
	}

	return allOne && K == 1, literals
}

func (t *Threshold) IsNormalized() (yes bool) {
	yes = true

	for _, e := range t.Entries {
		if e.Weight <= 0 {
			yes = false
		}
	}
	return t.K > 0 && yes && t.Typ == AtMost
}

// Normalize: AtLeast
// transform to AtLeast, only positive variables
// b=true: only positive variables
// b=false: only positive weights
func (t *Threshold) NormalizeAtLeast(posVariables bool) {
	posWeights := !posVariables

	// remove 0 weights?
	if t.Typ == AtMost {
		//set to AtMost, multiply by -1
		t.K = -t.K
		t.Typ = AtLeast
		for i, e := range t.Entries {
			t.Entries[i].Weight = -e.Weight
		}
	} else if t.Typ == Equal {
		panic("Equal type for threshold function not supported yet")
	}

	for i, e := range t.Entries {
		if (posWeights && t.Entries[i].Weight < 0) ||
			(posVariables && e.Literal.Sign == false) {
			t.Entries[i].Literal = sat.Neg(e.Literal)
			t.K -= t.Entries[i].Weight
			t.Entries[i].Weight = -t.Entries[i].Weight
		}
	}

	return
}

// Normalize: AtMost
// transform to AtMost, only positive weights
// transform negative weights
// check if maximum reaches K at all
// todo: sort by weight
// returns sum of weights
func (t *Threshold) NormalizeAtMost() (total int64) {

	total = 0

	// remove 0 weights?
	if t.Typ == AtLeast {
		//set to AtMost, multiply by -1
		for i, e := range t.Entries {
			t.Entries[i].Weight = -e.Weight
		}
		t.K = -t.K
		t.Typ = AtMost
	} else if t.Typ == Equal {
		panic("Equal type for threshold function not supported")
	}

	for i, e := range t.Entries {
		if e.Weight == 0 {
			panic("Threshold contains a 0 weight element, should not occur")
		}
		if e.Weight < 0 {
			t.Entries[i].Weight = -e.Weight
			t.Entries[i].Literal = sat.Neg(e.Literal)
			t.K -= e.Weight
		}
		total += t.Entries[i].Weight
	}
	return
}
