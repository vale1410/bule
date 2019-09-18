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

	p := parseProgramExpandGenerators()
	p.expandQuantifiers()

	p.Debug()

	for _, q := range p.Quantifiers {
		if q.Exist {
			fmt.Print("e ")
		} else {
			fmt.Print("a ")
		}
		for _, atom := range q.Atoms {
			fmt.Print(atom.s, " ")
		}
		fmt.Println()
	}

	for _, r := range p.Rules {
				//fmt.Println(r)
		for _, s := range ground(r.s, p.Domains, p.Constants) {
			fmt.Println(s)
		}
	}
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

func NewAtom(pred string, terms []string) Atom {
	s := pred+"("
	for i, x := range terms {
		s += x
		if i < len(terms)-1 {
			s += ","
		}
	}
	return Atom{s + ")"}
}

// instantiates atom by replacing variable with values and creates a new copy of atom
// move(X,Y,4) and Y->3 -> move(X,3,4)
// move(X,Y+3,4) and Y->3 -> move(X,6,4)
// Also evaluates math expressions
// If variable does not exist in move, then just a new copy is created.
func (a Atom) instantiate(variable string, val int) (newA Atom) {
	var newTerms []string
	pred, terms := decomposeAtom(a)
	for _, term := range terms {
		tmp := strings.ReplaceAll(term, variable, strconv.Itoa(val))
		if mathExpression(tmp) {
			tmp = strconv.Itoa(evaluateExpression(tmp))
		}
		newTerms = append(newTerms,tmp)
	}
	return NewAtom(pred,newTerms)
}

func (p *Program) expandQuantifiers() {
	//TODO restriction, only ONE free variable allowed
	freeVar := ""
	var quantifierWithFree []int
	var quantifierNoFree []int
	for i, r := range p.Quantifiers {
		containsFree := false
		for _, atom := range r.Atoms {
			_, expressions := decomposeAtom(atom)
			for _, expr := range expressions {
				globals := p.containsGlobals(expr)
				//fmt.Println(expr, "globals", globals)
				if len(globals) == 0 {
					asserts(number(expr), "must be a number.")
				} else if len(globals) == 1 {
					// remember the unique free variable
					if freeVar == "" {
						freeVar = globals[0]
					} else if freeVar != globals[0] {
						fmt.Println("We only allow one free variable in quantifiers. This var: ", expr, " free", freeVar)
						panic("Too many free Variable")
					} else {
						containsFree = true
					}
				} else {
					fmt.Println("This predicate contains too many free variables", atom.s, "in term",expr)
					panic("")
				}
			}
		}
		if containsFree {
			quantifierWithFree = append(quantifierWithFree, i)
		} else {
			quantifierNoFree = append(quantifierNoFree, i)
		}
	}

	if freeVar == "" {
		return
	}

	var quantifiers []Quantifier
	if Dom, ok := p.Domains[freeVar]; ok {
		for _, Val := range Dom {
			for _, i := range quantifierWithFree {
				quantifier := p.Quantifiers[i]
				quantifier.Atoms = []Atom{}
				for _, atom := range p.Quantifiers[i].Atoms {
					quantifier.Atoms = append(quantifier.Atoms, atom.instantiate(freeVar,Val))
				}
				quantifiers = append(quantifiers, quantifier)
			}
		}
	} else { // is some kind of expression
		asserts(false, "Wrong free variables in quantifier.")
	}
	// All the remaining predicates go innermost
	for _, i := range quantifierNoFree {
		quantifiers = append(quantifiers, p.Quantifiers[i])
	}
	p.Quantifiers = quantifiers
}

func parseProgramExpandGenerators() (p Program) {
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
				x := replaceConstants(interval[1], p.Constants)
				i2 := evaluateExpression(x)
				p.Domains[def[0]] = makeSet(i1, i2)
			} else { // this is a constant
				term := replaceConstants(def[1],p.Constants)
			 	if !mathExpression(term) {
			 		panic("is not ground" + term)
				}
				p.Constants[def[0]] =evaluateExpression(term)
			}
			continue
		}

		{ // A usual clause expression.
			literals := strings.Fields(s)
			quant := literals[0]
			isQuantifier := false
			if quant == "a" || quant == "e" {
				literals = literals[1:]
				isQuantifier = true
			}

			var clause []Atom
			for _, literal := range literals {

				if !strings.Contains(literal, ":") {
					a := simplifyAtom(Atom{literal}, p.Constants)
					clause = append(clause,Atom{a})
					continue
				}

				// This predicate has generators
				// Move from out to in and replace constants
				// and evaluates expressions
				xs := strings.Split(literal, ":")
				atom := Atom{xs[0]}

				var variables []string
				var constraints []string

				//fmt.Println("xs",xs)
				for i := len(xs) - 1; i > 0; i-- {
					if _, ok := p.Domains[xs[i]]; ok {
						variables = append(variables, xs[i])
					} else { // is some kind of expression
						constraints = append(constraints,xs[i])
					}
				}
				//fmt.Println("constraints",constraints)
				//fmt.Println("variables",variables)

				assignments := make([]map[string]int, 1,32)
				assignments[0] = make(map[string]int)
				for _, variable := range variables {
					if dom, ok := p.Domains[variable]; ok {
						newAssignments := make([]map[string]int,0,len(assignments) * len(dom))
						for _, val := range dom {
							for _,assignment := range assignments {
								newAssignment := make(map[string]int)
								for key, value := range assignment {
									newAssignment[key] = value
								}
								newAssignment[variable] = val
								newAssignments = append(newAssignments, newAssignment)
							}
						}
						assignments = newAssignments
					} else {
						panic("variable doesnt have domain " + variable)
					}
				}
				for _, assignment := range assignments {
					//fmt.Println(assignment)
					// check all constraints
					allConstraintsTrue := true
					for _,cons := range constraints {
						tmp := replaceConstants(cons,assignment)
						//fmt.Println(cons,tmp,evaluateBoolExpression(tmp))
						tmp = strings.ReplaceAll(tmp, "#mod", "%")
						asserts(boolMathExpression(tmp),"Must be bool expression " + tmp)
						allConstraintsTrue = allConstraintsTrue && evaluateBoolExpression(tmp)
					}
					if allConstraintsTrue {
						a := simplifyAtom(atom,assignment)
						//fmt.Println(a)
						clause = append(clause,Atom{a})
					}
				}
			}
			//fmt.Println("clause finished", clause)

			sClause := ""
			for _, atom := range clause {
				sClause += atom.s + " "
			}

			if isQuantifier {
				var q Quantifier
				q.s = quant + " " + sClause
				q.Exist = quant == "e"
				q.Atoms = clause
				p.Quantifiers = append(p.Quantifiers, q)
			} else {
				p.Rules = append(p.Rules, Rule{sClause, clause})
			}
		}
	}
	return
}

type Program struct {
	Quantifiers []Quantifier
	Rules       []Rule
	Domains     map[string][]int
	Constants   map[string]int
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
	s     string
	Atoms []Atom
}

type Quantifier struct {
	s     string
	Exist bool
	Atoms []Atom
}

type Atom struct {
	s string
	//	Name  string
	//	Neg   bool
	//	Arity int
	//	Terms []Term
}

type Term struct {
	s   string
	Val int
}

func number(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func mathExpression(s string) bool {
	//	r, _ := regexp.MatchString("[0-9+*/%]+", s)
	return "" == strings.Trim(s, "0123456789+*%-/()")
}

func boolMathExpression(s string) bool {
	//	r, _ := regexp.MatchString("[0-9+*/%!=><]+", s)
	return "" == strings.Trim(s, "0123456789+*%-=><")
}

//assumption:Space only between literals.
func replaceConstants(term string, constants map[string]int) string {
	for Const, Val := range constants {
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


func ground(s string, domain map[string][]int, constants map[string]int) []string {

	cls := []string{s}
	for Var := range domain {
		if strings.Count(s, Var) > 0 {
			newcls := []string{}
			for _, cl := range cls {
				for _, Val := range domain[Var] {
					cl3 := strings.ReplaceAll(cl, Var, strconv.Itoa(Val))
					newcls = append(newcls, cl3)
				}
			}
			cls = newcls
		}
	}

	rcls := make([]string, len(cls))
	for i, cl := range cls {
		literals := strings.Fields(cl)
		newcl := ""
		for _, literal := range literals {
			newcl += simplifyAtom(Atom{literal}, constants)
		}
		rcls[i] = newcl
	}

	return rcls
}

//  move(X,Y,T+1) ->  move, [X,Y,T+1]
func decomposeAtom(atom Atom) (name string, terms []string) {
	literalString := atom.s
	name = literalString[:strings.Index(literalString, "(")]
	literalString = literalString[strings.Index(literalString, "(")+1:]
	par := literalString[:strings.LastIndex(literalString, ")")]
	terms = strings.Split(par, ",")
	return
}

// move(X,5+1,c,k+1) -> move(X,6,10,11). % c=10 and k=11 are constants
func simplifyAtom(atom Atom, constants map[string]int) (newcl string) {
	pre, terms := decomposeAtom(atom)
	newcl = pre + "("
	for i, expr := range terms {
		expr = replaceConstants(expr, constants)
		expr = strings.ReplaceAll(expr, "#mod", "%")
		if mathExpression(expr) {
			r := evaluateExpression(expr)
			newcl += strconv.Itoa(r)
		} else {
			newcl += expr
		}
		if i == len(terms)-1 {
			newcl += ") "
		} else {
			newcl += ","
		}
	}
	return
}

func main2() {

	c := flag.Int("c", 3, "columns of board.")
	r := flag.Int("r", 3, "rows of board.")
	q := flag.Int("q", 3, "connect-Z.")
	prog := flag.String("f", "connect.lp", "Encoding.")
	flag.Parse()

	maxT := *r * *c
	if maxT%2 == 0 {
		maxT--
	}

	//domain := make(map[string][]int)
	//constants := make(map[string]int)

	// open a file or stream
	var scanner *bufio.Scanner
	file, err := os.Open(*prog)
	if err != nil {
		scanner = bufio.NewScanner(os.Stdin)
	} else {
		defer file.Close()
		scanner = bufio.NewScanner(file)
	}

	cls := []string{}

	for scanner.Scan() {

		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "%") {
			continue
		}
		fmt.Println(line)

		//	cls = append(cls, fmt.Sprintf("\nc %s\n", line))

		//ground(line, domain, constants, &cls)
	}

	W := genPositions(*c, *r, *q)

	if false {
		fmt.Print("e ")
		for T := 0; T < maxT; T++ {

			if T > 0 {
				if T%2 == 0 { // white moves, UNIVERSAL
					fmt.Println("")
					fmt.Print("a ")
					for i := 1; i < *c; i++ {
						fmt.Printf("move(%v,%v) ", i, T)
					}
					fmt.Println("")
					fmt.Print("e ")
				} else { // back moves
					for i := 1; i < *c; i++ {
						fmt.Printf("move(%v,%v) ", i, T)
					}
				}
				fmt.Printf("move(%v,%v) ", *c, T)

				for i := 1; i <= *c; i++ {
					fmt.Printf("first(%v,%v) ", i, T)
				}
			}

			for i := 1; i <= *c; i++ {
				for j := 0; j <= *r; j++ {
					fmt.Printf("count(%v,%v,%v) ", i, j, T)
				}
			}

			for i := 0; i <= *c; i++ {
				fmt.Printf("step(%v,%v) ", i, T)
			}

		}

		fmt.Print("\ne ") // NOT NEEDED BECAUSE ALREADY SPECIFIED.
		for i := 1; i <= *c; i++ {
			for j := 1; j <= *r; j++ {
				fmt.Printf("board(%v,%v,%v) ", i, j, 0)
				fmt.Printf("board(%v,%v,%v) ", i, j, 1)
			}
		}
		for i := 0; i < len(W); i++ {
			fmt.Printf("win(%v) ", i)
		}
	}
	fmt.Println()

	if false { // ground all winning positions

		{
			cls = append(cls, "c ~board(I,J,1):(I,J) in W.")
			for _, V := range W {
				s := ""
				for _, R := range V {
					s += fmt.Sprintf("~board(%v,%v,0) ", R[0], R[1])
				}
				cls = append(cls, s)
			}
		}

		{
			cls = append(cls, "c win(W) : W.")
			s := ""
			for i, _ := range W {
				s += fmt.Sprintf("win(%v) ", i)
			}
			cls = append(cls, s)
		}
		{
			cls = append(cls, "c ~win(W), board(I,J,0):(I,J) in W.")
			for i, V := range W {
				s := fmt.Sprintf("~win(%v) ", i)
				for _, R := range V {
					s += fmt.Sprintf("board(%v,%v,1) ", R[0], R[1])
				}
				cls = append(cls, s)
			}
		}
	}

	for _, cl := range cls {
		fmt.Println(cl)
	}

}

func neg(s string) string {
	if strings.HasPrefix(s, "~") {
		return strings.TrimLeft(s, "~")
	}
	return "~" + s
}

func genPositions(c, r, q int) [][][]int {
	W := make([][][]int, 0, 3*r*c)
	for i := 1; i <= c; i++ {
		for j := 1; j <= r-q+1; j++ {
			row := make([][]int, q)
			for d := 0; d < q; d++ {
				row[d] = []int{i, j + d}
			}
			W = append(W, row)
		}
	}
	for i := 1; i <= c-q+1; i++ {
		for j := 1; j <= r; j++ {
			row := make([][]int, q)
			for d := 0; d < q; d++ {
				row[d] = []int{i + d, j}
			}
			W = append(W, row)
		}
	}
	for i := 1; i <= c-q+1; i++ {
		for j := 1; j <= r-q+1; j++ {
			{
				row := make([][]int, q)
				for d := 0; d < q; d++ {
					row[d] = []int{i + d, j + d}
				}
				W = append(W, row)
			}
			{
				row := make([][]int, q)
				for d := 0; d < q; d++ {
					row[d] = []int{i + d, j + q - d - 1}
				}
				W = append(W, row)
			}
		}
	}
	return W
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
