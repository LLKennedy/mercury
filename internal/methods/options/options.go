// Package options handles HTTP OPTIONS method parsing
package options

import "github.com/LLKennedy/httpgrpc/internal/methods/generic"

// Handler handles OPTIONS methods
type Handler struct{}

// Match returns true if a method name begins with "Options" and contains at least one further character
func Match(methodName string) bool {
	return generic.MatchInsensitive(methodName, "OPTIONS")
}
