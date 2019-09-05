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
		//		fmt.Println(r)
		for _, s := range ground(r.s, p.Domains, p.Constants) {
			fmt.Println(s)
		}
	}
}

func (p *Program) expandQuantifiers() {
	//TODO restriction, only ONE free variable allowed
	freeVar := ""
	quantifierWithFree := []int{}
	quantifierNoFree := []int{}
	for i, r := range p.Quantifiers {
		containsFree := false
		for _, literal := range r.Atoms {
			_, Vars := extract(literal.s)
			for _, Var := range strings.Split(Vars, ",") {
				if number(Var) {
					continue
				} else if _, ok := p.Domains[Var]; ok {
					if freeVar == "" {
						freeVar = Var
					} else if freeVar != Var {
						fmt.Println("We only allow one free variable in quantifiers. This var: ", Var, " free", freeVar)
					} else {
						containsFree = true
					}
				} else {
					fmt.Println("This is not a variable, What is this", Var)
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

	quantifiers := []Quantifier{}
	if Dom, ok := p.Domains[freeVar]; ok {
		for _, Val := range Dom {
			for _, i := range quantifierWithFree {
				quantifier := p.Quantifiers[i]
				quantifier.Atoms = []Atom{}
				for _, atom := range p.Quantifiers[i].Atoms {
					var natom Atom // TODO DEEP COPY
					natom.s = strings.ReplaceAll(atom.s, freeVar, strconv.Itoa(Val))
					quantifier.Atoms = append(quantifier.Atoms, natom)
				}
				quantifiers = append(quantifiers, quantifier)
			}
		}
	} else { // is some kind of expression
		asserts(false, "Wrong free variables in quantifier.")
	}

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
		s = strings.Replace(s, ").", ")", -1)
		s = strings.Replace(s, "),", ") ", -1)

		if s == "" || strings.HasPrefix(s, "%") {
			continue
		}

		// s is a global definition
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
			} else {
				i, _ := strconv.Atoi(strings.Trim(def[1], "."))
				p.Constants[def[0]] = i
			}
			continue
		}

		{
			literals := strings.Fields(s)
			quant := literals[0]
			isQuantifier := false
			if quant == "a" || quant == "e" {
				literals = literals[1:]
				isQuantifier = true
			}

			ss := ""
			for _, literal := range literals {

				if !strings.Contains(literal, ":") {
					ss += literal + " "
					continue
				}
				xs := strings.Split(literal, ":")
				lits := xs[0]
				for i := len(xs) - 1; i > 0; i-- {
					newlits := ""
					if Dom, ok := p.Domains[xs[i]]; ok {
						for _, Val := range Dom {
							l := strings.ReplaceAll(lits, xs[i], strconv.Itoa(Val))
							newlits += l + " "
						}
					} else { // is some kind of expression
						fmt.Println(literal)
						asserts(false, "Generators must be global unaries.")
					}
					lits = newlits
				}
				ss += lits
			}

			atomStrings := strings.Fields(ss)
			atoms := []Atom{}
			for _, x := range atomStrings {
				var a Atom
				a.s = simplify(x, p.Constants)
				atoms = append(atoms, a)
			}

			if isQuantifier {
				var q Quantifier
				q.s = s
				q.Exist = quant == "e"
				q.Atoms = atoms
				p.Quantifiers = append(p.Quantifiers, q)
			} else {
				p.Rules = append(p.Rules, Rule{s, atoms})
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
	return "" == strings.Trim(s, "0123456789+*%-")
}

func boolMathExpression(s string) bool {
	//	r, _ := regexp.MatchString("[0-9+*/%!=><]+", s)
	return "" == strings.Trim(s, "0123456789+*%-=><")
}

//assumption:Space only between literals.
func replaceConstants(term string, constants map[string]int) (s string) {
	for Const, Val := range constants {
		term = strings.ReplaceAll(term, Const, strconv.Itoa(Val))
	}
	return term
}

func evaluateExpression(term string) int {
	term = strings.ReplaceAll(term, "#mod", "%")
	expression, err := govaluate.NewEvaluableExpression(term)
	assertx(err, term)
	result, err := expression.Evaluate(nil)
	assertx(err, term)
	return int(result.(float64))
}

type Literal struct {
	id     string
	par    []string
	ground []int
}

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
			newcl += simplify(literal, constants)
		}
		rcls[i] = newcl
	}

	return rcls
}

func extract(literal string) (pre string, par string) {
	pre = literal[:strings.Index(literal, "(")]
	literal = literal[strings.Index(literal, "(")+1:]
	par = literal[:strings.LastIndex(literal, ")")]
	return
}

func simplify(literal string, constants map[string]int) (newcl string) {
	pre, par := extract(literal)
	newcl = pre + "("
	terms := strings.Split(par, ",")
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
