package grounder

import (
	"fmt"
	"testing"
)

func checkLexing(input string, expected string, t *testing.T) {
	ts := lexRule(input)
	computed := ts.Debug()
	if expected != computed {
		fmt.Println("input\t:", input)
		fmt.Println("computed\t:", computed)
		fmt.Println("expected\t:", expected)
		t.Fail()
	}
}
func TestParser1c(t *testing.T) {
	in := "move[((A*B))]."
	out := "ATOM{move}-AtomBL{[}-TERM{((A*B))}-AtomBR{]}-DOT{.}-"
	checkLexing(in, out, t)
}

func TestParser1a(t *testing.T) {
	in := "move[A,b+1,4*(a*b)]."
	out := "ATOM{move}-AtomBL{[}-TERM{A}-TERMCOMMA{,}-TERM{b+1}-TERMCOMMA{,}-TERM{4*(a*b)}-AtomBR{]}-DOT{.}-"
	checkLexing(in, out, t)
}

func TestParser1b(t *testing.T) {
	in := "move(A,b+1,4*(a*b))."
	out := "ATOM{move}-ATOMPL{(}-TERM{A}-TERMCOMMA{,}-TERM{b+1}-TERMCOMMA{,}-TERM{4*(a*b)}-ATOMPR{)}-DOT{.}-"
	checkLexing(in, out, t)
}

func TestParser2(t *testing.T) {
	in := "mo[4],ab[5,A]."
	out := "ATOM{mo}-AtomBL{[}-TERM{4}-AtomBR{]}-RULECOMMA{,}-ATOM{ab}-AtomBL{[}-TERM{5}-TERMCOMMA{,}-TERM{A}-AtomBR{]}-DOT{.}-"
	checkLexing(in, out, t)
}

func TestParser3(t *testing.T) {

	in := "~mo[4],X<Y,Y+1>=5,~ab[X,Y+1]."
	out := "NEGATION{~}-ATOM{mo}-AtomBL{[}-TERM{4}-AtomBR{]}-RULECOMMA{,}-TERM{X}-LT{<}-TERM{Y}-RULECOMMA{,}-TERM{Y+1}-GE{>=}-TERM{5}-RULECOMMA{,}-NEGATION{~}-ATOM{ab}-AtomBL{[}-TERM{X}-TERMCOMMA{,}-TERM{Y+1}-AtomBR{]}-DOT{.}-"
	checkLexing(in, out, t)
}

func TestParser4(t *testing.T) {

	in := "abc[1..4]."
	out := "ATOM{abc}-AtomBL{[}-TERM{1}-DOUBLEDOT{..}-TERM{4}-AtomBR{]}-DOT{.}-"
	checkLexing(in, out, t)
}

func TestParser5(t *testing.T) {
	in := "abc[X]:ab[X]:X<7."
	out := "ATOM{abc}-AtomBL{[}-TERM{X}-AtomBR{]}-COLON{:}-ATOM{ab}-AtomBL{[}-TERM{X}-AtomBR{]}-COLON{:}-TERM{X}-LT{<}-TERM{7}-DOT{.}-"
	checkLexing(in, out, t)
}

func TestParser6(t *testing.T) {
	in := "asd12[A]->~fd[B]."
	out := "ATOM{asd12}-AtomBL{[}-TERM{A}-AtomBR{]}-IMPLICATION{->}-NEGATION{~}-ATOM{fd}-AtomBL{[}-TERM{B}-AtomBR{]}-DOT{.}-"
	checkLexing(in, out, t)
}

func TestParser7(t *testing.T) {
	in := "gena[A],sd12[(A+1/2)]<->~fd[B]:gen[B]."
	out := "ATOM{gena}-AtomBL{[}-TERM{A}-AtomBR{]}-RULECOMMA{,}-ATOM{sd12}-AtomBL{[}-TERM{(A+1/2)}-AtomBR{]}-EQUIVALENCE{<->}-NEGATION{~}-ATOM{fd}-AtomBL{[}-TERM{B}-AtomBR{]}-COLON{:}-ATOM{gen}-AtomBL{[}-TERM{B}-AtomBR{]}-DOT{.}-"
	checkLexing(in, out, t)
}

func TestParser8(t *testing.T) {
	in := "board[X+Z*D,Y+(1-Zas)*D*V+((-1)**Z)*D*(1-V),P]."
	out := "ATOM{board}-AtomBL{[}-TERM{X+Z*D}-TERMCOMMA{,}-TERM{Y+(1-Zas)*D*V+((-1)**Z)*D*(1-V)}-TERMCOMMA{,}-TERM{P}-AtomBR{]}-DOT{.}-"
	checkLexing(in, out, t)
}

func TestParser9(t *testing.T) {
	in := "X+3>Y+3,(A-1)<=Bds-1."
	out := "TERM{X+3}-GT{>}-TERM{Y+3}-RULECOMMA{,}-TERM{(A-1)}-LE{<=}-TERM{Bds-1}-DOT{.}-"
	checkLexing(in, out, t)
}

func TestParser10(t *testing.T) {
	in := "X+3>Y+3,(A-1)<=Bds-1."
	out := "TERM{X+3}-GT{>}-TERM{Y+3}-RULECOMMA{,}-TERM{(A-1)}-LE{<=}-TERM{Bds-1}-DOT{.}-"
	checkLexing(in, out, t)
}

func TestParser11(t *testing.T) {
	in := "move[3,2],1<3,0<2."
	out := "ATOM{move}-AtomBL{[}-TERM{3}-TERMCOMMA{,}-TERM{2}-AtomBR{]}-RULECOMMA{,}-TERM{1}-LT{<}-TERM{3}-RULECOMMA{,}-TERM{0}-LT{<}-TERM{2}-DOT{.}-"
	checkLexing(in, out, t)
}

func TestParser12(t *testing.T) {
	in := "#forall[3],a[1]."
	out := "ATOM{#forall}-AtomBL{[}-TERM{3}-AtomBR{]}-RULECOMMA{,}-ATOM{a}-AtomBL{[}-TERM{1}-AtomBR{]}-DOT{.}-"
	checkLexing(in, out, t)
}

func TestParser13(t *testing.T) {
	in := "move[X,Y#mod2,4]."
	out := "ATOM{move}-AtomBL{[}-TERM{X}-TERMCOMMA{,}-TERM{Y#mod2}-TERMCOMMA{,}-TERM{4}-AtomBR{]}-DOT{.}-"
	checkLexing(in, out, t)
}

// TODO:
//"#CONST q=5."
//
