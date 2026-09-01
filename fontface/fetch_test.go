package fontface

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestIsVariableFont(t *testing.T) {
	t.Run("fvar via arbitrary tag encoding", func(t *testing.T) {
		if !isVariableFont(buildWoff2(t, "fvar")) {
			t.Error("expected fvar to be detected")
		}
	})

	t.Run("fvar via known-tag index", func(t *testing.T) {
		idx := indexOf(t, "fvar")
		if !isVariableFont(buildWoff2KnownTag(t, idx)) {
			t.Error("expected fvar to be detected via its known-tag index")
		}
	})

	t.Run("an ordinary table is not variable", func(t *testing.T) {
		if isVariableFont(buildWoff2(t, "cmap")) {
			t.Error("cmap should not be reported as variable")
		}
	})

	t.Run("known tag other than fvar is not variable", func(t *testing.T) {
		idx := indexOf(t, "cmap")
		if isVariableFont(buildWoff2KnownTag(t, idx)) {
			t.Error("cmap (known-tag) should not be reported as variable")
		}
	})

	t.Run("not a woff2 file", func(t *testing.T) {
		if isVariableFont([]byte("not a font at all")) {
			t.Error("garbage input should not be reported as variable")
		}
	})

	t.Run("too short to have a header", func(t *testing.T) {
		if isVariableFont([]byte("wOF2")) {
			t.Error("a truncated header should not be reported as variable")
		}
	})

	t.Run("empty input", func(t *testing.T) {
		if isVariableFont(nil) {
			t.Error("nil input should not be reported as variable")
		}
	})
}

// indexOf returns tag's position in woff2KnownTags, failing the test if tag
// is not one of the fixed table names the WOFF2 spec assigns an index to.
func indexOf(t *testing.T, tag string) byte {
	t.Helper()
	for i, k := range woff2KnownTags {
		if k == tag {
			return byte(i)
		}
	}
	t.Fatalf("%q is not a known WOFF2 table tag", tag)
	return 0
}

// fontFaceServer serves a webfont stylesheet whose @font-face rules point back
// at itself, plus one file per (weight, style) group.
//
// It cannot exercise the Google Fonts variable-to-static swap in
// FetchStylesheet: that path only triggers for a stylesheet URL whose host is
// literally fonts.googleapis.com, which a test server's URL never is. That
// swap is covered by isGoogleFontsHost and googleFontsSingleWeightURL directly
// (the two pure pieces it is built from) and was verified end to end against
// the real API during development.
type fontFaceServer struct {
	*httptest.Server
	variable bool // whether the multi-weight stylesheet's files carry fvar
}

func newFontFaceServer(t *testing.T, variable bool) *fontFaceServer {
	t.Helper()
	s := &fontFaceServer{variable: variable}
	mux := http.NewServeMux()

	mux.HandleFunc("/css", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `
@font-face {
  font-family: 'Test';
  font-style: normal;
  font-weight: 400;
  src: url(%[1]s/font/400) format('woff2');
}
@font-face {
  font-family: 'Test';
  font-style: normal;
  font-weight: 600;
  src: url(%[1]s/font/600) format('woff2');
}`, s.URL)
	})

	mux.HandleFunc("/font/", func(w http.ResponseWriter, r *http.Request) {
		tag := "cmap"
		if s.variable {
			tag = "fvar"
		}
		w.Write(buildWoff2(t, tag))
	})

	s.Server = httptest.NewServer(mux)
	t.Cleanup(s.Close)
	return s
}

func TestFetchStylesheetStaticFiles(t *testing.T) {
	srv := newFontFaceServer(t, false)
	dir := t.TempDir()

	fetched, err := FetchStylesheet(srv.URL+"/css", dir, "test")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(fetched) != 2 {
		t.Fatalf("got %d files, want 2", len(fetched))
	}

	weights := map[int]bool{}
	for _, f := range fetched {
		weights[f.Weight] = true
		if _, err := os.Stat(f.Path); err != nil {
			t.Errorf("file not written: %v", err)
		}
		if filepath.Dir(f.Path) != dir {
			t.Errorf("Path = %q, want it inside %q", f.Path, dir)
		}
	}
	if !weights[400] || !weights[600] {
		t.Errorf("expected weights 400 and 600, got %+v", fetched)
	}
}

func TestFetchStylesheetNonGoogleVariableFontIsKept(t *testing.T) {
	// A test server is never fonts.googleapis.com, so a variable file from it
	// cannot be swapped for a static instance; FetchStylesheet should still
	// succeed and hand back the (variable) file rather than erroring out.
	srv := newFontFaceServer(t, true)
	dir := t.TempDir()

	fetched, err := FetchStylesheet(srv.URL+"/css", dir, "test")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(fetched) != 2 {
		t.Fatalf("got %d files, want 2", len(fetched))
	}
	for _, f := range fetched {
		data, err := os.ReadFile(f.Path)
		if err != nil {
			t.Fatal(err)
		}
		if !isVariableFont(data) {
			t.Errorf("expected %q to still carry fvar (no non-Google swap path)", f.Path)
		}
	}
}

func TestFetchStylesheetNoDownloadableFaces(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/css", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `@font-face { font-family: 'X'; src: url(data:font/woff2;base64,AAAA) format('woff2'); }`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	if _, err := FetchStylesheet(srv.URL+"/css", t.TempDir(), "test"); err == nil {
		t.Error("expected an error when no @font-face has a downloadable src")
	}
}

func TestFetchStylesheetBadURL(t *testing.T) {
	if _, err := FetchStylesheet("http://127.0.0.1:1/css", t.TempDir(), "test"); err == nil {
		t.Error("expected an error for an unreachable stylesheet URL")
	}
}

func TestFetchStylesheetNon200(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/css", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	if _, err := FetchStylesheet(srv.URL+"/css", t.TempDir(), "test"); err == nil {
		t.Error("expected an error for a non-200 stylesheet response")
	}
}
