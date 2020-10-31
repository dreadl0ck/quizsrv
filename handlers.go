package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

func quiz(w http.ResponseWriter, r *http.Request) {
	homeTemplate := template.Must(template.ParseFiles(filepath.Join(*configFolder, "pages/quiz.html")))

	err := homeTemplate.Execute(w, &TemplateData{
		//Host:       "ws://"+r.Host+"/connect", // TODO: unused
		Version:  Version,
		Category: strings.ToUpper(filepath.Base(r.URL.Path)),
		CourseID: strings.Split(r.URL.Path, "/")[2],
	})
	if err != nil {
		log.Println(err)
	}
}

func courseHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.RemoteAddr, "course")
	var (
		slides     []string
		categories = make(map[string]int)
		total      int
	)

	cou, ok := data.Courses[filepath.Base(r.RequestURI)]
	if !ok {
		http.Error(w, "invalid course name", http.StatusBadRequest)
		return
	}

	for c := range cou.Categories {
		categories[c] = len(cou.Categories[c])
		total += len(cou.Categories[c])
	}
	for c := range data.Slides {
		slides = append(slides, c)
	}

	var quote, quoteAuthor string
	for q, a := range cou.Info.Quotes {
		quote, quoteAuthor = q, a
		break
	}

	var courseNames []string
	for n := range data.Courses {
		courseNames = append(courseNames, n)
	}

	courseID := filepath.Base(r.RequestURI)
	courseTemplate := template.Must(template.ParseFiles(filepath.Join(*configFolder, "pages/course.html")))
	err := courseTemplate.Execute(w, &TemplateData{
		Categories:  categories,
		Slides:      slides,
		Version:     Version,
		Quote:       quote,
		QuoteAuthor: quoteAuthor,
		Total:       total,
		Courses:     courseNames,
		CourseID: courseID,
		CourseName: cou.Info.Name,
	})
	if err != nil {
		log.Println(err)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.RemoteAddr, "home")
	var (
		//slides     []string
		//categories = make(map[string]int)
		total      int
		//courseName = filepath.Base(r.RequestURI)
	)

	//cou, ok := data.Courses[courseName]
	//if !ok {
	//	http.Error(w, "invalid courseName", http.StatusBadRequest)
	//	return
	//}
	//
	//for c := range cou.Categories {
	//	categories[c] = len(cou.Categories[c])
	//	total += len(cou.Categories[c])
	//}
	//for c := range data.Slides {
	//	slides = append(slides, c)
	//}

	//var quote, quoteAuthor string
	//for q, a := range cou.Info.Q {
	//	quote, quoteAuthor = q, a
	//	break
	//}

	var courseNames []string
	for c := range data.Courses {
		courseNames = append(courseNames, c)
	}
	homeTemplate := template.Must(template.ParseFiles(filepath.Join(*configFolder, "pages/index.html")))
	err := homeTemplate.Execute(w, &TemplateData{
		//Slides:      slides,
		Version:     Version,
		//Quote:       quote,
		//QuoteAuthor: quoteAuthor,
		Total:       total,
		Courses:     courseNames,
	})
	if err != nil {
		log.Println(err)
	}
}
//
//func pdf(w http.ResponseWriter, r *http.Request) {
//	var (
//		slides, categories []string
//	)
//	for c := range data.Categories {
//		categories = append(categories, c)
//	}
//	for c := range data.Slides {
//		slides = append(slides, c)
//	}
//	homeTemplate := template.Must(template.ParseFiles("pages/pdf.html"))
//	err := homeTemplate.Execute(w, strings.TrimSuffix(filepath.Base(r.RequestURI), ".pdf"))
//	if err != nil {
//		log.Println(err)
//	}
//}
