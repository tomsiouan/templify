package main

import "testing"

func TestResolveVersion(t *testing.T) {
	t.Run("uses the ldflags-injected version when set", func(t *testing.T) {
		orig := version
		defer func() { version = orig }()

		version = "v1.2.3"
		if got := resolveVersion(); got != "v1.2.3" {
			t.Errorf("resolveVersion() = %q, want v1.2.3", got)
		}
	})

	t.Run("falls back to dev with no ldflags and no module version", func(t *testing.T) {
		orig := version
		defer func() { version = orig }()

		version = ""
		// go test builds without -ldflags and without a tagged module
		// version, so debug.ReadBuildInfo() reports "(devel)" here — this
		// exercises the same fallback path a plain `go build .` takes.
		if got := resolveVersion(); got == "" {
			t.Error("resolveVersion() returned empty, want a non-empty fallback")
		}
	})
}
