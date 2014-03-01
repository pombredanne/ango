package main

import (
	"fmt"
	"github.com/GeertJohan/go.incremental"
	"github.com/GeertJohan/go.rice"
	"net/http"
	"os"
)

var idInc = &incremental.Int{}

// Handler implements ChatServiceHandler
type ChatServiceSession struct {
	id int
}

// NewChatService creates and returns a new ChatServiceHandler instance
func NewChatServiceSession() ChatServiceSessionInterface {
	// Create new ChatService instance with next id
	return &ChatServiceSession{
		id: idInc.Next(),
	}
}

func (cs *ChatServiceSession) Stop(err error) {
	fmt.Printf("Stopping session %d with error: %s\n", cs.id, err)
}

func (cs *ChatServiceSession) Add(a int, b int) (c int, err error) {
	c = a + b
	fmt.Printf("Call to Add(%d, %d) will return %d\n", a, b, c)
	return c, nil
}

func (cs *ChatServiceSession) Notify(text string) {
	fmt.Printf("instance %d have notification: %s\n", cs.id, text)
}

var server = &ChatServiceServer{
	NewSession: NewChatServiceSession,
	ErrorIncommingConnection: func(err error) {
		fmt.Printf("Error setting up connection: %s\n", err)
	},
}

func main() {
	httpFiles, err := rice.FindBox("http-files")
	if err != nil {
		fmt.Printf("Error opening http filex box: %s\n", err)
		os.Exit(1)
	}

	http.Handle("/", http.FileServer(httpFiles.HTTPBox()))
	http.Handle("/websocket-ango-chatService", server)

	err = http.ListenAndServe(":8123", nil)
	if err != nil {
		fmt.Printf("Error listenAndServe: %s\n", err)
		os.Exit(1)
	}
}
