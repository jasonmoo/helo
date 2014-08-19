package helo

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/smtp"
	"testing"
)

func TestSendSmtp(t *testing.T) {

	c, err := smtp.Dial(":25")
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

func TestSendSmtps(t *testing.T) {

	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         ":465",
		RootCAs:            x509.NewCertPool(),
	}

	conn, err := tls.Dial("tcp", tlsconfig.ServerName, tlsconfig)
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

func BenchmarkSendSmtp(b *testing.B) {

	for i := 0; i < b.N; i++ {
		c, err := smtp.Dial(":25")
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
		ServerName:         ":465",
		RootCAs:            x509.NewCertPool(),
	}

	for i := 0; i < b.N; i++ {
		conn, err := tls.Dial("tcp", tlsconfig.ServerName, tlsconfig)
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
