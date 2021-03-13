package codegen

import (
	"strings"

	"github.com/LLKennedy/httpgrpc/internal/version"
)

type file struct {
	sourceName string
}

// String complies with the Stringer interface
func (f file) String() string {
	var builder strings.Builder
	builder.WriteString(getCodeGenmarker(version.GetVersionString(), protocVersion, f.sourceName))
	// TODO: the rest of the file
	return builder.String()
}
