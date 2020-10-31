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
	"github.com/davecgh/go-spew/spew"
	"github.com/flimzy/anki"
	"github.com/foomo/simplecert"
	"github.com/foomo/tlsconfig"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var (
	Version = "master"

	addr         = flag.String("addr", "localhost:8080", "http service address")
	configFolder = flag.String("c", "", "path to the configuration files")
	tls          = flag.Bool("tls", false, "use TLS")

	upgrader = websocket.Upgrader{}
	data     *Data
)

type TemplateData struct {
	Host        string
	Categories  map[string]int // name to num questions
	Slides      []string
	Version     string
	Quote       string
	QuoteAuthor string
	Total       int
	Category    string
	Courses     []string
	CourseName string
	CourseID string
}

func main() {

	flag.Parse()
	rand.Seed(time.Now().UnixNano())

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
					//strip.StripTags(
						re.ReplaceAllStringFunc(in, func(m string) string{
							return strings.Replace(m, ".", ".\n", 1)
						}),
					//),
				)
			}

			if strings.HasPrefix(n.FieldValues[0], "What are the five different OSPF") {
				spew.Dump(n)
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
		f, err := os.Create(inrReplacer.Replace("courses/inr/Anki/" +f.Name()))
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
		data.Courses[file.Name()] = cou

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

	log.SetFlags(0)

	http.HandleFunc("/mkexam", genExam)
	http.HandleFunc("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(filepath.Join(*configFolder, "files")))).ServeHTTP)
	http.HandleFunc("/connect", connect)
	http.HandleFunc("/cia", http.RedirectHandler("/courses/cia", http.StatusMovedPermanently).ServeHTTP)
	http.HandleFunc("/courses", home)
	http.HandleFunc("/", home)

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