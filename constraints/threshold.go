package constraints

import (
	//	"fmt"
	"github.com/vale1410/bule/sat"
	"sort"
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
	Id   int // unique id to reference Threshold in encodings
	Desc string

	Entries []Entry
	K       int64
	Typ     EquationType
	Pred    sat.Pred
}

// works with any threshold
func (t *Threshold) OnlyFacts() (is bool, cs sat.ClauseSet) {

	is = false

	if t.Typ == Equal {
		t.Normalize(Equal, true)
	} else {
		t.Normalize(AtLeast, true)
	}

	if t.SumWeights() == t.K {
		is = true
		for _, x := range t.Entries {
			cs.AddTaggedClause("Fact", x.Literal)
		}
	}

	return
}

// all weights are the same; performs rounding
// if is is true, then all weights are 1, and K is the cardinality
func (t *Threshold) Cardinality() (allSame bool, literals []sat.Literal) {

	t.NormalizePositiveCoefficients()
	allSame = true

	coeff := t.Entries[0].Weight
	for _, x := range t.Entries {
		if x.Weight != coeff {
			allSame = false
			break
		}
	}

	if allSame {
		literals = make([]sat.Literal, len(t.Entries))
		t.K = t.K / coeff
		for i, x := range t.Entries {
			t.Entries[1].Weight = 1
			literals[i] = x.Literal
		}

	}

	return allSame, literals
}

func (t *Threshold) NormalizePositiveCoefficients() {

	for i, e := range t.Entries {
		if t.Entries[i].Weight < 0 {
			t.Entries[i].Literal = sat.Neg(e.Literal)
			t.K -= t.Entries[i].Weight
			t.Entries[i].Weight = -t.Entries[i].Weight
		}
	}
}

func (t *Threshold) NormalizePositiveLiterals() {

	for i, e := range t.Entries {
		if t.Entries[i].Literal.Sign == false {
			t.Entries[i].Literal = sat.Neg(e.Literal)
			t.K -= t.Entries[i].Weight
			t.Entries[i].Weight = -t.Entries[i].Weight
		}
	}
}

// works only on AtMost/AtLeast
func (t *Threshold) ChangeTo(typ EquationType) {
	if typ != AtMost && typ != AtLeast {
		panic("EquationType not correct")
	}

	if t.Typ != typ {
		for i, e := range t.Entries {
			t.Entries[i].Weight = -e.Weight
		}
		t.K = -t.K
		t.Typ = typ
	}

}

// normalizes the threshold
func (t *Threshold) Normalize(typ EquationType, posWeights bool) {

	if t.Typ == Equal {
		if typ != Equal {
			panic("cant normalize Equal on threshold that is not Equal")
		}
	} else {
		t.ChangeTo(typ)
	}

	if posWeights {
		t.NormalizePositiveCoefficients()
	} else {
		t.NormalizePositiveLiterals()
	}

	return
}

// sums up all weights
func (t *Threshold) SumWeights() (total int64) {
	for _, e := range t.Entries {
		total += e.Weight
	}
	return
}

type ThresholdAscending Threshold

func (a ThresholdAscending) Len() int           { return len(a.Entries) }
func (a ThresholdAscending) Swap(i, j int)      { a.Entries[i], a.Entries[j] = a.Entries[j], a.Entries[i] }
func (a ThresholdAscending) Less(i, j int) bool { return a.Entries[i].Weight > a.Entries[j].Weight }

func (t *Threshold) Sort() {
	sort.Sort(ThresholdAscending(*t))
}
