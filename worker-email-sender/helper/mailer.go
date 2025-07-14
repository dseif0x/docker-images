package helper

import (
	"fmt"
	mail "github.com/xhit/go-simple-mail/v2"
	"os"
	"strconv"
	"time"
)

func GetSmtpClient() (*mail.SMTPClient, error) {
	server := mail.NewSMTPClient()
	port, err := strconv.ParseInt(os.Getenv("SMTP_PORT"), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SMTP_PORT: %v", err)
	}

	server.Host = os.Getenv("SMTP_SERVER")
	server.Port = int(port)
	server.Username = os.Getenv("SMTP_USER")
	server.Password = os.Getenv("SMTP_PASSWORD")
	server.KeepAlive = true
	server.Encryption = mail.EncryptionSSLTLS
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 30 * time.Second

	smtpClient, err := server.Connect()
	if err != nil {
		fmt.Printf("Failed to connect to SMTP server: %v\n", err)
		return nil, fmt.Errorf("failed to connect to SMTP server: %v", err)
	}
	return smtpClient, nil
}
