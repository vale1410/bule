package translation

import (
	"fmt"
	"github.com/vale1410/bule/constraints"
	"testing"
)

func TestCategorize(test *testing.T) {
	fmt.Println("TestCategorize")

	pb1 := constraints.CreatePB([]int64{1, 1, 1}, 1)
	t := Categorize(&pb1)
	if t.Typ != AtMostOne {
		pb1.Print10()
		test.Errorf("1: Does not classify atmostOne")
	}

	pb2 := constraints.CreatePB([]int64{1, 1, 1}, 1)
	pb2.Typ = constraints.AtLeast
	t = Categorize(&pb2)
	if t.Typ != Clause {
		pb2.Print10()
		test.Errorf("2: Does not classify a clause")
	}

	pb3 := constraints.CreatePB([]int64{1, 1, 1}, 1)
	pb3.Typ = constraints.Equal
	t = Categorize(&pb3)

	if t.Typ != ExactlyOne {
		pb3.Print10()
		test.Errorf("3: Does not classify ExactlyOne")
	}

	pb4 := constraints.CreatePB([]int64{1, 1, -1}, 0)
	pb4.Typ = constraints.Equal
	t = Categorize(&pb4)

	if t.Typ != ExactlyOne {
		pb4.Print10()
		test.Errorf("4: Does not classify ExactlyOne")
	}

	pb5 := constraints.CreatePB([]int64{-3, 3, -3}, 0)
	pb5.Typ = constraints.AtMost
	t = Categorize(&pb5)

	if t.Typ != Clause { // should be different
		pb5.Print10()
		test.Errorf("5: Does not classify clause", pb5)
	}

	pb6 := constraints.CreatePB([]int64{1, 1, 1, 1, 1}, 4)
	pb6.Typ = constraints.Equal

	t = Categorize(&pb6)

	if t.Typ != ExactlyOne {
		test.Errorf("6: Does not classify ExactlyOne", pb6)
	}

}
