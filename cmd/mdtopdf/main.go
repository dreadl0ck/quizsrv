// SPDX-License-Identifier: GPL-3.0-only
package main

import (
	"flag"
	"log"
	"os"

	"github.com/mandolyte/mdtopdf"
)

func main() {
	input := flag.String("i", "", "input Markdown file")
	output := flag.String("o", "", "output PDF file")
	flag.Parse()
	if *input == "" || *output == "" {
		log.Fatal("both -i and -o are required")
	}
	content, err := os.ReadFile(*input)
	if err != nil {
		log.Fatal(err)
	}
	renderer := mdtopdf.NewPdfRenderer("", "", *output, "")
	if err := renderer.Process(content); err != nil {
		log.Fatal(err)
	}
}
