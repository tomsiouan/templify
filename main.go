package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tomsiouan/templify/pkg/parser"
	"github.com/tomsiouan/templify/pkg/renderer"
	"github.com/tomsiouan/templify/pkg/tmpl"
)

func main() {
	input := flag.String("input", "", "path to the input Markdown file (required)")
	output := flag.String("output", "output.pdf", "path for the generated PDF")
	template := flag.String("template", "default", "built-in template name or path to a .html file")
	flag.Parse()

	if *input == "" {
		fmt.Fprintln(os.Stderr, "error: -input is required")
		flag.Usage()
		os.Exit(1)
	}

	doc, err := parser.ParseFile(*input)
	if err != nil {
		log.Fatalf("parse: %v", err)
	}

	html, err := tmpl.Render(*template, doc)
	if err != nil {
		log.Fatalf("template: %v", err)
	}

	if err := renderer.ToPDF(html, *output); err != nil {
		log.Fatalf("render: %v", err)
	}

	fmt.Printf("generated: %s\n", *output)
}
