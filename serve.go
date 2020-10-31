package main

import (
	"context"
	"log"
	"net/http"
	"time"
)

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

