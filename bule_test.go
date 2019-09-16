package main

import (
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
		"X":5,
		"Y":10,
	}
	cc2 := map[string]int{
		"X":2,
		"Y":1,
	}
	exprFalse := replaceConstants(expr, cc)
	exprTrue := replaceConstants(expr, cc2)
	valF := evaluateBoolExpression(exprFalse)
	valT := evaluateBoolExpression(exprTrue)
	if valF ||!valT  {
		t.Error("Evaluation is wrong.")
	}
}



func TestInstantiate(t *testing.T) {

	{
		a := Atom{"move(X,Y,4)"}
		b := a.instantiate("Y", 3)
		if b.s != "move(X,3,4)" {
			t.Fail()
		}
	}

	{
		a := Atom{"move(X,Y+3,4)"}
		b := a.instantiate("Y", 3)
		if b.s != "move(X,6,4)" {
			t.Fail()
		}
	}

	{
		a := Atom{"move(X,Y+3,4)"}
		b := a.instantiate("Z", 3)
		if b.s != "move(X,Y+3,4)" {
			t.Fail()
		}
	}
}