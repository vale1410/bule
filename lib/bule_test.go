package bule

import (
	"fmt"
	"testing"
)

func TestEvaluateExpression(t *testing.T) {

	expr := Term("X+Y==3")
	cc := map[string]int{
		"X": 5,
		"Y": 10,
	}
	cc2 := map[string]int{
		"X": 2,
		"Y": 1,
	}
	exprFalse, _ := assign(expr, cc)
	exprTrue, _ := assign(expr, cc2)
	valF := evaluateBoolExpression(exprFalse.String())
	valT := evaluateBoolExpression(exprTrue.String())
	if valF || !valT {
		t.Error("Evaluation is wrong.")
	}
}

func TestInstantiate1(t *testing.T) {

	{
		rule := parseRule("move[X,Y,4].")
		assignment := make(map[string]int, 0)
		assignment["Y"] = 3
		a := rule.Literals[0]
		b := a.assign(assignment)
		if b.String() != "move[X,3,4]" {
			fmt.Println("a:", a, "\nb:", b)
			t.Fail()
		}
	}
}

func TestBule1(t *testing.T) {

	lines := []string{"#const a=5.", "fact[1,2].", "fact[2,a].", "search[A,B]:A+1==B:fact[A,B]."}
	p := ParseProgramFromStrings(lines)
	p.ReplaceConstants()
	p.CollectFacts()
	p.ExpandGenerators()
	if p.Rules[0].String() == "search[1,2]" {
		t.Fail()
	}
}

func TestBule2(t *testing.T) {

	lines := []string{"a[1..7,4].", "b[2].", "search[A,B,_C] : A==B : a[A,C] : b[B]."}
	fmt.Println(lines)
	p := ParseProgramFromStrings(lines)
	p.ReplaceConstants()
	p.Print()
	p.ExpandIntervals()
	p.Print()
	p.CollectFacts()
	p.ExpandGenerators()
	p.Print()
}

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
