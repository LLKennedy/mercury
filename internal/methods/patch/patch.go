// Package patch handles HTTP PATCH method parsing
package patch

import "github.com/LLKennedy/httpgrpc/internal/methods/generic"

// Handler handles PATCH methods
type Handler struct{}

// Match returns true if a method name begins with "Patch" and contains at least one further character
func Match(methodName string) bool {
	return generic.MatchInsensitive(methodName, "PATCH")
}
