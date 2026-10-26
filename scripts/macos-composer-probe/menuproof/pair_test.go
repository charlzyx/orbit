package menuproof

import (
	"reflect"
	"testing"
)

func TestNativePairConstructionFailurePostsNothing(t *testing.T) {
	for _, pair := range [][2]uintptr{{0, 22}, {11, 0}, {0, 0}} {
		var constructed []int
		var posted []uintptr
		receipts := 0
		e := postPair(func(i int) uintptr { constructed = append(constructed, i); return pair[i] }, func(p uintptr) { posted = append(posted, p) }, func() { receipts++ })
		if e == nil || len(posted) != 0 || receipts != 0 || !reflect.DeepEqual(constructed, []int{0, 1}) {
			t.Fatal(pair, e, constructed, posted, receipts)
		}
	}
}
func TestNativePairSuccessPostsTwoInOrderThenReceipt(t *testing.T) {
	var operations []string
	e := postPair(func(i int) uintptr {
		if i == 0 {
			operations = append(operations, "make-down")
			return 11
		}
		operations = append(operations, "make-up")
		return 22
	}, func(p uintptr) {
		if p == 11 {
			operations = append(operations, "post-down")
		} else if p == 22 {
			operations = append(operations, "post-up")
		} else {
			t.Fatal(p)
		}
	}, func() { operations = append(operations, "receipt") })
	want := []string{"make-down", "make-up", "post-down", "post-up", "receipt"}
	if e != nil || !reflect.DeepEqual(operations, want) {
		t.Fatal(e, operations)
	}
}
