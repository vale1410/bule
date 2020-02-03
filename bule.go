package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/scylladb/go-set/strset"
	"github.com/vale1410/bule/parser"
	"os"
	"strconv"
	"strings"
	"unicode"
)

var (
	debugFlag = flag.Int("d", 0, "Debug Level .")
	progFlag  = flag.String("f", "", "Path to file.")
)

func debug(level int, s ...interface{}) {
	if level <= *debugFlag {
		fmt.Println(s...)
	}
}

func (p *Program) Debug() {
	fmt.Println("constants:", p.Constants)
	fmt.Println("domains:", p.Domains)
	fmt.Println("globalVars", p.GlobalVariables())
	for i, r := range p.Rules {
		fmt.Println("\nrule", i, r.debugString)
		if r.hasHead() {
			fmt.Println("head", r.Head)
		}
		fmt.Println("atoms", r.Atoms)
		fmt.Println("atomGenerators", r.AtomGenerators)
		fmt.Println("constraints", r.Constraints)
		fmt.Println()
	}
}

func (p *Program) RewriteEquivalences() {
	// Make rules from the head equivalences
	// Current assumption: head is only one literal, body is a conjunction!
	newRules := make([]Rule, 0)
	for _, r := range p.Rules {
		if r.hasHead() {
			// Check that freeVars are the same for right and left side (assumption)
			for i, atom := range r.Atoms {
				newRule := Rule{}
				newRule.Atoms = []Literal{r.Head.makeNeg(), atom.Copy()}
				newRule.Constraints = r.Constraints
				newRules = append(newRules, newRule)
				r.Atoms[i] = atom.makeNeg()
			}
			r.Atoms = append(r.Atoms, r.Head)
		}
		newRules = append(newRules, r)
	}
	p.Rules = newRules
}

func (p *Program) ExpandGenerators() {

	for i, r := range p.Rules {
		for _, atomG := range r.AtomGenerators {
			assignments := p.generateAssignments(
				atomG.variables,
				atomG.constraints)
			for _, assignment := range assignments {
				p.Rules[i].Atoms = append(p.Rules[i].Atoms,
					atomG.literal.simplifyAtom(assignment))
			}
		}
		if len(r.AtomGenerators) > 0 {
			debug(2, "after generation of atoms:")
			debug(2, "atoms", r.Atoms)
		}
	}
}

func (p *Program) Ground() (gRules []GroundRule,
	existQ map[int][]Literal,
	forallQ map[int][]Literal,
	maxIndex int) {

	gRules = make([]GroundRule, 0)
	existQ = make(map[int][]Literal)
	forallQ = make(map[int][]Literal)
	maxIndex = 0
	globals := p.GlobalVariables()

	for _, r := range p.Rules {

		debug(2, "freevariables", r.FreeVars())

		asserts(globals.IsSubset(r.FreeVars()), "There are free variables that are not bound by globals.")
		assignments := p.generateAssignments(r.FreeVars().List(), r.Constraints)

		debug(2, "Ground Rules:")
		for _, assignment := range assignments {
			gRule := GroundRule{}
			for _, atom := range r.Atoms {
				gRule.literals = append(gRule.literals, atom.simplifyAtom(assignment))
			}
			debug(2, "gRule", gRule)
			if len(gRule.literals) > 0 {
				if gRule.literals[0].Name == "#forall" {
					asserts(len(gRule.literals[0].Terms) == 1, "Wrong arity for forall")
					val, err := strconv.Atoi(string(gRule.literals[0].Terms[0]))
					asserte(err)
					forallQ[val] = append(forallQ[val], gRule.literals[1:]...)
					if val > maxIndex {
						maxIndex = val
					}
				} else if gRule.literals[0].Name == "#exist" {
					asserts(len(gRule.literals[0].Terms) == 1, "Wrong arity for exist")
					val, err := strconv.Atoi(string(gRule.literals[0].Terms[0]))
					asserte(err)
					existQ[val] = append(existQ[val], gRule.literals[1:]...)
					if val > maxIndex {
						maxIndex = val
					}
				} else {
					gRules = append(gRules, gRule)
				}
			}
		}
		debug(2)
	}
	return
}

func main() {

	flag.Parse()

	p := parseProgram()

	if true {
		return
	}

	debug(2, "\nExpand generators")
	p.ExpandGenerators()

	// forget about heads now!
	debug(2, "\nRewrite Equivalences")
	p.RewriteEquivalences()

	// There are no equivalences and no generators anymore !

	{
		debug(2, "Grounding:")
		gRules, existQ, forallQ, maxIndex := p.Ground()

		// Do Unit Propagation

		// Find variables that need to be put in the quantifier alternation

		for i := 0; i <= maxIndex; i++ {

			if atoms, ok := forallQ[i]; ok {
				fmt.Print("a")
				for _, a := range atoms {
					fmt.Print(" ", a)
				}
				fmt.Println()
			}
			if atoms, ok := existQ[i]; ok {
				fmt.Print("e")
				for _, a := range atoms {
					fmt.Print(" ", a)
				}
				fmt.Println()
			}
		}

		for _, r := range gRules {
			for _, a := range r.literals {
				fmt.Print(a, " ")
			}
			fmt.Println()
		}
	}
}

func (name AtomName) String() string {
	return name.String()
}

func (atomG AtomGenerator) String() string {
	return fmt.Sprintf("%v:constraints%v:vars%v",
		atomG.literal.String(),
		atomG.constraints,
		atomG.variables)
}

func (literal Literal) String() string {
	var s string
	if literal.Neg == false {
		s = "~"
	}
	s = s + literal.Name.String() + "["
	for i, x := range literal.Terms {
		s += string(x)
		if i < len(literal.Terms)-1 {
			s += ","
		}
	}
	return s + "]"
}

func (literal Literal) FreeVars()  *strset.Set {
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
	variables := strings.FieldsFunc(string(term), f)
	set := strset.New()
	for _, x := range variables {
		if !number(x) {
			set.Add(x)
		}
	}
	return set
}


func (rule *Rule) FreeVars() *strset.Set  {
	set := rule.Head.FreeVars()
	for _, a := range rule.Atoms {
		set.Merge(a.FreeVars())
	}
	return set
}


func (p *Program) generateAssignments(variables []string, constraints []Constraint) []map[string]int {

	allPossibleAssignments := make([]map[string]int, 1, 32)
	allPossibleAssignments[0] = make(map[string]int)

	for _, variable := range variables {
		if dom, ok := p.Domains[variable]; ok {
			newAssignments := make([]map[string]int, 0, len(allPossibleAssignments)*len(dom))
			for _, val := range dom {
				for _, assignment := range allPossibleAssignments {
					newAssignment := make(map[string]int)
					for key, value := range assignment {
						newAssignment[key] = value
					}
					newAssignment[variable] = val
					newAssignments = append(newAssignments, newAssignment)
				}
			}
			allPossibleAssignments = newAssignments
		} else {
			panic("variable doesnt have domain " + variable)
		}
	}

	assignments := make([]map[string]int, 0, 32)

	for _, assignment := range allPossibleAssignments {
		fmt.Println(assignment)
		fmt.Println(constraints)
		// check all constraints
		allConstraintsTrue := true
		for _, cons := range constraints {
			tmp := assign(cons.BoolExpr, assignment)
			tmp = assign(tmp, p.Constants)
			tmp = strings.ReplaceAll(string(tmp), "#mod", "%")
			asserts(groundBoolLogicalMathExpression(tmp), "Must be bool expression", tmp, "from", cons.BoolExpr)
			allConstraintsTrue = allConstraintsTrue && evaluateBoolExpression(tmp)
		}
		if allConstraintsTrue {
			assignments = append(assignments, assignment)
		}
	}
	return assignments
}

func parseProgram() (p Program) {
	// open a file or stream
	var scanner *bufio.Scanner
	file, err := os.Open(*progFlag)
	if err != nil {
		scanner = bufio.NewScanner(os.Stdin)
	} else {
		defer file.Close()
		scanner = bufio.NewScanner(file)
	}

	p.Domains = make(map[string][]int, 0)
	p.Constants = make(map[string]int)
	// Math operator replacement

	for scanner.Scan() {

		s := strings.TrimSpace(scanner.Text())
		if pos := strings.Index(s, "%"); pos >= 0 {
			s = s[:pos]
		}
		//s = strings.Trim(s, ".")
		s = strings.Replace(s, " ", "", -1)
		//s = strings.Replace(s, "].", "]", -1)
		//s = strings.Replace(s, "],", "] ", -1)
		//s = strings.Replace(s, ", ", " ", -1)
		//		debugString := s

		if s == "" || strings.HasPrefix(s, "%") {
			continue
		}


		// parsing a global definition like "
		// X = {4..5}.
		// or c = 5.
		// or k  = c*2.
		if strings.Contains(s, "=") &&
			!isConstraint(s) &&
			!strings.Contains(s, "<=>") {

			def := strings.Split(s, "=")
			asserts(len(def) == 2, s)
			if strings.Contains(def[1], "..") {
				set := strings.Trim(def[1], "{}.")
				interval := strings.Split(set, "..")
				i1, _ := strconv.Atoi(interval[0])
				x := assign(interval[1], p.Constants)
				i2 := evaluateExpression(x)
				p.Domains[def[0]] = makeSet(i1, i2)
			} else { // this is a constant
				term := assign(def[1], p.Constants)
				if !groundMathExpression(term) {
					panic("is not ground" + term)
				}
				p.Constants[def[0]] = evaluateExpression(term)
				p.Domains[def[0]] = []int{evaluateExpression(term)}
			}
			continue
		}


		{
			var head Literal
			if strings.Contains(s, "<=>") {
				ht := strings.Split(s, "<=>")
				asserts(len(ht) == 2, "Parsing equivalence wrong", s)
				head, _ = parseAtom(ht[0])
				s = ht[1]
			}

			ruleElements, err := parser.RuleElements(s)
			asserte(err)

			fmt.Println("RuleElements", ruleElements)

			var atoms []Literal
			var atomGenerators []AtomGenerator
			var ruleConstraints []Constraint

			for _, token := range ruleElements {

				if !strings.Contains(ruleElement, ":") {
					if isConstraint(ruleElement) {
						ruleConstraints = append(ruleConstraints, Constraint{ruleElement})
					} else {
						literal, _ := parseAtom(ruleElement)
						atoms = append(atoms, literal)

					}
					continue
				}

				atomG := AtomGenerator{}

				// This predicate has generators
				// Move from out to in and replace constants
				// and evaluates expressions
				xs := strings.Split(ruleElement, ":")

				for i := len(xs) - 1; i > 0; i-- {
					if _, ok := p.Domains[xs[i]]; ok {
						atomG.variables = append(atomG.variables, xs[i])
					} else { // is some kind of expression
						atomG.constraints = append(atomG.constraints, Constraint{xs[i]})
					}
				}

				atomG.literal, _ = parseAtom(xs[0])
				atomGenerators = append(atomGenerators, atomG)
			}
			p.Rules = append(p.Rules,
				Rule{debugString,
					head,
					atoms,
					atomGenerators,
					ruleConstraints})
		}
	}
	return
}

type Program struct {
	Rules     []Rule
	Domains   map[string][]int
	Constants map[string]int
}

func (p *Program) GlobalVariables() *strset.Set {
	set := strset.New()
	for k := range p.Domains {
		set.Add(k)
	}
	for k := range p.Constants {
		set.Add(k)
	}
	return set
}

type GroundRule struct {
	literals []Literal
}

type Rule struct {
	debugString    string
	Head           Literal
	Atoms          []Literal
	AtomGenerators []AtomGenerator
	Constraints    []Constraint
}

func (r *Rule) hasHead() bool {
	return r.Head.Name != ""
}

// is a math expression that evaluates to true or false
// Constraints can contain variables
// supported are <,>,<=,>=,==
// z.B.: A*3<=5-2*R/7#mod3.
type Constraint struct {
	BoolExpr string
}

type AtomGenerator struct {
	variables   []string
	constraints []Constraint
	literal     Literal
}

type Literal struct {
	debugString string
	Neg         bool
	Name        AtomName
	Terms       []TermExpression
}

type TermExpression string
type AtomName string

func (t TermExpression) String() string {
	return string(t)
}

// assuming it is not a constraint
// ~a4gDH[123,a*b,432-43#mod2]
func parseAtom(literalString string) (Literal, error) {
	// Check EBNF of Literal
	// Check for if has [] or not .
	asserts(strings.Contains(literalString, "["), "doesnt contain [", literalString)
	asserts(strings.Contains(literalString, "]"), "doesnt contain ]", literalString)
	name := literalString[:strings.Index(literalString, "[")]
	literalString = literalString[strings.Index(literalString, "[")+1:]
	par := literalString[:strings.LastIndex(literalString, "]")]
	ts := strings.Split(par, ",")
	terms := make([]TermExpression, len(ts))
	for i, expr := range ts {
		expr = strings.ReplaceAll(expr, "#mod", "%")
		terms[i] = TermExpression(expr)
	}
	n := true
	if strings.HasPrefix(name, "~") {
		name = strings.TrimLeft(name, "~")
		n = false
	}
	return Literal{literalString, n, AtomName(name), terms}, nil
}

// Makes a deep copy
func (literal Literal) simplifyAtom(assignment map[string]int) (newAtom Literal) {
	newAtom = literal.Copy()
	for i, term := range literal.Terms {
		expr := assign(string(term), assignment)
		if groundMathExpression(expr) {
			r := evaluateExpression(expr)
			newAtom.Terms[i] = TermExpression(strconv.Itoa(r))
		} else {
			newAtom.Terms[i] = TermExpression(expr)
		}
	}
	return
}

func isConstraint(s string) bool {
	return (strings.Contains(s, "==") ||
		strings.Contains(s, "!=") ||
		strings.Contains(s, "<=") ||
		strings.Contains(s, ">=") ||
		strings.Contains(s, ">") ||
		strings.Contains(s, "<")) &&
		!strings.Contains(s, "<=>") &&
		!strings.Contains(s, "=>")
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

func (literal *Literal) makeNeg() Literal {
	a := literal.Copy()
	a.Neg = !a.Neg
	if strings.HasPrefix(a.debugString, "~") {
		a.debugString = strings.TrimLeft(a.debugString, "~")
	} else {
		a.debugString = "~" + a.debugString
	}
	return a
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
