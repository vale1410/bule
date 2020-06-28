package lib

import (
	"fmt"
	"strconv"
	"strings"
)

type Program struct {
	Rules                 []Rule
	Constants             map[string]int
	FinishCollectingFacts map[Predicate]bool
	PredicateToTuples     map[Predicate][][]int // Contains a slice for all tuples for predicate
	PredicateTupleMap     map[string]bool       // hashmap that contains all positive and negative ground atoms in the program
	PredicateToArity      map[Predicate]int     //
	PredicateExplicit     map[Predicate]bool    // If there is a explicit quantification for predicate
//	ZeroArity             map[Predicate]bool // TODO: do we still need this or can be removed?
	Alternation           [][]Literal
	existQ                map[int][]Literal
	forallQ               map[int][]Literal
}

type Rule struct {
	initialTokens  []Token
	LineNumber     int
	Parent         *Rule
	Generators     []Literal
	Constraints    []Constraint
	Literals       []Literal
	Iterators      []Iterator
	IsQuestionMark bool // if final token is tokenQuestionMark then it generates, otherwise tokenDot
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

type Iterator struct {
	Constraints  []Constraint
	Conditionals []Literal
	Head         Literal
}

type Literal struct {
	Neg   bool
	Fact  bool // if true then search variable with parenthesis () otherwise a fact with brackets []
	Name  Predicate
	Terms []Term
}

func (literal *Literal) IsGround() bool {
	return literal.FreeVars().IsEmpty()
}

type Term string

type Predicate string

func (rule *Rule) IsGround() bool {
	return rule.FreeVars().IsEmpty()
}

func (rule *Rule) IsFact() bool {
	return len(rule.Literals) == 1 && rule.Literals[0].Fact
}

func IsMarkedAsFree(v string) bool {
	return strings.HasPrefix(v, "_")
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
func (iterator Iterator) Copy() (newGen Iterator) {
	newGen = iterator
	newGen.Head = iterator.Head.Copy()
	newGen.Constraints = []Constraint{}
	newGen.Conditionals = []Literal{}
	for _, c := range iterator.Constraints {
		newGen.Constraints = append(newGen.Constraints, c.Copy())
	}
	for _, l := range iterator.Conditionals {
		newGen.Conditionals = append(newGen.Conditionals, l.Copy())
	}
	return
}

// Deep Copy
func (rule Rule) Copy() (newRule Rule) {
	newRule = rule
	newRule.Constraints = []Constraint{}
	newRule.Generators = []Literal{}
	newRule.Literals = []Literal{}
	newRule.Iterators = []Iterator{}
	for _, g := range rule.Generators {
		newRule.Generators = append(newRule.Generators, g.Copy())
	}
	for _, c := range rule.Constraints {
		newRule.Constraints = append(newRule.Constraints, c.Copy())
	}
	for _, l := range rule.Literals {
		newRule.Literals = append(newRule.Literals, l.Copy())
	}
	for _, i := range rule.Iterators {
		newRule.Iterators = append(newRule.Iterators, i.Copy())
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

func (iterator Iterator) String() string {
	sb := strings.Builder{}
	sb.WriteString(iterator.Head.String())
	sb.WriteString(":")
	for _, c := range iterator.Constraints {
		sb.WriteString(c.String())
		sb.WriteString(":")
	}
	for _, l := range iterator.Conditionals {
		sb.WriteString(l.String())
		sb.WriteString(":")
	}
	return strings.TrimSuffix(sb.String(), ":")
}

func (literal Literal) String() string {
	var s string
	if literal.Neg == true {
		s += "~"
	}
	s += literal.Name.String()

	// is 0-arity atom
	if len(literal.Terms) == 0 {
		return s
	}

	var opening string
	var closing string
	if literal.Fact {
		opening = "["
		closing = "]"
	} else {
		opening = "("
		closing = ")"
	}

	s += opening
	for i, x := range literal.Terms {
		s += x.String()
		if i < len(literal.Terms)-1 {
			s += ","
		}
	}
	return s + closing
}

func (rule *Rule) Debug() string {
	sb := strings.Builder{}
	sb.WriteString(rule.String())
	p := rule.Parent
	s := "\n  "
	for p != nil {
		sb.WriteString(s + "╚ " + p.String())
		s += "  "
		p = p.Parent
	}
	sb.WriteString(" %% l:" + strconv.Itoa(rule.LineNumber))
	return sb.String()
}

func (rule *Rule) String() string {

	sb := strings.Builder{}

	if len(rule.Generators) > 0 || len(rule.Constraints) > 0 {

		for _, l := range rule.Generators {
			sb.WriteString(l.String())
			sb.WriteString(", ")
		}

		for _, c := range rule.Constraints {
			sb.WriteString(c.String())
			sb.WriteString(", ")
		}
		tmp := strings.TrimSuffix(sb.String(), ", ")
		sb.Reset()
		sb.WriteString(tmp)

		sb.WriteString(" => ")
	}

	for _, g := range rule.Iterators {
		sb.WriteString(g.String())
		sb.WriteString(", ")
	}

	for _, l := range rule.Literals {
		sb.WriteString(l.String())
		sb.WriteString(", ")
	}
	tmp := strings.TrimSuffix(sb.String(), ", ")
	sb.Reset()
	sb.WriteString(tmp)
	if rule.IsQuestionMark {
		sb.WriteString("?")
	} else {
		sb.WriteString(".")
	}
	return sb.String()
}

func (p *Program) PrintDebug(level int) {
	if DebugLevel >= level {
		p.PrintFacts()
		p.PrintRules()
	}
}

func (p *Program) PrintTuples() {

	for pred, tuples := range p.PredicateToTuples {
		fmt.Println(pred.String(), ": ")
		for _, t := range tuples {
			fmt.Println("\t", t)
		}
	}

}

func (p *Program) Print() {
	p.PrintQuantification()
	p.PrintRules()
}

type RuleError struct {
	R       Rule
	Message string
	Err     error
}

func (err RuleError) Error() string {
	var sb strings.Builder
	sb.WriteString(err.Message + ":\n")
	if err.Err != nil {
		sb.WriteString(err.Err.Error() + ":\n")
	}
	sb.WriteString("Rule \n" + err.R.Debug() + "\n")
	return sb.String()
}

type LiteralError struct {
	L       Literal
	R       Rule
	Message string
}

func (err LiteralError) Error() string {
	var sb strings.Builder
	sb.WriteString(err.Message + ":\n")
	sb.WriteString("Literal " + err.L.String() + "\n")
	sb.WriteString("Rule " + err.R.Debug() + "\n")
	return sb.String()
}

func (p *Program) CheckNoExplicitDeclarationAndNonGroundExplicit() error {
	for _, r := range p.Rules {
		for _, l := range r.AllLiterals() {
			if p.PredicateExplicit[l.Name] && !l.IsGround()  {
				return LiteralError{
					*l,
					r,
					"Every explicit literal should be ground now!",
				}
			}
			if l.Name.String() == "#exist" || l.Name.String() == "#forall" {
				return LiteralError{
					*l,
					r,
					"Should not have any exist literals anymore!",
				}
			}
		}
	}
	return nil
}

func (p *Program) CheckNoGeneratorsOrIterators() error {
	for _, r := range p.Rules {
		if len(r.Generators) > 0 {
			return RuleError{
				r,
				"Should not have any generators anymore!",
				fmt.Errorf("Rule is not free of Generators"),
			}
		}
		if len(r.Iterators) > 0 {
			return RuleError{
				r,
				"Should not have any Iterators anymore!",
				fmt.Errorf("Rule is not free of Iterators: %v", r.Iterators),
			}
		}
	}
	return nil
}

func (p *Program) CheckFactsInIterators() error {
	for _, r := range p.Rules {
		for _, g := range r.Iterators {
			for _, l := range g.Conditionals {
				if !l.Fact {
					return LiteralError{
						l,
						r,
						fmt.Sprintf("In iterator there is a search literal used as a iterator but has to be fact!\n"),
					}
				}
			}
		}
	}
	return nil
}

//func (p *Program) TreatZeroArityOfLiterals() error {
//	// Fix zero arity predicates to a pseudo 1 arity with value 0
//	for _, r := range p.Rules {
//		all := r.AllLiterals()
//		for _, l := range all {
//			if len(l.Terms) == 0 {
//				p.ZeroArity[l.Name] = true
//				l.Terms = []Term{"0"}
//				if len(all) == 1 {
//					l.Fact = true
//
//				}
//			}
//		}
//	}
//	return nil
//}

func (p *Program) CheckArityOfLiterals() error {

	p.PredicateToArity = make(map[Predicate]int)
	p.FinishCollectingFacts = make(map[Predicate]bool)
	for _, r := range p.Rules {
		for _, l := range r.AllLiterals() {
			if l.Fact {
				p.FinishCollectingFacts[l.Name] = false
			}
			if n, ok := p.PredicateToArity[l.Name]; ok {
				if n != len(l.Terms) {
					return LiteralError{*l, r,
						fmt.Sprintf("Literal with arity %d already occurs in program with arity %d. \n "+
							"Bule predicat to arity has to be unique.", len(l.Terms), n)}
				}
			} else {
				p.PredicateToArity[l.Name] = len(l.Terms)
			}

		}
	}
	return nil
}

func (p *Program) CheckSearch() error {
	p.PredicateToArity = make(map[Predicate]int)
	for _, r := range p.Rules {
		for _, l := range r.Literals {
			if n, ok := p.PredicateToArity[l.Name]; ok {
				if n != len(l.Terms) {
					return LiteralError{l, r,
						fmt.Sprintf("Literal with arity %d already occurs in program with arity %d. \n "+
							"Bule predicat to arity has to be unique.", len(l.Terms), n)}
				}
			} else {
				p.PredicateToArity[l.Name] = len(l.Terms)
			}
		}
	}
	return nil
}

func (p *Program) PrintRules() {
	if len(p.Rules) == 0 {
		return
	}
	fmt.Println("%% Rules:")
	for _, r := range p.Rules {
		fmt.Print(r.String())
		if DebugLevel > 0 {
			fmt.Print(" % line: ", r.LineNumber)
		}
		fmt.Println()
	}
}

func (p *Program) PrintFacts() {
	//	fmt.Println("%% Collected Facts:")
	for pred := range p.FinishCollectingFacts {
		for _, tuple := range p.PredicateToTuples[pred] {
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
	for i, quantifier := range p.Alternation {
		if i%2 == 0 {
			fmt.Print("e ")
		} else {
			fmt.Print("a ")
		}
		for _, v := range quantifier {
			fmt.Print(v, " ")
		}
		fmt.Println()
	}
}

// Translates forallQ and existQ into quantification
func (p *Program) MergeConsecutiveQuantificationLevels() {

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

	last := "e"
	var acc []Literal

	for i := -1; i <= maxIndex; i++ {

		if atoms, ok := p.forallQ[i]; ok {
			if last == "a" {
				acc = append(acc, atoms...)
			} else {
				p.Alternation = append(p.Alternation, acc)
				last = "a"
				acc = atoms
			}
		}
		if atoms, ok := p.existQ[i]; ok {
			if last == "e" {
				acc = append(acc, atoms...)
			} else {
				p.Alternation = append(p.Alternation, acc)
				last = "e"
				acc = atoms
			}
		}
	}
	if len(acc) > 0 {
		p.Alternation = append(p.Alternation, acc)
	}
}
