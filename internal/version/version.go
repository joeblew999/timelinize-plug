package version

import "runtime/debug"

var version = "dev"

// String returns the build version if embedded via -ldflags "-X".
func String() string {
	if version != "" && version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}
