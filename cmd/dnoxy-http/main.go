package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/fardog/dnoxy"
	"github.com/fardog/dnoxy/cmd"
	log "github.com/sirupsen/logrus"
)

var (
	listenAddress = flag.String(
		"listen", ":80", "listen address, as `[host]:port`",
	)
	handlerPath = flag.String(
		"path",
		"/dns-query",
		"DNS handler path",
	)

	logLevel = flag.String(
		"level",
		"info",
		"Log level, one of: debug, info, warn, error, fatal, panic",
	)

	shutdownTimeout = flag.Duration(
		"shutdown-timeout",
		10*time.Second,
		"Time to wait for requests to finish before shutting down",
	)

	endpoints = make(cmd.Values, 0)
)

var endpointRex = regexp.MustCompile(`:[\d]+$`)

func startHTTP(ex dnoxy.Exchanger, shutdown, done chan struct{}) {
	log := log.New()
	handler, err := dnoxy.NewHTTPHandler(ex, nil)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	// lazy hack: add handler for path with and without the trailing slash
	with := *handlerPath
	without := strings.TrimSuffix(with, "/")
	if !strings.HasSuffix(with, "/") {
		with = fmt.Sprintf("%s/", with)
	}
	mux.Handle(with, handler)
	if without != "" {
		mux.Handle(without, handler)
	}

	server := &http.Server{
		Addr:    *listenAddress,
		Handler: mux,
	}
	go func() {
		select {
		case <-shutdown:
			ctx, cancel := context.WithTimeout(context.Background(), *shutdownTimeout)
			defer cancel()
			server.Shutdown(ctx)
		}
	}()

	log.Printf("listening on %s", *listenAddress)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}

	done <- struct{}{}
}

func main() {
	flag.Var(
		&endpoints,
		"endpoint",
		`DNS Endpoint to be used; specify multiple as:
    -endpoint 1.0.0.1 -endpoint 1.1.1.1 `,
	)

	flag.Usage = func() {
		_, exe := filepath.Split(os.Args[0])
		fmt.Fprint(os.Stderr, "A DNS-over-HTTP server.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\n  %s [options]\n\nOptions:\n\n", exe)
		flag.PrintDefaults()
	}
	flag.Parse()

	// seed the global random number generator, used in some utilities and the
	// google provider
	rand.Seed(time.Now().UTC().UnixNano())

	if len(endpoints) == 0 {
		endpoints = []string{"1.0.0.1:53", "1.1.1.1:53"}
	}

	var eps []string
	for _, e := range endpoints {
		if !endpointRex.MatchString(e) {
			e = fmt.Sprintf("%s:53", e)
		}
		eps = append(eps, e)
	}
	fmt.Println(eps)

	ex, err := dnoxy.NewDNSExchanger(eps, nil)
	if err != nil {
		panic(err)
	}

	shutdown := make(chan struct{})
	done := make(chan struct{})
	allDone := make(chan struct{})
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint)

	go func() {
		startHTTP(ex, shutdown, done)
	}()

	go func() {
		<-done
		close(allDone)
	}()

	<-sigint
	close(shutdown)

	to, cancel := context.WithTimeout(context.Background(), *shutdownTimeout)
	defer cancel()

	select {
	case <-allDone:
	case <-to.Done():
		log.Fatal("shutdown timeout reached; exiting unclean")
	}

	log.Print("shutdown cleanly")
}
