package lib

import (
	"testing"
)

//func TestEvaluateExpression(t *testing.T) {
//
//	expr := Term("X+Y==3")
//	cc := map[string]int{
//		"X": 5,
//		"Y": 10,
//	}
//	cc2 := map[string]int{
//		"X": 2,
//		"Y": 1,
//	}
//	exprFalse, _ := assign(expr, cc)
//	exprTrue, _ := assign(expr, cc2)
//	valF := evaluateBoolExpression(exprFalse.String())
//	valT := evaluateBoolExpression(exprTrue.String())
//	if valF || !valT {
//		t.Error("Evaluation is wrong.")
//	}
//}
//
//func TestInstantiate1(t *testing.T) {
//
//	{
//		rule,_ := parseRule("move[X,Y,4].")
//		assignment := make(map[string]int, 0)
//		assignment["Y"] = 3
//		a := rule.Literals[0]
//		b := a.assign(assignment)
//		if b.String() != "move[X,3,4]" {
//			fmt.Println("a:", a, "\nb:", b)
//			t.Fail()
//		}
//	}
//}
//
//func TestBule1(t *testing.T) {
//
//	lines := []string{"#const a=5.", "fact[1,2].", "fact[2,a].", "search[A,B]:A+1==B:fact[A,B]."}
//	p := ParseProgramFromStrings(lines)
//	p.ReplaceConstantsAndMathFunctions()
//	p.CollectGroundFacts()
//	p.ExpandConditionals()
//	if p.Rules[0].String() == "search[1,2]" {
//		t.Fail()
//	}
//}
//
//func TestBule2(t *testing.T) {
//
//	lines := []string{"a[1..7,4].", "b[3].", "search[A,B,C] : A>B : a[A,C] : b[B]."}
////	fmt.Println(lines)
//	p := ParseProgramFromStrings(lines)
//	p.ReplaceConstantsAndMathFunctions()
//	p.ExpandGroundRanges()
//	p.CollectGroundFacts()
//	p.ExpandConditionals()
//	p.Print()
//}

func TestMathExpressions(t *testing.T) {

	mathExpression := "1-(2)"
	t.Log(mathExpression)
	t.Log(evaluateTermExpression(mathExpression))

}
