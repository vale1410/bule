package constraints

import (
	"fmt"
	"testing"
)

func TestTranslate(test *testing.T) {
	fmt.Println("TestTranslate")

	pb1 := CreatePB([]int64{1, 1, 1}, 1)
	t := Translate(&pb1)
	if t.Typ != AtMostOne {
		pb1.Print10()
		test.Errorf("1: Does not classify atmostOne")
	}

	pb2 := CreatePB([]int64{1, 1, 1}, 1)
	pb2.Typ = AtLeast
	t = Translate(&pb2)
	if t.Typ != Clause {
		pb2.Print10()
		test.Errorf("2: Does not classify a clause")
	}

	pb3 := CreatePB([]int64{1, 1, 1}, 1)
	pb3.Typ = Equal
	t = Translate(&pb3)

	if t.Typ != ExactlyOne {
		pb3.Print10()
		test.Errorf("3: Does not classify ExactlyOne")
	}

	pb4 := CreatePB([]int64{1, 1, -1}, 0)
	pb4.Typ = Equal
	t = Translate(&pb4)

	if t.Typ != ExactlyOne {
		pb4.Print10()
		test.Errorf("4: Does not classify ExactlyOne")
	}

	pb5 := CreatePB([]int64{-3, 3, -3}, 0)
	pb5.Typ = AtMost
	t = Translate(&pb5)

	if t.Typ != Clause { // should be different
		pb5.Print10()
		test.Errorf("5: Does not classify clause", pb5)
	}

	pb6 := CreatePB([]int64{1, 1, 1, 1, 1}, 4)
	pb6.Typ = Equal

	t = Translate(&pb6)

	if t.Typ != ExactlyOne {
		test.Errorf("6: Does not classify ExactlyOne", pb6)
	}

}
