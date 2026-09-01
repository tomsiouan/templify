package renderer

import (
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

//go:embed assets/*.js
var assets embed.FS

const pagedJSConfig = `<script>
window.PagedConfig = {
    auto: true,

    // paged.js starts as soon as the DOM is interactive, which is well before a
    // webfont has arrived. Laying out against fallback metrics and then
    // reflowing once the real font lands leaves the PDF text layer in pieces:
    // copied text comes back with stray spaces and letters out of order. The
    // polyfill awaits this hook before paginating, so the layout is measured
    // against the final fonts. It also makes pagination reproducible, which it
    // is not when a font sometimes wins the race and sometimes doesn't.
    before: async () => {
        if (document.fonts && document.fonts.ready) {
            await document.fonts.ready;
        }
    },

    after: () => {
        // TOC page numbers
        document.querySelectorAll('a.toc-entry[href], a.tof-entry[href]').forEach(entry => {
            const id = entry.getAttribute('href').slice(1);
            const target = document.getElementById(id);
            if (target) {
                const page = target.closest('.pagedjs_page');
                if (page) entry.setAttribute('data-page', page.dataset.pageNumber);
            }
        });

        // Footnotes: group hidden inline spans by page, then anchor them at page bottom
        const pageFootnotes = new Map();
        document.querySelectorAll('.pagedjs_page .footnote-note').forEach(note => {
            const page = note.closest('.pagedjs_page');
            if (!page) return;
            if (!pageFootnotes.has(page)) pageFootnotes.set(page, []);
            pageFootnotes.get(page).push(note.innerHTML);
            note.remove();
        });
        pageFootnotes.forEach((items, page) => {
            const area = page.querySelector('.pagedjs_area');
            if (!area) return;
            const box = document.createElement('div');
            box.className = 'footnote-area';
            box.innerHTML = items.map(c => '<p class="fn-item">' + c + '</p>').join('');
            area.appendChild(box);
        });

        window.__pagedDone = true;
    }
};
</script>`

// ToPDF renders html to a PDF file. baseDir is the directory of the source
// markdown file; the temp HTML file is created there so relative image paths resolve.
func ToPDF(html string, outputPath string, baseDir string) error {
	pagedJS, err := assets.ReadFile("assets/paged.polyfill.js")
	if err != nil {
		return fmt.Errorf("read paged.js: %w", err)
	}

	tmpPath, cleanup, err := writeTempHTML(injectPagedJS(html, pagedJS), baseDir)
	if err != nil {
		return err
	}
	defer cleanup()

	return launchAndRender(tmpPath, outputPath)
}

// injectPagedJS inserts the paged.js config and polyfill script before </head>.
func injectPagedJS(html string, pagedJS []byte) string {
	tag := "<script>\n" + string(pagedJS) + "\n</script>"
	return strings.Replace(html, "</head>", pagedJSConfig+tag+"</head>", 1)
}

// writeTempHTML writes html to a temp file in baseDir and returns its path and a cleanup func.
func writeTempHTML(html string, baseDir string) (string, func(), error) {
	tmp, err := os.CreateTemp(baseDir, "templify-*.html")
	if err != nil {
		return "", nil, fmt.Errorf("create temp file: %w", err)
	}
	cleanup := func() { os.Remove(tmp.Name()) }
	if _, err := tmp.WriteString(html); err != nil {
		tmp.Close()
		cleanup()
		return "", nil, fmt.Errorf("write temp file: %w", err)
	}
	tmp.Close()
	return tmp.Name(), cleanup, nil
}

// launchAndRender opens tmpPath in headless Chromium, waits for paged.js, and writes a PDF.
func launchAndRender(tmpPath, outputPath string) error {
	chromiumPath, err := findChromium()
	if err != nil {
		return err
	}

	u, err := launcher.New().Bin(chromiumPath).Headless(true).Launch()
	if err != nil {
		return fmt.Errorf("launch browser: %w", err)
	}

	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()

	absHTML, err := filepath.Abs(tmpPath)
	if err != nil {
		return fmt.Errorf("resolve temp path: %w", err)
	}

	data, err := renderHTMLToPDF(browser, absHTML)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, data, 0o644)
}

// findChromium returns the path to a local Chromium binary, downloading one if needed.
func findChromium() (string, error) {
	if path, ok := launcher.LookPath(); ok {
		return path, nil
	}
	path, err := launcher.NewBrowser().Get()
	if err != nil {
		return "", fmt.Errorf("download chromium: %w", err)
	}
	return path, nil
}

// renderHTMLToPDF opens absHTML in browser, waits for paged.js layout, and returns PDF bytes.
func renderHTMLToPDF(browser *rod.Browser, absHTML string) ([]byte, error) {
	p, err := browser.Page(proto.TargetCreateTarget{URL: "file://" + absHTML})
	if err != nil {
		return nil, fmt.Errorf("create page: %w", err)
	}

	if err := p.WaitLoad(); err != nil {
		return nil, fmt.Errorf("wait load: %w", err)
	}

	if _, err = p.Eval(`async () => {
		await new Promise((resolve, reject) => {
			const start = Date.now();
			const check = () => {
				if (window.__pagedDone) { resolve(); return; }
				if (Date.now() - start > 60000) { reject(new Error('paged.js timeout')); return; }
				setTimeout(check, 100);
			};
			check();
		});
	}`); err != nil {
		return nil, fmt.Errorf("wait paged.js: %w", err)
	}

	reader, err := p.PDF(&proto.PagePrintToPDF{
		PrintBackground:   true,
		PreferCSSPageSize: true,
	})
	if err != nil {
		return nil, fmt.Errorf("print PDF: %w", err)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read PDF: %w", err)
	}

	return data, nil
}
