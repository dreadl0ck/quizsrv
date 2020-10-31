package main

// YAML types

type courseInfo struct {
	Name string
	Description string
	Quotes  map[string]string
	Exam map[string]int
}

type Data struct {
	Slides     map[string]string
	Courses    map[string]*course
}

type course struct {
	Categories map[string][]question
	Info *courseInfo
}

type question struct {
	Question string `yaml:"question"`
	Answer   string `yaml:"answer"`
}
