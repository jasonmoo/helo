package helo

import (
	"crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

type (
	SmtpServer struct {
		host    string
		logger  *log.Logger
		running bool
	}
	SmtpsServer struct {
		*SmtpServer
		cert string
		key  string
	}
)

var (
	AlreadyRunningError = errors.New("helo already running")
)

func NewSmtpServer(host string) *SmtpServer {
	return &SmtpServer{
		host:   host,
		logger: log.New(os.Stdout, "helo: ", log.LstdFlags|log.Llongfile),
	}
}

func NewSmtpsServer(host, cert, key string) *SmtpsServer {
	return &SmtpsServer{NewSmtpServer(host), cert, key}
}

func (s *SmtpServer) SetLogger(logger *log.Logger) {
	s.logger = logger
}

func (s *SmtpServer) Start() error {

	if s.running {
		return AlreadyRunningError
	}

	l, err := net.Listen("tcp", s.host)
	if err != nil {
		return err
	}

	s.logger.Println("helo starting up.")
	s.logger.Printf("Listening on %s", s.host)

	s.running = true
	go func() {
		for s.running {
			conn, err := l.Accept()
			if err != nil {
				s.logger.Println(err)
				continue
			}
			go s.handleSession(conn)
		}
	}()

	return nil

}

func (s *SmtpsServer) StartTLS() error {

	if s.running {
		return AlreadyRunningError
	}

	certificate, err := tls.LoadX509KeyPair(s.cert, s.key)
	if err != nil {
		return err
	}

	config := &tls.Config{
		Rand:         rand.Reader,
		Certificates: []tls.Certificate{certificate},
		ClientAuth:   tls.NoClientCert,
	}

	tlsl, err := tls.Listen("tcp", s.host, config)
	if err != nil {
		return err
	}

	s.logger.Println("helo starting up.")
	s.logger.Printf("Listening on %s", s.host)

	s.running = true

	go func() {
		for {
			conn, err := tlsl.Accept()
			if err != nil {
				s.logger.Println(err)
				continue
			}
			go handleSession(conn)
		}
	}()

	return nil

}

func (s *SmtpServer) Stop() error {
	s.logger.Println("helo shutting down")
	s.running = false
}

func (s *SmtpServer) handleSession(conn net.Conn) {
	defer conn.Close()

	r := NewReader(conn)
	w := NewWriter(conn)

	// SMTP COMMANDS
	// http://tools.ietf.org/html/rfc821#page-29

	// SEQUENCING OF COMMANDS AND REPLIES
	// http://tools.ietf.org/html/rfc821#page-37

	// CONNECTION ESTABLISHMENT
	// S: 220
	// F: 421
	w.WriteStatus(StatusServiceReady)

	for {
		command, data, err := r.ReadCommand()
		if err != nil {
			s.logger.Println(err)
			w.WriteStatus(StatusRequestedActionAbortedInProcessing)
			return
		}

		switch command {
		// smtp

		// HELO <SP> <domain> <CRLF>
		// S: 250
		// E: 500, 501, 504, 421
		case CommandHelo:
			if len(data) == 0 {
				w.WriteStatus(StatusSyntaxErrorInParametersOrArguments)
			} else {
				w.WriteReply(StatusOk, "helo at your service")
			}

		// MAIL <SP> FROM:<reverse-path> <CRLF>
		// S: 250
		// F: 552, 451, 452
		// E: 500, 501, 421
		case CommandMail:

		// RCPT <SP> TO:<forward-path> <CRLF>
		// S: 250, 251
		// F: 550, 551, 552, 553, 450, 451, 452
		// E: 500, 501, 503, 421
		case CommandRcpt:

		// DATA <CRLF>
		// I: 354 -> data -> S: 250
		//                   F: 552, 554, 451, 452
		// F: 451, 554
		// E: 500, 501, 503, 421
		case CommandData:

		// RSET <CRLF>
		// S: 250
		// E: 500, 501, 504, 421
		case CommandRset:

		// SEND <SP> FROM:<reverse-path> <CRLF>
		// S: 250
		// F: 552, 451, 452
		// E: 500, 501, 502, 421
		case CommandSend:

		// SOML <SP> FROM:<reverse-path> <CRLF>
		// S: 250
		// F: 552, 451, 452
		// E: 500, 501, 502, 421
		case CommandSoml:

		// SAML <SP> FROM:<reverse-path> <CRLF>
		// S: 250
		// F: 552, 451, 452
		// E: 500, 501, 502, 421
		case CommandSaml:

		// VRFY <SP> <string> <CRLF>
		// S: 250, 251
		// F: 550, 551, 553
		// E: 500, 501, 502, 504, 421
		case CommandVrfy:

		// EXPN <SP> <string> <CRLF>
		// S: 250
		// F: 550
		// E: 500, 501, 502, 504, 421
		case CommandExpn:

		// HELP [<SP> <string>] <CRLF>
		// S: 211, 214
		// E: 500, 501, 502, 504, 421
		case CommandHelp:

		// NOOP <CRLF>
		// S: 250
		// E: 500, 421
		case CommandNoop:

		// QUIT <CRLF>
		// S: 221
		// E: 500
		case CommandQuit:

		// TURN <CRLF>
		// S: 250
		// F: 502
		// E: 500, 503
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

		default:
			w.WriteStatus(StatusSyntaxErrorCommandUnrecognized)
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
