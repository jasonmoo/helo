package helo

import (
	"crypto/rand"
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
	// S: 220 helo Service ready
	// F: 421 helo Service not available
	w.WriteStatus(StatusServiceReady)

	for {
		command, arg, err := r.ReadCommand()
		if err != nil {
			if err == BadSyntaxError {
				w.WriteStatus(StatusSyntaxErrorCommandUnrecognized)
				continue
			}
			s.logger.Println(err)
			w.WriteStatus(StatusRequestedActionAbortedInProcessing)
			return
		}

		switch command {

		// smtp
		case CommandHelo:
			// HELO <SP> <domain> <CRLF>
			//
			// This command is used to identify the sender-SMTP to the
			// receiver-SMTP.  The argument field contains the host name of
			// the sender-SMTP.
			//
			// The receiver-SMTP identifies itself to the sender-SMTP in
			// the connection greeting reply, and in the response to this
			// command.
			//
			// This command and an OK reply to it confirm that both the
			// sender-SMTP and the receiver-SMTP are in the initial state,
			// that is, there is no transaction in progress and all state
			// tables and buffers are cleared.
			//
			// S: 250 OK
			// E: 421 helo Service not available
			// E: 500 Syntax error, command unrecognized
			// E: 501 Syntax error in parameters or arguments
			// E: 504 Command parameter not implemented
			w.WriteReply(StatusOk, "helo at your service")

		case CommandMail:
			// MAIL <SP> FROM:<reverse-path> <CRLF>
			//
			// This command is used to initiate a mail transaction in which
			// the mail data is delivered to one or more mailboxes.  The
			// argument field contains a reverse-path.
			//
			// The reverse-path consists of an optional list of hosts and
			// the sender mailbox.  When the list of hosts is present, it
			// is a "reverse" source route and indicates that the mail was
			// relayed through each host on the list (the first host in the
			// list was the most recent relay).  This list is used as a
			// source route to return non-delivery notices to the sender.
			// As each relay host adds itself to the beginning of the list,
			// it must use its name as known in the IPCE to which it is
			// relaying the mail rather than the IPCE from which the mail
			// came (if they are different).  In some types of error
			// reporting messages (for example, undeliverable mail
			// notifications) the reverse-path may be null (see Example 7).
			//
			// This command clears the reverse-path buffer, the
			// forward-path buffer, and the mail data buffer; and inserts
			// the reverse-path information from this command into the
			// reverse-path buffer.
			//
			// S: 250 OK
			// E: 421 helo Service not available
			// E: 500 Syntax error, command unrecognized
			// E: 501 Syntax error in parameters or arguments
			// F: 451 Requested action aborted: error in processing
			// F: 452 Requested action not taken: insufficient system storage
			// F: 552 Requested mail action aborted: exceeded storage allocation

		case CommandRcpt:
			// RCPT <SP> TO:<forward-path> <CRLF>
			//
			// This command is used to identify an individual recipient of
			// the mail data; multiple recipients are specified by multiple
			// use of this command.
			//
			// The forward-path consists of an optional list of hosts and a
			// required destination mailbox.  When the list of hosts is
			// present, it is a source route and indicates that the mail
			// must be relayed to the next host on the list.  If the
			// receiver-SMTP does not implement the relay function it may
			// user the same reply it would for an unknown local user
			// (550).
			//
			// When mail is relayed, the relay host must remove itself from
			// the beginning forward-path and put itself at the beginning
			// of the reverse-path.  When mail reaches its ultimate
			// destination (the forward-path contains only a destination
			// mailbox), the receiver-SMTP inserts it into the destination
			// mailbox in accordance with its host mail conventions.
			//
			//    For example, mail received at relay host A with arguments
			//
			//       FROM:<USERX@HOSTY.ARPA>
			//       TO:<@HOSTA.ARPA,@HOSTB.ARPA:USERC@HOSTD.ARPA>
			//
			//    will be relayed on to host B with arguments
			//
			//       FROM:<@HOSTA.ARPA:USERX@HOSTY.ARPA>
			//       TO:<@HOSTB.ARPA:USERC@HOSTD.ARPA>.
			//
			// This command causes its forward-path argument to be appended
			// to the forward-path buffer.
			//
			// S: 250 OK
			// S: 251 User not local; will forward to %s
			// E: 421 helo Service not available
			// E: 500 Syntax error, command unrecognized
			// E: 501 Syntax error in parameters or arguments
			// E: 503 Bad sequence of commands
			// F: 450 Requested mail action not taken: mailbox unavailable
			// F: 451 Requested action aborted: error in processing
			// F: 452 Requested action not taken: insufficient system storage
			// F: 550 Requested action not taken: mailbox unavailable
			// F: 551 User not local; please try %s
			// F: 552 Requested mail action aborted: exceeded storage allocation
			// F: 553 Requested action not taken: mailbox name not allowed

		case CommandData:
			// DATA <CRLF>
			//
			// The receiver treats the lines following the command as mail
			// data from the sender.  This command causes the mail data
			// from this command to be appended to the mail data buffer.
			// The mail data may contain any of the 128 ASCII character
			// codes.
			//
			// The mail data is terminated by a line containing only a
			// period, that is the character sequence "<CRLF>.<CRLF>" (see
			// Section 4.5.2 on Transparency).  This is the end of mail
			// data indication.
			//
			// The end of mail data indication requires that the receiver
			// must now process the stored mail transaction information.
			// This processing consumes the information in the reverse-path
			// buffer, the forward-path buffer, and the mail data buffer,
			// and on the completion of this command these buffers are
			// cleared.  If the processing is successful the receiver must
			// send an OK reply.  If the processing fails completely the
			// receiver must send a failure reply.
			//
			// When the receiver-SMTP accepts a message either for relaying
			// or for final delivery it inserts at the beginning of the
			// mail data a time stamp line.  The time stamp line indicates
			// the identity of the host that sent the message, and the
			// identity of the host that received the message (and is
			// inserting this time stamp), and the date and time the
			// message was received.  Relayed messages will have multiple
			// time stamp lines.
			//
			// When the receiver-SMTP makes the "final delivery" of a
			// message it inserts at the beginning of the mail data a
			//
			// return path line.  The return path line preserves the
			// information in the <reverse-path> from the MAIL command.
			// Here, final delivery means the message leaves the SMTP
			// world.  Normally, this would mean it has been delivered to
			// the destination user, but in some cases it may be further
			// processed and transmitted by another mail system.
			//
			//    It is possible for the mailbox in the return path be
			//    different from the actual sender's mailbox, for example,
			//    if error responses are to be delivered a special error
			//    handling mailbox rather than the message senders.
			//
			// The preceding two paragraphs imply that the final mail data
			// will begin with a  return path line, followed by one or more
			// time stamp lines.  These lines will be followed by the mail
			// data header and body [2].  See Example 8.
			//
			// Special mention is needed of the response and further action
			// required when the processing following the end of mail data
			// indication is partially successful.  This could arise if
			// after accepting several recipients and the mail data, the
			// receiver-SMTP finds that the mail data can be successfully
			// delivered to some of the recipients, but it cannot be to
			// others (for example, due to mailbox space allocation
			// problems).  In such a situation, the response to the DATA
			// command must be an OK reply.  But, the receiver-SMTP must
			// compose and send an "undeliverable mail" notification
			// message to the originator of the message.  Either a single
			// notification which lists all of the recipients that failed
			// to get the message, or separate notification messages must
			// be sent for each failed recipient (see Example 7).  All
			// undeliverable mail notification messages are sent using the
			// MAIL command (even if they result from processing a SEND,
			// SOML, or SAML command).
			//
			// I: 354 Start mail input; end with <CRLF>.<CRLF>
			//      -> data ->
			//                S: 250 OK
			//                F: 451 Requested action aborted: error in processing
			//                F: 452 Requested action not taken: insufficient system storage
			//                F: 552 Requested mail action aborted: exceeded storage allocation
			//                F: 554 Transaction failed
			// E: 421 helo Service not available
			// E: 500 Syntax error, command unrecognized
			// E: 501 Syntax error in parameters or arguments
			// E: 503 Bad sequence of commands
			// F: 451 Requested action aborted: error in processing
			// F: 554 Transaction failed

		case CommandRset:
			// RSET <CRLF>
			//
			// This command specifies that the current mail transaction is
			// to be aborted.  Any stored sender, recipients, and mail data
			// must be discarded, and all buffers and state tables cleared.
			// The receiver must send an OK reply.
			//
			// S: 250 OK
			// E: 421 helo Service not available
			// E: 500 Syntax error, command unrecognized
			// E: 501 Syntax error in parameters or arguments
			// E: 504 Command parameter not implemented

		case CommandSend:
			// SEND <SP> FROM:<reverse-path> <CRLF>
			//
			// This command is used to initiate a mail transaction in which
			// the mail data is delivered to one or more terminals.  The
			// argument field contains a reverse-path.  This command is
			// successful if the message is delivered to a terminal.
			//
			// The reverse-path consists of an optional list of hosts and
			// the sender mailbox.  When the list of hosts is present, it
			// is a "reverse" source route and indicates that the mail was
			// relayed through each host on the list (the first host in the
			// list was the most recent relay).  This list is used as a
			// source route to return non-delivery notices to the sender.
			// As each relay host adds itself to the beginning of the list,
			// it must use its name as known in the IPCE to which it is
			// relaying the mail rather than the IPCE from which the mail
			// came (if they are different).
			//
			// This command clears the reverse-path buffer, the
			// forward-path buffer, and the mail data buffer; and inserts
			// the reverse-path information from this command into the
			// reverse-path buffer.
			//
			// S: 250 OK
			// E: 421 helo Service not available
			// E: 500 Syntax error, command unrecognized
			// E: 501 Syntax error in parameters or arguments
			// E: 502 Command not implemented
			// F: 451 Requested action aborted: error in processing
			// F: 452 Requested action not taken: insufficient system storage
			// F: 552 Requested mail action aborted: exceeded storage allocation

		case CommandSoml:
			// SOML <SP> FROM:<reverse-path> <CRLF>
			//
			// This command is used to initiate a mail transaction in which
			// the mail data is delivered to one or more terminals or
			//
			// mailboxes. For each recipient the mail data is delivered to
			// the recipient's terminal if the recipient is active on the
			// host (and accepting terminal messages), otherwise to the
			// recipient's mailbox.  The argument field contains a
			// reverse-path.  This command is successful if the message is
			// delivered to a terminal or the mailbox.
			//
			// The reverse-path consists of an optional list of hosts and
			// the sender mailbox.  When the list of hosts is present, it
			// is a "reverse" source route and indicates that the mail was
			// relayed through each host on the list (the first host in the
			// list was the most recent relay).  This list is used as a
			// source route to return non-delivery notices to the sender.
			// As each relay host adds itself to the beginning of the list,
			// it must use its name as known in the IPCE to which it is
			// relaying the mail rather than the IPCE from which the mail
			// came (if they are different).
			//
			// This command clears the reverse-path buffer, the
			// forward-path buffer, and the mail data buffer; and inserts
			// the reverse-path information from this command into the
			// reverse-path buffer.
			//
			// S: 250 OK
			// E: 421 helo Service not available
			// E: 500 Syntax error, command unrecognized
			// E: 501 Syntax error in parameters or arguments
			// E: 502 Command not implemented
			// F: 451 Requested action aborted: error in processing
			// F: 452 Requested action not taken: insufficient system storage
			// F: 552 Requested mail action aborted: exceeded storage allocation

		case CommandSaml:
			// SAML <SP> FROM:<reverse-path> <CRLF>
			//
			// This command is used to initiate a mail transaction in which
			// the mail data is delivered to one or more terminals and
			// mailboxes. For each recipient the mail data is delivered to
			// the recipient's terminal if the recipient is active on the
			// host (and accepting terminal messages), and for all
			// recipients to the recipient's mailbox.  The argument field
			// contains a reverse-path.  This command is successful if the
			// message is delivered to the mailbox.
			//
			// The reverse-path consists of an optional list of hosts and
			// the sender mailbox.  When the list of hosts is present, it
			// is a "reverse" source route and indicates that the mail was
			// relayed through each host on the list (the first host in the
			// list was the most recent relay).  This list is used as a
			// source route to return non-delivery notices to the sender.
			// As each relay host adds itself to the beginning of the list,
			// it must use its name as known in the IPCE to which it is
			// relaying the mail rather than the IPCE from which the mail
			// came (if they are different).
			//
			// This command clears the reverse-path buffer, the
			//
			// forward-path buffer, and the mail data buffer; and inserts
			// the reverse-path information from this command into the
			// reverse-path buffer.
			//
			// S: 250 OK
			// E: 421 helo Service not available
			// E: 500 Syntax error, command unrecognized
			// E: 501 Syntax error in parameters or arguments
			// E: 502 Command not implemented
			// F: 451 Requested action aborted: error in processing
			// F: 452 Requested action not taken: insufficient system storage
			// F: 552 Requested mail action aborted: exceeded storage allocation

		case CommandVrfy:
			// VRFY <SP> <string> <CRLF>
			//
			// This command asks the receiver to confirm that the argument
			// identifies a user.  If it is a user name, the full name of
			// the user (if known) and the fully specified mailbox are
			// returned.
			//
			// This command has no effect on any of the reverse-path
			// buffer, the forward-path buffer, or the mail data buffer.
			//
			// S: 250 OK
			// S: 251 User not local; will forward to %s
			// E: 421 helo Service not available
			// E: 500 Syntax error, command unrecognized
			// E: 501 Syntax error in parameters or arguments
			// E: 502 Command not implemented
			// E: 504 Command parameter not implemented
			// F: 550 Requested action not taken: mailbox unavailable
			// F: 551 User not local; please try %s
			// F: 553 Requested action not taken: mailbox name not allowed

		case CommandExpn:
			// EXPN <SP> <string> <CRLF>
			//
			// This command asks the receiver to confirm that the argument
			// identifies a mailing list, and if so, to return the
			// membership of that list.  The full name of the users (if
			// known) and the fully specified mailboxes are returned in a
			// multiline reply.
			//
			// This command has no effect on any of the reverse-path
			// buffer, the forward-path buffer, or the mail data buffer.
			//
			// S: 250 OK
			// E: 421 helo Service not available
			// E: 500 Syntax error, command unrecognized
			// E: 501 Syntax error in parameters or arguments
			// E: 502 Command not implemented
			// E: 504 Command parameter not implemented
			// F: 550 Requested action not taken: mailbox unavailable

		case CommandHelp:
			// HELP [<SP> <string>] <CRLF>
			//
			// This command causes the receiver to send helpful information
			// to the sender of the HELP command.  The command may take an
			// argument (e.g., any command name) and return more specific
			// information as a response.
			//
			// This command has no effect on any of the reverse-path
			// buffer, the forward-path buffer, or the mail data buffer.
			//
			// S: 211 System status, or system help reply
			// S: 214 Help message
			// E: 421 helo Service not available
			// E: 500 Syntax error, command unrecognized
			// E: 501 Syntax error in parameters or arguments
			// E: 502 Command not implemented
			// E: 504 Command parameter not implemented

		case CommandNoop:
			// NOOP <CRLF>
			//
			// This command does not affect any parameters or previously
			// entered commands.  It specifies no action other than that
			// the receiver send an OK reply.
			//
			// This command has no effect on any of the reverse-path
			// buffer, the forward-path buffer, or the mail data buffer.
			//
			// S: 250 OK
			// E: 421 helo Service not available
			// E: 500 Syntax error, command unrecognized

		case CommandQuit:
			// QUIT <CRLF>
			//
			// This command specifies that the receiver must send an OK
			// reply, and then close the transmission channel.
			//
			// The receiver should not close the transmission channel until
			// it receives and replies to a QUIT command (even if there was
			// an error).  The sender should not close the transmission
			// channel until it send a QUIT command and receives the reply
			// (even if there was an error response to a previous command).
			// If the connection is closed prematurely the receiver should
			// act as if a RSET command had been received (canceling any
			// pending transaction, but not undoing any previously
			// completed transaction), the sender should act as if the
			// command or transaction in progress had received a temporary
			// error (4xx).
			//
			// S: 221 helo Service closing transmission channel
			// E: 500 Syntax error, command unrecognized

		case CommandTurn:
			// TURN <CRLF>
			//
			// This command specifies that the receiver must either (1)
			// send an OK reply and then take on the role of the
			// sender-SMTP, or (2) send a refusal reply and retain the role
			// of the receiver-SMTP.
			//
			// If program-A is currently the sender-SMTP and it sends the
			// TURN command and receives an OK reply (250) then program-A
			// becomes the receiver-SMTP.  Program-A is then in the initial
			// state as if the transmission channel just opened, and it
			// then sends the 220 service ready greeting.
			//
			// If program-B is currently the receiver-SMTP and it receives
			// the TURN command and sends an OK reply (250) then program-B
			// becomes the sender-SMTP.  Program-B is then in the initial
			// state as if the transmission channel just opened, and it
			// then expects to receive the 220 service ready greeting.
			//
			// To refuse to change roles the receiver sends the 502 reply.
			//
			// S: 250 OK
			// E: 500 Syntax error, command unrecognized
			// E: 503 Bad sequence of commands
			// F: 502 Command not implemented

		// esmtp:
		case CommandEHLO:
			// EHLO <domain> <CRLF>
		case Command8BITMIME:
			// 8BITMIME — 8 bit data transmission, RFC 6152
		case CommandATRN:
			// ATRN — Authenticated TURN for On-Demand Mail Relay, RFC 2645
		case CommandAUTH:
			// AUTH — Authenticated SMTP, RFC 4954
		case CommandCHUNKING:
			// CHUNKING — Chunking, RFC 3030
		case CommandDSN:
			// DSN — Delivery status notification, RFC 3461 (See Variable envelope return path)
		case CommandETRN:
			// ETRN — Extended version of remote message queue starting command TURN, RFC 1985
		case CommandPIPELINING:
			// PIPELINING — Command pipelining, RFC 2920
		case CommandSIZE:
			// SIZE — Message size declaration, RFC 1870
		case CommandSTARTTLS:
			// STARTTLS — Transport layer security, RFC 3207 (2002)
		case CommandSMTPUTF8:
			// SMTPUTF8 — Allow UTF-8 encoding in mailbox names and header fields, RFC 6531

		default:
			w.WriteStatus(StatusSyntaxErrorCommandUnrecognized)
		}
	}

}
