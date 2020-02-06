package proxy

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type exposedThingA struct{}

func (a *exposedThingA) PostDoThing(context.Context, *thingAIn) (*thingAOut, error) {
	return &thingAOut{
		res: false,
	}, nil
}

type unrelatedThing struct{}

func (u *unrelatedThing) Thing(x, y int) bool {
	return false
}

func Test_validateMethod(t *testing.T) {
	tests := []struct {
		name            string
		apiMethod       reflect.Method
		serverType      reflect.Type
		expectedErr     string
		expectedType    string
		expectedName    string
		expectedPattern apiMethodPattern
	}{
		{
			name:            "matching",
			apiMethod:       reflect.TypeOf(&exposedThingA{}).Method(0),
			serverType:      reflect.TypeOf(&thingA{}),
			expectedType:    "POST",
			expectedName:    "DoThing",
			expectedPattern: apiMethodPatternStructStruct,
		},
		{
			name:        "no method at the start",
			apiMethod:   reflect.TypeOf(&thingA{}).Method(0),
			serverType:  reflect.TypeOf(&thingA{}),
			expectedErr: "DoThing does not begin with a valid HTTP method",
		},
		{
			name:        "no matching server method",
			apiMethod:   reflect.TypeOf(&exposedThingA{}).Method(0),
			serverType:  reflect.TypeOf(&unrelatedThing{}),
			expectedErr: "server is missing method DoThing",
		},
		{
			name:        "invalid matching server method",
			apiMethod:   reflect.TypeOf(&exposedThingA{}).Method(0),
			serverType:  reflect.TypeOf(&thingB{}),
			expectedErr: "validation of PostDoThing to DoThing mapping: argments mismatch in position 0: interface vs int",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotName, gotPattern, err := validateMethod(tt.apiMethod, tt.serverType)
			if tt.expectedErr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedType, gotType)
				assert.Equal(t, tt.expectedName, gotName)
				assert.Equal(t, tt.expectedPattern, gotPattern)
			} else {
				assert.EqualError(t, err, tt.expectedErr)
			}
		})
	}
}
