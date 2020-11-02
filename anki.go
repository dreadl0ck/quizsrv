package main

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/flimzy/anki"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func ankiImport() {
	// TODO: make configurable
	files, err := ioutil.ReadDir("courses/inr/Anki")
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {

		var q []*question

		if f.IsDir() || !(filepath.Ext(f.Name()) == ".apkg") {
			continue
		}

		apkg, err := anki.ReadFile(filepath.Join("courses/inr/Anki", f.Name()))
		if err != nil {
			log.Fatal(err)
		}
		defer apkg.Close()

		os.MkdirAll("files/anki", 0700)
		for _, fn := range apkg.ListFiles() {
			data, err := apkg.ReadMediaFile(fn)
			if err != nil {
				log.Fatal(err)
			}

			fi, err := os.Create(filepath.Join("files/anki", fn))
			if err != nil {
				log.Fatal(err)
			}
			defer fi.Close()

			fi.Write(data)

			fmt.Println("extracted", fn)
		}

		notes, err := apkg.Notes()
		if err != nil {
			log.Fatal(err)
		}
		for {
			notes.Next()
			n, err := notes.Note()
			if err != nil {
				fmt.Println(err)
				break
			}

			re := regexp.MustCompile("\\.([A-Z]+)")
			r := strings.NewReplacer("\\n", "\n", "</div>", "", "<div>", "", "&gt;", ">", "&lt;", "<", ". ", ".\n", "<br />", "\n", "&nbsp;", " ", "<br>", "\n", "- ", "\n- ", "src=\"paste-", "src=\"/files/anki/paste-")
			f := func(in string) string {
				return r.Replace(
					re.ReplaceAllStringFunc(in, func(m string) string {
						return strings.Replace(m, ".", ".\n", 1)
					}),
				)
			}

			q = append(q, &question{
				Question: f(n.FieldValues[0]),
				Answer:   f(n.FieldValues[1]),
			})

			if len(n.FieldValues) != 2 {
				spew.Dump(n)
			}
		}

		inrReplacer := strings.NewReplacer("INR-", "", "INR_", "", ".apkg", ".yml")
		f, err := os.Create(inrReplacer.Replace("courses/inr/Anki/" + f.Name()))
		if err != nil {
			log.Fatal(err)
		}

		data, err := yaml.Marshal(q)
		if err != nil {
			log.Fatal(err)
		}

		f.Write(data)

		err = f.Close()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("done with", f.Name())
	}
}
