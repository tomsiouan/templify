package renderer

import (
	"fmt"
	"io"
	"os"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

func ToPDF(html string, outputPath string) error {
	tmp, err := os.CreateTemp("", "templify-*.html")
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

	p, err := browser.Page(proto.TargetCreateTarget{URL: "file://" + tmp.Name()})
	if err != nil {
		return fmt.Errorf("create page: %w", err)
	}
	if err := p.WaitLoad(); err != nil {
		return fmt.Errorf("wait load: %w", err)
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
