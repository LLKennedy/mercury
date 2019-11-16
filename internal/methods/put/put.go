// Package put handles HTTP PUT method parsing
package put

import "github.com/LLKennedy/httpgrpc/internal/methods/generic"

// Handler handles PUT methods
type Handler struct{}

// Match returns true if a method name begins with "Put" and contains at least one further character
func Match(methodName string) bool {
	return generic.MatchInsensitive(methodName, "PUT")
}
