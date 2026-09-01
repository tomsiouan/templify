package fontface

import (
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// chromeUA makes font services hand back the same woff2 files Chromium would
// get. Google Fonts in particular serves older formats to unknown clients.
const chromeUA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) " +
	"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

// Fetched is one downloaded font file and the descriptors its @font-face rule
// carried, ready to be written back as a config.FontFace.
type Fetched struct {
	Path         string
	Weight       int
	Style        string
	UnicodeRange string
}

var (
	reFontFace    = regexp.MustCompile(`(?s)@font-face\s*\{(.*?)\}`)
	reSrcURL      = regexp.MustCompile(`url\(\s*['"]?([^'")]+)['"]?\s*\)`)
	reFontWeight  = regexp.MustCompile(`font-weight\s*:\s*([0-9]+)`)
	reFontStyle   = regexp.MustCompile(`font-style\s*:\s*([a-zA-Z]+)`)
	reFontFamily  = regexp.MustCompile(`font-family\s*:\s*['"]?([^'";]+)['"]?`)
	reUnicodeRang = regexp.MustCompile(`unicode-range\s*:\s*([^;]+)`)
)

// block is one @font-face rule parsed out of a webfont stylesheet.
type block struct {
	family       string
	weight       int
	style        string
	unicodeRange string
	srcURL       string
}

// parseFontFaceBlocks extracts every @font-face rule with a downloadable src
// from css. Rules already embedding a data URI are skipped: there is nothing
// to fetch.
func parseFontFaceBlocks(css string) []block {
	var out []block
	for _, m := range reFontFace.FindAllStringSubmatch(css, -1) {
		body := m[1]

		src := reSrcURL.FindStringSubmatch(body)
		if src == nil {
			continue
		}
		srcURL := src[1]
		if strings.HasPrefix(srcURL, "data:") {
			continue
		}

		weight := 400
		if m := reFontWeight.FindStringSubmatch(body); m != nil {
			if n, err := strconv.Atoi(m[1]); err == nil {
				weight = n
			}
		}
		style := "normal"
		if m := reFontStyle.FindStringSubmatch(body); m != nil {
			style = strings.ToLower(m[1])
		}
		family := ""
		if m := reFontFamily.FindStringSubmatch(body); m != nil {
			family = strings.TrimSpace(m[1])
		}
		unicodeRange := ""
		if m := reUnicodeRang.FindStringSubmatch(body); m != nil {
			unicodeRange = strings.TrimSpace(m[1])
		}

		out = append(out, block{
			family:       family,
			weight:       weight,
			style:        style,
			unicodeRange: unicodeRange,
			srcURL:       srcURL,
		})
	}
	return out
}

// FetchStylesheet downloads the webfont stylesheet at stylesheetURL, then
// downloads every font file it references into dir. Files are named after
// prefix, the weight, the style and their position, so repeated runs overwrite
// rather than pile up.
//
// Google Fonts hands out a variable font (one file covering a whole weight
// range) even when a stylesheet's @font-face rule is scoped to a single
// weight — the file just carries a font-weight descriptor telling the browser
// which instance to render. Chromium's PDF export mis-handles these: the text
// layer comes out with glyphs split into individual runs, and copied text is
// garbled. Requesting one weight/style per Google Fonts stylesheet reliably
// returns a static instance instead, so a variable file coming from
// fonts.googleapis.com is swapped for that before it is saved. Variable files
// from any other host are kept, with a warning: there is no general way to
// turn an arbitrary third-party variable font into a static one here.
func FetchStylesheet(stylesheetURL, dir, prefix string) ([]Fetched, error) {
	css, err := get(stylesheetURL)
	if err != nil {
		return nil, fmt.Errorf("fetch stylesheet %q: %w", stylesheetURL, err)
	}

	blocks := parseFontFaceBlocks(string(css))
	if len(blocks) == 0 {
		return nil, fmt.Errorf("no @font-face with a downloadable src found at %q", stylesheetURL)
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create %q: %w", dir, err)
	}

	// Group by (weight, style): a variable file is detected and swapped for a
	// static instance once per group, not once per unicode-range subset.
	type key struct {
		weight int
		style  string
	}
	var order []key
	groups := map[key][]block{}
	for _, b := range blocks {
		k := key{b.weight, b.style}
		if _, ok := groups[k]; !ok {
			order = append(order, k)
		}
		groups[k] = append(groups[k], b)
	}

	googleFonts := isGoogleFontsHost(stylesheetURL)
	warnedVariable := false
	fileIndex := 0
	var out []Fetched

	for _, k := range order {
		group := groups[k]

		probe, err := get(group[0].srcURL)
		if err != nil {
			return nil, fmt.Errorf("fetch font %q: %w", group[0].srcURL, err)
		}

		swapped := false
		if isVariableFont(probe) && googleFonts && group[0].family != "" {
			staticURL := googleFontsSingleWeightURL(group[0].family, k.weight, k.style)
			if staticCSS, err := get(staticURL); err == nil {
				if staticBlocks := parseFontFaceBlocks(string(staticCSS)); len(staticBlocks) > 0 {
					group = staticBlocks
					swapped = true
				}
			}
		}

		for i, b := range group {
			data := probe
			if i > 0 || swapped {
				data, err = get(b.srcURL)
				if err != nil {
					return nil, fmt.Errorf("fetch font %q: %w", b.srcURL, err)
				}
			}

			if isVariableFont(data) && !warnedVariable {
				slog.Warn("font file is a variable font; copied PDF text may come out garbled",
					"family", b.family, "weight", b.weight, "style", b.style,
					"hint", "Chromium mis-handles variable fonts in PDF export; supply a static instance for this weight/style if this happens")
				warnedVariable = true
			}

			ext := filepath.Ext(b.srcURL)
			if _, ok := mimeTypes[strings.ToLower(ext)]; !ok {
				ext = ".woff2" // Google always serves woff2 to a modern user agent
			}
			name := fmt.Sprintf("%s-%d%s-%02d%s", prefix, b.weight, styleSuffix(b.style), fileIndex, ext)
			fileIndex++

			path := filepath.Join(dir, name)
			if err := os.WriteFile(path, data, 0o644); err != nil {
				return nil, fmt.Errorf("write %q: %w", path, err)
			}

			out = append(out, Fetched{
				Path:         path,
				Weight:       b.weight,
				Style:        b.style,
				UnicodeRange: b.unicodeRange,
			})
		}
	}

	return out, nil
}

// isGoogleFontsHost reports whether rawURL points at the Google Fonts CSS2 API,
// the only host this package knows how to re-query for a static instance.
func isGoogleFontsHost(rawURL string) bool {
	u, err := url.Parse(rawURL)
	return err == nil && u.Host == "fonts.googleapis.com"
}

// googleFontsSingleWeightURL builds a Google Fonts CSS2 request scoped to one
// weight and style, which Google resolves to a static font instance instead of
// the variable file a multi-weight request returns.
func googleFontsSingleWeightURL(family string, weight int, style string) string {
	axis := fmt.Sprintf("wght@%d", weight)
	if style == "italic" {
		axis = fmt.Sprintf("ital,wght@1,%d", weight)
	}
	return fmt.Sprintf("https://fonts.googleapis.com/css2?family=%s:%s&display=swap",
		url.QueryEscape(family), axis)
}

// woff2KnownTags is the WOFF2 spec's fixed table tag list: table directory
// entries reference it by index instead of spelling the tag out.
var woff2KnownTags = []string{
	"cmap", "head", "hhea", "hmtx", "maxp", "name", "OS/2", "post", "cvt ", "fpgm",
	"glyf", "loca", "prep", "CFF ", "VORG", "EBDT", "EBLC", "gasp", "hdmx", "kern",
	"LTSH", "PCLT", "VDMX", "vhea", "vmtx", "BASE", "GDEF", "GPOS", "GSUB", "EBSC",
	"JSTF", "MATH", "CBDT", "CBLC", "COLR", "CPAL", "SVG ", "sbix", "acnt", "avar",
	"bdat", "bloc", "bsln", "cvar", "fdsc", "feat", "fmtx", "fvar", "gvar", "hsty",
	"just", "lcar", "mort", "morx", "opbd", "prop", "trak", "Zapf", "Silf", "Glat",
	"Gloc", "Feat", "Sill",
}

// isVariableFont reports whether a WOFF2 font file carries an `fvar` table,
// which marks it as a variable font. Chromium's PDF export corrupts the text
// layer for these: see the FetchStylesheet doc comment. Non-WOFF2 data (or
// anything too short to parse) is reported as not variable, since this check
// only needs to catch the Google Fonts case, which always serves WOFF2.
func isVariableFont(data []byte) bool {
	if len(data) < 48 || string(data[:4]) != "wOF2" {
		return false
	}
	numTables := int(binary.BigEndian.Uint16(data[12:14]))
	o := 48
	for i := 0; i < numTables && o < len(data); i++ {
		flags := data[o]
		o++

		idx := int(flags & 0x3f)
		var tag string
		switch {
		case idx == 0x3f:
			if o+4 > len(data) {
				return false
			}
			tag = string(data[o : o+4])
			o += 4
		case idx < len(woff2KnownTags):
			tag = woff2KnownTags[idx]
		}

		// Table entries carry a UIntBase128-encoded original length that must
		// be skipped to reach the next entry; its value itself is unused here.
		for j := 0; j < 5 && o < len(data); j++ {
			b := data[o]
			o++
			if b&0x80 == 0 {
				break
			}
		}

		if tag == "fvar" {
			return true
		}
	}
	return false
}

// styleSuffix returns the filename fragment for a font style, empty for normal.
func styleSuffix(style string) string {
	if style == "" || style == "normal" {
		return ""
	}
	return "-" + style
}

// get retrieves target with a browser user agent and returns its body.
func get(target string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", chromeUA)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %s", resp.Status)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 32<<20))
}

// FacesYAML renders fetched files as the `faces:` block to paste into a config.
// The config is not rewritten in place: yaml.v2 cannot round-trip comments, and
// these files are usually commented.
func FacesYAML(key string, fetched []Fetched, relativeTo string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s:\n  faces:\n", key)
	for _, f := range fetched {
		path := f.Path
		if rel, err := filepath.Rel(relativeTo, f.Path); err == nil {
			path = "./" + filepath.ToSlash(rel)
		}
		fmt.Fprintf(&sb, "    - file: %s\n", path)
		fmt.Fprintf(&sb, "      weight: %d\n", f.Weight)
		if f.Style != "" && f.Style != "normal" {
			fmt.Fprintf(&sb, "      style: %s\n", f.Style)
		}
		if f.UnicodeRange != "" {
			fmt.Fprintf(&sb, "      unicode_range: %q\n", f.UnicodeRange)
		}
	}
	return sb.String()
}

// Prefix returns a filename-safe prefix derived from a font family name.
func Prefix(family string) string {
	f := strings.TrimSpace(strings.ToLower(family))
	f = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			return r
		}
		return '-'
	}, f)
	f = strings.Trim(f, "-")
	if f == "" {
		return "font"
	}
	return f
}
