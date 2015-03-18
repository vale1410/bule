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

// assumption is that pb2 is already a subsequece of pb1
func CommonSlice(e []Entry, l []sat.Literal) (bool, []Entry) {
	for i, x := range e {
		if x.Literal == l[0] {
			return true, e[i : i+len(l)]
		}
	}
	return false, []Entry{}
}

// assumption is that pb2 is already a subsequece of pb1
func PositionSlice(e1 []Entry, e2 []Entry) (bool, []int) {
	//find min coefficient, to subtract
	pos := make([]int, len(e2))

	j := 0
	for i, x := range e1 {
		if j == len(pos) {
			break
		}
		if x.Literal == e2[j].Literal {
			pos[j] = i
			j++
		}
	}
	if j != len(pos) {
		return false, []int{}
	}
	return false, pos
}

// creates an AtMost constraint
// with coefficients in weights,
// variables x1..xm
func CreatePB(weights []int64, K int64) (pb Threshold) {

	pb.Entries = make([]Entry, len(weights))
	pb.Typ = AtMost
	pb.K = K

	p := sat.Pred("x")
	for i := 0; i < len(weights); i++ {
		l := sat.Literal{true, sat.NewAtomP1(p, i)}
		pb.Entries[i] = Entry{l, weights[i]}
	}
	return
}

// finds trivially implied facts, returns set of facts
// removes such entries from the pb
// threshold can become empty!
func (t *Threshold) RemoveZeros() {
	c := len(t.Entries)

	for i := 0; i < c; i++ {
		if t.Entries[i].Weight == 0 {
			//fmt.Println(i, c)
			c--
			t.Entries[i] = t.Entries[c]
			i--
		}
	}
	t.Entries = t.Entries[:c]
}

// finds trivially implied facts, returns set of facts
// removes such entries from the pb
// threshold can become empty!
func (t *Threshold) Simplify() (cs sat.ClauseSet) {

	if t.Typ == Equal {
		t.Normalize(Equal, true)
	} else {
		t.Normalize(AtMost, true)
	}

	entries := make([]Entry, 0, len(t.Entries))

	for _, x := range t.Entries {
		if x.Weight > t.K {
			cs.AddTaggedClause("Trivial", sat.Neg(x.Literal))
		} else {
			entries = append(entries, x)
		}
	}
	t.Entries = entries

	if t.Typ == Equal {
		t.Normalize(Equal, true)
	} else {
		t.Normalize(AtLeast, true)
	}

	if t.SumWeights() == t.K {
		for _, x := range t.Entries {
			cs.AddTaggedClause("Fact", x.Literal)
		}
		t.Entries = []Entry{}
		t.K = 0
	}

	return
}

// all weights are the same; performs rounding
// if this is true, then all weights are 1, and K is the cardinality
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
			t.Entries[i].Weight = 1
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

func (t *Threshold) Multiply(c int64) {
	if c == 0 {
		panic("multiplyer is 0")
	}
	for i, e := range t.Entries {
		t.Entries[i].Weight = c * e.Weight
	}

	t.K = c * t.K

	if c < 0 {
		switch t.Typ {
		case AtMost:
			t.Typ = AtLeast
		case AtLeast:
			t.Typ = AtMost
		default:
			//nothing
		}
	}
}

// normalizes the threshold
func (t *Threshold) Normalize(typ EquationType, posWeights bool) {

	if t.Typ == Equal {
		if typ != Equal {
			panic("cant normalize Equal on threshold that is not Equal")
		}
	} else if typ != t.Typ {
		t.Multiply(-1)
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

func (t *Threshold) SortVar() {
	sort.Sort(EntriesVariables(t.Entries))
}
func (t *Threshold) Sort() {
	sort.Sort(EntriesAscending(t.Entries))
}

type EntriesVariables []Entry
type EntriesAscending []Entry

func (a EntriesVariables) Len() int      { return len(a) }
func (a EntriesVariables) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a EntriesVariables) Less(i, j int) bool {
	return a[i].Literal.A.Id() <= a[j].Literal.A.Id()
}

func (a EntriesAscending) Len() int           { return len(a) }
func (a EntriesAscending) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a EntriesAscending) Less(i, j int) bool { return a[i].Weight > a[j].Weight }
