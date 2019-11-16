// Package trace handles HTTP TRACE method parsing
package trace

import "github.com/LLKennedy/httpgrpc/internal/methods/generic"

// Handler handles TRACE methods
type Handler struct{}

// Match returns true if a method name begins with "Trace" and contains at least one further character
func Match(methodName string) bool {
	return generic.MatchInsensitive(methodName, "TRACE")
}
