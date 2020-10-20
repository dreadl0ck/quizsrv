/*
 * QUIZSRV - A simple quiz server
 * Copyright (c) 2020 Philipp Mieden <dreadl0ck [at] protonmail [dot] ch>
 * This code is licensed under the GPLv3.
 */

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/dreadl0ck/cryptoutils"
	"github.com/foomo/simplecert"
	"github.com/foomo/tlsconfig"
	"github.com/russross/blackfriday/v2"
	"gopkg.in/yaml.v2"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var Version = "master"

type TemplateData struct {
	Host        string
	Categories  map[string]int // name to num questions
	Slides      []string
	Version     string
	Quote       string
	QuoteAuthor string
	Total       int
	Category    string
}

type Data struct {
	Categories map[string][]question
	Slides     map[string]string
}

type question struct {
	Question string `yaml:"question"`
	Answer   string `yaml:"answer"`
}

var (
	addr         = flag.String("addr", "localhost:8080", "http service address")
	configFolder = flag.String("c", "", "path to the configuration files")
	tls          = flag.Bool("tls", false, "use TLS")

	upgrader = websocket.Upgrader{}
	data     *Data
)

var quotes = map[string]string{
	"We will never do this again":                         "John Postel",
	"CIA next summer":                                     "OS3 students that failed the exam",
	"One week, one week, one week and we had Unix":        "Ken Thompson",
	"If you can read assembly, everything is open source": "Unknown",
	"UNIX is basically a simple operating system, but you have to be a genius to understand the simplicity.":                                                                                "Dennis Ritchie",
	"I think the major good idea in Unix was its clean and simple interface: open, close, read, and write.":                                                                                 "Ken Thompson",
	"Pretty much everything on the web uses those two things: C and UNIX. The browsers are written in C. The UNIX kernel - that pretty much the entire Internet runs on - is written in C.": "Rob Pike",
	"The one thing I stole was the hierarchical file system because it was a really good idea ...":                                                                                          "Ken Thompson",
	"If you want to travel around the world and be invited to speak at a lot of different places, just write a Unix operating system.":                                                      "Linus Torvalds",
	"Contrary to popular belief, Unix is user friendly. It just happens to be very selective about who it decides to make friends with.":                                                    "Unknown",
	"UNIX is not so much an operating system as a way of thinking.":                                                                                                                         "Unknown",
	"UNIX was not designed to stop its users from doing stupid things, as that would also stop them from doing clever things.":                                                              "Unknown",
	"One of my most productive days was throwing away 1,000 lines of code.":                                                                                                                 "Ken Thompson",
	"Be conservative in what you send, and liberal in what you accept":                                                                                                                      "John Postel",
}

var exam = map[string]int{
	"architecture": 20,
	"dns":          10,
	"dnssec":       10,
	"email":        20,
	"history":      19,
	"web":          20,
}

func genExam(w http.ResponseWriter, r *http.Request) {

	rand.Seed(time.Now().UnixNano())

	var (
		examQuestions string
		examSolutions string
		count         int
	)

	id, err := cryptoutils.RandomString(10)
	if err != nil {
		log.Fatal(err)
	}

	id = strings.TrimSuffix(id, "==")

	// write header
	examQuestions += "# CIA Test Exam" + "\n"
	examQuestions += "\n> ID: " + id + "\n\n"

	examSolutions += "# CIA Test Exam" + "\n"
	examSolutions += "\n> ID: " + id + "\n\n"

	for c, n := range exam {
		var (
			category = data.Categories[c]
			current  int
			done     []int
		)

		// write category
		examQuestions += "\n### " + strings.ToUpper(c) + "\n"
		examSolutions += "\n### " + strings.ToUpper(c) + "\n"

		if c == "history" {
			count++
			// always add a quote for history part
			for q, author := range quotes {
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

		// write questions
		for i := 0; i < n; i++ {

			// pick a random one
			for {
				current = rand.Intn(len(category))
				// exclude light bulb jokes
				if !hasBeenAsked(current, done) && !strings.Contains(category[current].Question, "light bulb") {
					break
				}
			}

			count++

			examQuestions += strconv.Itoa(count) + ") " + fixLinks(category[current].Question)
			examQuestions += "    \n"

			examSolutions += strconv.Itoa(count) + ") " + fixLinks(category[current].Question)
			examSolutions += "    \n"

			for _, line := range strings.Split(category[current].Answer, "\n") {
				examSolutions += "    " + fixLinks(line)
			}

			done = append(done, current)
		}
	}

	var base string
	if *tls {
		base = filepath.Join("/etc/quizsrv", "files", id)
	} else {
		base = filepath.Join("files", id)
	}
	_ = os.Mkdir(base, 0777)

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

	err = exec.Command(e, "-i", filepath.Join(base, "examSolutions.md"), "-o", filepath.Join(base, "examSolutions.pdf")).Run()
	if err != nil {
		log.Fatal(err)
	}

	err = exec.Command(e, "-i", filepath.Join(base, "examQuestions.md"), "-o", filepath.Join(base, "examQuestions.pdf")).Run()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("done! generated exam", id)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("<a target='_blank' href='/cia/files/" + id + "/examQuestions.pdf'>examQuestions.pdf</a><br><a target='_blank' href='/cia/files/" + id + "/examSolutions.pdf'>examSolutions.pdf</a>"))
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
				out += strings.ReplaceAll(line, "/cia/files/img/", "./etc/quizsrv/files/img/") + "\n"
			} else {
				out += strings.ReplaceAll(line, "/cia/files/img/", "./files/img/")
			}
			continue
		}

		out += line + "\n"
	}

	return out
}

func quiz(w http.ResponseWriter, r *http.Request) {
	homeTemplate := template.Must(template.ParseFiles(filepath.Join(*configFolder, "pages/quiz.html")))
	err := homeTemplate.Execute(w, &TemplateData{
		//Host:       "ws://"+r.Host+"/cia/connect", // TODO: unused
		Version:  Version,
		Category: strings.ToUpper(filepath.Base(r.RequestURI)),
	})
	if err != nil {
		log.Println(err)
	}
}

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

	s := strings.Split(string(helloMsg), ";")
	location := s[0]
	flagged := strings.Split(s[1], ",")

	fmt.Println("flagged", flagged)

	var (
		category      = data.Categories[filepath.Base(string(location))]
		done          = initDoneSlice(category, flagged)
		current       int
		previousIndex int
	)

	c.WriteMessage(websocket.TextMessage, []byte("total="+strconv.Itoa(len(category)-len(done))))

	fmt.Println(len(category), filepath.Base(string(location)), string(location))

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
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)

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

func home(w http.ResponseWriter, r *http.Request) {
	var (
		slides     []string
		categories = make(map[string]int)
		total      int
	)
	for c := range data.Categories {
		categories[c] = len(data.Categories[c])
		total += len(data.Categories[c])
	}
	for c := range data.Slides {
		slides = append(slides, c)
	}

	var quote, quoteAuthor string
	for q, a := range quotes {
		quote, quoteAuthor = q, a
		break
	}
	homeTemplate := template.Must(template.ParseFiles(filepath.Join(*configFolder, "pages/index.html")))
	err := homeTemplate.Execute(w, &TemplateData{
		Categories:  categories,
		Slides:      slides,
		Version:     Version,
		Quote:       quote,
		QuoteAuthor: quoteAuthor,
		Total:       total,
	})
	if err != nil {
		log.Println(err)
	}
}

func pdf(w http.ResponseWriter, r *http.Request) {
	var (
		slides, categories []string
	)
	for c := range data.Categories {
		categories = append(categories, c)
	}
	for c := range data.Slides {
		slides = append(slides, c)
	}
	homeTemplate := template.Must(template.ParseFiles("pages/pdf.html"))
	err := homeTemplate.Execute(w, strings.TrimSuffix(filepath.Base(r.RequestURI), ".pdf"))
	if err != nil {
		log.Println(err)
	}
}

func main() {

	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	root := filepath.Join(*configFolder, "categories")
	files, err := ioutil.ReadDir(root)
	if err != nil {
		log.Fatal(err)
	}

	data = new(Data)
	data.Categories = make(map[string][]question)
	data.Slides = make(map[string]string)

	for _, f := range files {

		category := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))

		switch filepath.Ext(f.Name()) {
		case ".yml", ".yaml":
			c, err := ioutil.ReadFile(filepath.Join(root, f.Name()))
			if err != nil {
				log.Fatal(err)
			}

			var questions []question

			err = yaml.UnmarshalStrict(c, &questions)
			if err != nil {
				log.Fatal(err, " file: ", f.Name())
			}

			data.Categories[category] = questions
			fmt.Println("loaded category", category, len(questions))
			http.HandleFunc("/cia/"+category, quiz)
		case ".pdf":
			data.Slides[category] = f.Name()
			http.HandleFunc("/cia/"+category, pdf)
		default:
			log.Println("unsupported format", filepath.Ext(f.Name()), " file: ", f.Name())
		}
	}

	log.SetFlags(0)

	http.HandleFunc("/cia/mkexam", genExam)
	http.HandleFunc("/cia/files/", http.StripPrefix("/cia/files/", http.FileServer(http.Dir(filepath.Join(*configFolder, "files")))).ServeHTTP)
	http.HandleFunc("/cia/connect", connect)
	http.HandleFunc("/cia", home)
	http.HandleFunc("/", http.RedirectHandler("/cia", http.StatusMovedPermanently).ServeHTTP)

	fmt.Println("serving at", *addr)
	if *tls {

		fmt.Println("production mode")

		var (
			certReloader *simplecert.CertReloader
			numRenews    int

			ctx, cancel = context.WithCancel(context.Background())

			// init strict tlsConfig
			tlsconf = tlsconfig.NewServerTLSConfig(tlsconfig.TLSModeServerStrict)

			makeServer = func() *http.Server {
				return &http.Server{
					Addr:      ":443",
					Handler:   nil,
					TLSConfig: tlsconf,
				}
			}

			// init server
			srv = makeServer()

			// init simplecert configuration
			cfg = simplecert.Default
		)

		cfg.Domains = []string{"www.dreadl0ck.net", "dreadl0ck.net"}
		cfg.CacheDir = "/etc/letsencrypt/quizsrv"
		cfg.SSLEmail = "pmieden@os3.nl"

		// disable renewal over HTTP challenge
		// we'll stick to using the TLS challenge exclusively for now
		cfg.HTTPAddress = ""

		// set hooks to shutdown and restart
		cfg.WillRenewCertificate = func() {
			// stop server
			cancel()
		}

		cfg.DidRenewCertificate = func() {
			numRenews++
			// restart server after renewing cert
			fmt.Println("finished renew number", numRenews, "restarting service...")

			// restart server: both context and server instance need to be recreated!
			ctx, cancel = context.WithCancel(context.Background())
			srv = makeServer()

			// force reload the updated cert from disk
			certReloader.ReloadNow()

			go serve(ctx, srv, numRenews)
		}

		// init simplecert
		certReloader, err = simplecert.Init(cfg, func() {
			// simplecert cleanup function: nothing to cleanup yet, exit cleanly.
			os.Exit(0)
		})
		if err != nil {
			log.Fatal("simplecert init failed: ", err)
		}

		// redirect HTTP to HTTPS
		fmt.Println("starting HTTP Listener on Port 80")
		go http.ListenAndServe(":80", http.HandlerFunc(simplecert.Redirect))

		// enable hot reload
		// now set GetCertificate to the reloaders GetCertificateFunc to enable hot reload
		tlsconf.GetCertificate = certReloader.GetCertificateFunc()

		// start serving
		log.Println("will serve at: https://" + cfg.Domains[0])
		serve(ctx, srv, numRenews)

		fmt.Println("initial serve stopped, waiting forever")

		// block forever, to avoid having main quit after the first renewal
		<-make(chan struct{})
	} else {
		log.Fatal(http.ListenAndServe(*addr, nil))
	}
}

// serve uses the passed in context and the http server to start accepting connections
// this function
// - starts the passed in http.Server in a new goroutine
// - blocks until the server received the initial shutdown signal
// - gives the service 5 seconds to finish handling connections before stopping
// - can be called again after stopping the service via the context
func serve(ctx context.Context, srv *http.Server, instance int) {

	// start serving in the background
	go func() {
		// certFile and keyFile left blank because they will be retrieve via our GetCertificateFunc
		if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen:%s\n", err)
		}
	}()

	// wait for context to finish
	log.Println("server started, instance:", instance)
	<-ctx.Done()
	log.Println("server shutdown initiated, instance:", instance)

	// graceful stop: give the server 5 seconds to finish all connections using the passed in context
	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	// shutdown service
	err := srv.Shutdown(ctxShutDown)
	if err == http.ErrServerClosed {
		log.Printf("server exited properly, instance %d\n", instance)
	} else if err != nil {
		log.Printf("server encountered an error on exit: %s, instance: %d\n", err, instance)
	}
}
