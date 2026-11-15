package main

import (
	"context"
	"fmt"
	"go.mau.fi/whatsmeow/types"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type MessageHandler struct {
	sender       *Sender
	sourceChat   string
	targetChat   string
	allChats     bool
	enableGroups bool
}

func (h *MessageHandler) eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		if h.allChats || v.Info.Chat.String() == h.sourceChat {
			text := messageText(v)
			link := findReelLink(text)
			if link == "" {
				if strings.Contains(text, "instagram.com") {
					fmt.Printf("Ignored an Instagram link that is not a reel/post: %q\n", text)
				}
				return
			}
			fmt.Println("Received an Instagram reel link '", link)
			recipient := v.Info.Chat
			if !h.allChats {
				var err error
				recipient, err = types.ParseJID(h.targetChat)
				if err != nil {
					fmt.Println("Error parsing target JID:", err)
					return
				}
			}
			if !h.enableGroups && recipient.Server == types.GroupServer {
				fmt.Println("Group sending is disabled, skipping message to", recipient.String())
				return
			}
			h.sender.SendReel(link, recipient)

		}
	}
}

// messageText returns the text of a message: a plain message, a message with
// a link preview, or the caption of an image or video.
func messageText(v *events.Message) string {
	m := v.Message
	for _, t := range []string{
		m.GetConversation(),
		m.GetExtendedTextMessage().GetText(),
		m.GetImageMessage().GetCaption(),
		m.GetVideoMessage().GetCaption(),
	} {
		if t != "" {
			return t
		}
	}
	return ""
}

// reelLinkRegex finds an Instagram reel/post link anywhere in a text, with or
// without www./m., including /share/ links (resolved by a redirect later).
var reelLinkRegex = regexp.MustCompile(`(?i)(?:https?://)?(?:www\.|m\.)?instagram\.com/(?:share/)?(?:[A-Za-z0-9_.]+/)?(?:reels?|p|tv)/[A-Za-z0-9_-]+[^\s]*`)

// findReelLink returns the first Instagram reel/post link in text, or "".
func findReelLink(text string) string {
	link := reelLinkRegex.FindString(text)
	if link != "" && !strings.HasPrefix(strings.ToLower(link), "http") {
		link = "https://" + link
	}
	return link
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, continuing...")
	}

	sourceChat := os.Getenv("SOURCE_CHAT")
	targetChat := os.Getenv("TARGET_CHAT")
	enableGroups, _ := strconv.ParseBool(os.Getenv("ENABLE_GROUP_SENDING"))

	fmt.Println("Source chat:", sourceChat)
	fmt.Println("Target chat:", targetChat)
	fmt.Println("Group sending enabled:", enableGroups)

	dbLog := waLog.Stdout("Database", "DEBUG", true)
	ctx := context.Background()
	// Create data/ if it does not exist
	if _, err := os.Stat("data"); os.IsNotExist(err) {
		err = os.Mkdir("data", 0755)
		if err != nil {
			panic(err)
		}
	}
	container, err := sqlstore.New(ctx, "sqlite3", "file:data/store.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}
	// If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		panic(err)
	}
	clientLog := waLog.Stdout("Client", "INFO", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)
	sender := NewSender(client)
	eventHandler := &MessageHandler{sender: sender, sourceChat: sourceChat, targetChat: targetChat, enableGroups: enableGroups}
	if sourceChat == "" || targetChat == "" {
		eventHandler.allChats = true
	}

	client.AddEventHandler(eventHandler.eventHandler)

	if client.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				// Render the QR code here
				// e.g. qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				// or just manually `echo 2@... | qrencode -t ansiutf8` in a terminal
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}

	sender.Start()

	// Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	sender.Stop()
	client.Disconnect()
}
