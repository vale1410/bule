package constraints

import (
	"fmt"
	"github.com/vale1410/bule/sat"
	"testing"
)

func TestCategorize(test *testing.T) {
	fmt.Println("TestCategorize")

	pb1 := CreatePB([]int64{1, 1, 1}, 1)
	pb1.Cardinality()

}

//func createIgnasi1() (t Threshold) {
//	weights := []int64{4, 3, 1, 1, 1, 1}
//	return createPB(weights, 5)
//}
//
//func createIgnasi2() (t Threshold) {
//	weights := []int64{7, 6, 2, 2, 2, 2, 1, 1, 1, 1, 1}
//	return createPB(weights, 12)
//}
