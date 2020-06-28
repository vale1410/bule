package lib

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
