package main

import (
	"log"
	"regexp"
	"testing"
)

////assumption:Space only between literals.
//func replaceConstants(term string, constants map[string]int) string {
//	for Const, Val := range constants {
//		term = strings.ReplaceAll(term, Const, strconv.Itoa(Val))
//	}
//	return term
//}

//
//
//// Evaluates a ground math expression, needs to path mathExpression
//func evaluateExpression(term string) int {
//	term = strings.ReplaceAll(term, "#mod", "%")
//	expression, err := govaluate.NewEvaluableExpression(term)
//	assertx(err, term)
//	result, err := expression.Evaluate(nil)
//	assertx(err, term)
//	return int(result.(float64))
//}

func TestEvaluateExpression(t *testing.T) {

	expr := "X+Y==3"
	cc := map[string]int{
		"X": 5,
		"Y": 10,
	}
	cc2 := map[string]int{
		"X": 2,
		"Y": 1,
	}
	exprFalse := assign(expr, cc)
	exprTrue := assign(expr, cc2)
	valF := evaluateBoolExpression(exprFalse)
	valT := evaluateBoolExpression(exprTrue)
	if valF || !valT {
		t.Error("Evaluation is wrong.")
	}
}

func TestInstantiate1(t *testing.T) {

	{
		a, _ := parseAtom("move[X,Y,4]")
		assignment := make(map[string]int, 0)
		assignment["Y"] = 3
		b := a.simplifyAtom(assignment)
		if b.String() != "move[X,3,4]" {
			log.Println(a," ",b)
			t.Fail()
		}
	}
}
func TestInstantiate2(t *testing.T) {

	{
		a, _ := parseAtom("move[X,Y+3,4]")
		assignment := make(map[string]int, 0)
		assignment["Y"] = 3
		b := a.simplifyAtom(assignment)
		if b.String() != "move[X,6,4]" {
			log.Println(a," ",b)
			t.Fail()
		}
	}
}

func TestInstantiate3(t *testing.T) {

	{
		a, _ := parseAtom("move[X,Y#mod2,4]")
		assignment := make(map[string]int, 0)
		assignment["Y"] = 3
		b := a.simplifyAtom(assignment)
		if b.String() != "move[X,1,4]" {
			log.Println(a," ",b)
			t.Fail()
		}
	}
}

//board(X+Z*D,Y+(1-Z)*D*V+((-1)**Z)*D*(1-V),P)
func TestDecompose(t *testing.T) {

	{
		a, _ := parseAtom("board[X+Z*D,Y+(1-Zas)*D*V+((-1)**Z)*D*(1-V),P]")
		if a.Name != "board" {
			t.Fail()
			log.Println(a.Terms)
			log.Println(a.FreeVars())
		}
		if len(a.Terms) != 3 {
			t.Fail()
			log.Println(a.Terms)
			log.Println(a.FreeVars())
		}
	}
}



func TestRuleElement(t *testing.T) {

}
func TestRegexp(t *testing.T) {

//	buleVariable:= `^[A-Z][a-zA-Z0-9_]*`
//	numberExpression := `^[a-z][a-zA-Z0-9_]*` // for variables and constants
//	buleIdentifier := `^[a-z][a-zA-Z0-9_]*` // for variables and constants
//	buleDefinition := buleVariable + `=\{\}`
//	buleAtomParameter := `^[a-z][a-zA-Z0-9]*`
//	termExpression := ``
//	atomExpression := regexp.MustCompile(`^[a-z][a-zA-Z0-9]*\[.*\]`)

	atom := regexp.MustCompile(`^[a-z][a-zA-Z0-9]*\[`)
	if 	atom.FindString("aJG4G[43,2,54],") !=  "aJG4G[" {
		t.Fail()
	}
}
