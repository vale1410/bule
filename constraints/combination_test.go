package constraints

import (
	"fmt"
	"github.com/vale1410/bule/sat"
	"testing"
)

func TestRewriteExactly1(test *testing.T) {
	fmt.Println("TestExactly1")

	//+2 x1 +2 x2 +3 x3 +4 x4 +1 x5 +1 x6 <= 6 ;
	//+1 x1 +1 x2 +1 x3 +1 x4 = 1 ;

	pb1 := CreatePB([]int64{2, 2, 3, 4, 1, 1}, 6)
	pb1.Typ = AtMost

	pb2 := CreatePB([]int64{1, 1, 1, 1}, 1)
	pb2.Typ = Equal

	//pb1.Print10()
	//pb2.Print10()

	b := PreprocessExactly(&pb1, &pb2)
	if !b && len(pb1.Entries) != 4 {
		test.Fail()
	}

	//pb1.Print10()

}

func TestRewriteExactly2(test *testing.T) {
	fmt.Println("TestExactly2")

	//+2 x1 +2 x2 +3 x3 +4 x4 +1 x5 +1 x6 <= 6 ;
	//+1 x1 +1 x2 +1 x3 +1 x4 = 2 ;

	pb1 := CreatePB([]int64{2, 2, 3, 4, 1, 1}, 6)
	pb1.Typ = AtMost

	pb2 := CreatePB([]int64{1, 1, 1, 1}, 2)
	pb2.Typ = Equal

	//pb1.Print10()
	//pb2.Print10()

	b := PreprocessExactly(&pb1, &pb2)

	//pb1.Print10()

	if !b && len(pb1.Entries) != 4 {
		test.Fail()
	}

}

func TestRewriteExactly3(test *testing.T) {
	fmt.Println("TestExactly3")

	pb1 := CreatePB([]int64{2, 2, 3, 4, 1, 1}, 6)
	pb1.Typ = AtMost
	pb1.Sort()

	pb2 := CreatePB([]int64{1, 1, 1, 1}, 2)
	pb2.Typ = Equal

	//pb1.Print10()
	//pb2.Print10()

	b := PreprocessExactly(&pb1, &pb2)

	//pb1.Print10()
	//pb1.SortVar()
	//pb1.Print10()
	if !b && len(pb1.Entries) != 4 {
		test.Fail()
	}

}

func TestRewriteExactly4(test *testing.T) {
	fmt.Println("TestExactly3")

	pb1 := CreatePB([]int64{2, 2, 3, 4, 1, 1}, 6)
	pb1.Typ = AtMost

	pb2 := CreatePB([]int64{1, 1, 1, 1}, 2)
	pb2.Typ = Equal
	pb2.Entries[2].Literal = sat.Neg(pb2.Entries[2].Literal)

	b := PreprocessExactly(&pb1, &pb2)

	if b {
		test.Fail()
	}

}

func TestRewriteAMO(test *testing.T) {
	fmt.Println("TestRewriteAMO1")

	pb1 := CreatePB([]int64{2, 2, 3, 4, 1, 1}, 6)

	pb2 := CreatePB([]int64{1, 1, 1, 1}, 1)

	pb1.Print10()
	pb2.Print10()

	//translate AMO, i.e. pb2
	//b, literals := pb2.Cardinality()
	//	amo := AtMostOne(Count, "count", literals)
	//
	//	t := TranslatePBwithAMO(&pb1, amo)
	//	fmt.Println(t)
	//
	//	if !b || len(pb1.Entries) != 5 {
	//		test.Fail()
	//	}

}
