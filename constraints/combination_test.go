package constraints

import (
	"fmt"
	"testing"

	"github.com/vale1410/bule/glob"
	"github.com/vale1410/bule/sat"
)

func TestRewriteExactly1(test *testing.T) {
	glob.D("TestExactly1")

	//+2 x1 +2 x2 +3 x3 +4 x4 +1 x5 +1 x6 <= 6 ;
	//+1 x1 +1 x2 +1 x3 +1 x4 = 1 ;

	pb1 := CreatePB([]int64{2, 2, 3, 4, 1, 1}, 6)
	pb1.Typ = LE

	pb2 := CreatePB([]int64{1, 1, 1, 1}, 1)
	pb2.Typ = EQ

	//pb1.Print10()
	//pb2.Print10()

	b := PreprocessPBwithExactly(&pb1, &pb2)
	if !b && len(pb1.Entries) != 4 {
		test.Fail()
	}

	//pb1.Print10()

}

func TestRewriteExactly2(test *testing.T) {
	glob.D("TestExactly2")

	//+2 x1 +2 x2 +3 x3 +4 x4 +1 x5 +1 x6 <= 6 ;
	//+1 x1 +1 x2 +1 x3 +1 x4 = 2 ;

	pb1 := CreatePB([]int64{2, 2, 3, 4, 1, 1}, 6)
	pb1.Typ = LE

	pb2 := CreatePB([]int64{1, 1, 1, 1}, 2)
	pb2.Typ = EQ

	//pb1.Print10()
	//pb2.Print10()

	b := PreprocessPBwithExactly(&pb1, &pb2)

	//pb1.Print10()

	if !b && len(pb1.Entries) != 4 {
		test.Fail()
	}

}

func TestRewriteExactly3(test *testing.T) {
	glob.D("TestExactly3")

	pb1 := CreatePB([]int64{2, 2, 3, 4, 1, 1}, 6)
	pb1.Typ = LE
	pb1.SortDescending()

	pb2 := CreatePB([]int64{1, 1, 1, 1}, 2)
	pb2.Typ = EQ

	//pb1.Print10()
	//pb2.Print10()

	b := PreprocessPBwithExactly(&pb1, &pb2)

	//pb1.Print10()
	//pb1.SortVar()
	//pb1.Print10()
	if !b && len(pb1.Entries) != 4 {
		test.Fail()
	}

}

func TestRewriteExactly4(test *testing.T) {
	glob.D("TestExactly3")

	pb1 := CreatePB([]int64{2, 2, 3, 4, 1, 1}, 6)
	pb1.Typ = LE

	pb2 := CreatePB([]int64{1, 1, 1, 1}, 2)
	pb2.Typ = EQ
	pb2.Entries[2].Literal = sat.Neg(pb2.Entries[2].Literal)

	b := PreprocessPBwithExactly(&pb1, &pb2)

	if b {
		test.Fail()
	}

}

func TestRewriteAMO(test *testing.T) {
	glob.D("TestRewriteAMO1")

	pb1 := CreatePB([]int64{2, 2, 3, 4, 1, 1}, 6)
	pb2 := CreatePB([]int64{1, 1, 1, 1}, 1)

	//pb1.Print10()
	//pb2.Print10()

	//translate AMO, i.e. pb2
	b, literals := pb2.Cardinality()
	amo := TranslateAtMostOne(Count, "count", literals)
	amo.PB = &pb2

	b = PreprocessPBwithAMO(&pb1, amo)

	if !b || len(pb1.Entries) != 5 {
		test.Fail()
	}

	if sumE(pb1) != 6 {
		test.Fail()
	}
}

func TestTranslateAMO0(test *testing.T) {
	glob.D("TestTranslateAMO0")
	*glob.MDD_max_flag = 300000
	*glob.MDD_redundant_flag = false

	pb1 := CreatePB([]int64{2, 3, 4, 3}, 5)
	pb2 := CreatePB([]int64{1, 1, 1}, 1)

	//pb1.Print10()
	//pb2.Print10()

	//translate AMO, i.e. pb2
	b, literals := pb2.Cardinality()
	amo := TranslateAtMostOne(Count, "c", literals)
	amo.PB = &pb2

	TranslatePBwithAMO(&pb1, amo)

	if !b || pb1.Clauses.Size() != 8 {
		fmt.Println("translation size incorrect", pb1.Clauses.Size())
		pb1.Clauses.PrintDebug()
		test.Fail()
	}
}

func TestTranslateAMO1(test *testing.T) {
	glob.D("TestTranslateAMO1")
	*glob.MDD_max_flag = 300000
	*glob.MDD_redundant_flag = false

	pb1 := CreatePB([]int64{2, 2, 3, 4, 2, 3}, 6)
	pb2 := CreatePB([]int64{1, 1, 1, 1}, 1)

	//pb1.Print10()
	//pb2.Print10()

	//translate AMO, i.e. pb2
	b, literals := pb2.Cardinality()
	amo := TranslateAtMostOne(Count, "c", literals)
	amo.PB = &pb2

	TranslatePBwithAMO(&pb1, amo)

	if !b || pb1.Clauses.Size() != 13 {
		fmt.Println("translation size incorrect", pb1.Clauses.Size(), "should be:", 13)
		pb1.Clauses.PrintDebug()
		test.Fail()
	}
}

func TestTranslateAMO2(test *testing.T) {
	glob.D("TestTranslateAMO2")
	*glob.MDD_max_flag = 300000
	*glob.MDD_redundant_flag = false

	results := []int{40, 33, 29}

	for i := 0; i < 3; i++ {

		//fmt.Println()
		pb1 := CreatePB([]int64{2, 2, 3, 4, 4, 5, 2, 1}, 8)
		pb2 := CreatePBOffset(i, []int64{1, 1, 1, 1}, 1)

		//pb1.Print10()
		//pb2.Print10()

		b, literals := pb2.Cardinality()
		amo := TranslateAtMostOne(Count, "c", literals)
		amo.PB = &pb2

		TranslatePBwithAMO(&pb1, amo)
		//t.Clauses.PrintDebug()

		if !b || pb1.Clauses.Size() != results[i] {
			fmt.Println("translation size incorrect", pb1.Clauses.Size(), " should be", results[i])
			//t.Clauses.PrintDebug()
			test.Fail()
		}
	}

}

func sumE(pb Threshold) (r int64) {
	for _, x := range pb.Entries {
		r += x.Weight
	}
	return
}
