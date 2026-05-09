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

	pagedJSTag := "<script>\n" + string(pagedJS) + "\n</script>"
	html = strings.Replace(html, "</head>", pagedJSConfig+pagedJSTag+"</head>", 1)

	tmp, err := os.CreateTemp(baseDir, "templify-*.html")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(html); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	tmp.Close()

	path, ok := launcher.LookPath()
	if !ok {
		path, err = launcher.NewBrowser().Get()
		if err != nil {
			return fmt.Errorf("download chromium: %w", err)
		}
	}

	u, err := launcher.New().Bin(path).Headless(true).Launch()
	if err != nil {
		return fmt.Errorf("launch browser: %w", err)
	}

	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()

	absHTML, err := filepath.Abs(tmp.Name())
	if err != nil {
		return fmt.Errorf("resolve temp path: %w", err)
	}
	p, err := browser.Page(proto.TargetCreateTarget{URL: "file://" + absHTML})
	if err != nil {
		return fmt.Errorf("create page: %w", err)
	}

	if err := p.WaitLoad(); err != nil {
		return fmt.Errorf("wait load: %w", err)
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
		return fmt.Errorf("wait paged.js: %w", err)
	}

	reader, err := p.PDF(&proto.PagePrintToPDF{
		PrintBackground:   true,
		PreferCSSPageSize: true,
	})
	if err != nil {
		return fmt.Errorf("print PDF: %w", err)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("read PDF: %w", err)
	}

	return os.WriteFile(outputPath, data, 0644)
}
