package jsonpath

// Version is the current version of jsonpath
const Version = "2.0.2"

// VersionWithPrefix returns the version with v prefix
func VersionWithPrefix() string {
	return "v" + Version
}
