package main

import (
	"flag"
	"log"
	"runtime"

	"github.com/jasonmoo/helo"
)

// SMTP by default uses TCP port 25.
// The protocol for mail submission is the same,
// but uses port 587. SMTP connections secured by SSL,
// known as SMTPS, default to port 465.
var (
	dev = flag.Bool("dev", false, "output dev signals")

	smtp_host  = flag.String("smtp_host", ":25", "host:port to listen on")
	smtps_host = flag.String("smtps_host", ":465", "host:port to listen on")

	tls_cert = flag.String("tls_cert", "cert/cert.pem", "cert for tls server")
	tls_key  = flag.String("tls_key", "cert/key.pem", "key for tls server")
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	s := helo.NewSmtpServer(*smtp_host)
	ss := helo.NewSmtpsServer(*smtps_host, *tls_cert, *tls_key)

	err := s.Start()
	if err != nil {
		log.Fatal(err)
	}

	err = ss.Start()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("server starting up")
	select {}

}
