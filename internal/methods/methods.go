// Package methods provides general-purpose parsing of HTTP methods
package methods

import (
	"fmt"
	"strings"
)

var methodStrings = []string{
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

// MatchAndStrip returns the method name stripped of its HTTP method and a success flag for that operation
func MatchAndStrip(methodName string) (string, bool) {
	for _, methodType := range methodStrings {
		if MatchInsensitive(methodName, methodType) {
			return StripInsensitive(methodName, methodType), true
		}
	}
	return "", false
}

// MatchInsensitive returns true if methodName starts with methodType (case insensitive), is at least one character longer, and follows methodType with an uppercase letter (exported method)
func MatchInsensitive(methodName, methodType string) bool {
	nameLength := len(methodName)
	typeLength := len(methodType)
	return nameLength > typeLength && // method name is at least one character longer than method type
		strings.ToLower(methodName[:typeLength]) == strings.ToLower(methodType) && // method name starts with method type
		strings.ContainsAny(methodName[typeLength:typeLength+1], "ABCDEFGHIJKLMNOPQRSTUVWXYZ") // first character after method type in method name is uppercase alphabetical
}

// StripInsensitive returns the method name stripped of a prepending method type, panics if that isn't possible
func StripInsensitive(methodName, methodType string) string {
	if !MatchInsensitive(methodName, methodType) {
		panic(fmt.Sprintf("httpgrpc: cannot strip invalid method name/type combination %s/%s", methodName, methodType))
	}
	return methodName[len(methodType):]
}
