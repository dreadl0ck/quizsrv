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
	"github.com/foomo/simplecert"
	"github.com/foomo/tlsconfig"
	"gopkg.in/yaml.v3"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var Version = "master"

type TemplateData struct {
	Host string
	Categories map[string]int // name to num questions
	Slides []string
	Version string
	Quote string
	QuoteAuthor string
	Total int
	Category string
}

type Data struct {
	Categories map[string][]question
	Slides map[string]string
}

type question struct {
	Question string `yaml:"question"`
	Answer   string `yaml:"answer"`
}

var (
	addr = flag.String("addr", "localhost:8080", "http service address")
	configFolder = flag.String("c", "", "path to the configuration files")
	tls = flag.Bool("tls", false, "use TLS")

	upgrader = websocket.Upgrader{}
	data *Data
)

var quotes = map[string]string{
	"We will never do this again": "John Postel",
	"CIA next summer": "OS3 students that failed the exam",
	"One week, one week, one week and we had Unix": "Ken Thompson",
	"If you can read assembly, everything is open source": "Unknown",
	"UNIX is basically a simple operating system, but you have to be a genius to understand the simplicity.": "Dennis Ritchie",
	"I think the major good idea in Unix was its clean and simple interface: open, close, read, and write.": "Ken Thompson",
	"Pretty much everything on the web uses those two things: C and UNIX. The browsers are written in C. The UNIX kernel - that pretty much the entire Internet runs on - is written in C.": "Rob Pike",
	"The one thing I stole was the hierarchical file system because it was a really good idea ...": "Ken Thompson",
	"If you want to travel around the world and be invited to speak at a lot of different places, just write a Unix operating system.": "Linus Torvalds",
	"Contrary to popular belief, Unix is user friendly. It just happens to be very selective about who it decides to make friends with.": "Unknown",
	"UNIX is not so much an operating system as a way of thinking.": "Unknown",
	"UNIX was not designed to stop its users from doing stupid things, as that would also stop them from doing clever things.": "Unknown",
	"One of my most productive days was throwing away 1,000 lines of code.": "Ken Thompson",
	"Be conservative in what you send, and liberal in what you accept": "John Postel",
}

func quiz(w http.ResponseWriter, r *http.Request) {
	homeTemplate := template.Must(template.ParseFiles(filepath.Join(*configFolder, "pages/quiz.html")))
	err := homeTemplate.Execute(w, &TemplateData{
		Host:       "ws://"+r.Host+"/cia/connect",
		Version: Version,
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

func connect(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	_, location, err := c.ReadMessage()
	if err != nil {
		log.Println("read:", err)
		return
	}

	var (
		category = data.Categories[filepath.Base(string(location))]
		done []int
		current = rand.Intn(len(category))
	)

	c.WriteMessage(websocket.TextMessage, []byte("total="+strconv.Itoa(len(category))))

	fmt.Println(len(category), filepath.Base(string(location)), string(location))
	c.WriteMessage(websocket.TextMessage, []byte("Q: " + category[current].Question))

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)

		if len(done) == len(category) {
			fmt.Println("client DONE", r.RemoteAddr)
			c.WriteMessage(websocket.TextMessage, []byte("congrats you are done! reload to restart"))
			c.Close()
			return
		}

		switch string(message) {
			case "next":

				for {
					current = rand.Intn(len(category))
					if !hasBeenAsked(current, done) {
						break
					}
				}

				c.WriteMessage(websocket.TextMessage, []byte("Q: " + category[current].Question))

			case "answer":
				c.WriteMessage(websocket.TextMessage, []byte(category[current].Answer))
				done = append(done, current)
		}
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	var (
		slides []string
		categories = make(map[string]int)
		total int
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
		Categories: categories,
		Slides: slides,
		Version: Version,
		Quote: quote,
		QuoteAuthor: quoteAuthor,
		Total: total,
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

			err = yaml.Unmarshal(c, &questions)
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

	http.HandleFunc("/cia/files/", http.StripPrefix("/cia/files/", http.FileServer(http.Dir(filepath.Join(*configFolder, "files")))).ServeHTTP)
	http.HandleFunc("/cia/connect", connect)
	http.HandleFunc("/cia", home)
	http.HandleFunc("/", http.RedirectHandler("/cia", http.StatusMovedPermanently).ServeHTTP)

	fmt.Println("serving at", *addr)
	if *tls {

		fmt.Println("production mode")

		var (
			certReloader *simplecert.CertReloader
			numRenews   int

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
