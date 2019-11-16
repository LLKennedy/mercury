package proxy

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type exposedThingA struct{}

func (a *exposedThingA) PostDoThing(x, y int) bool {
	return true
}

type unrelatedThing struct{}

func (u *unrelatedThing) Thing(x, y int) bool {
	return false
}

func Test_validateMethod(t *testing.T) {
	tests := []struct {
		name        string
		apiMethod   reflect.Method
		serverType  reflect.Type
		expectedErr string
	}{
		{
			name:       "matching",
			apiMethod:  reflect.TypeOf(&exposedThingA{}).Method(0),
			serverType: reflect.TypeOf(&thingA{}),
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
			expectedErr: "api/server arguments do not match for method (PostDoThing/DoThing): api and server arguments mismatch: int vs string",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMethod(tt.apiMethod, tt.serverType)
			if tt.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedErr)
			}
		})
	}
}
