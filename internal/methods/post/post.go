// Package post handles HTTP POST method parsing
package post

import "github.com/LLKennedy/httpgrpc/internal/methods/generic"

// Handler handles POST methods
type Handler struct{}

// Match returns true if a method name begins with "Post" and contains at least one further character
func Match(methodName string) bool {
	return generic.MatchInsensitive(methodName, "POST")
}
