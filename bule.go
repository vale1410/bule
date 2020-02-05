package bule

import (
	"errors"
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/scylladb/go-set/strset"
	"strconv"
	"strings"
	"unicode"
)

var (
	DebugLevel int
)

func debug(level int, s ...interface{}) {
	if level <= DebugLevel {
		fmt.Println(s...)
	}
}

func (p *Program) Debug() {
	fmt.Println("constants:", p.Constants)
	fmt.Println("GroundFacts", p.GroundFacts)
	fmt.Println("globalVars", p.AtomTuples)
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
	fmt.Println("generators", r.Generators)
	fmt.Println("Constraints", r.Constraints)
	fmt.Println("Logical Connection", r.Typ)
	fmt.Println("Open Head", r.GeneratingHead)
}

func (p *Program) Print() {
	for i, r := range p.Rules {
		fmt.Print(i, ": ")
		fmt.Println(r.String())
	}
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

func (r *Rule) String() string {

	sb := strings.Builder{}

	for _, c := range r.Constraints {
		sb.WriteString(c.BoolExpr())
		sb.WriteString(",")
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

func (p *Program) RuleExpansion(check func(r Rule) bool, expand func(Rule) []Rule) (changed bool) {
	newRules := []Rule{}
	for _, rule := range p.Rules {
		if check(rule) {
			changed = true
			for _, newRule := range expand(rule) {
				newRules = append(newRules, newRule)
			}
		} else {
			newRules = append(newRules, rule)

		}
	}
	p.Rules = newRules
	return
}

func (p *Program) TermExpansionOnlyLiterals(check func(r Term) bool, expand func(Term) []Term) (changed bool) {

	// Make rules from the head equivalences
	// Current assumption: head is only one literal, body is a conjunction!
	checkRule := func(r Rule) bool {
		for _, l := range r.Literals {
			for _, t := range l.Terms {
				if check(t) {
					return true
				}
			}
		}
		return false
	}

	expandRule := func(r Rule) (newRules []Rule) {
		for il, literal := range r.Literals {
			for it, term := range literal.Terms {
				if check(term) {
					for _, newTerm := range expand(term) {
						newRule := r.Copy()
						newRule.Literals[il].Terms[it] = newTerm
						newRules = append(newRules, newRule)
					}
					return
				}
			}
		}
		return
	}
	return p.RuleExpansion(checkRule, expandRule)
}

func (p *Program) TermTranslation(transform func(Term) (Term, bool)) (changed bool) {
	var ok bool
	for _, term := range p.AllTerms() {
		*term, ok = transform(*term)
		changed = ok || changed
	}
	return
}

func (p *Program) ExpandIntervals() (changed bool) {
	transform := func(term Term) (newTerms []Term) {
		interval := strings.Split(string(term), "..")
		i1 := evaluateExpression(interval[0])
		i2 := evaluateExpression(interval[1])
		for _, newValue := range makeSet(i1, i2) {
			newTerms = append(newTerms, Term(strconv.Itoa(newValue)))
		}
		return
	}
	check := func(term Term) bool {
		interval := strings.Split(string(term), "..")
		return len(interval) == 2 && groundMathExpression(interval[0]) && groundMathExpression(interval[1])
	}
	return p.TermExpansionOnlyLiterals(check, transform)
}

func (p *Program) RewriteEquivalencesAndImplications() bool {
	// Make rules from the head equivalences
	// Current assumption: head is only one literal, body is a conjunction!
	check := func(r Rule) bool {
		return r.hasHead()
	}
	transform := func(r Rule) (newRules []Rule) {
		assert(r.Typ == ruleTypeEquivalence || r.Typ == ruleTypeImplication)
		newRules = make([]Rule, 0)
		for i, literal := range r.Literals {
			if r.Typ == ruleTypeEquivalence {
				newRule := Rule{}
				newRule.Typ = ruleTypeDisjunction
				newRule.Literals = []Literal{literal.Copy(), r.Head.createNegatedLiteral()}
				newRule.Constraints = r.Constraints
				newRules = append(newRules, newRule)
			}
			r.Literals[i] = literal.createNegatedLiteral()
			r.Typ = ruleTypeDisjunction
		}
		r.Literals = append(r.Literals, r.Head)
		return append(newRules, r)
	}
	return p.RuleExpansion(check, transform)
}

func (p *Program) CollectFacts() (changed bool) {
	check := func(r Rule) bool {
		return r.Typ == ruleTypeDisjunction &&
			len(r.Literals) == 1 &&
			len(r.Generators) == 0 &&
			len(r.Constraints) == 0 &&
			r.FreeVars().IsEmpty()
	}
	transform := func(rule Rule) (empty []Rule) {
		lit := rule.Literals[0]
		p.AtomTuples[lit.Name] = append(p.AtomTuples[lit.Name], evaluateExpressionTuples(lit.Terms))
		p.GroundFacts[lit.Name] = true
		return
	}
	return p.RuleExpansion(check, transform)
}

func (p *Program) ReplaceConstants() bool {
	return p.Simplify(p.Constants)
}

func (p *Program) Simplify(assignment map[string]int) bool {

	transform := func(term Term) (Term, bool) {
		return assign(term, assignment)
	}

	return p.TermTranslation(transform)
}

func (p *Program) AllTerms() (terms []*Term) {
	for _, r := range p.Rules {
		for i := range r.Head.Terms {
			terms = append(terms, &r.Head.Terms[i])
		}
		for _, l := range r.Literals {
			for i := range l.Terms {
				terms = append(terms, &l.Terms[i])
			}
		}
		for _, c := range r.Constraints {
			terms = append(terms, &c.LeftTerm)
			terms = append(terms, &c.RightTerm)
		}
		for _, g := range r.Generators {
			for _, l := range g.Literals {
				for i := range l.Terms {
					terms = append(terms, &l.Terms [i])
				}
			}
			for _, c := range g.Constraints {
				terms = append(terms, &c.LeftTerm)
				terms = append(terms, &c.RightTerm)
			}
		}
	}
	return
}

func (p *Program) ExpandGenerators() {

	for i, r := range p.Rules {
		for _, generator := range r.Generators {
			assignments := p.generateAssignments(
				generator.Literals,
				generator.Constraints)
			for _, assignment := range assignments {
				p.Rules[i].Literals = append(p.Rules[i].Literals,
					generator.Head.assign(assignment))
			}
		}
		p.Rules[i].Generators = []Generator{}
	}
}

func (p *Program) Ground() (clauses []Clause, existQ map[int][]Literal, forallQ map[int][]Literal, maxIndex int) {

	clauses = make([]Clause, 0)
	existQ = make(map[int][]Literal)
	forallQ = make(map[int][]Literal)
	maxIndex = 0

	for _, r := range p.Rules {

		debug(2, "Free Variables:", r.FreeVars())

		//(globals.IsSubset(r.FreeVars()), "There are free variables that are not bound by globals.")
		// TODO NEEDS TO BE FIXED> Literals wont work!
		assignments := p.generateAssignments(r.Literals, r.Constraints)

		debug(2, "Ground Clause:")
		for _, assignment := range assignments {
			clause := Clause{}
			for _, literal := range r.Literals {
				clause = append(clause, literal.assign(assignment))
			}
			debug(2, "clause", clause)
			if len(clause) > 0 {
				if clause[0].Name == "#forall" {
					asserts(len(clause[0].Terms) == 1, "Wrong arity for forall")
					val, err := strconv.Atoi(string(clause[0].Terms[0].String()))
					asserte(err)
					forallQ[val] = append(forallQ[val], clause[1:]...)
					if val > maxIndex {
						maxIndex = val
					}
				} else if clause[0].Name == "#exist" {
					asserts(len(clause[0].Terms) == 1, "Wrong arity for exist")
					val, err := strconv.Atoi(string(clause[0].Terms[0].String()))
					asserte(err)
					existQ[val] = append(existQ[val], clause[1:]...)
					if val > maxIndex {
						maxIndex = val
					}
				} else {
					clauses = append(clauses, clause)
				}
			}
		}
		debug(2)
	}
	return
}

func (name AtomName) String() string {
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
		sb.WriteString(c.BoolExpr())
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

func (literal Literal) Copy() Literal {
	t := make([]Term, len(literal.Terms))
	copy(t, literal.Terms)
	literal.Terms = t
	return literal
}

// only works on disjunctions
func (rule *Rule) FreeVars() *strset.Set {
	assert(rule.IsDisjunction())
	set := strset.New()
	for _, a := range rule.Literals {
		set.Merge(a.FreeVars())
	}
	return set
}

func (literal Literal) FreeVars() *strset.Set {
	set := strset.New()
	for _, t := range literal.Terms {
		set.Merge(t.FreeVars())
	}
	return set
}

func (constraint Constraint) FreeVars() *strset.Set {
	set := constraint.LeftTerm.FreeVars()
	set.Merge(constraint.RightTerm.FreeVars())
	return set
}

func (term Term) FreeVars() *strset.Set {
	s := strings.ReplaceAll(term.String(), "#mod", "%")
	s = strings.ReplaceAll(s, "#log", "%")
	f := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c) && c != '_'
	}
	variables := strings.FieldsFunc(s, f)
	set := strset.New()
	for _, x := range variables {
		if !number(x) {
			set.Add(x)
		}
	}
	return set
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

func (p *Program) generateAssignments(literals []Literal, constraints []Constraint) []map[string]int {

	// Assumption:
	// 1) freevars of literals are all disjunct
	// 2) literal is GroundFact.
	// 3) literal is of form <name>[A,B].
	{
		set := strset.New()
		for _, lit := range literals {
			asserts(strset.Intersection(lit.FreeVars(), set).IsEmpty(),
				"freevars of literals are all disjunct", set.String(), lit.FreeVars().String())
			asserts(p.GroundFacts[lit.Name],
				"Is Ground fact", lit.String())
			set.Merge(lit.FreeVars())
		}
	}

	allPossibleAssignments := make([]map[string]int, 1, 32)
	allPossibleAssignments[0] = make(map[string]int)

	for _, literal := range literals {
		if termsDomain, ok := p.AtomTuples[literal.Name]; ok {
			newAssignments := make([]map[string]int, 0, len(allPossibleAssignments)*len(termsDomain))
			for _, tuple := range termsDomain {
				assert(len(tuple) == len(literal.Terms))
				for _, assignment := range allPossibleAssignments {
					newAssignment := make(map[string]int)
					for key, value := range assignment {
						newAssignment[key] = value
					}
					for i, value := range tuple {
						newAssignment[string(literal.Terms[i])] = value
					}
					newAssignments = append(newAssignments, newAssignment)

				}
			}
			allPossibleAssignments = newAssignments
		} else {
			panic("literal doesnt have domain " + literal.String())
		}
	}

	assignments := make([]map[string]int, 0, 32)

	for _, assignment := range allPossibleAssignments {
		fmt.Println(assignment)
		// check all constraints
		allConstraintsTrue := true
		for _, cons := range constraints {
			fmt.Println(assignment)
			fmt.Println(cons.BoolExpr())
			cons.LeftTerm, _ = assign(cons.LeftTerm, assignment)
			cons.RightTerm, _ = assign(cons.RightTerm, assignment)
			asserts(groundBoolLogicalMathExpression(cons.BoolExpr()), "Must be bool expression", cons.BoolExpr(), "from", cons.BoolExpr())
			allConstraintsTrue = allConstraintsTrue && evaluateBoolExpression(cons.BoolExpr())
		}
		if allConstraintsTrue {
			assignments = append(assignments, assignment)
		}
	}
	return assignments
}

//func (p *Program) generateAssignments(variables []string, constraints []Constraint) []map[string]int {
//
//	allPossibleAssignments := make([]map[string]int, 1, 32)
//	allPossibleAssignments[0] = make(map[string]int)
//
//	for _, variable := range variables {
//		if dom, ok := p.GlobalDefinitions[variable]; ok {
//			newAssignments := make([]map[string]int, 0, len(allPossibleAssignments)*len(dom))
//			for _, val := range dom {
//				for _, assignment := range allPossibleAssignments {
//					newAssignment := make(map[string]int)
//					for key, value := range assignment {
//						newAssignment[key] = value
//					}
//					newAssignment[variable] = val
//					newAssignments = append(newAssignments, newAssignment)
//				}
//			}
//			allPossibleAssignments = newAssignments
//		} else {
//			panic("variable doesnt have domain " + variable)
//		}
//	}
//
//	assignments := make([]map[string]int, 0, 32)
//
//	for _, assignment := range allPossibleAssignments {
//		fmt.Println(assignment)
//		fmt.Println(constraints)
//		// check all constraints
//		allConstraintsTrue := true
//		for _, cons := range constraints {
//			tmp := assign(cons.BoolExpr(), assignment)
//			//tmp = assign(tmp, p.Constants)
//			//tmp = strings.ReplaceAll(string(tmp), "#mod", "%")
//			asserts(groundBoolLogicalMathExpression(tmp), "Must be bool expression", tmp, "from", cons.BoolExpr())
//			allConstraintsTrue = allConstraintsTrue && evaluateBoolExpression(tmp)
//		}
//		if allConstraintsTrue {
//			assignments = append(assignments, assignment)
//		}
//	}
//	return assignments
//}

type Program struct {
	Rules       []Rule
	Constants   map[string]int
	AtomTuples  map[AtomName][][]int
	GroundFacts map[AtomName]bool
}

type Clause []Literal

type ruleType int

const (
	ruleTypeDisjunction ruleType = iota
	ruleTypeImplication
	ruleTypeEquivalence
)

type Rule struct {
	initialTokens  []Token
	Head           Literal
	Literals       []Literal
	Generators     []Generator
	Constraints    []Constraint
	GeneratingHead bool     // if final token is tokenQuestionMark then it generates, otherwise tokenDot
	Typ            ruleType // Can be Implication or Equivalence or RuleComma(normal rule)
}

func (r *Rule) hasHead() bool {
	return r.Head.Name != ""
}

// is a math expression that evaluates to true or false
// Constraints can contain variables
// supported are <,>,<=,>=,==
// E.g..: A*3v<=v5-2*R/7#mod3.
type Constraint struct {
	LeftTerm    Term
	Comparision tokenKind
	RightTerm   Term
}

func (constraint *Constraint) BoolExpr() string {
	return string(constraint.LeftTerm) + ComparisonString(constraint.Comparision) + string(constraint.RightTerm)
}

func (constraint Constraint) Copy() (cons Constraint) {
	cons = constraint
	cons.LeftTerm = constraint.LeftTerm
	cons.RightTerm = constraint.RightTerm
	return cons
}

type Generator struct {
	Constraints []Constraint
	Literals    []Literal
	Head        Literal
}

type Literal struct {
	Neg   bool
	Name  AtomName
	Terms []Term
}

type Term string

type AtomName string

// Makes a deep copy and creates a new Literal
func (literal Literal) assign(assignment map[string]int) (newLiteral Literal) {
	newLiteral = literal.Copy()
	for i, term := range literal.Terms {
		newLiteral.Terms[i], _ = assign(term, assignment)
	}
	return newLiteral
}

//returns true if term has been changed
func assign(term Term, assignment map[string]int) (Term, bool) {
	input := term.String()
	output := term.String()
	for Const, Val := range assignment {
		output = strings.ReplaceAll(output, Const, strconv.Itoa(Val))
	}
	if groundMathExpression(output) {
		output = strconv.Itoa(evaluateExpression(output))
	}
	return Term(output), input != output
}

func number(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func groundMathExpression(s string) bool {
	//	r, _ := regexp.MatchString("[0-9+*/%]+", s)
	return "" == strings.Trim(string(s), "0123456789+*%-/()")
}

func groundBoolLogicalMathExpression(s string) bool {
	//	r, _ := regexp.MatchString("[0-9+*/%!=><]+", s)
	return "" == strings.Trim(string(s), "0123456789+*%-=><()!&")
}

// Evaluates a ground math expression, needs to path mathExpression
func evaluateBoolExpression(term string) bool {
	term = strings.ReplaceAll(term, "#mod", "%")
	expression, err := govaluate.NewEvaluableExpression(term)
	assertx(err, term)
	result, err := expression.Evaluate(nil)
	assertx(err, term)
	return result.(bool)
}

// Evaluates a ground math expression, needs to path mathExpression
func evaluateExpression(termExpression string) int {
	termExpression = strings.ReplaceAll(termExpression, "#mod", "%")
	expression, err := govaluate.NewEvaluableExpression(termExpression)
	assertx(err, termExpression)
	result, err := expression.Evaluate(nil)
	assertx(err, termExpression)
	return int(result.(float64))
}

// Evaluates a ground math expression, needs to path mathExpression
func evaluateExpressionTuples(terms []Term) (result []int) {
	for _, t := range terms {
		result = append(result, evaluateExpression(string(t)))
	}
	return
}

func (literal *Literal) createNegatedLiteral() Literal {
	a := literal.Copy()
	a.Neg = !a.Neg
	return a
}

func (c *Constraint) createNegatedConstraint() Constraint {
	constraint := c.Copy()
	switch constraint.Comparision {
	case tokenComparisonLT:
		constraint.Comparision = tokenComparisonGE
	case tokenComparisonGT:
		constraint.Comparision = tokenComparisonLE
	case tokenComparisonEQ:
		constraint.Comparision = tokenComparisonNQ
	case tokenComparisonGE:
		constraint.Comparision = tokenComparisonLT
	case tokenComparisonLE:
		constraint.Comparision = tokenComparisonGT
	case tokenComparisonNQ:
		constraint.Comparision = tokenComparisonEQ
	}
	return constraint
}

func ComparisonString(tokenComparison tokenKind) (s string) {
	switch tokenComparison {
	case tokenComparisonLT:
		s = "<"
	case tokenComparisonGT:
		s = ">"
	case tokenComparisonEQ:
		s = "=="
	case tokenComparisonGE:
		s = ">="
	case tokenComparisonLE:
		s = "<="
	case tokenComparisonNQ:
		s = "!="
	}
	return
}

func RuleTypeString(typ ruleType) (s string) {
	switch typ {
	case ruleTypeImplication:
		s = " -> "
	case ruleTypeEquivalence:
		s = " <->"
	case ruleTypeDisjunction:
	}
	return
}

func assert(condition bool) {
	if !condition {
		panic("ASSERT FAILED")
	}
}

func asserts(condition bool, info ...string) {
	if !condition {
		s := ""
		for _, x := range info {
			s += x + " "
		}
		fmt.Println(s)
		panic(errors.New(s))
	}
}

func asserte(err error) {
	if err != nil {
		panic(err)
	}
}

func assertx(err error, info ...string) {
	if err != nil {
		for _, s := range info {
			fmt.Print(s, " ")
		}
		fmt.Println()
		panic(err)
	}
}

func makeSet(a, b int) (c []int) {
	if a >= b {
		return []int{}
	}
	c = make([]int, 0, b-a)
	for i := a; i <= b; i++ {
		c = append(c, i)
	}
	return
}
