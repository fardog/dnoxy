package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/fardog/dnoxy"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

const (
	cloudflareEndpoint = "https://cloudflare-dns.com/dns-query"
)

var (
	listenAddress = flag.String(
		"listen", ":53", "listen address, as `[host]:port`",
	)

	logLevel = flag.String(
		"level",
		"info",
		"Log level, one of: debug, info, warn, error, fatal, panic",
	)

	endpoint = flag.String(
		"endpoint",
		cloudflareEndpoint,
		"DNS-over-HTTPS endpoint url",
	)

	enableTCP = flag.Bool("tcp", true, "Listen on TCP")
	enableUDP = flag.Bool("udp", true, "Listen on UDP")
)

func serve(net string) {
	log.Infof("starting %s service on %s", net, *listenAddress)

	server := &dns.Server{Addr: *listenAddress, Net: net, TsigSecret: nil}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Failed to setup the %s server: %s\n", net, err.Error())
		}
	}()

	// serve until exit
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Infof("shutting down %s on interrupt\n", net)
	if err := server.Shutdown(); err != nil {
		log.Errorf("got unexpected error %s", err.Error())
	}
}

func main() {
	flag.Usage = func() {
		_, exe := filepath.Split(os.Args[0])
		fmt.Fprint(os.Stderr, "A DNS to DNS-over-HTTPS proxy.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\n  %s [options]\n\nOptions:\n\n", exe)
		flag.PrintDefaults()
	}
	flag.Parse()

	// seed the global random number generator, used in some utilities and the
	// google provider
	rand.Seed(time.Now().UTC().UnixNano())

	// set the loglevel
	level, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Fatalf("invalid log level: %s", err.Error())
	}
	log.SetLevel(level)

	ex, err := dnoxy.NewHTTPExchanger(*endpoint, nil)
	if err != nil {
		panic(err)
	}
	handler, err := dnoxy.NewDNSHandler(ex, nil)
	if err != nil {
		panic(err)
	}

	dns.HandleFunc(".", handler.Handle)

	// push the list of enabled protocols into an array
	var protocols []string
	if *enableTCP {
		protocols = append(protocols, "tcp")
	}
	if *enableUDP {
		protocols = append(protocols, "udp")
	}

	// start the servers
	servers := make(chan bool)
	for _, protocol := range protocols {
		go func(protocol string) {
			serve(protocol)
			servers <- true
		}(protocol)
	}

	// wait for servers to exit
	for i := 0; i < len(protocols); i++ {
		<-servers
	}

	log.Infoln("servers exited, stopping")
}
