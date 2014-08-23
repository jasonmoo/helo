package helo

import (
	"bufio"
	"crypto/tls"
	"errors"
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
		logger: log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds),
	}
}

func NewSmtpsServer(host, cert, key string) *SmtpsServer {
	return &SmtpsServer{NewSmtpServer(host), cert, key}
}

func (s *SmtpServer) SetLogger(logger *log.Logger) {
	s.logger = logger
}

func (s *SmtpServer) newReader(conn net.Conn) *Reader {
	return &Reader{bufio.NewReader(conn), s}
}

func (s *SmtpServer) newWriter(conn net.Conn) *Writer {
	return &Writer{conn, s}
}

func (s *SmtpServer) log(data interface{}) {
	if s.logger != nil {
		s.logger.Println(data)
	}
}
func (s *SmtpServer) logf(messagef string, data ...interface{}) {
	if s.logger != nil {
		s.logger.Printf(messagef, data...)
	}
}

func (s *SmtpServer) Start() error {

	if s.running {
		return AlreadyRunningError
	}

	l, err := net.Listen("tcp", s.host)
	if err != nil {
		return err
	}

	s.log("helo smtp starting up.")
	s.logf("Listening on %s", s.host)

	s.running = true

	go func() {
		for s.running {
			conn, err := l.Accept()
			if err != nil {
				s.log(err)
				continue
			}
			go s.handleSession(conn)
		}
	}()

	return nil

}

func (s *SmtpsServer) Start() error {

	if s.running {
		return AlreadyRunningError
	}

	certificate, err := tls.LoadX509KeyPair(s.cert, s.key)
	if err != nil {
		return err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{certificate},
	}

	tlsl, err := tls.Listen("tcp", s.host, config)
	if err != nil {
		return err
	}

	s.log("helo smtps starting up.")
	s.logf("Listening on %s", s.host)

	s.running = true

	go func() {
		for s.running {
			conn, err := tlsl.Accept()
			if err != nil {
				s.log(err)
				continue
			}
			go s.handleSession(conn)
		}
	}()

	return nil

}

func (s *SmtpServer) Stop() {
	s.log("helo shutting down")
	s.running = false
}
