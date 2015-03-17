package constraints

import (
	"fmt"
	"testing"
)

func TestRemoveZeros(test *testing.T) {
	fmt.Println("TestRemoveZeros")

	pb1 := CreatePB([]int64{1, 2, 3, 0, 321, 0, 0, -123, 0}, 1347)
	c := len(pb1.Entries)
	//pb1.Print10()
	pb1.RemoveZeros()
	//pb1.Print10()

	if len(pb1.Entries) != c-4 {
		test.Fail()
	}

}
