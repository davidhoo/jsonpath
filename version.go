package jsonpath

// Version is the current version of jsonpath
const Version = "1.0.4"

// VersionWithPrefix returns the version with v prefix
func VersionWithPrefix() string {
	return "v" + Version
}
