package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/scylladb/go-set/strset"
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
	for _, r := range p.Rules {
		fmt.Println("rule:", r.debugString)
		fmt.Println("head", r.Head)
		fmt.Println("atoms", r.Atoms)
		fmt.Println("atomGenerators", r.AtomGenerators)
		fmt.Println("constraints", r.Constraints)
		fmt.Println("constraints", r.Constraints)
		fmt.Println()
	}
}

func (p *Program) RewriteEquivalences() {
	// Make rules from the head equivalences
	// Current assumption: head is only one atom, body is a conjunction!
	newRules := make([]Rule, 0)
	for _, r := range p.Rules {
		if r.hasHead() {
			// Check that freeVars are the same for right and left side (assumption)
			for i, atom := range r.Atoms {
				newRule := Rule{}
				newRule.Atoms = []Atom{r.Head.makeNeg(), atom.Copy()}
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
	// Expand the generators
	for _, r := range p.Rules {
		for _, atomG := range r.AtomGenerators {
			assignments := p.generateAssignments(atomG.variables, atomG.constraints)
			for _, assignment := range assignments {
				r.Atoms = append(r.Atoms, atomG.atom.simplifyAtom(assignment))
			}
		}
		if len(r.AtomGenerators) > 0 {
			debug(2, "after generation of atoms:")
			debug(2, "atoms", r.Atoms)
		}
	}
}

func (p *Program) Ground() (gRules []GroundRule,
	existQ map[int][]Atom,
	forallQ map[int][]Atom,
	maxIndex int) {

	gRules = make([]GroundRule, 0)
	existQ = make(map[int][]Atom)
	forallQ = make(map[int][]Atom)
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
				gRule.Atoms = append(gRule.Atoms, atom.simplifyAtom(assignment))
			}
			debug(2, "gRule", gRule)
			if len(gRule.Atoms) > 0 {
				if gRule.Atoms[0].Name == "#forall" {
					asserts(len(gRule.Atoms[0].Terms) == 1, "Wrong arity for forall")
					val, err := strconv.Atoi(string(gRule.Atoms[0].Terms[0]))
					asserte(err)
					forallQ[val] = append(forallQ[val], gRule.Atoms[1:]...)
					if val > maxIndex {
						maxIndex = val
					}
				} else if gRule.Atoms[0].Name == "#exist" {
					asserts(len(gRule.Atoms[0].Terms) == 1, "Wrong arity for exist")
					val, err := strconv.Atoi(string(gRule.Atoms[0].Terms[0]))
					asserte(err)
					existQ[val] = append(existQ[val], gRule.Atoms[1:]...)
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

	// forget about Generators now!
	p.ExpandGenerators()

	// forget about heads now!
	p.RewriteEquivalences()

	// we only work with Atoms now !

	{
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
			for _, a := range r.Atoms {
				fmt.Print(a, " ")
			}
			fmt.Println()
		}
	}
}

func (atomG AtomGenerator) String() string {
	return fmt.Sprintf("%v:constraints%v:vars%v",
		atomG.atom.String(),
		atomG.constraints,
		atomG.variables)
}

func (atom Atom) String() string {
	var s string
	if atom.Neg == false {
		s = "~"
	}
	s = s + atom.Name + "("
	for i, x := range atom.Terms {
		s += string(x)
		if i < len(atom.Terms)-1 {
			s += ","
		}
	}
	return s + ")"
}

func (term Term) FreeVars() *strset.Set {
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

func (atom Atom) FreeVars() *strset.Set {
	set := strset.New()
	for _, t := range atom.Terms {
		set.Merge(t.FreeVars())
	}
	return set
}

func (rule *Rule) FreeVars() *strset.Set {
	set := strset.New()
	for _, a := range rule.Atoms {
		set.Merge(a.FreeVars())
	}
	set.Merge(rule.Head.FreeVars())
	return set
}

func (atom Atom) Copy() Atom {
	t := make([]Term, len(atom.Terms))
	copy(t, atom.Terms)
	atom.Terms = t
	return atom
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
		//fmt.Println(assignment)
		// check all constraints
		allConstraintsTrue := true
		for _, cons := range constraints {
			tmp := assign(cons.BoolExpr, assignment)
			tmp = assign(tmp, p.Constants)
			tmp = strings.ReplaceAll(tmp, "#mod", "%")
			asserts(groundBoolLogicalMathExpression(tmp), "Must be bool expression "+tmp)
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
		s = strings.Trim(s, ".")
		//s = strings.Replace(s, " ", "", -1)
		s = strings.Replace(s, ").", ")", -1)
		s = strings.Replace(s, "),", ") ", -1)
		s = strings.Replace(s, ", ", " ", -1)

		if s == "" || strings.HasPrefix(s, "%") {
			continue
		}

		// parsing a global definition like " X = {4..5}.
		// or c = 5. or k  = c*2.
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
			var head Atom
			if strings.Contains(s, "<=>") {
				ht := strings.Split(s, "<=>")
				asserts(len(ht) == 2, "Parsing equivalence wrong")
				s = ht[1]
				head, _ = parseAtom(ht[0])
			}
			ruleElements := strings.Fields(s)

			var atoms []Atom
			var atomGenerators []AtomGenerator
			var ruleConstraints []Constraint

			for _, ruleElement := range ruleElements {

				if !strings.Contains(ruleElement, ":") {
					if isConstraint(ruleElement) {
						ruleConstraints = append(ruleConstraints, Constraint{ruleElement})
					} else {
						atom, _ := parseAtom(ruleElement)
						atoms = append(atoms, atom)

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

				atomG.atom, _ = parseAtom(xs[0])
				atomGenerators = append(atomGenerators, atomG)
			}
			p.Rules = append(p.Rules,
				Rule{s,
					head,
					atoms,
					atomGenerators,
					ruleConstraints})
		}
	}
	return
}

type Program struct {
	//	Quantifiers []Quantifier
	Rules     []Rule
	Domains   map[string][]int
	Constants map[string]int
}

func (p *Program) GlobalVariables() *strset.Set {
	set := strset.New()
	for k := range p.Domains {
		set.Add(k)
	}
	for k, _ := range p.Constants {
		set.Add(k)
	}
	return set
}

type GroundRule struct {
	Atoms []Atom
}

type Rule struct {
	debugString    string
	Head           Atom
	Atoms          []Atom
	AtomGenerators []AtomGenerator
	Constraints    []Constraint
}

func (r *Rule) hasHead() bool {
	return r.Head.Name != ""
}

// is a math expression that evaluates to true or false
// Constraints can contain variables
// supported are <,>,<=,>=,==
// z.B.: A*3<=5-2*R/7#mode3.
type Constraint struct {
	BoolExpr string
}

type AtomGenerator struct {
	debugString string
	variables   []string
	constraints []Constraint
	atom        Atom
}

type Atom struct {
	debugString string
	Neg         bool
	Name        string
	Terms       []Term
}

type Term string

// assuming it is not a constraint
func parseAtom(literalString string) (Atom, bool) {
	name := literalString[:strings.Index(literalString, "(")]
	literalString = literalString[strings.Index(literalString, "(")+1:]
	par := literalString[:strings.LastIndex(literalString, ")")]
	ts := strings.Split(par, ",")
	terms := make([]Term, len(ts))
	for i, expr := range ts {
		expr = strings.ReplaceAll(expr, "#mod", "%")
		terms[i] = Term(expr)
	}
	n := true
	if strings.HasPrefix(name, "~") {
		name = strings.TrimLeft(name, "~")
		n = false
	}
	return Atom{literalString, n, name, terms}, true
}

// Makes a deep copy
func (atom Atom) simplifyAtom(assignment map[string]int) (newAtom Atom) {
	newAtom = atom.Copy()
	for i, term := range atom.Terms {
		expr := assign(string(term), assignment)
		if groundMathExpression(expr) {
			r := evaluateExpression(expr)
			newAtom.Terms[i] = Term(strconv.Itoa(r))
		} else {
			newAtom.Terms[i] = Term(expr)
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
	return "" == strings.Trim(s, "0123456789+*%-/()")
}

func groundBoolLogicalMathExpression(s string) bool {
	//	r, _ := regexp.MatchString("[0-9+*/%!=><]+", s)
	return "" == strings.Trim(s, "0123456789+*%-=><()")
}

//assumption:Space only between literals.
func assign(term string, assignment map[string]int) string {
	for Const, Val := range assignment {
		term = strings.ReplaceAll(term, Const, strconv.Itoa(Val))
	}
	return term
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
func evaluateExpression(term string) int {
	term = strings.ReplaceAll(term, "#mod", "%")
	expression, err := govaluate.NewEvaluableExpression(term)
	assertx(err, term)
	result, err := expression.Evaluate(nil)
	assertx(err, term)
	return int(result.(float64))
}

func (atom *Atom) makeNeg() Atom {
	a := atom.Copy()
	a.Neg = !a.Neg
	if strings.HasPrefix(a.debugString, "~") {
		a.debugString = strings.TrimLeft(a.debugString, "~")
	} else {
		a.debugString = "~" + a.debugString
	}
	return a
}

func asserts(condition bool, info string) {
	if !condition {
		fmt.Println(info)
		panic(errors.New(info))
	}
}

func asserte(err error) {
	if err != nil {
		panic(err)
	}
}

func assertx(err error, info string) {
	if err != nil {
		fmt.Println(info)
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
