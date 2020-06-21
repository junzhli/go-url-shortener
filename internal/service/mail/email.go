package mail

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"time"
)

type EmailServiceOptions struct {
	Email    string
	Password string
	Server   string
}

type SendEmailOptions struct {
	To      string
	Subject string
	Message string
}

func (s SendEmailOptions) toBodyBytes() []byte {
	return []byte(fmt.Sprintf("To: %v\r\n"+
		"Subject: %v\r\n"+
		"\r\n"+
		"%v\r\n", s.To, s.Subject, s.Message))
}

var emailServiceLogTag = "EmailService"

func logMessage(msg interface{}) {
	log.Printf("%v: %v\n", emailServiceLogTag, msg)
}

func StartEmailService(ctx context.Context, c *EmailServiceOptions, incoming <-chan SendEmailOptions) {
	if c == nil {
		logMessage("Service disabled")
		return
	}

	auth := smtp.PlainAuth("", c.Email, c.Password, strings.Split(c.Server, ":")[0])

	logMessage("Started...")

	for {
		logMessage("Awaiting another incoming request for sending email...")

		req := <-incoming
		go func(req SendEmailOptions) {
			logMessage("=== Requested ===")
			logMessage(fmt.Sprintf("Recipient: %v", req.To))
			logMessage(fmt.Sprintf("Content: %v", req.Message))
			logMessage("=================")

			_ctx, cancel := context.WithTimeout(ctx, time.Second*30)
			done := make(chan bool)

			go func(isDone chan<- bool) {
				defer cancel()

				err := smtp.SendMail(c.Server, auth, c.Email, []string{req.To}, req.toBodyBytes())
				if err != nil {
					logMessage(fmt.Sprintf("Sending email failed | Reason: %v", err))
					isDone <- false
				}
				isDone <- true
			}(done)

			completed := false

			for {
				select {
				case isDone := <-done:
					if isDone {
						logMessage("Sending email succeed")
					}

					completed = true
				case <-_ctx.Done():
					logMessage("Sending email timed out")

					completed = true
				default:
				}

				if completed {
					break
				}
			}
		}(req)
	}
}
