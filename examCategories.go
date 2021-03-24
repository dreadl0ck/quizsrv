package main

import (
	"encoding/json"
	"fmt"
	"github.com/dreadl0ck/cryptoutils"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func genExamCategories(w http.ResponseWriter, r *http.Request) {

	rand.Seed(time.Now().UnixNano())

	var (
		output   string
	)

	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("mkexam ALL", r.RemoteAddr, r.UserAgent())

	id, err := cryptoutils.RandomString(10)
	if err != nil {
		log.Fatal(err)
	}

	// generate exams until all questions have been used
	id = strings.TrimSuffix(id, "==")
	courseID := strings.ToLower(r.FormValue("course"))

	cou, ok := data.Courses[courseID]
	if !ok {
		http.Error(w, "invalid course id", http.StatusBadRequest)
		return
	}

	for c, _ := range cou.Categories {

		var (
			examQuestions string
			examSolutions string
			count         int
		)

		// write header
		examQuestions += "# " + strings.ToUpper(courseID) + " Test Exam" + "\n"
		examQuestions += "\n> ID: " + id + "\n\n"

		examSolutions += "# " + strings.ToUpper(courseID) + " Test Exam Solution" + "\n"
		examSolutions += "\n> ID: " + id + "\n\n"

		var (
			category = cou.Categories[strings.ToLower(c)]
			current  int
			val      = r.FormValue(c)
			flagged  []string
		)

		if len(category) == 0 {
			continue
		}

		if val != "" && val != "[]" {
			err = json.Unmarshal([]byte(val), &flagged)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("will only include", flagged, "for category", c)

			flagged = shuffle(flagged)
		}

		// write category
		examQuestions += "\n### " + strings.ToUpper(c) + "\n"
		examSolutions += "\n### " + strings.ToUpper(c) + "\n"

		if c == "history" {
			count++
			// always add a quote for history part
			for q, author := range cou.Info.Quotes {
				// skip CIA joke for exam questions
				if strings.HasPrefix(q, "CIA") {
					continue
				}
				examQuestions += strconv.Itoa(count) + ") Name the author of the following quote and explain his intentions:  \n"
				examQuestions += "    \n"
				examQuestions += "    " + q + "\n"
				examQuestions += "    \n\n"

				examSolutions += strconv.Itoa(count) + ") Name the author of the following quote and explain his intentions:  \n"
				examSolutions += "    \n"
				examSolutions += "    " + q + "\n"
				examSolutions += "    \n"
				examSolutions += "    " + author
				examSolutions += "    \n"
				break
			}
		}

		var (
			numFlagged = len(flagged)
		)

		// write questions
		for i := 0; i < len(category); i++ {

			if numFlagged > 0 {
				if i >= numFlagged {
					break
				}
				n, err := strconv.Atoi(flagged[i])
				if err == nil {
					current = n
				} else {
					fmt.Println("invalid integer in flagged values", flagged[i])
					continue
				}
			} else {
				current = i
			}

			count++

			examQuestions += strconv.Itoa(count) + ") " + fixLinks(category[current].Question)
			examQuestions += "    \n"

			examSolutions += strconv.Itoa(count) + ") " + fixLinks(category[current].Question)
			examSolutions += "    \n"

			var res string
			for _, line := range strings.Split(category[current].Answer, "\n") {
				res = fixLinks(line)
				if strings.HasPrefix(res, "!") {
					examSolutions += fixLinks(line)
				} else {
					examSolutions += "    " + fixLinks(line)
				}
			}
		}

		var base string
		if *tls {
			// TODO: dont hardcode paths
			base = filepath.Join("/etc/quizsrv", "files", id)
		} else {
			base = filepath.Join(*configFolder, "files", id)
		}
		_ = os.MkdirAll(base, 0777)

		fq, err := os.Create(filepath.Join(base, "examQuestions.md"))
		if err != nil {
			log.Fatal(err)
		}
		fq.WriteString(examQuestions)

		fs, err := os.Create(filepath.Join(base, "examSolutions.md"))
		if err != nil {
			log.Fatal(err)
		}
		fs.WriteString(examSolutions)

		err = fq.Close()
		if err != nil {
			log.Fatal(err)
		}

		err = fs.Close()
		if err != nil {
			log.Fatal(err)
		}

		var e string
		if *tls {
			e = "/usr/bin/mdtopdf"
		} else {
			e = "mdtopdf"
		}

		examName := strings.ToUpper(courseID) + "-Exam-Category-" + strings.ToUpper(c) + "-"

		err = exec.Command(e, "-i", filepath.Join(base, "examSolutions.md"), "-o", filepath.Join(base, "/" + examName + "Solutions.pdf")).Run()
		if err != nil {
			log.Fatal(err)
		}

		err = exec.Command(e, "-i", filepath.Join(base, "examQuestions.md"), "-o", filepath.Join(base, "/" + examName + "Questions.pdf")).Run()
		if err != nil {
			log.Fatal(err)
		}

		dir, _ := os.Getwd()
		fmt.Println("done! generated exam", id, "to", base, "pwd:", dir)
		output += "<a target='_blank' href='/files/" + id + "/" + examName + "Questions.pdf'>" + examName + "Questions.pdf</a><br><a target='_blank' href='/files/" + id + "/"+examName + "Solutions.pdf'>" + examName + "Solutions.pdf</a><br>"
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(output))
}
