package main

import (
	"crypto/rand"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"runtime"
)

// SMTP by default uses TCP port 25.
// The protocol for mail submission is the same,
// but uses port 587. SMTP connections secured by SSL,
// known as SMTPS, default to port 465.
var (
	dev = flag.Bool("dev", false, "output dev signals")

	smtp_port  = flag.String("smtp", ":25", "host:port to listen on")
	smtps_port = flag.String("smtps", ":465", "host:port to listen on")
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {

	certificate, err := tls.LoadX509KeyPair("cert/cert.pem", "cert/key.pem")
	if err != nil {
		log.Fatal(err)
	}

	config := &tls.Config{
		Rand:         rand.Reader,
		Certificates: []tls.Certificate{certificate},
		ClientAuth:   tls.NoClientCert,
	}

	tlsl, err := tls.Listen("tcp", *smtps_port, config)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for {
			conn, err := tlsl.Accept()
			if err != nil {
				log.Println(err)
				continue
			}
			go handleSession(conn)
		}
	}()

	l, err := net.Listen("tcp", *smtp_port)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Println(err)
				continue
			}
			go handleSession(conn)
		}
	}()

	log.Println("ehlo server online")

	select {}
}

func handleSession(conn net.Conn) {
	defer conn.Close()

	r := NewReader(conn)
	w := NewWriter(conn)

	for {
		command, data, err := r.ReadCommand()
		if err != nil {
			log.Println(err)
			w.WriteString(err.Error())
			return
		}

		switch command {
		// smtp
		// HELO <SP> <domain> <CRLF>
		case CommandHelo:
			w.WriteStatus()
			// MAIL <SP> FROM:<reverse-path> <CRLF>
		case CommandMail:
			// RCPT <SP> TO:<forward-path> <CRLF>
		case CommandRcpt:
			// DATA <CRLF>
		case CommandData:
			// RSET <CRLF>
		case CommandRset:
			// SEND <SP> FROM:<reverse-path> <CRLF>
		case CommandSend:
			// SOML <SP> FROM:<reverse-path> <CRLF>
		case CommandSoml:
			// SAML <SP> FROM:<reverse-path> <CRLF>
		case CommandSaml:
			// VRFY <SP> <string> <CRLF>
		case CommandVrfy:
			// EXPN <SP> <string> <CRLF>
		case CommandExpn:
			// HELP [<SP> <string>] <CRLF>
		case CommandHelp:
			// NOOP <CRLF>
		case CommandNoop:
			// QUIT <CRLF>
		case CommandQuit:
			// TURN <CRLF>
		case CommandTurn:

		// esmtp:
		// EHLO <domain> <CRLF>
		case CommandEHLO:
			// 8BITMIME — 8 bit data transmission, RFC 6152
		case Command8BITMIME:
			// ATRN — Authenticated TURN for On-Demand Mail Relay, RFC 2645
		case CommandATRN:
			// AUTH — Authenticated SMTP, RFC 4954
		case CommandAUTH:
			// CHUNKING — Chunking, RFC 3030
		case CommandCHUNKING:
			// DSN — Delivery status notification, RFC 3461 (See Variable envelope return path)
		case CommandDSN:
			// ETRN — Extended version of remote message queue starting command TURN, RFC 1985
		case CommandETRN:
			// PIPELINING — Command pipelining, RFC 2920
		case CommandPIPELINING:
			// SIZE — Message size declaration, RFC 1870
		case CommandSIZE:
			// STARTTLS — Transport layer security, RFC 3207 (2002)
		case CommandSTARTTLS:
			// SMTPUTF8 — Allow UTF-8 encoding in mailbox names and header fields, RFC 6531
		case CommandSMTPUTF8:
		}
	}

	fmt.Fprintf(conn, "220 mx.google.com ESMTP z11sm13237475pdl.8 - gsmtp\r\n")

	for _, resp := range []string{"250 ok\r\n", "250 ok\r\n", "250 ok\r\n", "354 data\r\n", "250 Message queued\r\n", "221 Goodbye\r\n"} {
		data := make([]byte, 1<<10)
		_, err := conn.Read(data)
		if err != nil && err != io.EOF {
			log.Println(err)
			return
		}
		// log.Printf("data:   %q\n", data[:n])
		// log.Printf("resp:   %q\n", resp)
		fmt.Fprint(conn, resp)
	}
}
