package generic

import "strings"

// MatchInsensitive returns true if methodName starts with methodType (case insensitive) and is at least one character longer
func MatchInsensitive(methodName, methodType string) bool {
	nameLength := len(methodName)
	typeLength := len(methodType)
	return nameLength > typeLength && strings.ToLower(methodName[:typeLength]) == strings.ToLower(methodType)
}
