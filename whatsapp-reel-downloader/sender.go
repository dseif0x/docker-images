package main

import (
	"context"
	"fmt"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

type Reel struct {
	Link      string
	Recipient types.JID
}
type Sender struct {
	client     *whatsmeow.Client
	toSendChan chan Reel
}

func NewSender(client *whatsmeow.Client) *Sender {
	return &Sender{
		client:     client,
		toSendChan: make(chan Reel),
	}
}

func (s *Sender) Start() {
	go func() {
		for link := range s.toSendChan {
			fmt.Println("Processing link:", link)
			if err := s.downloadReelAndSendTo(link.Recipient, link.Link); err != nil {
				fmt.Println("Error sending reel:", err)
			}
		}
	}()
}

func (s *Sender) SendReel(link string, recipient types.JID) {
	s.toSendChan <- Reel{
		Link:      link,
		Recipient: recipient,
	}
}

func (s *Sender) Stop() {
	close(s.toSendChan)
}

func (s *Sender) downloadReelAndSendTo(targetJID types.JID, link string) error {
	videoBytes, _, err := DownloadInstagramReelBytes(link)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	resp, err := s.client.Upload(context.Background(), videoBytes, whatsmeow.MediaVideo)
	if err != nil {
		fmt.Println("Upload error:", err)
		return err
	}

	videoMsg := &waE2E.VideoMessage{
		Caption:       nil,
		Mimetype:      proto.String("video/mp4"), // likely mime type for Instagram
		URL:           &resp.URL,
		DirectPath:    &resp.DirectPath,
		MediaKey:      resp.MediaKey,
		FileEncSHA256: resp.FileEncSHA256,
		FileSHA256:    resp.FileSHA256,
		FileLength:    &resp.FileLength,
		// Thumbnail, ContextInfo, etc. optional
	}

	_, err = s.client.SendMessage(context.Background(), targetJID, &waE2E.Message{
		VideoMessage: videoMsg,
	})
	if err != nil {
		fmt.Println("Send error:", err)
		return err
	}
	return nil
}
