package version

import (
	"testing"
)

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Fatal("version must not be empty")
	}
}
