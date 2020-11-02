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
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
)

var (
	Version = "master"

	addr         = flag.String("addr", "localhost:8080", "http service address")
	configFolder = flag.String("c", "", "path to the configuration files")
	tls          = flag.Bool("tls", false, "use TLS")
	importAnki   = flag.Bool("anki", false, "import anki")

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
	CourseName  string
	CourseID    string
}

func main() {

	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	if *importAnki {
		ankiImport()
		return
	}

	courseImport()

	log.SetFlags(0)

	http.HandleFunc("/mkexam", genExam)
	http.HandleFunc("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(filepath.Join(*configFolder, "files")))).ServeHTTP)
	http.HandleFunc("/connect", connect)
	http.HandleFunc("/cia", http.RedirectHandler("/courses/cia", http.StatusMovedPermanently).ServeHTTP)
	http.HandleFunc("/courses", home)
	http.HandleFunc("/robots.txt", nobots)
	http.HandleFunc("/", home)

	fmt.Println("serving at", *addr)
	if *tls {

		fmt.Println("production mode")

		var (
			certReloader *simplecert.CertReloader
			numRenews    int
			err          error

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
