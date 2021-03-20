package codegen

import "strings"

type filename struct {
	pathParts            []string
	name                 string
	fullWithoutExtension string
}

// Takes input like "sampleproto/test.proto" and splits it into path components and the final name stripped of the .proto extension
// This works even on windows systems where the path may be passed with a backslash, since protoc fixes this for us
//
// It will break if the file name includes ".proto" as part of the actual name, e.g. "test.proto.stuff.proto" will become "test" and discard the rest
// That's a pretty strange naming pattern though, and for now I'm OK with not supporting it.
func filenameFromProto(in string) (out filename) {
	out.fullWithoutExtension = strings.Split(in, ".proto")[0]
	// Split on all slashes first
	parts := strings.Split(in, "/")
	// Name will always be the last component, whether that's the 0th or 500th, and strings.Split always returns at least 1 element
	lastPart := parts[len(parts)-1]
	if len(parts) > 1 {
		// More than one part means leading path elements, store them without the final name
		out.pathParts = parts[:len(parts)-1]
	}
	// Strip the extension from the filename
	out.name = strings.Split(lastPart, ".proto")[0]
	return
}
