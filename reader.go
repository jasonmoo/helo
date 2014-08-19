package main

import (
	"bufio"
	"errors"
	"io"
	"net"
	"regexp"
)

type (
	Reader struct {
		*bufio.Reader
	}
)

const (
	MaxMessageSize = 32 << 20 // 32 mb

	// smtp
	CommandHelo = "HELO"
	CommandMail = "MAIL"
	CommandRcpt = "RCPT"
	CommandData = "DATA"
	CommandRset = "RSET"
	CommandSend = "SEND"
	CommandSoml = "SOML"
	CommandSaml = "SAML"
	CommandVrfy = "VRFY"
	CommandExpn = "EXPN"
	CommandHelp = "HELP"
	CommandNoop = "NOOP"
	CommandQuit = "QUIT"
	CommandTurn = "TURN"

	// esmtp
	CommandEhlo       = "EHLO"
	Command8bitmime   = "8BITMIME"
	CommandAtrn       = "ATRN"
	CommandAuth       = "AUTH"
	CommandChunking   = "CHUNKING"
	CommandDsn        = "DSN"
	CommandEtrn       = "ETRN"
	CommandPipelining = "PIPELINING"
	CommandSize       = "SIZE"
	CommandStarttls   = "STARTTLS"
	CommandSmtputf8   = "SMTPUTF8"
)

var (
	command_regexp = regexp.MustCompile("^([A-Za-z0-9]+) ?(.*)$")
	email_regexp   = regexp.MustCompile("^(?:[Ff][Rr][Oo][Mm]|[Tt][Oo]):<([^>]+)>$")
	data_regexp    = regexp.MustCompile("^(?s.+?)\r\n.\r\n$")

	BadSyntaxError   = errors.New("bad syntax error")
	MessageSizeError = errors.New("max message size exceeded")
)

func NewReader(conn net.Conn) *Reader {
	return Reader{bufio.NewReader(conn)}
}

func (r *Reader) ReadCommand() (string, string, error) {

	data, err := r.consume(64)
	if err != nil {
		return "", "", err
	}

	if matches := command_regexp.FindSubmatch(data); len(matches) == 3 {
		return string(matches[1]), string(matches[2]), nil
	}
	return "", "", BadSyntaxError

}

func (r *Reader) ReadData() (string, error) {

	data, err := r.consume(1 << 10)
	if err != nil {
		return "", err
	}

	// match against expected data format (<data\r\n.\r\n>)
	if matches := data_regexp.FindSubmatch(data, 1); len(matches) == 2 {
		return string(matches[1]), nil
	}
	return "", BadSyntaxError

}

func (r *Reader) consume(chunk_size int) ([]byte, error) {

	var (
		data  []byte
		total int
	)

	// consume data until end of tcp buffer
	for {
		d := make([]byte, chunk_size)
		n, err := r.Read(d)
		if err != nil {
			if err == io.EOF {
				break
			}
			return data, err
		}
		total += n
		if total > MaxMessageSize {
			return data, MessageSizeError
		}
		data = append(data, d[:n]...)
	}

	return data
}
