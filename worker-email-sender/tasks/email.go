package tasks

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go/jetstream"
	mail "github.com/xhit/go-simple-mail/v2"
	"log"
	"os"
	"worker-email-sender/helper"
)

const (
	TypeEmailDelivery = "emails.deliver"
)

type EmailDeliveryPayload struct {
	Email   string
	Body    string
	Subject string
}

func HandleEmailDeliveryMessage(msg jetstream.Msg) {
	if err := handleEmailDeliveryTask(msg.Data()); err != nil {
		log.Printf("Error handling email delivery: %v", err)
		if err := msg.Nak(); err != nil {
			log.Printf("Error sending Nak for message: %v", err)
		}
	} else {
		err := msg.Ack()
		if err != nil {
			log.Printf("Error acknowledging message: %v", err)
		}
	}
}

func handleEmailDeliveryTask(data []byte) error {
	var p EmailDeliveryPayload
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v", err)
	}
	log.Printf("Sending Email to User: email=%s, subject=%s, body=%d", p.Email, p.Subject, len(p.Body))
	email := mail.NewMSG()
	email.SetFrom(os.Getenv("SMTP_SENDER")).
		AddTo(p.Email).
		SetSubject(p.Subject).
		SetBody(mail.TextHTML, p.Body)

	client, err := helper.GetSmtpClient()
	if err != nil {
		return fmt.Errorf("failed to get SMTP client: %w", err)
	}
	err = email.Send(client)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	log.Printf("Processed email delivery to: %s", p.Email)
	return nil
}
