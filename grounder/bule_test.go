package grounder

import (
	"testing"
)

func TestMathExpressions(t *testing.T) {

	mathExpression := "1-(2)"
	t.Log(mathExpression)
	x, err := evaluateTermExpression(mathExpression)
	if err != nil {
		t.Log(err)
		return
	}
	t.Log(x)

}

func TestMathExpressions3(t *testing.T) {
	ter := Term("sd-42*fhd(8%43+asdBN%sd+=123**dasd+sd")
	assignment := make(map[string]string, 0)
	assignment["asdBN"] = "1111111"
	assignment["sd"] = "666"
	ts, _, _ := assign(ter, assignment)
	if ts != "666-42*fhd(8%43+1111111%666+=123**dasd+666" {
		t.Fail()
	}
}

func TestMathExpressions2(t *testing.T) {
	ter := Term("0..n")
	assignment := make(map[string]string, 0)
	assignment["n"] = "123"
	ts, _, _ := assign(ter, assignment)
	if ts != "0..123" {
		t.Fail()
	}
}
