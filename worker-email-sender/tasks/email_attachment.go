package tasks

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go/jetstream"
	"log"
)

const (
	TypeEmailDeliveryAttachment = "emails.deliver_with_attachment"
)

type EmailAttachment struct {
	Filename string
	Content  string
}

type EmailDeliveryAttachmentPayload struct {
	Email       string
	Body        string
	Subject     string
	Attachments []EmailAttachment
}

func HandleEmailAttachmentDeliveryMessage(msg jetstream.Msg) {
	if err := handleEmailAttachmentDeliveryTask(msg.Data()); err != nil {
		log.Printf("Error handling email attachment delivery: %v", err)
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

func handleEmailAttachmentDeliveryTask(data []byte) error {
	var p EmailDeliveryAttachmentPayload
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v", err)
	}
	log.Printf("Sending Attachment Email to User: email=%s, subject=%s, body=%s", p.Email, p.Subject, p.Body)
	//panic("not implemented")
	return fmt.Errorf("not implemented")
}
