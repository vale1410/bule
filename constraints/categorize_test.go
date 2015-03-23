package constraints

import (
	"testing"

	"github.com/vale1410/bule/glob"
)

func TestTranslate(test *testing.T) {
	glob.D("TestTranslate")

	pb1 := CreatePB([]int64{1, 1, 1}, 1)
	t := Categorize1(&pb1)
	if t.Typ != AMO {
		pb1.Print10()
		test.Errorf("1: Does not classify atmostOne")
	}

	pb2 := CreatePB([]int64{1, 1, 1}, 1)
	pb2.Typ = GE
	t = Categorize1(&pb2)
	if t.Typ != Clause {
		pb2.Print10()
		test.Errorf("2: Does not classify a clause")
	}

	pb3 := CreatePB([]int64{1, 1, 1}, 1)
	pb3.Typ = EQ
	t = Categorize1(&pb3)

	if t.Typ != EX1 {
		pb3.Print10()
		test.Errorf("3: Does not classify ExactlyOne")
	}

	pb4 := CreatePB([]int64{1, 1, -1}, 0)
	pb4.Typ = EQ
	t = Categorize1(&pb4)

	if t.Typ != EX1 {
		pb4.Print10()
		test.Errorf("4: Does not classify ExactlyOne")
	}

	pb5 := CreatePB([]int64{-3, 3, -3}, 0)
	pb5.Typ = LE
	t = Categorize1(&pb5)

	if t.Typ != Clause { // should be different
		pb5.Print10()
		test.Errorf("5: Does not classify clause", pb5)
	}

	pb6 := CreatePB([]int64{1, 1, 1, 1, 1}, 4)
	pb6.Typ = EQ

	t = Categorize1(&pb6)

	if t.Typ != EX1 {
		test.Errorf("6: Does not classify ExactlyOne", pb6)
	}

}
