// Package connect handles HTTP CONNECT method parsing
package connect

import "github.com/LLKennedy/httpgrpc/internal/methods/generic"

// Handler handles CONNECT methods
type Handler struct{}

// Match returns true if a method name begins with "Connect" and contains at least one further character
func Match(methodName string) bool {
	return generic.MatchInsensitive(methodName, "CONNECT")
}
