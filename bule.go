package bule

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/scylladb/go-set/strset"
	"os"
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
	//fmt.Println("domains:", p.Domains)
	//fmt.Println("globalVars", p.GlobalVariables())
	for i, r := range p.Rules {
		fmt.Println("\nrule", i)
		r.Debug()
	}
}

func (r *Rule) Debug() {
	fmt.Println( r.debugging)
	if r.hasHead() {
		fmt.Println("head", r.Head)
	}
	fmt.Println("literals", r.Literals)
	fmt.Println("literalGenerators", r.LiteralGenerator)
	fmt.Println("Constraints", r.Constraints)
	fmt.Println("Logical Connection",r.LogicalConnection)
	fmt.Println("Open Head",r.OpenHead)
}


func (p *Program) RewriteEquivalences() {
	// Make rules from the head equivalences
	// Current assumption: head is only one literal, body is a conjunction!
	newRules := make([]Rule, 0)
	for _, r := range p.Rules {
		if r.hasHead() {
			// Check that freeVars are the same for right and left side (assumption)
			for i, atom := range r.Literals {
				newRule := Rule{}
				newRule.Literals = []Literal{r.Head.createNegatedLiteral(), atom.Copy()}
				newRule.Constraints = r.Constraints
				newRules = append(newRules, newRule)
				r.Literals[i] = atom.createNegatedLiteral()
			}
			r.Literals = append(r.Literals, r.Head)
		}
		newRules = append(newRules, r)
	}
	p.Rules = newRules
}

func (p *Program) ExpandGenerators() {

	//for i, r := range p.Rules {
	//	for _, atomG := range r.LiteralGenerator {
	//		assignments := p.generateAssignments(
	//			atomG.variables,
	//			atomG.constraints)
	//		for _, assignment := range assignments {
	//			p.Rules[i].Literals = append(p.Rules[i].Literals,
	//				atomG.literal.simplifyAtom(assignment))
	//		}
	//	}
	//	if len(r.LiteralGenerator) > 0 {
	//		debug(2, "after generation of atoms:")
	//		debug(2, "atoms", r.Literals)
	//	}
	//}
}

func (p *Program) Ground() (
	//gRules []GroundRule,
	existQ map[int][]Literal,
	forallQ map[int][]Literal,
	maxIndex int) {
	//
	//gRules = make([]GroundRule, 0)
	//existQ = make(map[int][]Literal)
	//forallQ = make(map[int][]Literal)
	//maxIndex = 0
	////globals := p.GlobalVariables()
	//
	//for _, r := range p.Rules {
	//
	//	debug(2, "freevariables", r.FreeVars())
	//
	//	//asserts(globals.IsSubset(r.FreeVars()), "There are free variables that are not bound by globals.")
	//	assignments := p.generateAssignments(r.FreeVars().List(), r.Constraints)
	//
	//	debug(2, "Ground Rules:")
	//	for _, assignment := range assignments {
	//		gRule := GroundRule{}
	//		for _, atom := range r.Literals {
	//			gRule.literals = append(gRule.literals, atom.simplifyAtom(assignment))
	//		}
	//		debug(2, "gRule", gRule)
	//		if len(gRule.literals) > 0 {
	//			if gRule.literals[0].Name == "#forall" {
	//				asserts(len(gRule.literals[0].Terms) == 1, "Wrong arity for forall")
	//				val, err := strconv.Atoi(string(gRule.literals[0].Terms[0].String()))
	//				asserte(err)
	//				forallQ[val] = append(forallQ[val], gRule.literals[1:]...)
	//				if val > maxIndex {
	//					maxIndex = val
	//				}
	//			} else if gRule.literals[0].Name == "#exist" {
	//				asserts(len(gRule.literals[0].Terms) == 1, "Wrong arity for exist")
	//				val, err := strconv.Atoi(string(gRule.literals[0].Terms[0].String()))
	//				asserte(err)
	//				existQ[val] = append(existQ[val], gRule.literals[1:]...)
	//				if val > maxIndex {
	//					maxIndex = val
	//				}
	//			} else {
	//				gRules = append(gRules, gRule)
	//			}
	//		}
	//	}
	//	debug(2)
	//}
	return
}

func (name AtomName) String() string {
	return name.String()
}
//
//func (atomG Generator) String() string {
//	return fmt.Sprintf("%v:constraints%v:vars%v",
//		atomG.literal.String(),
//		atomG.constraints,
//		atomG.variables)
//}
//
//func (literal Literal) String() string {
//	var s string
//	if literal.Neg == false {
//		s = "~"
//	}
//	s = s + literal.Name.String() + "["
//	for i, x := range literal.Terms {
//		s += x.String()
//		if i < len(literal.Terms)-1 {
//			s += ","
//		}
//	}
//	return s + "]"
//}

func (literal Literal) FreeVars() *strset.Set {
	set := strset.New()
	for _, t := range literal.Terms {
		set.Merge(t.FreeVars())
	}
	return set
}

func (literal Literal) Copy() Literal {
	t := make([]TermExpression, len(literal.Terms))
	copy(t, literal.Terms)
	literal.Terms = t
	return literal
}

func (term TermExpression) FreeVars() *strset.Set {
	f := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c) && c != '_'
	}
	variables := strings.FieldsFunc(Tokens(term).String(), f)
	set := strset.New()
	for _, x := range variables {
		if !number(x) {
			set.Add(x)
		}
	}
	return set
}

func (rule *Rule) FreeVars() *strset.Set {
	set := rule.Head.FreeVars()
	for _, a := range rule.Literals {
		set.Merge(a.FreeVars())
	}
	return set
}

func (p *Program) generateAssignments(variables []string, constraints []Constraint) []map[string]int {

	//allPossibleAssignments := make([]map[string]int, 1, 32)
	//allPossibleAssignments[0] = make(map[string]int)
	//
	//for _, variable := range variables {
	//	if dom, ok := p.Domains[variable]; ok {
	//		newAssignments := make([]map[string]int, 0, len(allPossibleAssignments)*len(dom))
	//		for _, val := range dom {
	//			for _, assignment := range allPossibleAssignments {
	//				newAssignment := make(map[string]int)
	//				for key, value := range assignment {
	//					newAssignment[key] = value
	//				}
	//				newAssignment[variable] = val
	//				newAssignments = append(newAssignments, newAssignment)
	//			}
	//		}
	//		allPossibleAssignments = newAssignments
	//	} else {
	//		panic("variable doesnt have domain " + variable)
	//	}
	//}
	//
	assignments := make([]map[string]int, 0, 32)
	//
	//for _, assignment := range allPossibleAssignments {
	//	fmt.Println(assignment)
	//	fmt.Println(constraints)
	//	// check all constraints
	//	allConstraintsTrue := true
	//	for _, cons := range constraints {
	//		tmp := assign(cons.BoolExpr, assignment)
	//		tmp = assign(tmp, p.Constants)
	//		tmp = strings.ReplaceAll(string(tmp), "#mod", "%")
	//		asserts(groundBoolLogicalMathExpression(tmp), "Must be bool expression", tmp, "from", cons.BoolExpr)
	//		allConstraintsTrue = allConstraintsTrue && evaluateBoolExpression(tmp)
	//	}
	//	if allConstraintsTrue {
	//		assignments = append(assignments, assignment)
	//	}
	//}
	return assignments
}

func ParseProgram(path string) (p Program) {
	// open a file or stream
	var scanner *bufio.Scanner
	file, err := os.Open(path)
	if err != nil {
		scanner = bufio.NewScanner(os.Stdin)
	} else {
		defer file.Close()
		scanner = bufio.NewScanner(file)
	}
	for scanner.Scan() {

		s := strings.TrimSpace(scanner.Text())
		if pos := strings.Index(s, "%"); pos >= 0 {
			s = s[:pos]
		}

		s = strings.Replace(s, " ", "", -1)

		if s == "" {
			continue
		}
	}
	return
}

	//p.Domains = make(map[string][]int, 0)
	//p.Constants = make(map[string]int)
	//// Math operator replacement
	//

	//
	//	// parsing a global definition like "
	//	// X = {4..5}.
	//	// or c = 5.
	//	// or k  = c*2.
	//	if strings.Contains(s, "=") &&
	//		!isConstraint(s) &&
	//		!strings.Contains(s, "<=>") {
	//
	//		def := strings.Split(s, "=")
	//		asserts(len(def) == 2, s)
	//		if strings.Contains(def[1], "..") {
	//			set := strings.Trim(def[1], "{}.")
	//			interval := strings.Split(set, "..")
	//			i1, _ := strconv.Atoi(interval[0])
	//			x := assign(interval[1], p.Constants)
	//			i2 := evaluateExpression(x)
	//			p.Domains[def[0]] = makeSet(i1, i2)
	//		} else { // this is a constant
	//			term := assign(def[1], p.Constants)
	//			if !groundMathExpression(term) {
	//				panic("is not ground" + term)
	//			}
	//			p.Constants[def[0]] = evaluateExpression(term)
	//			p.Domains[def[0]] = []int{evaluateExpression(term)}
	//		}
	//		continue
	//	}
	//}

type Program struct {
	Rules []Rule
	//Domains   map[string][]int
	Constants      map[string]int
	GroundAtom     map[AtomName][]int
	AtomToRulesMap map[AtomName]*Rule
}

//func (p *Program) GlobalVariables() *strset.Set {
//	set := strset.New()
//	for k := range p.Domains {
//		set.Add(k)
//	}
//	for k := range p.Constants {
//		set.Add(k)
//	}
//	return set
//}


type Clause struct {
	literals []Literal
}

type Rule struct {
	debugging         []Token
	Head              Literal
	Literals          []Literal
	LiteralGenerator  []Generator
	Constraints       []Constraint
	OpenHead          bool      // if final token is tokenQuestionMark then it generates, otherwise tokenDot
	LogicalConnection tokenKind // Can be tokenImplication or tokenEquivalence or tokenRuleComma(normal rule)
	// GroundAtoms
	// OpenAtoms
}

func (r *Rule) hasHead() bool {
	return r.Head.Name != ""
}

// is a math expression that evaluates to true or false
// Constraints can contain variables
// supported are <,>,<=,>=,==
// E.g..: A*3v<=v5-2*R/7#mod3.
type Constraint struct {
	debugging   Tokens
	Neg         bool
	LeftTerm    TermExpression
	Comparision tokenKind
	RightTerm   TermExpression
}

func (constraint Constraint) Copy() (cons Constraint) {
	cons = constraint
	copy(cons.LeftTerm, constraint.LeftTerm)
	copy(cons.RightTerm, constraint.RightTerm)
	return cons
}

type Generator struct {
	debugging   Tokens
	constraints []Constraint
	generators  []Literal
	head        Literal
}

type Literal struct {
	debugging Tokens
	Neg       bool
	Name      AtomName
	Terms     []TermExpression
}

type TermExpression Tokens

type AtomName string

// Makes a deep copy and creates a new Literal
func (literal Literal) simplifyAtom(assignment map[string]int) (newLiteral Literal) {
	newLiteral = literal.Copy()
	//for i, term := range literal.Terms {
	for _, term := range literal.Terms {
		expr := assign(Tokens(term).String(), assignment)
		if groundMathExpression(expr) {
		//	r := evaluateExpression(expr)
		//	newLiteral.Terms[i] = TermExpression{Token(strconv.Itoa(r)),tokenTermExpression)}
		//} else {
		//	newLiteral.Terms[i] = TermExpression(expr)
		}
	}
	return
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

//assumption:Space only between literals.
func assign(termExpression string, assignment map[string]int) string {
	for Const, Val := range assignment {
		termExpression = strings.ReplaceAll(termExpression, Const, strconv.Itoa(Val))
	}
	return termExpression
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
