package helo

import (
	"bufio"
	"bytes"
	"errors"
	"regexp"
	"strings"
)

type (
	Reader struct {
		*bufio.Reader
		s *SmtpServer
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
	command_regexp  = regexp.MustCompile("^([A-Za-z0-9]+) ?(.*)\r\n$")
	to_email_regexp = regexp.MustCompile("^[Tt][Oo]:<([^>]+)>$")
	// SIZE unused in this context
	from_email_regexp = regexp.MustCompile(`^[Ff][Rr][Oo][Mm]:<([^>]+)>(?: [Ss][Ii][Zz][Ee]=\d+)?$`)

	BadSyntaxError   = errors.New("bad syntax error")
	MessageSizeError = errors.New("max message size exceeded")
)

func (r *Reader) ReadCommand() (string, string, error) {

	data := make([]byte, 1<<10)
	n, err := r.Read(data)
	if err != nil {
		return "", "", err
	}
	data = data[:n]

	r.s.logf("<<< %q", data)

	if matches := command_regexp.FindSubmatch(data); len(matches) == 3 {
		return strings.ToUpper(string(matches[1])), string(matches[2]), nil
	}
	return "", "", BadSyntaxError

}

func (r *Reader) ReadData() (string, error) {

	var (
		data  []byte
		total int
	)

	for {
		d := make([]byte, 24<<10)
		n, err := r.Read(d)
		if err != nil {
			return "", err
		}
		total += n
		if total > MaxMessageSize {
			return "", MessageSizeError
		}
		data = append(data, d[:n]...)
		if bytes.HasSuffix(data, []byte("\r\n.\r\n")) {
			break
		}
	}

	r.s.logf("<<< %q", data)

	dataString := string(data)

	return strings.TrimSuffix(dataString, "\r\n.\r\n"), nil

}
