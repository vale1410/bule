package lib

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


func (p *Program) Print() {
	p.PrintFacts()
	p.PrintRules()
}

func (p *Program) PrintRules() {
	for i, r := range p.Rules {
		fmt.Print(r.String())
		if DebugLevel > 1 {
			fmt.Print(" % rule ", i)
		}
		fmt.Println()
	}
}

func (p *Program) PrintFacts() {
	for pred,_ := range p.GroundFacts {
		for _,tuple := range p.PredicatToTuples[pred] {
			fmt.Print(pred)
			for i,t := range  tuple {
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

	maxIndex := 0

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

	for i := 0; i <= maxIndex; i++ {

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

// goes through all rules and expands expands if check is true.
// Note that this does not expand the generated rules. (i.e. run until fixpoint)
func (p *Program) RuleExpansion(check func(r Rule) bool, expand func(Rule) []Rule) (changed bool) {
	var newRules []Rule
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

func (r *Rule) TermTranslation(transform func(Term) (Term, bool)) (changed bool) {
	var ok bool
	for _, term := range r.AllTerms() {
		*term, ok = transform(*term)
		changed = ok || changed
	}
	return
}

func (p *Program) ExpandRanges() (changed bool) {
	transform := func(term Term) (newTerms []Term) {
		interval := strings.Split(string(term), "..")
		i1 := evaluateTermExpression(interval[0])
		i2 := evaluateTermExpression(interval[1])
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

func (p *Program) InstanciateAndRemoveFacts() (changed bool) {
	// Find rule with fact
	check := func(r Rule) bool {
		for _, lit := range r.Literals {
			if p.GroundFacts[lit.Name] {
				return true
			}
		}
		return false
	}

	transform := func(rule Rule) (generatedRules []Rule) {

		var fact Literal
		var i int
		for i, fact = range rule.Literals {
			if p.GroundFacts[fact.Name] {
				break
			}
		}
		rule.Literals = append(rule.Literals[:i], rule.Literals[i+1:]...)

		for _, tuple := range p.PredicatToTuples[fact.Name] {
			newRule := rule.Copy()
			for j, val := range tuple {
				newConstraint := Constraint{
					LeftTerm:   fact.Terms[j],
					RightTerm:  Term(strconv.Itoa(val)), // TODO Could be simpler ...
					Comparison: tokenComparisonEQ,
				}
				newRule.Constraints = append(newRule.Constraints, newConstraint)
			}
			generatedRules = append(generatedRules, newRule)
		}
		return generatedRules
	}
	return p.RuleExpansion(check, transform)
}

func (p *Program) FindNewFacts() (changed bool) {
	// All literals are facts but one!
	// No generators
	check := func(r Rule) bool {
		if len(r.Generators) != 0 || r.Typ != ruleTypeDisjunction {
			return false
		}
		numberOfNoneFacts := len(r.Literals)
		if numberOfNoneFacts < 2 {
			return false
		}
		for _, lit := range r.Literals {
			if p.GroundFacts[lit.Name] {
				numberOfNoneFacts--
			}
		}
		return numberOfNoneFacts == 1
	}

	transform := func(rule Rule) (empty []Rule) {
		var facts []Literal
		var newFact Literal
		for _, lit := range rule.Literals {
			if p.GroundFacts[lit.Name] {
				facts = append(facts, lit)
			} else {
				newFact = lit
			}
		}
		assignments := p.generateAssignments(facts, rule.Constraints)
		for _, assignment := range assignments {
			newLit := newFact.Copy()
			for i, Term := range newLit.Terms {
				newLit.Terms[i], _ = assign(Term, assignment)
			}
			p.InsertTuple(newLit)
		}
		p.GroundFacts[newFact.Name] = true
		return // remove rule
	}
	return p.RuleExpansion(check, transform)
}

func (p *Program) InsertTuple(lit Literal) {
	groundTerms := evaluateExpressionTuples(lit.Terms)
	tuples := p.PredicatToTuples[lit.Name]
	for _, tuple := range tuples {
		isSame := true
		for i, t := range tuple {
			if groundTerms[i] != t {
				isSame = false
				break
			}
		}
		if isSame {
			return
		}
	}
	p.PredicatToTuples[lit.Name] = append(p.PredicatToTuples[lit.Name], groundTerms)
}

func (p *Program) CollectGroundFacts() (changed bool) {
	check := func(r Rule) bool {
		return r.Typ == ruleTypeDisjunction &&
			len(r.Literals) == 1 &&
			len(r.Generators) == 0 &&
			len(r.Constraints) == 0 &&
			r.FreeVars().IsEmpty()
	}
	transform := func(rule Rule) (empty []Rule) {
		lit := rule.Literals[0]
		p.PredicatToTuples[lit.Name] = append(p.PredicatToTuples[lit.Name], evaluateExpressionTuples(lit.Terms))
		p.GroundFacts[lit.Name] = true
		return // remove rule
	}
	return p.RuleExpansion(check, transform)
}

func (constraint Constraint) IsInstantiation() (is bool, variable string, value int) {
	freeVars := constraint.FreeVars()
	if freeVars.Size() == 1 && freeVars.Pop() == constraint.LeftTerm.String() && groundMathExpression(string(constraint.RightTerm)) {
		return true, constraint.LeftTerm.String(), evaluateTermExpression(string(constraint.RightTerm))
	} else if freeVars.Size() == 1 && freeVars.Pop() == constraint.RightTerm.String() && groundMathExpression(string(constraint.LeftTerm)) {
		return true, constraint.RightTerm.String(), evaluateTermExpression(string(constraint.LeftTerm))

	}
	return false, "", 0
}

// Remove Rules with false constraint
// Remove true constraints from Rule
// This is essentially Unit Propagation on Constraint Instantiation
func (p *Program) CleanRules() bool {

	check := func(r Rule) bool {
		for _, cons := range r.Constraints {
			re, _ := cons.GroundBoolExpression()
			if re {
				return true
			}
		}
		return false
	}

	transform := func(rule Rule) (result []Rule) {
		newRule := rule
		newRule.Constraints = []Constraint{}
		for _, cons := range rule.Constraints {
			isGround , boolResult := cons.GroundBoolExpression()
			if isGround {
				if boolResult {
					result = []Rule{newRule}
				} else {
					result = []Rule{}
					break
				}
			} else {
				newRule.Constraints = append(rule.Constraints, cons)
			}
		}
		return
	}
	return p.RuleExpansion(check, transform)
}

// for each Constraint X==<Value>
// Rewrite all Terms with X <- <Value>
func (p *Program) TransformConstraintsToInstantiation() bool {

	check := func(r Rule) bool {
		for _, cons := range r.Constraints {
			is, _, _ := cons.IsInstantiation()
			if is {
				return true
			}
		}
		return false
	}

	transform := func(rule Rule) (empty []Rule) {
		var i int
		var cons Constraint
		var is bool
		var variable string
		var value int
		for i, cons = range rule.Constraints {
			is, variable, value = cons.IsInstantiation()
			if is {
				break
			}
		}
		assignment := map[string]int{variable: value}
		rule.Constraints = append(rule.Constraints[:i], rule.Constraints[i+1:]...)
		rule.Simplify(assignment)
		return []Rule{rule}
	}
	return p.RuleExpansion(check, transform)
}

func (p *Program) ReplaceConstantsAndMathFunctions() {

	transform := func(term Term) (Term, bool) {
		out := strings.ReplaceAll(string(term),"#mod", "%")
		return Term(out), out != string(term)
	}

	for i := range p.Rules {
		p.Rules[i].TermTranslation(transform)
		p.Rules[i].Simplify(p.Constants)
	}
}

func (r *Rule) Simplify(assignment map[string]int) bool {

	transform := func(term Term) (Term, bool) {
		return assign(term, assignment)
	}

	return r.TermTranslation(transform)
}

func (r *Rule) AllTerms() (terms []*Term) {
	for i := range r.Head.Terms {
		terms = append(terms, &r.Head.Terms[i])
	}
	for _, l := range r.Literals {
		for i := range l.Terms {
			terms = append(terms, &l.Terms[i])
		}
	}
	for i := range r.Constraints {
		terms = append(terms, &r.Constraints[i].LeftTerm)
		terms = append(terms, &r.Constraints[i].RightTerm)
	}
	for _, g := range r.Generators {
		for i := range g.Head.Terms {
			terms = append(terms, &g.Head.Terms[i])
		}
		for _, l := range g.Literals {
			for i := range l.Terms {
				terms = append(terms, &l.Terms[i])
			}
		}
		for j := range g.Constraints {
			terms = append(terms, &g.Constraints[j].LeftTerm)
			terms = append(terms, &g.Constraints[j].RightTerm)
		}
	}
	return
}

func (p *Program) ExpandConditionals() {

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

func (p* Program) CollectGroundTuples() {

	for _, r := range p.Rules {
		for _, literal := range r.Literals {
			if literal.FreeVars().IsEmpty() {
				p.InsertTuple(literal)
			}
		}
	}
}

func (p* Program) GroundFromTuples() bool {
	check := func(r Rule) bool {
		return !r.IsGround()
	}

	transform := func(rule Rule) (result []Rule) {
		assignments := p.generateAssignments(rule.Literals, rule.Constraints)
		for _, assignment := range assignments {
			newRule := rule.Copy()
			newRule.Constraints = []Constraint{}
			for i, lit := range newRule.Literals {
				newRule.Literals[i] =  lit.assign(assignment)
			}
			result = append(result, newRule)
		}
		return
	}
	return p.RuleExpansion(check, transform)
}

func (p *Program) ExtractQuantors() {

	p.forallQ = make(map[int][]Literal, 0)
	p.existQ = make(map[int][]Literal, 0)

	checkA := func(r Rule) bool {
		return r.Literals[0].Name == "#forall"
	}

	transformA := func(rule Rule) (remove []Rule) {
		lit := rule.Literals[0]
		asserts(len(lit.Terms) == 1, "Wrong arity for forall")
		val, err := strconv.Atoi(string(lit.Terms[0].String()))
		asserte(err)
		p.forallQ[val] = append(p.forallQ[val], rule.Literals[1:]...)
		return
	}

	checkE := func(r Rule) bool {
		return r.Literals[0].Name == "#exist"
	}

	transformE := func(rule Rule) (remove []Rule) {
		lit := rule.Literals[0]
		asserts(len(lit.Terms) == 1, "Wrong arity for exist")
		val, err := strconv.Atoi(string(lit.Terms[0].String()))
		asserte(err)
		p.existQ[val] = append(p.existQ[val], rule.Literals[1:]...)
		return
	}

	p.RuleExpansion(checkA,transformA)
	p.RuleExpansion(checkE,transformE)
	return
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
	s := term.String()
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
	// 3) literal is of form <name>[A,B..].
	{
		set := strset.New()
		for _, lit := range literals {
			asserts(strset.Intersection(lit.FreeVars(), set).IsEmpty(),
				"freevars of literals are all disjunct", set.String(), lit.FreeVars().String())
//			asserts(p.GroundFacts[lit.Name],
//				"Is ExtractQuantors fact", lit.String())
			set.Merge(lit.FreeVars())
		}
	}

	allPossibleAssignments := make([]map[string]int, 1, 32)
	allPossibleAssignments[0] = make(map[string]int)

	for _, literal := range literals {
		if termsDomain, ok := p.PredicatToTuples[literal.Name]; ok {
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
		allConstraintsTrue := true
		for _, cons := range constraints {
			debug(2, "assignment:", assignment)
			debug(2, "BoolExpression before assignment:", cons.String())
			cons.LeftTerm, _ = assign(cons.LeftTerm, assignment)
			cons.RightTerm, _ = assign(cons.RightTerm, assignment)
			isGround, result := cons.GroundBoolExpression()
			asserts(isGround, "Must be bool expression ", cons.String())
			allConstraintsTrue = allConstraintsTrue && result
		}
		if allConstraintsTrue {
			assignments = append(assignments, assignment)
		}
	}
	return assignments
}

func (p *Program) PrintTuples() {

	for pred, tuples := range p.PredicatToTuples {
		fmt.Println(pred.String(), ": ")
		for _, t := range tuples {
			fmt.Println("\t", t)
		}
	}

}

type Program struct {
	Rules            []Rule
	Constants        map[string]int
	PredicatToTuples map[Predicate][][]int
	GroundFacts      map[Predicate]bool
	Search           map[Predicate]bool
	existQ 			 map[int][]Literal
	forallQ 		 map[int][]Literal
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
	LeftTerm   Term
	Comparison tokenKind
	RightTerm  Term
}

func (constraint *Constraint) String() string {
	return string(constraint.LeftTerm) + ComparisonString(constraint.Comparison) + string(constraint.RightTerm)
}

func (constraint *Constraint) GroundBoolExpression() (isGround bool, result bool) {
	isGround = groundMathExpression(string(constraint.LeftTerm)) && groundMathExpression(string(constraint.RightTerm))
	if !isGround {
		return
	}
	result = evaluateBoolExpression(constraint.String())
	return
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
	Name  Predicate
	Terms []Term
}

type Term string

type Predicate string

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
	output := term.String()
	for Const, Val := range assignment {
		// TODO currenlty variables need to be prefix free, i.e. X, Xa, will make problems :\
		// Use Term parser for getting proper FreeVariables. Or the FreeVariables function
		output = strings.ReplaceAll(output, Const, strconv.Itoa(Val))
	}
	if groundMathExpression(output) {
		output = strconv.Itoa(evaluateTermExpression(output))
	}
	return Term(output), term.String() != output
}

func number(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func groundMathExpression(s string) bool {
	//	r, _ := regexp.MatchString("[0-9+*/%]+", s)
	return "" == strings.Trim(string(s), "0123456789+*%-/()")
}

// Evaluates a ground math expression, needs to path mathExpression
func evaluateBoolExpression(termComparison string) bool {
//	termComparison = strings.ReplaceAll(termComparison, "#mod", "%")
	expression, err := govaluate.NewEvaluableExpression(termComparison)
	assertx(err, termComparison)
	result, err := expression.Evaluate(nil)
	assertx(err, termComparison)
	return result.(bool)
}

// Evaluates a ground math expression, needs to path mathExpression
func evaluateTermExpression(termExpression string) int {
//	termExpression = strings.ReplaceAll(termExpression, "#mod", "%")
	expression, err := govaluate.NewEvaluableExpression(termExpression)
	assertx(err, termExpression)
	result, err := expression.Evaluate(nil)
	assertx(err, termExpression)
	return int(result.(float64))
}

// Evaluates a ground math expression, needs to pass mathExpression
func evaluateExpressionTuples(terms []Term) (result []int) {
	for _, t := range terms {
		result = append(result, evaluateTermExpression(string(t)))
	}
	return
}

func (literal *Literal) createNegatedLiteral() Literal {
	a := literal.Copy()
	a.Neg = !a.Neg
	return a
}

func (constraint *Constraint) createNegatedConstraint() Constraint {
	negatedConstraint := constraint.Copy()
	switch constraint.Comparison {
	case tokenComparisonLT:
		constraint.Comparison = tokenComparisonGE
	case tokenComparisonGT:
		constraint.Comparison = tokenComparisonLE
	case tokenComparisonEQ:
		constraint.Comparison = tokenComparisonNQ
	case tokenComparisonGE:
		constraint.Comparison = tokenComparisonLT
	case tokenComparisonLE:
		constraint.Comparison = tokenComparisonGT
	case tokenComparisonNQ:
		constraint.Comparison = tokenComparisonEQ
	}
	return negatedConstraint
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
