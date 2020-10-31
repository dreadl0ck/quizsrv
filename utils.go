package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/russross/blackfriday/v2"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func hasBeenAsked(current int, done []int) bool {
	for _, d := range done {
		if d == current {
			return true
		}
	}
	return false
}

func writeQuestion(c *websocket.Conn, category []question, current int) {
	c.WriteMessage(websocket.TextMessage, append([]byte("index="+strconv.Itoa(current)+":"), blackfriday.Run([]byte(category[current].Question))...))
}

func writeAnswer(c *websocket.Conn, category []question, current int) {
	c.WriteMessage(websocket.TextMessage, append([]byte("index="+strconv.Itoa(current)+":"), blackfriday.Run([]byte(category[current].Answer))...))
}

func writeDone(c *websocket.Conn, r *http.Request) {
	fmt.Println("client DONE", r.RemoteAddr)
	c.WriteMessage(websocket.TextMessage, []byte("index=XXX:congrats you are done! reload to restart"))
	c.Close()
}

func initDoneSlice(category []question, flagged []string) (out []int) {

	if len(flagged) == 0 {
		return []int{}
	}

	if flagged[0] == "" {
		return []int{}
	}

	// init flagged map
	var flaggedMap = make(map[int]struct{})
	for _, f := range flagged {
		i, err := strconv.Atoi(f)
		if err != nil {
			fmt.Println(i, err)
			continue
		}
		flaggedMap[i] = struct{}{}
	}

	// populate out slice only with all indices that are NOT flagged
	for i := range category {
		if _, ok := flaggedMap[i]; !ok {
			out = append(out, i)
		}
	}

	return out
}

func shuffle(src []string) []string {
	final := make([]string, len(src))
	rand.Seed(time.Now().UTC().UnixNano())
	perm := rand.Perm(len(src))

	for i, v := range perm {
		final[v] = src[i]
	}
	return final
}

func fixLinks(in string) string {

	var out string
	for _, line := range strings.Split(in, "\n") {

		if line == "" {
			out += "    \n"
			continue
		}

		if strings.HasPrefix(line, "!") {
			if *tls {
				out += strings.ReplaceAll(line, "/files/img/", "./etc/quizsrv/files/img/") + "\n"
			} else {
				out += strings.ReplaceAll(line, "/files/img/", "./files/img/")
			}
			continue
		}

		out += line + "\n"
	}

	return out
}
