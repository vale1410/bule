package bule

import (
	//"log"
	"testing"
)

func TestEvaluateExpression(t *testing.T) {

	expr := "X+Y==3."
	rule, err := parseRule(expr)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	rule.Debug()


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
//
//func TestInstantiate1(t *testing.T) {
//
//	{
//		a, _ := parseAtom("move[X,Y,4]")
//		assignment := make(map[string]int, 0)
//		assignment["Y"] = 3
//		b := a.simplifyAtom(assignment)
//		if b.String() != "move[X,3,4]" {
//			log.Println(a," ",b)
//			t.Fail()
//		}
//	}
//}
//
//
//func TestInstantiate2(t *testing.T) {
//
//	{
//		a, _ := parseAtom("move[X,Y+3,4]")
//		assignment := make(map[string]int, 0)
//		assignment["Y"] = 3
//		b := a.simplifyAtom(assignment)
//		if b.String() != "move[X,6,4]" {
//			log.Println(a," ",b)
//			t.Fail()
//		}
//	}
//}
//
//func TestInstantiate3(t *testing.T) {
//
//	{
//		a, _ := parseAtom("move[X,Y#mod2,4]")
//		assignment := make(map[string]int, 0)
//		assignment["Y"] = 3
//		b := a.simplifyAtom(assignment)
//		if b.String() != "move[X,1,4]" {
//			log.Println(a," ",b)
//			t.Fail()
//		}
//	}
//}

