package helo

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"testing"
)

const (
	SmtpTestHost  = ":9991"
	SmtpsTestHost = ":9992"

	Cert = "server/cert/cert.pem"
	Key  = "server/cert/key.pem"
)

var (
	s  *SmtpServer
	ss *SmtpsServer
)

func init() {
	s = NewSmtpServer(SmtpTestHost)
	ss = NewSmtpsServer(SmtpsTestHost, Cert, Key)

	err := s.Start()
	if err != nil {
		log.Fatal(err)
	}

	err = ss.Start()
	if err != nil {
		log.Fatal(err)
	}
}

func TestSendSmtp(t *testing.T) {

	c, err := smtp.Dial(SmtpTestHost)
	if err != nil {
		t.Error(err)
	}

	// Set the sender and recipient first
	if err := c.Mail("sender@example.org"); err != nil {
		t.Error(err)
	}
	if err := c.Rcpt("recipient@example.net"); err != nil {
		t.Error(err)
	}

	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		t.Error(err)
	}
	_, err = fmt.Fprintf(wc, "This is the email body")
	if err != nil {
		t.Error(err)
	}
	err = wc.Close()
	if err != nil {
		t.Error(err)
	}

	// Send the QUIT command and close the connection.
	err = c.Quit()
	if err != nil {
		t.Error(err)
	}

}

func TestSendSmtps(t *testing.T) {

	conn, err := tls.Dial("tcp", SmtpsTestHost, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		t.Error(err)
	}

	c, err := smtp.NewClient(conn, SmtpsTestHost)
	if err != nil {
		t.Error(err)
	}

	// Set the sender and recipient first
	if err := c.Mail("sender@example.org"); err != nil {
		t.Error(err)
	}
	if err := c.Rcpt("recipient@example.net"); err != nil {
		t.Error(err)
	}

	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		t.Error(err)
	}
	_, err = fmt.Fprintf(wc, "This is the email body")
	if err != nil {
		t.Error(err)
	}
	err = wc.Close()
	if err != nil {
		t.Error(err)
	}

	// Send the QUIT command and close the connection.
	err = c.Quit()
	if err != nil {
		t.Error(err)
	}

}

func BenchmarkSendSmtp(b *testing.B) {

	s.SetLogger(nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c, err := smtp.Dial(SmtpTestHost)
		if err != nil {
			log.Fatal(err)
		}

		// Set the sender and recipient first
		if err := c.Mail("sender@example.org"); err != nil {
			log.Fatal(err)
		}
		if err := c.Rcpt("recipient@example.net"); err != nil {
			log.Fatal(err)
		}

		// Send the email body.
		wc, err := c.Data()
		if err != nil {
			log.Fatal(err)
		}
		_, err = fmt.Fprintf(wc, "This is the email body")
		if err != nil {
			log.Fatal(err)
		}
		err = wc.Close()
		if err != nil {
			log.Fatal(err)
		}

		// Send the QUIT command and close the connection.
		err = c.Quit()
		if err != nil {
			log.Fatal(err)
		}
	}

}

func BenchmarkSendSmtps(b *testing.B) {

	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	ss.SetLogger(nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		conn, err := tls.Dial("tcp", SmtpsTestHost, tlsconfig)
		if err != nil {
			log.Panic(err)
		}

		c, err := smtp.NewClient(conn, tlsconfig.ServerName)
		if err != nil {
			log.Panic(err)
		}

		// Set the sender and recipient first
		if err := c.Mail("sender@example.org"); err != nil {
			log.Fatal(err)
		}
		if err := c.Rcpt("recipient@example.net"); err != nil {
			log.Fatal(err)
		}

		// Send the email body.
		wc, err := c.Data()
		if err != nil {
			log.Fatal(err)
		}
		_, err = fmt.Fprintf(wc, "This is the email body")
		if err != nil {
			log.Fatal(err)
		}
		err = wc.Close()
		if err != nil {
			log.Fatal(err)
		}

		// Send the QUIT command and close the connection.
		err = c.Quit()
		if err != nil {
			log.Fatal(err)
		}
	}

}
