package lib

import (
	"fmt"
	"strings"
)

type Program struct {
	Rules            []Rule
	Constants        map[string]int
	PredicatToTuples map[Predicate][][]int
	GroundFacts      map[Predicate]bool
	Search           map[Predicate]bool
	existQ           map[int][]Literal
	forallQ          map[int][]Literal
}

type Rule struct {
	initialTokens  []Token
	Head           Literal
	Literals       []Literal
	Generators     []Generator
	Constraints    []Constraint
	GeneratingHead bool     // if final token is tokenQuestionMark then it generates, otherwise tokenDot
	Typ            ruleType // Can be Implication or Equivalence or RuleComma(normal rule)
}

// is a math expression that evaluates to true or false
// Constraints can contain variables
// supported are <,>,<=,>=,==
// E.g..: A*3v<=v5-2*R/7#mod3.
type Constraint struct {
	LeftTerm   Term
	Comparison tokenKind
	RightTerm  Term
}

type Clause []Literal

type ruleType int

const (
	ruleTypeDisjunction ruleType = iota
	ruleTypeImplication
	ruleTypeEquivalence
)

type Generator struct {
	Constraints []Constraint
	Literals    []Literal
	Head        Literal
}

type Literal struct {
	Neg   bool
	Name  Predicate
	Terms []Term
}

type Term string

type Predicate string


func (r *Rule) hasHead() bool {
	return r.Head.Name != ""
}

func (r *Rule) IsDisjunction() bool {
	return len(r.Generators) == 0 && !r.hasHead() && r.Typ == ruleTypeDisjunction
}

func (r *Rule) IsGround() bool {
	return r.FreeVars().IsEmpty()
}

func (r *Rule) IsFact() bool {
	return !r.hasHead() && r.Typ == ruleTypeDisjunction && len(r.Literals) == 1
}


func (constraint Constraint) Copy() (cons Constraint) {
	cons = constraint
	cons.LeftTerm = constraint.LeftTerm
	cons.RightTerm = constraint.RightTerm
	return cons
}

func (literal Literal) Copy() Literal {
	t := make([]Term, len(literal.Terms))
	copy(t, literal.Terms)
	literal.Terms = t
	return literal
}

// Deep Copy
func (gen Generator) Copy() (newGen Generator) {
	newGen = gen
	newGen.Head = gen.Head.Copy()
	newGen.Constraints = []Constraint{}
	newGen.Literals = []Literal{}
	for _, c := range gen.Constraints {
		newGen.Constraints = append(newGen.Constraints, c.Copy())
	}
	for _, l := range gen.Literals {
		newGen.Literals = append(newGen.Literals, l.Copy())
	}
	return
}

// Deep Copy
func (rule Rule) Copy() (newRule Rule) {
	newRule = rule
	if rule.hasHead() {
		newRule.Head = rule.Head.Copy()
	}
	newRule.Constraints = []Constraint{}
	newRule.Literals = []Literal{}
	newRule.Generators = []Generator{}
	for _, c := range rule.Constraints {
		newRule.Constraints = append(newRule.Constraints, c.Copy())
	}
	for _, l := range rule.Literals {
		newRule.Literals = append(newRule.Literals, l.Copy())
	}
	for _, g := range rule.Generators {
		newRule.Generators = append(newRule.Generators, g.Copy())
	}
	return
}

func (constraint *Constraint) String() string {
	return string(constraint.LeftTerm) + ComparisonString(constraint.Comparison) + string(constraint.RightTerm)
}

func (name Predicate) String() string {
	return string(name)
}

func (term Term) String() string {
	return string(term)
}

func (g Generator) String() string {
	sb := strings.Builder{}
	sb.WriteString(g.Head.String())
	sb.WriteString(":")
	for _, c := range g.Constraints {
		sb.WriteString(c.String())
		sb.WriteString(":")
	}
	for _, l := range g.Literals {
		sb.WriteString(l.String())
		sb.WriteString(":")
	}
	return strings.TrimSuffix(sb.String(), ":")
}

func (literal Literal) String() string {
	var s string
	if literal.Neg == true {
		s = "~"
	}
	s = s + literal.Name.String() + "["
	for i, x := range literal.Terms {
		s += x.String()
		if i < len(literal.Terms)-1 {
			s += ","
		}
	}
	return s + "]"
}

func (r *Rule) String() string {

	sb := strings.Builder{}

	for _, c := range r.Constraints {
		sb.WriteString(c.String())
		sb.WriteString(", ")
	}

	for _, g := range r.Generators {
		sb.WriteString(g.String())
		sb.WriteString(", ")
	}

	for _, l := range r.Literals {
		sb.WriteString(l.String())
		sb.WriteString(", ")
	}
	tmp := strings.TrimSuffix(sb.String(), ", ")
	sb.Reset()
	sb.WriteString(tmp)

	if !r.IsDisjunction() {
		sb.WriteString(RuleTypeString(r.Typ))
		sb.WriteString(r.Head.String())
	}
	if r.GeneratingHead {
		sb.WriteString("?")
	} else {
		sb.WriteString(".")
	}
	return sb.String()
}

func (p *Program) Debug() {
	fmt.Println("constants:", p.Constants)
	fmt.Println("groundFacts", p.GroundFacts)
	fmt.Println("PredicatsToTuples", p.PredicatToTuples)
	for i, r := range p.Rules {
		fmt.Println("\nrule", i)
		r.Debug()
	}
}

func (r *Rule) Debug() {
	fmt.Println(r.initialTokens)
	if r.hasHead() {
		fmt.Println("head", r.Head)
	}
	fmt.Println("literals", r.Literals)
	fmt.Println("generator", r.Generators)
	fmt.Println("constraints", r.Constraints)
	fmt.Println("logical connection", r.Typ)
	fmt.Println("open Head", r.GeneratingHead)
}

func (p *Program) PrintDebug(level int) {
	if DebugLevel >= level {
		p.PrintFacts()
		p.PrintRules()
	}
}

func (p *Program) PrintTuples() {

	for pred, tuples := range p.PredicatToTuples {
		fmt.Println(pred.String(), ": ")
		for _, t := range tuples {
			fmt.Println("\t", t)
		}
	}

}

func (p *Program) Print() {
	p.PrintFacts()
	p.PrintRules()
}

func (p *Program) PrintRules() {
	for i, r := range p.Rules {
		fmt.Print(r.String())
		if DebugLevel > 0 {
			fmt.Print(" % rule ", i)
		}
		fmt.Println()
	}
}

func (p *Program) PrintFacts() {
	for pred, _ := range p.GroundFacts {
		for _, tuple := range p.PredicatToTuples[pred] {
			fmt.Print(pred)
			for i, t := range tuple {
				if i == 0 {
					fmt.Print("[")
				}
				fmt.Print(t)
				if i == len(tuple)-1 {
					fmt.Print("]")
				} else {
					fmt.Print(",")
				}
			}
			fmt.Println(".")
		}
	}
}

func (p *Program) PrintQuantification() {

	maxIndex := -1

	for k := range p.forallQ {
		if maxIndex < k {
			maxIndex = k
		}
	}
	for k := range p.existQ {
		if maxIndex < k {
			maxIndex = k
		}
	}

	for i := -1; i <= maxIndex; i++ {

		if atoms, ok := p.forallQ[i]; ok {
			fmt.Print("a")
			for _, a := range atoms {
				fmt.Print(" ", a)
			}
			fmt.Println()
		}
		if atoms, ok := p.existQ[i]; ok {
			fmt.Print("e")
			for _, a := range atoms {
				fmt.Print(" ", a)
			}
			fmt.Println()
		}
	}
}
