package proxy

import (
	"fmt"
	"reflect"
)

func argsMatch(expected, found reflect.Type) (bool, error) {
	if expected.NumIn() != found.NumIn() || expected.NumOut() != found.NumOut() {
		return false, fmt.Errorf("api and server argument/return counts do not match")
	}
	for j := 0; j < expected.NumIn(); j++ {
		if expected.In(j) != found.In(j) {
			return false, fmt.Errorf("api argument %s is not the same type as server argument %s", expected.In(j).Name(), found.In(j).Name())
		}
	}
	for j := 0; j < expected.NumOut(); j++ {
		if expected.Out(j) != found.Out(j) {
			return false, fmt.Errorf("api return %s is not the same type as server return %s", expected.Out(j).Name(), found.Out(j).Name())
		}
	}
	return true, nil
}
