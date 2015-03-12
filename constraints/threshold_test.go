package constraints

import (
	"fmt"
	"testing"
)

func TestCategorize(test *testing.T) {
	fmt.Println("TestCategorize")

	pb1 := CreatePB([]int64{1, 1, 1}, 1)
	pb1.Cardinality()

}
