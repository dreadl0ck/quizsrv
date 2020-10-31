package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

func connect(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	_, helloMsg, err := c.ReadMessage()
	if err != nil {
		log.Println("read:", err)
		return
	}

	var (
		s = strings.Split(string(helloMsg), ";")
		location = s[0]
		flagged = strings.Split(s[1], ",")
		courseName = r.FormValue("course")
	)

	fmt.Println("flagged", flagged)
	fmt.Println("courseName", courseName)

	cou, ok := data.Courses[courseName]
	if !ok {
		c.WriteMessage(websocket.TextMessage, []byte("index=XXX:invalid course name"))
		c.Close()
		return
	}

	var catName = filepath.Base(location)
	fmt.Println("received categoryName", catName)

	var (
		category      = cou.Categories[catName]
		done          = initDoneSlice(category, flagged)
		current       int
		previousIndex int
	)

	if len(category) == 0 {
		c.Close()
		return
	}

	c.WriteMessage(websocket.TextMessage, []byte("total="+strconv.Itoa(len(category)-len(done))))

	fmt.Println(len(category), filepath.Base(location), location)

	for {
		current = rand.Intn(len(category))
		if !hasBeenAsked(current, done) {
			break
		}
	}

	writeQuestion(c, category, current)

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println(r.RemoteAddr, "read:", err)
			break
		}
		log.Println(r.RemoteAddr, "recv:", string(message))

		if len(done) == len(category) {
			writeDone(c, r)
			return
		}

		switch string(message) {
		case "next":

			if previousIndex == 0 {
				done = append(done, current)

				if len(done) == len(category) {
					writeDone(c, r)
					return
				}
			}

			if previousIndex > 0 {

				previousIndex--

				if previousIndex == 0 {
					for {
						current = rand.Intn(len(category))
						if !hasBeenAsked(current, done) {
							break
						}
					}
				} else {
					current = done[len(done)-previousIndex]
				}

			} else {
				for {
					current = rand.Intn(len(category))
					if !hasBeenAsked(current, done) {
						break
					}
				}
			}

			writeQuestion(c, category, current)

		case "answer":
			writeAnswer(c, category, current)

		case "previous":

			if len(done)-previousIndex+1 >= 0 {

				previousIndex++

				// last entry in done was the previous question
				current = done[len(done)-previousIndex]

				writeQuestion(c, category, current)
			}
		}
	}
}
