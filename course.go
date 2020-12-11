package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

func courseImport() {
	root := filepath.Join(*configFolder, "courses")
	courses, err := ioutil.ReadDir(root)
	if err != nil {
		log.Fatal(err)
	}

	data = new(Data)
	data.Courses = make(map[string]*course)
	data.Slides = make(map[string]string)

	for _, file := range courses {

		if !file.IsDir() {
			continue
		}

		http.HandleFunc("/courses/"+file.Name(), courseHandler)
		cou := &course{
			Categories: make(map[string][]question),
		}
		data.Courses[strings.ToLower(file.Name())] = cou

		files, err := ioutil.ReadDir(filepath.Join(root, file.Name()))
		if err != nil {
			log.Fatal(err)
		}

		for _, f := range files {

			if f.IsDir() {
				continue
			}

			if f.Name() == "course.yml" {
				c, err := ioutil.ReadFile(filepath.Join(root, file.Name(), f.Name()))
				if err != nil {
					log.Fatal(err)
				}

				var i = new(courseInfo)

				err = yaml.UnmarshalStrict(c, &i)
				if err != nil {
					log.Fatal(err, " file: ", f.Name())
				}

				cou.Info = i
				continue
			}

			category := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))

			switch filepath.Ext(f.Name()) {
			case ".yml", ".yaml":
				c, err := ioutil.ReadFile(filepath.Join(root, file.Name(), f.Name()))
				if err != nil {
					log.Fatal(err)
				}

				var questions []question

				err = yaml.UnmarshalStrict(c, &questions)
				if err != nil {
					log.Fatal(err, " file: ", f.Name())
				}

				// TODO: categories have to be unique for now
				if _, ok := cou.Categories[category]; ok {
					log.Fatal("categories have to be unique for now:", category)
				}
				cou.Categories[category] = questions

				fmt.Println("loaded category", category, len(questions), "for course", file.Name())

				http.HandleFunc("/courses/"+file.Name()+"/"+category, quiz)
			//case ".pdf":
			//	data.Slides[category] = f.Name()
			//	http.HandleFunc("/cia/"+category, pdf)
			default:
				log.Println("unsupported format", filepath.Ext(f.Name()), " file: ", f.Name())
			}
		}
	}
}
