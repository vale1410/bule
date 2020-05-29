package lib

import (
	"testing"
)

func TestMathExpressions(t *testing.T) {

	mathExpression := "1-(2)"
	t.Log(mathExpression)
	t.Log(evaluateTermExpression(mathExpression))

}
