package parser

import (
	"fmt"
	"log"
	"testing"
)

func TestParser1(t *testing.T) {
	a, _ := DebugString("move[A,b+1,4*(a*b)].")
	if a != "ATOM{move}-PL{[}-TERM{A}-TERMCOMMA{,}-TERM{b+1}-TERMCOMMA{,}-TERM{4*(a*b)}-PR{]}-DOT{.}" {
		fmt.Println(a)
		t.Fail()
	}
}

func TestParser2(t *testing.T) {
	a, _ := DebugString("mo[4],ab[5,A].")
	if a != "ATOM{mo}-PL{[}-TERM{4}-PR{]}-RULECOMMA{,}-ATOM{ab}-PL{[}-TERM{5}-TERMCOMMA{,}-TERM{A}-PR{]}-DOT{.}" {
		log.Println(a)
		t.Fail()
	}
}

func TestParser3(t *testing.T) {
	a, err := DebugString("~mo[4],X<Y,Y+1>=5,~ab[X,Y+1].")

	if err != nil {
		fmt.Println(err)
	}
	if a != "NEGATION{~}-ATOM{mo}-PL{[}-TERM{4}-PR{]}-RULECOMMA{,}-CONSTRAINT{X<Y}-RULECOMMA{,}-CONSTRAINT{Y+1>=5}-RULECOMMA{,}-NEGATION{~}-ATOM{ab}-PL{[}-TERM{X}-TERMCOMMA{,}-TERM{Y+1}-PR{]}-DOT{.}" {
		log.Println(a)
		t.Fail()
	}
}

func TestParser4(t *testing.T) {
	a, err := DebugString("abc[1..4].")
	if err != nil {
		fmt.Println(err)
	}
	if a != "ATOM{abc}-PL{[}-TERM{1}-DOUBLEDOT{..}-TERM{4}-PR{]}-DOT{.}" {
		log.Println(a)
		t.Fail()
	}
}

func TestParser5(t *testing.T) {
	a, err := DebugString("abc[X]:ab[X]:X<7.")
	if err != nil {
		fmt.Println(err)
	}
	if a != "ATOM{abc}-PL{[}-TERM{X}-PR{]}-COLON{:}-ATOM{ab}-PL{[}-TERM{X}-PR{]}-COLON{:}-CONSTRAINT{X<7}-DOT{.}" {
		log.Println(a)
		t.Fail()
	}
}

//func TestParser6(t *testing.T) {
//	a, err := DebugString("#CONST q=5.")
//	if err != nil {
//		fmt.Println(err)
//	}
//	if a != "ATOM[abc]-PL[(]-TERM[X]-PR[)]-COLON[:]-ATOM[ab]-PL[(]-TERM[X]-PR[)]-COLON[:]-CONSTRAINT[X<7]-DOT{.}\n" {
//		log.Println(a)
//		t.Fail()
//	}
//	fmt.Println()
//}

func TestParser7(t *testing.T) {
	a, err := DebugString("asd12[A]=>~fd[B].")
	if err != nil {
		fmt.Println(err)
	}
	if a != "ATOM{asd12}-PL{[}-TERM{A}-PR{]}-IMPLICATION{=>}-NEGATION{~}-ATOM{fd}-PL{[}-TERM{B}-PR{]}-DOT{.}" {
		log.Println(a)
		t.Fail()
	}
	fmt.Println()
}

func TestParser8(t *testing.T) {
	a, err := DebugString("gena[A],sd12[A]<=>~fd[B]:gen[B].")
	if err != nil {
		fmt.Println(err)
	}
	if a != "ATOM{gena}-PL{[}-TERM{A}-PR{]}-RULECOMMA{,}-ATOM{sd12}-PL{[}-TERM{A}-PR{]}-EQUIVALENCE{<=>}-NEGATION{~}-ATOM{fd}-PL{[}-TERM{B}-PR{]}-COLON{:}-ATOM{gen}-PL{[}-TERM{B}-PR{]}-DOT{.}" {
		log.Println(a)
		t.Fail()
	}
	fmt.Println()
}

func TestParser9(t *testing.T) {
	a, err := DebugString("board[X+Z*D,Y+(1-Zas)*D*V+((-1)**Z)*D*(1-V),P].")
	if err != nil {
		fmt.Println(err)
	}
	if a != "ATOM{board}-PL{[}-TERM{X+Z*D}-TERMCOMMA{,}-TERM{Y+(1-Zas)*D*V+((-1)**Z)*D*(1-V)}-TERMCOMMA{,}-TERM{P}-PR{]}-DOT{.}" {
		log.Println(a)
		t.Fail()
	}

}


func TestParser10(t *testing.T) {

	a, err := DebugString("move[3,2],(1<3),(0<2).")
	if err != nil {
		fmt.Println(err)
	}
	if a != "ATOM{move}-PL{[}-TERM{3}-TERMCOMMA{,}-TERM{2}-PR{]}-RULECOMMA{,}-CONSTRAINT{(1<3)}-RULECOMMA{,}-CONSTRAINT{(0<2)}-DOT{.}" {
		log.Println(a)
		t.Fail()
	}

}

func TestParser11(t *testing.T) {

	a, err := DebugString("move[X,Y#mod2,4].")
	if err != nil {
		fmt.Println(err)
	}
	if a != "ATOM{move}-PL{[}-TERM{X}-TERMCOMMA{,}-TERM{Y#mod2}-TERMCOMMA{,}-TERM{4}-PR{]}-DOT{.}" {
		log.Println(a)
		t.Fail()
	}
}


func TestParser12(t *testing.T) {

	a, err := DebugString("#forall[3],a[1].")
	if err != nil {
		fmt.Println(err)
	}
	if a != "ATOM{#forall}-PL{[}-TERM{3}-PR{]}-RULECOMMA{,}-ATOM{a}-PL{[}-TERM{1}-PR{]}-DOT{.}" {
		log.Println(a)
		t.Fail()
	}
}

