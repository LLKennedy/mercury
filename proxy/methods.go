package proxy

import (
	"fmt"
	"reflect"

	"github.com/LLKennedy/httpgrpc/internal/methods"
)

func validateMethod(apiMethod reflect.Method, serverType reflect.Type) error {
	name := apiMethod.Name
	trueName, valid := methods.MatchAndStrip(name)
	if !valid {
		return fmt.Errorf("httpgrpc: %s does not begin with a valid HTTP method", name)
	}
	serverMethod, found := serverType.MethodByName(trueName)
	if !found {
		return fmt.Errorf("httpgrpc: server is missing method %s", trueName)
	}
	expectedType := apiMethod.Type
	foundType := serverMethod.Type
	err := validateArgs(expectedType, foundType)
	if err != nil {
		return fmt.Errorf("httpgrpc: api/server arguments do not match for method (%s/%s): %v", name, trueName, err)
	}
	return nil
}
