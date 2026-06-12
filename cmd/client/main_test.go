package main

import (
	"reflect"
	"testing"
)

func TestSplitCSV(t *testing.T) {
	got := splitCSV("1.1.1.1, 8.8.8.8,,9.9.9.9")
	want := []string{"1.1.1.1", "8.8.8.8", "9.9.9.9"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("splitCSV() = %#v, want %#v", got, want)
	}
}
