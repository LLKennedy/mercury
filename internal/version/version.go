package version

import (
	"fmt"

	"google.golang.org/protobuf/types/pluginpb"
)

const (
	// Major is the major version of the tool
	Major = 0
	// Minor is the minor version of the tool
	Minor = 1
	// Patch is the patch version of the tool
	Patch = 2
)

// GetVersionString generates a version string based on the constants in this package
func GetVersionString() string {
	return fmt.Sprintf("v%d.%d.%d", Major, Minor, Patch)
}

// FormatProtocVersion formats a protoc version into a version string
func FormatProtocVersion(v *pluginpb.Version) string {
	base := fmt.Sprintf("v%d.%d.%d", v.GetMajor(), v.GetMinor(), v.GetPatch())
	if v.GetSuffix() != "" {
		return fmt.Sprintf("%s-%s", base, v.GetSuffix())
	}
	return base
}
