package proxy

import (
	"fmt"
	"reflect"
	"strings"
)

func validateMethod(apiMethod reflect.Method, serverType reflect.Type) error {
	name := apiMethod.Name
	trueName, valid := matchAndStrip(name)
	if !valid {
		return fmt.Errorf("%s does not begin with a valid HTTP method", name)
	}
	serverMethod, found := serverType.MethodByName(trueName)
	if !found {
		return fmt.Errorf("server is missing method %s", trueName)
	}
	expectedType := apiMethod.Type
	foundType := serverMethod.Type
	err := validateArgs(expectedType, foundType)
	if err != nil {
		return fmt.Errorf("validation of %s to %s mapping: %v", name, trueName, err)
	}
	return nil
}

var httpStrings = []string{
	"GET",
	"HEAD",
	"POST",
	"PUT",
	"DELETE",
	"CONNECT",
	"OPTIONS",
	"TRACE",
	"PATCH",
}

// matchAndStrip returns the method name stripped of its HTTP method and a success flag for that operation
func matchAndStrip(methodName string) (string, bool) {
	for _, httpType := range httpStrings {
		if matchInsensitive(methodName, httpType) {
			return stripInsensitive(methodName, httpType), true
		}
	}
	return "", false
}

// matchInsensitive returns true if methodName starts with httpType (case insensitive), is at least one character longer, and follows httpType with an uppercase letter (exported method)
func matchInsensitive(methodName, httpType string) bool {
	nameLength := len(methodName)
	typeLength := len(httpType)
	return nameLength > typeLength && // method name is at least one character longer than httpType
		strings.ToLower(methodName[:typeLength]) == strings.ToLower(httpType) && // method name starts with httpType
		strings.ContainsAny(methodName[typeLength:typeLength+1], "ABCDEFGHIJKLMNOPQRSTUVWXYZ") // first character after httpType in method name is uppercase alphabetical
}

// StripInsensitive returns the method name stripped of a prepending httpType, panics if that isn't possible
func stripInsensitive(methodName, httpType string) string {
	if !matchInsensitive(methodName, httpType) {
		panic(fmt.Sprintf("httpgrpc: cannot strip invalid method name/type combination %s/%s", methodName, httpType))
	}
	return methodName[len(httpType):]
}