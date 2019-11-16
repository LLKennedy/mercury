package proxy

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type thingA struct{}

func (a *thingA) DoThing(x, y int) bool {
	return false
}

type thingB struct{}

func (b *thingB) DoThing(x int, y string) bool {
	return false
}

type thingC struct{}

func (c *thingC) DoThing(x int, y int) bool {
	return true
}

type thingD struct{}

func (d *thingD) DoThing(x, y int) int {
	return 3
}

type thingE struct{}

func (e *thingE) DoThing(x, y *int) bool {
	return false
}

func DoThing(x, y int) bool {
	return true
}

func Unrelated(a, b, c bool) string {
	return ""
}

func TestArgsMatch(t *testing.T) {
	type test struct {
		name, expectedErr string
		a, b              interface{}
	}
	tests := []test{
		{
			name: "matching perfectly",
			a:    &thingA{},
			b:    &thingA{},
		},
		{
			name: "matching, different definition",
			a:    &thingA{},
			b:    &thingC{},
		},
		{
			name:        "wrong argument",
			a:           &thingA{},
			b:           &thingB{},
			expectedErr: "api and server arguments mismatch: int vs string",
		},
		{
			name:        "no receiver on one",
			a:           &thingA{},
			b:           DoThing,
			expectedErr: "api and server argument/return counts do not match",
		},
		{
			name:        "no receiver on both",
			a:           DoThing,
			b:           DoThing,
			expectedErr: "no receiver",
		},
		{
			name:        "total mismatch of same in-length",
			a:           &thingA{},
			b:           Unrelated,
			expectedErr: "no receiver",
		},
		{
			name:        "invalid pointer arguments",
			a:           &thingE{},
			b:           &thingA{},
			expectedErr: "api and server arguments mismatch: ptr vs int",
		},
		{
			name: "valid pointer arguments",
			a:    &thingE{},
			b:    &thingE{},
		},
		{
			name:        "wrong return",
			a:           &thingA{},
			b:           &thingD{},
			expectedErr: "api and server returns mismatch: bool vs int",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			funcA, validA := reflect.TypeOf(tt.a).MethodByName("DoThing")
			funcB, validB := reflect.TypeOf(tt.b).MethodByName("DoThing")
			var aType, bType reflect.Type
			if validA {
				aType = funcA.Type
			} else {
				aType = reflect.TypeOf(tt.a)
			}
			if validB {
				bType = funcB.Type
			} else {
				bType = reflect.TypeOf(tt.b)
			}
			err := validateArgs(aType, bType)
			if tt.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedErr)
			}
		})
	}
}
