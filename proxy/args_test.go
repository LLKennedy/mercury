package proxy

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type thingAIn struct {
	x, y int
}

type thingAOut struct {
	res bool
}

type thingA struct{}

func (a *thingA) DoThing(context.Context, *thingAIn) (*thingAOut, error) {
	return &thingAOut{
		res: false,
	}, nil
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

type thingF struct{}

func (c *thingF) DoThing(x int, y int) bool {
	return true
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
		pattern           apiMethodPattern
	}
	tests := []test{
		{
			name: "matching perfectly",
			a:    &thingA{},
			b:    &thingA{},
		},
		{
			name: "matching, different definition",
			a:    &thingF{},
			b:    &thingC{},
		},
		{
			name:        "wrong argument",
			a:           &thingC{},
			b:           &thingB{},
			expectedErr: "argments mismatch in position 1: int vs string",
		},
		{
			name:        "no receiver on one",
			a:           &thingC{},
			b:           DoThing,
			expectedErr: "no receiver",
		},
		{
			name:        "no receiver on both",
			a:           DoThing,
			b:           DoThing,
			expectedErr: "no receiver",
		},
		{
			name:        "total mismatch of same in-length",
			a:           &thingC{},
			b:           Unrelated,
			expectedErr: "no receiver",
		},
		{
			name:        "invalid pointer arguments",
			a:           &thingE{},
			b:           &thingC{},
			expectedErr: "argments mismatch in position 0: ptr vs int",
		},
		{
			name: "valid pointer arguments",
			a:    &thingE{},
			b:    &thingE{},
		},
		{
			name:        "wrong return",
			a:           &thingC{},
			b:           &thingD{},
			expectedErr: "argments mismatch in position 0: bool vs int",
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
			err := validateArgs(aType, bType, tt.pattern)
			if tt.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedErr)
			}
		})
	}
}
