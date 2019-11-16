// Package head handles HTTP HEAD method parsing
package head

import "github.com/LLKennedy/httpgrpc/internal/methods/generic"

// Handler handles HEAD methods
type Handler struct{}

// Match returns true if a method name begins with "Head" and contains at least one further character
func Match(methodName string) bool {
	return generic.MatchInsensitive(methodName, "HEAD")
}
