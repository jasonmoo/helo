package main

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
	response_codes = map[Status]string{
		StatusSystemStatus:                      "211 System status, or system help reply",
		StatusHelpMessage:                       "214 Help message",
		StatusServiceReady:                      "220 %s Service ready",
		StatusServiceClosingTransmissionChannel: "221 %s Service closing transmission channel",
		StatusOk:                                                  "250 OK",
		StatusUserNotLocalWillForwardTo:                           "251 User not local; will forward to %s",
		StatusStartMailInputEndWith:                               "354 Start mail input; end with <CRLF>.<CRLF>",
		StatusServiceNotAvailable:                                 "421 %s Service not available",                              // closing transmission channel [This may be a reply to any command if the service knows it must shut down]
		StatusRequestedMailActionNotTakenMailboxUnavailable:       "450 Requested mail action not taken: mailbox unavailable ", // [E.g., mailbox busy]
		StatusRequestedActionAbortedInProcessing:                  "451 Requested action aborted: error in processing",
		StatusRequestedActionNotTakenInsufficientSystemStorage:    "452 Requested action not taken: insufficient system storage",
		StatusSyntaxErrorCommandUnrecognized:                      "500 Syntax error, command unrecognized", // [This may include errors such as command line too long]
		StatusSyntaxErrorInParametersOrArguments:                  "501 Syntax error in parameters or arguments",
		StatusCommandNotImplemented:                               "502 Command not implemented",
		StatusBadSequenceOfCommands:                               "503 Bad sequence of commands",
		StatusCommandParameterNotImplemented:                      "504 Command parameter not implemented",
		StatusRequestedActionNotTakenMailboxUnavailable:           "550 Requested action not taken: mailbox unavailable", // [E.g., mailbox not found, no access]
		StatusUserNotLocalPleaseTry:                               "551 User not local; please try %s",
		StatusRequestedMailActionAbortedExceededStorageAllocation: "552 Requested mail action aborted: exceeded storage allocation",
		StatusRequestedActionNotTakenMailboxNameNotAllowed:        "553 Requested action not taken: mailbox name not allowed", // Requested action not taken: mailbox name not allowed
		StatusTransactionFailed:                                   "554 Transaction failed",
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
