package bundle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("loads builtin bundle by name", func(t *testing.T) {
		b, err := Load("report")
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if b.Main == "" {
			t.Error("expected non-empty Main")
		}
	})

	t.Run("returns error for unknown builtin name", func(t *testing.T) {
		_, err := Load("nonexistent")
		if err == nil {
			t.Error("expected error for unknown bundle name")
		}
	})

	t.Run("loads bundle from directory", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "main.html"), []byte("<html>main</html>"), 0o644); err != nil {
			t.Fatal(err)
		}
		b, err := Load(dir)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if b.Main != "<html>main</html>" {
			t.Errorf("Main = %q", b.Main)
		}
	})
}

func TestLoadBuiltin(t *testing.T) {
	t.Run("report bundle has main", func(t *testing.T) {
		b, err := loadBuiltin("report")
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if b.Main == "" {
			t.Error("expected non-empty Main")
		}
	})

	t.Run("report bundle has cover", func(t *testing.T) {
		b, err := loadBuiltin("report")
		if err != nil {
			t.Fatal(err)
		}
		if b.Cover == "" {
			t.Error("expected non-empty Cover for report bundle")
		}
	})

	t.Run("invoice bundle has config", func(t *testing.T) {
		b, err := loadBuiltin("invoice")
		if err != nil {
			t.Fatal(err)
		}
		if len(b.Config) == 0 {
			t.Error("expected non-nil Config for invoice bundle")
		}
	})

	t.Run("unknown name returns error", func(t *testing.T) {
		_, err := loadBuiltin("nonexistent")
		if err == nil {
			t.Error("expected error")
		}
	})
}

func TestLoadDir(t *testing.T) {
	t.Run("loads main.html from directory", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "main.html"), []byte("<main>"), 0o644)
		b, err := loadDir(dir)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if b.Main != "<main>" {
			t.Errorf("Main = %q", b.Main)
		}
	})

	t.Run("loads optional cover and config", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "main.html"), []byte("<main>"), 0o644)
		os.WriteFile(filepath.Join(dir, "cover.html"), []byte("<cover>"), 0o644)
		os.WriteFile(filepath.Join(dir, "default.yml"), []byte("page:\n  size: A4\n"), 0o644)
		b, err := loadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		if b.Cover != "<cover>" {
			t.Errorf("Cover = %q", b.Cover)
		}
		if len(b.Config) == 0 {
			t.Error("expected non-empty Config")
		}
	})

	t.Run("missing main.html returns error", func(t *testing.T) {
		dir := t.TempDir()
		_, err := loadDir(dir)
		if err == nil {
			t.Error("expected error for missing main.html")
		}
	})

	t.Run("absent optional files leave fields empty", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "main.html"), []byte("<main>"), 0o644)
		b, err := loadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		if b.Cover != "" {
			t.Errorf("Cover should be empty, got %q", b.Cover)
		}
		if b.Config != nil {
			t.Errorf("Config should be nil, got %v", b.Config)
		}
	})
}

func TestIsDir(t *testing.T) {
	t.Run("existing directory returns true", func(t *testing.T) {
		if !isDir(t.TempDir()) {
			t.Error("expected true for temp dir")
		}
	})

	t.Run("non-existent path returns false", func(t *testing.T) {
		if isDir("/nonexistent/path") {
			t.Error("expected false for non-existent path")
		}
	})

	t.Run("file returns false", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "file.txt")
		os.WriteFile(f, []byte("x"), 0o644)
		if isDir(f) {
			t.Error("expected false for file")
		}
	})
}
