package jsonpath

import (
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}

	// 版本号格式应该是 x.y.z
	parts := strings.Split(Version, ".")
	if len(parts) != 3 {
		t.Errorf("Version should be in format x.y.z, got %s", Version)
	}

	for _, part := range parts {
		if part == "" {
			t.Errorf("Version number parts should not be empty, got %s", Version)
		}
	}
}

func TestVersionWithPrefix(t *testing.T) {
	version := VersionWithPrefix()

	// 应该以 "v" 开头
	if !strings.HasPrefix(version, "v") {
		t.Errorf("VersionWithPrefix() should start with 'v', got %s", version)
	}

	// 去掉 "v" 后应该等于 Version
	if strings.TrimPrefix(version, "v") != Version {
		t.Errorf("VersionWithPrefix() without 'v' should equal Version, got %s, want %s",
			strings.TrimPrefix(version, "v"), Version)
	}
}
