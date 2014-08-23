package helo

import (
	"fmt"
	"net"
	"strconv"
)

type (
	Writer struct {
		net.Conn
		s *SmtpServer
	}
	Reply int
)

const (
	ReplySystemReply                                         Reply = 211
	ReplyHelpMessage                                         Reply = 214
	ReplyServiceReady                                        Reply = 220
	ReplyServiceClosingTransmissionChannel                   Reply = 221
	ReplyOk                                                  Reply = 250
	ReplyUserNotLocalWillForwardTo                           Reply = 251
	ReplyStartMailInputEndWith                               Reply = 354
	ReplyServiceNotAvailable                                 Reply = 421
	ReplyRequestedMailActionNotTakenMailboxUnavailable       Reply = 450
	ReplyRequestedActionAbortedInProcessing                  Reply = 451
	ReplyRequestedActionNotTakenInsufficientSystemStorage    Reply = 452
	ReplySyntaxErrorCommandUnrecognized                      Reply = 500
	ReplySyntaxErrorInParametersOrArguments                  Reply = 501
	ReplyCommandNotImplemented                               Reply = 502
	ReplyBadSequenceOfCommands                               Reply = 503
	ReplyCommandParameterNotImplemented                      Reply = 504
	ReplyRequestedActionNotTakenMailboxUnavailable           Reply = 550
	ReplyUserNotLocalPleaseTry                               Reply = 551
	ReplyRequestedMailActionAbortedExceededStorageAllocation Reply = 552
	ReplyRequestedActionNotTakenMailboxNameNotAllowed        Reply = 553
	ReplyTransactionFailed                                   Reply = 554
)

var (
	// REPLY CODES
	// http://tools.ietf.org/html/rfc821#page-35
	reply_codes = map[Reply]string{
		ReplySystemReply:                       "211 System status, or system help reply\r\n",
		ReplyHelpMessage:                       "214 http://www.google.com/search?btnI&q=RFC+2821\r\n",
		ReplyServiceReady:                      "220 helo Service ready\r\n",
		ReplyServiceClosingTransmissionChannel: "221 helo Service closing transmission channel\r\n",
		ReplyOk: "250 OK\r\n",
		ReplyUserNotLocalWillForwardTo:                           "251 User not local; will forward to %s\r\n",
		ReplyStartMailInputEndWith:                               "354 Start mail input; end with <CRLF>.<CRLF>\r\n",
		ReplyServiceNotAvailable:                                 "421 helo Service not available\r\n",                           // closing transmission channel [This may be a reply to any command if the service knows it must shut down]
		ReplyRequestedMailActionNotTakenMailboxUnavailable:       "450 Requested mail action not taken: mailbox unavailable\r\n", // [E.g., mailbox busy]
		ReplyRequestedActionAbortedInProcessing:                  "451 Requested action aborted: error in processing\r\n",
		ReplyRequestedActionNotTakenInsufficientSystemStorage:    "452 Requested action not taken: insufficient system storage\r\n",
		ReplySyntaxErrorCommandUnrecognized:                      "500 Syntax error, command unrecognized\r\n", // [This may include errors such as command line too long]
		ReplySyntaxErrorInParametersOrArguments:                  "501 Syntax error in parameters or arguments\r\n",
		ReplyCommandNotImplemented:                               "502 Command not implemented\r\n",
		ReplyBadSequenceOfCommands:                               "503 Bad sequence of commands\r\n",
		ReplyCommandParameterNotImplemented:                      "504 Command parameter not implemented\r\n",
		ReplyRequestedActionNotTakenMailboxUnavailable:           "550 Requested action not taken: mailbox unavailable\r\n", // [E.g., mailbox not found, no access]
		ReplyUserNotLocalPleaseTry:                               "551 User not local; please try %s\r\n",
		ReplyRequestedMailActionAbortedExceededStorageAllocation: "552 Requested mail action aborted: exceeded storage allocation\r\n",
		ReplyRequestedActionNotTakenMailboxNameNotAllowed:        "553 Requested action not taken: mailbox name not allowed\r\n", // Requested action not taken: mailbox name not allowed
		ReplyTransactionFailed:                                   "554 Transaction failed\r\n",
	}
)

func (w *Writer) WriteReplyCode(code Reply, args ...interface{}) error {
	if w.s.logger != nil {
		w.s.logf(">>> %q", fmt.Sprintf(reply_codes[code], args...))
	}
	_, err := fmt.Fprintf(w, reply_codes[code], args...)
	return err
}

func (w *Writer) WriteReply(code Reply, message string, args ...interface{}) error {
	if w.s.logger != nil {
		w.s.logf(">>> %q", fmt.Sprintf(strconv.Itoa(int(code))+" "+message+"\r\n", args...))
	}
	_, err := fmt.Fprintf(w, strconv.Itoa(int(code))+" "+message+"\r\n", args...)
	return err
}

func (w *Writer) WriteContinuedReply(code Reply, message string, args ...interface{}) error {
	if w.s.logger != nil {
		w.s.logf(">>> %q", fmt.Sprintf(strconv.Itoa(int(code))+"-"+message+"\r\n", args...))
	}
	_, err := fmt.Fprintf(w, strconv.Itoa(int(code))+"-"+message+"\r\n", args...)
	return err
}
