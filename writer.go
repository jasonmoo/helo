package helo

import (
	"bufio"
	"fmt"
	"net"
)

type (
	Writer struct {
		*bufio.Writer
	}
	Status int
)

const (
	StatusSystemStatus                                        Status = 211
	StatusHelpMessage                                         Status = 214
	StatusServiceReady                                        Status = 220
	StatusServiceClosingTransmissionChannel                   Status = 221
	StatusOk                                                  Status = 250
	StatusUserNotLocalWillForwardTo                           Status = 251
	StatusStartMailInputEndWith                               Status = 354
	StatusServiceNotAvailable                                 Status = 421
	StatusRequestedMailActionNotTakenMailboxUnavailable       Status = 450
	StatusRequestedActionAbortedInProcessing                  Status = 451
	StatusRequestedActionNotTakenInsufficientSystemStorage    Status = 452
	StatusSyntaxErrorCommandUnrecognized                      Status = 500
	StatusSyntaxErrorInParametersOrArguments                  Status = 501
	StatusCommandNotImplemented                               Status = 502
	StatusBadSequenceOfCommands                               Status = 503
	StatusCommandParameterNotImplemented                      Status = 504
	StatusRequestedActionNotTakenMailboxUnavailable           Status = 550
	StatusUserNotLocalPleaseTry                               Status = 551
	StatusRequestedMailActionAbortedExceededStorageAllocation Status = 552
	StatusRequestedActionNotTakenMailboxNameNotAllowed        Status = 553
	StatusTransactionFailed                                   Status = 554
)

var (
	// REPLY CODES
	// http://tools.ietf.org/html/rfc821#page-35
	response_codes = map[Status]string{
		StatusSystemStatus:                      "211 System status, or system help reply\r\n",
		StatusHelpMessage:                       "214 Help message\r\n",
		StatusServiceReady:                      "220 helo Service ready\r\n",
		StatusServiceClosingTransmissionChannel: "221 helo Service closing transmission channel\r\n",
		StatusOk:                                                  "250 OK\r\n",
		StatusUserNotLocalWillForwardTo:                           "251 User not local; will forward to %s\r\n",
		StatusStartMailInputEndWith:                               "354 Start mail input; end with <CRLF>.<CRLF>\r\n",
		StatusServiceNotAvailable:                                 "421 helo Service not available\r\n",                           // closing transmission channel [This may be a reply to any command if the service knows it must shut down]
		StatusRequestedMailActionNotTakenMailboxUnavailable:       "450 Requested mail action not taken: mailbox unavailable\r\n", // [E.g., mailbox busy]
		StatusRequestedActionAbortedInProcessing:                  "451 Requested action aborted: error in processing\r\n",
		StatusRequestedActionNotTakenInsufficientSystemStorage:    "452 Requested action not taken: insufficient system storage\r\n",
		StatusSyntaxErrorCommandUnrecognized:                      "500 Syntax error, command unrecognized\r\n", // [This may include errors such as command line too long]
		StatusSyntaxErrorInParametersOrArguments:                  "501 Syntax error in parameters or arguments\r\n",
		StatusCommandNotImplemented:                               "502 Command not implemented\r\n",
		StatusBadSequenceOfCommands:                               "503 Bad sequence of commands\r\n",
		StatusCommandParameterNotImplemented:                      "504 Command parameter not implemented\r\n",
		StatusRequestedActionNotTakenMailboxUnavailable:           "550 Requested action not taken: mailbox unavailable\r\n", // [E.g., mailbox not found, no access]
		StatusUserNotLocalPleaseTry:                               "551 User not local; please try %s\r\n",
		StatusRequestedMailActionAbortedExceededStorageAllocation: "552 Requested mail action aborted: exceeded storage allocation\r\n",
		StatusRequestedActionNotTakenMailboxNameNotAllowed:        "553 Requested action not taken: mailbox name not allowed\r\n", // Requested action not taken: mailbox name not allowed
		StatusTransactionFailed:                                   "554 Transaction failed\r\n",
	}
)

func NewWriter(conn net.Conn) *Writer {
	return &Writer{bufio.NewWriter(conn)}
}

func (w *Writer) WriteStatus(code Status) error {
	_, err := w.WriteString(response_codes[code])
	return err
}
func (w *Writer) WriteStatusf(code Status, val string) error {
	_, err := fmt.Fprintf(w, response_codes[code], val)
	return err
}

func (w *Writer) WriteReply(code Status, message string) error {
	_, err := fmt.Fprintf(w, "%d %s\r\n", code, message)
	return err
}
func (w *Writer) WriteReplyf(code Status, message string, args ...interface{}) error {
	_, err := fmt.Fprintf(w, "%d "+message+"\r\n", code, args...)
	return err
}

func (w *Writer) WriteContinuedReply(code Status, message string) error {
	_, err := fmt.Fprintf(w, "%d-%s\r\n", code, message)
	return err
}
func (w *Writer) WriteContinuedReplyf(code Status, message string, args ...interface{}) error {
	_, err := fmt.Fprintf(w, "%d-"+message+"\r\n", code, args...)
	return err
}
