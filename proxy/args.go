package proxy

import (
	"context"
	"fmt"
	"reflect"
)

func validateArgs(expected, found reflect.Type) error {
	expectedInLen := expected.NumIn()
	expectedOutLen := expected.NumOut()
	foundInLen := found.NumIn()
	foundOutLen := found.NumOut()
	expectedIn := []reflect.Type{}
	for i := 0; i < expectedInLen; i++ {
		expectedIn = append(expectedIn, expected.In(i))
	}
	expectedOut := []reflect.Type{}
	for i := 0; i < expectedOutLen; i++ {
		expectedOut = append(expectedOut, expected.Out(i))
	}
	foundIn := []reflect.Type{}
	for i := 0; i < foundInLen; i++ {
		foundIn = append(foundIn, found.In(i))
	}
	foundOut := []reflect.Type{}
	for i := 0; i < foundOutLen; i++ {
		foundOut = append(foundOut, found.Out(i))
	}
	if expectedInLen < 2 || foundInLen < 2 {
		return fmt.Errorf("cannot exclude receiver from argument checks if receiver is the only argument: expected >= 2 input argments, found %d and %d", expectedInLen, foundInLen)
	}
	// Don't check receivers, those don't have to be the same type
	err := typesMatch(expectedIn[1:], foundIn[1:])
	if err != nil {
		return err
	}
	err = typesMatch(expectedOut, foundOut)
	return err
}

func typesMatch(expected, found []reflect.Type) error {
	if len(expected) != len(found) {
		return fmt.Errorf("argument lengths did not match: expected %d but found %d", len(expected), len(found))
	}
	for i := range expected {
		if expected[i].Kind() != found[i].Kind() {
			return fmt.Errorf("argments mismatch in position %d: %s vs %s", i, expected[i].Kind(), found[i].Kind())
		}
	}
	return nil
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

func getPattern(in reflect.Type) apiMethodPattern {
	if !(expected.In(1).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) && isStructPtr(expected.In(2))) {
		return fmt.Errorf("expected (context, *request), got (%s, %s)", expected.In(1).Kind(), expected.In(2).Kind())
	}
}
