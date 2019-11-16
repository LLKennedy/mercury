// Package get handles HTTP GET method parsing
package get

import "github.com/LLKennedy/httpgrpc/internal/methods/generic"

// Handler handles GET methods
type Handler struct{}

// Match returns true if a method name begins with "Get" and contains at least one further character
func Match(methodName string) bool {
	return generic.MatchInsensitive(methodName, "GET")
}
