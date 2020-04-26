package constraints

import (
	"testing"

	"github.com/vale1410/bule/glob"
)

func TestTranslate(test *testing.T) {
	glob.D("TestTranslate")

	pb1 := CreatePB([]int64{1, 1, 1}, 1)
	pb1.CategorizeTranslate1()
	if pb1.TransTyp != AMO {
		pb1.Print10()
		test.Errorf("1: Does not classify atmostOne")
	}

	pb2 := CreatePB([]int64{1, 1, 1}, 1)
	pb2.Typ = GE
	pb2.CategorizeTranslate1()
	if pb2.TransTyp != Clause {
		pb2.Print10()
		test.Errorf("2: Does not classify a clause")
	}

	pb3 := CreatePB([]int64{1, 1, 1}, 1)
	pb3.Typ = EQ
	pb3.CategorizeTranslate1()

	if pb3.TransTyp != EX1 {
		pb3.Print10()
		test.Errorf("3: Does not classify ExactlyOne")
	}

	pb4 := CreatePB([]int64{1, 1, -1}, 0)
	pb4.Typ = EQ
	pb4.CategorizeTranslate1()

	if pb4.TransTyp != EX1 {
		pb4.Print10()
		test.Errorf("4: Does not classify ExactlyOne")
	}

	pb5 := CreatePB([]int64{-3, 3, -3}, 0)
	pb5.Typ = LE
	pb5.CategorizeTranslate1()

	if pb5.TransTyp != Clause { // should be different
		pb5.Print10()
		test.Errorf("5: Does not classify clause", pb5)
	}

	pb6 := CreatePB([]int64{1, 1, 1, 1, 1}, 4)
	pb6.Typ = EQ
	pb6.CategorizeTranslate1()

	if pb6.TransTyp != EX1 {
		test.Errorf("6: Does not classify ExactlyOne", pb6)
	}

}
