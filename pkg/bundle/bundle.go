package bundle

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed builtin/*
var builtinFS embed.FS

// Bundle holds the raw template and config content for a document type.
type Bundle struct {
	Main   string // content of main.html
	Cover  string // content of cover.html, empty if absent
	Config []byte // content of default.yml, nil if absent
}

// Load resolves a bundle by name (built-in) or directory path (local).
func Load(nameOrPath string) (*Bundle, error) {
	if isDir(nameOrPath) {
		return loadDir(nameOrPath)
	}
	return loadBuiltin(nameOrPath)
}

func loadBuiltin(name string) (*Bundle, error) {
	prefix := "builtin/" + name

	mainData, err := builtinFS.ReadFile(prefix + "/main.html")
	if err != nil {
		return nil, fmt.Errorf("built-in bundle %q not found", name)
	}

	b := &Bundle{Main: string(mainData)}

	if data, err := builtinFS.ReadFile(prefix + "/cover.html"); err == nil {
		b.Cover = string(data)
	}
	if data, err := builtinFS.ReadFile(prefix + "/default.yml"); err == nil {
		b.Config = data
	}

	return b, nil
}

func loadDir(path string) (*Bundle, error) {
	mainData, err := os.ReadFile(filepath.Join(path, "main.html"))
	if err != nil {
		return nil, fmt.Errorf("read bundle main.html in %q: %w", path, err)
	}

	b := &Bundle{Main: string(mainData)}

	if data, err := os.ReadFile(filepath.Join(path, "cover.html")); err == nil {
		b.Cover = string(data)
	}
	if data, err := os.ReadFile(filepath.Join(path, "default.yml")); err == nil {
		b.Config = data
	}

	return b, nil
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
