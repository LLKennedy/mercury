package proxy

import (
	"fmt"
	"reflect"
	"strings"
)

func validateMethod(apiMethod reflect.Method, serverType reflect.Type) (methodType string, procedureName string, pattern apiMethodPattern, err error) {
	name := apiMethod.Name
	httpType, trueName, valid := MatchAndStripMethodName(name)
	if !valid {
		err = fmt.Errorf("%s does not begin with a valid HTTP method", name)
		return
	}
	serverMethod, found := serverType.MethodByName(trueName)
	if !found {
		err = fmt.Errorf("server is missing method %s", trueName)
		return
	}
	expectedType := apiMethod.Type
	foundType := serverMethod.Type
	pattern = getPattern(expectedType)
	if pattern == apiMethodPatternUnknown {
		err = fmt.Errorf("method %s did not match standard GRPC patterns (stream or struct pointer in and out, plus error out where applicable)", name)
		return
	}
	err = validateArgs(expectedType, foundType, pattern)
	if err != nil {
		err = fmt.Errorf("validation of %s to %s mapping: %v", name, trueName, err)
		return
	}
	methodType = httpType
	procedureName = trueName
	return
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

// MatchAndStripMethodName returns the method name stripped of its HTTP method and a success flag for that operation
func MatchAndStripMethodName(methodName string) (string, string, bool) {
	for _, httpType := range httpStrings {
		if matchInsensitive(methodName, httpType) {
			return httpType, stripInsensitive(methodName, httpType), true
		}
	}
	return "", "", false
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
		panic(fmt.Sprintf("mercury: cannot strip invalid method name/type combination %s/%s", methodName, httpType))
	}
	return methodName[len(httpType):]
}
