package proxy

import (
	"fmt"
	"reflect"
)

func validateArgs(expected, found reflect.Type) error {
	expectedInLen := expected.NumIn()
	expectedOutLen := expected.NumOut()
	foundInLen := found.NumIn()
	foundOutLen := found.NumOut()
	if expectedInLen != foundInLen || expectedOutLen != foundOutLen {
		return fmt.Errorf("api and server argument/return counts do not match")
	}
	if expectedInLen == 0 || !isStructPtr(expected.In(0)) || !isStructPtr(found.In(0)) {
		return fmt.Errorf("no receiver")
	}
	for j := 1; j < expectedInLen; j++ {
		expectedIn := expected.In(j)
		foundIn := found.In(j)
		if !typesMatch(expectedIn, foundIn) {
			return fmt.Errorf("api and server arguments mismatch: %s vs %s", expectedIn.Kind(), foundIn.Kind())
		}
	}
	for j := 0; j < expectedOutLen; j++ {
		expectedOut := expected.Out(j)
		foundOut := found.Out(j)
		if !typesMatch(expectedOut, foundOut) {
			return fmt.Errorf("api and server returns mismatch: %s vs %s", expectedOut.Kind(), foundOut.Kind())
		}
	}
	return nil
}

func typesMatch(expected, found reflect.Type) bool {
	for expected.Kind() == reflect.Ptr {
		if found.Kind() != reflect.Ptr {
			return false
		}
		expected = expected.Elem()
		found = found.Elem()
	}
	return expected == found
}

// isStructPtr returns true if the pointer stack exists and resolves to a struct
func isStructPtr(in reflect.Type) bool {
	for in.Kind() == reflect.Ptr {
		in = in.Elem()
		if in.Kind() == reflect.Struct {
			return true
		}
	}
	return false
}
