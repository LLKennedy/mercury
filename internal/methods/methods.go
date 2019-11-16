// Package methods provides general-purpose parsing of HTTP methods
package methods

import (
	"fmt"
	"strings"
)

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

// MatchAndStrip returns the method name stripped of its HTTP method and a success flag for that operation
func MatchAndStrip(methodName string) (string, bool) {
	for _, httpType := range httpStrings {
		if MatchInsensitive(methodName, httpType) {
			return StripInsensitive(methodName, httpType), true
		}
	}
	return "", false
}

// MatchInsensitive returns true if methodName starts with httpType (case insensitive), is at least one character longer, and follows httpType with an uppercase letter (exported method)
func MatchInsensitive(methodName, httpType string) bool {
	nameLength := len(methodName)
	typeLength := len(httpType)
	return nameLength > typeLength && // method name is at least one character longer than httpType
		strings.ToLower(methodName[:typeLength]) == strings.ToLower(httpType) && // method name starts with httpType
		strings.ContainsAny(methodName[typeLength:typeLength+1], "ABCDEFGHIJKLMNOPQRSTUVWXYZ") // first character after httpType in method name is uppercase alphabetical
}

// StripInsensitive returns the method name stripped of a prepending httpType, panics if that isn't possible
func StripInsensitive(methodName, httpType string) string {
	if !MatchInsensitive(methodName, httpType) {
		panic(fmt.Sprintf("httpgrpc: cannot strip invalid method name/type combination %s/%s", methodName, httpType))
	}
	return methodName[len(httpType):]
}
