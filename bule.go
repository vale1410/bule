package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/Knetic/govaluate"
	//	"github.com/jinzhu/copier"
	"os"
	//	"regexp"
	"strconv"
	"strings"
)

var (
	progFlag = flag.String("f", "", "Path to file.")
)

func main() {

	flag.Parse()

	p := parseProgram()
	for _, r := range p.Rules {
		fmt.Println("rule:",r.debugString)
		fmt.Println("atoms", r.Atoms)
		fmt.Println("atomGenerators", r.AtomGenerators)
		fmt.Println("constraints", r.Constraints)
		for _,atomG := range r.AtomGenerators {
			assignments := p.generateAssignments(atomG.variables, atomG.constraints)
			for _,assignment := range assignments {
				r.Atoms = append (r.Atoms, atomG.atom.simplifyAtom(assignment))
			}
		}
		fmt.Println("after generation of atoms:")
		fmt.Println("atoms", r.Atoms)
		fmt.Println()
	}


//	p.expandQuantifiers()
//
//	p.Debug()
//
//	for _, q := range p.Quantifiers {
//		if q.Exist {
//			fmt.Print("e ")
//		} else {
//			fmt.Print("a ")
//		}
//		for _, atom := range q.Atoms {
//			fmt.Print(atom, " ")
//		}
//		fmt.Println()
//	}

//	for _, r := range p.Rules {
//		//fmt.Println(r)
//		for _, s := range ground(r.debugString, p.Domains, p.Constants) {
//			fmt.Println(s)
//		}
//	}
}

func (p *Program) containsGlobals(term string) (globals []string) {

	globals = make([]string, 0)
	for _, g := range p.GlobalVariables() {
		if strings.Contains(term, g) {
			globals = append(globals, g)
		}
	}
	return
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
			//fmt.Println(cons,tmp,evaluateBoolExpression(tmp))
			tmp = strings.ReplaceAll(tmp, "#mod", "%")
			asserts(groundBoolLogicalMathExpression(tmp), "Must be bool expression "+tmp)
			allConstraintsTrue = allConstraintsTrue && evaluateBoolExpression(tmp)
		}
		if allConstraintsTrue {
			assignments = append(assignments, assignment)
//			a := simplifyAtom(atom, assignment)
//			//fmt.Println(a)
//			atoms = append(atoms, Atom{a})
		}
	}
	return assignments
}

//func (p *Program) expandQuantifiers() {
//	//TODO restriction, only ONE free variable allowed
//	freeVar := ""
//	var quantifierWithFree []int
//	var quantifierNoFree []int
//	for i, r := range p.Quantifiers {
//		containsFree := false
//		for _, atom := range r.Atoms {
//			_, expressions := decomposeAtom(atom)
//			for _, expr := range expressions {
//				globals := p.containsGlobals(expr)
//				//fmt.Println(expr, "globals", globals)
//				if len(globals) == 0 {
//					asserts(number(expr), "must be a number.")
//				} else if len(globals) == 1 {
//					// remember the unique free variable
//					if freeVar == "" {
//						freeVar = globals[0]
//					} else if freeVar != globals[0] {
//						fmt.Println("We only allow one free variable in quantifiers. This var: ", expr, " free", freeVar)
//						panic("Too many free Variable")
//					} else {
//						containsFree = true
//					}
//				} else {
//					fmt.Println("This predicate contains too many free variables", atom.s, "in term", expr)
//					panic("")
//				}
//			}
//		}
//		if containsFree {
//			quantifierWithFree = append(quantifierWithFree, i)
//		} else {
//			quantifierNoFree = append(quantifierNoFree, i)
//		}
//	}
//
//	if freeVar == "" {
//		return
//	}
//
//	var quantifiers []Quantifier
//	if Dom, ok := p.Domains[freeVar]; ok {
//		for _, Val := range Dom {
//			for _, i := range quantifierWithFree {
//				quantifier := p.Quantifiers[i]
//				quantifier.Atoms = []Atom{}
//				for _, atom := range p.Quantifiers[i].Atoms {
//					quantifier.Atoms = append(quantifier.Atoms, atom.instantiate(freeVar, Val))
//				}
//				quantifiers = append(quantifiers, quantifier)
//			}
//		}
//	} else { // is some kind of expression
//		asserts(false, "Wrong free variables in quantifier.")
//	}
//	// All the remaining predicates go innermost
//	for _, i := range quantifierNoFree {
//		quantifiers = append(quantifiers, p.Quantifiers[i])
//	}
//	p.Quantifiers = quantifiers
//}

func parse() (p Program) {
	return
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
		if strings.Contains(s, "=") && !strings.Contains(s, "==") {
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
			}
			continue
		}

		{
			ruleElements := strings.Fields(s)

			var atoms []Atom
			var atomGenerators []AtomGenerator
			var ruleConstraints []Constraint

			for _, ruleElement := range ruleElements {

				if !strings.Contains(ruleElement, ":") {
					if constraint, ok := parseConstraint(ruleElement); ok {
						ruleConstraints = append(ruleConstraints, constraint)

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

				atomG.atom,_ = parseAtom(xs[0])
				atomGenerators = append(atomGenerators, atomG)
			}
			p.Rules = append(p.Rules,
				Rule{s,
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

func (p *Program) GlobalVariables() []string {
	keys := make([]string, 0, len(p.Domains))
	for k := range p.Domains {
		keys = append(keys, k)
	}
	return keys
}

func (p *Program) Debug() {

	fmt.Println("Constants")
	for k, v := range p.Constants {
		fmt.Println(k, "=", v)
	}
	fmt.Println("Domains")
	for k, v := range p.Domains {
		fmt.Println(k, "in", v)
	}
	//	fmt.Println("Quantifiers")
	//	for i, q := range p.Quantifiers {
	//		fmt.Println("s \t", q.s)
	//		fmt.Println(i, "\t", q.Atoms)
	//	}
	//	fmt.Println("Rules")
	//	for i, r := range p.Rules {
	//		fmt.Println("s \t", r.s)
	//		fmt.Println(i, "\t", r.Atoms)
	//	}
}

type Rule struct {
	debugString    string
	Atoms          []Atom
	AtomGenerators []AtomGenerator
	Constraints    []Constraint
}

// is a math expression that evaluates to true or false
// Constraints can contain variables
// supported are <,>,<=,>=,==
// z.B.: A*3<=5-2*R/7#mode3.
type Constraint struct {
	BoolExpr string
}

type Quantifier struct {
	s     string
	Exist bool
	Atoms []Atom
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
		expr = strings.ReplaceAll(expr, "#mod", "%")
		if groundMathExpression(expr) {
			r := evaluateExpression(expr)
			newAtom.Terms[i] = Term(strconv.Itoa(r))
		} else {
			newAtom.Terms[i] = Term(expr)
		}
	}
	return
}

// instantiates atom by replacing variable with values and creates a new copy of atom
// move(X,Y,4) and Y->3 -> move(X,3,4)
// move(X,Y+3,4) and Y->3 -> move(X,6,4)
// Also evaluates math expressions
// If variable does not exist in move, then just a new copy is created.
func (a Atom) instantiate(variable string, val int) Atom {
	b := a.Copy()
	for i, term := range a.Terms {
		tmp := strings.ReplaceAll(string(term), variable, strconv.Itoa(val))
		if groundMathExpression(tmp) {
			tmp = strconv.Itoa(evaluateExpression(tmp))
		}
		b.Terms[i] = Term(tmp)
	}
	return b
}

func parseConstraint(s string) (Constraint, bool) {
	if (strings.Contains(s, "==") ||
		strings.Contains(s, "<=") ||
		strings.Contains(s, ">=") ||
		strings.Contains(s, ">") ||
		strings.Contains(s, ">")) &&
		!strings.Contains(s, "<=>") &&
		!strings.Contains(s, "=>") {
		return Constraint{s}, true
	}
	return Constraint{}, false
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
	return "" == strings.Trim(s, "0123456789+*%-=><")
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

//type Term string
//type Variable string

//func ground(s string, domain map[string][]int, constants map[string]int) []string {
//
//	cls := []string{s}
//	for Var := range domain {
//		if strings.Count(s, Var) > 0 {
//			newcls := []string{}
//			for _, cl := range cls {
//				for _, Val := range domain[Var] {
//					cl3 := strings.ReplaceAll(cl, Var, strconv.Itoa(Val))
//					newcls = append(newcls, cl3)
//				}
//			}
//			cls = newcls
//		}
//	}
//
//	rcls := make([]string, len(cls))
//	for i, cl := range cls {
//		literals := strings.Fields(cl)
//		newcl := ""
//		for _, literal := range literals {
//			newcl += simplifyAtom(Atom{literal}, constants)
//		}
//		rcls[i] = newcl
//	}
//
//	return rcls
//}

func neg(s string) string {
	if strings.HasPrefix(s, "~") {
		return strings.TrimLeft(s, "~")
	}
	return "~" + s
}

func assert(condition bool) {
	if !condition {
		panic(errors.New(""))
	}
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
