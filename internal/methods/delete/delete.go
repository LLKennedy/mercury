// Package delete handles HTTP DELETE method parsing
package delete

import "github.com/LLKennedy/httpgrpc/internal/methods/generic"

// Handler handles DELETE methods
type Handler struct{}

// Match returns true if a method name begins with "Delete" and contains at least one further character
func Match(methodName string) bool {
	return generic.MatchInsensitive(methodName, "DELETE")
}
