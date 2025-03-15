package main

import (
	"context"
	"embed"
	"fmt"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"signal-chat/client/apiclient"
	"signal-chat/client/database"
	"signal-chat/client/encryption"
	"signal-chat/client/models"
	"time"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	ac := apiclient.NewFake()
	ac.SetErrorHandler(func(err error) {
		fmt.Println("Error:", err)
	})

	// Create fake user1
	db1 := database.NewFake()
	encryptor1 := encryption.NewEncryptionManager(db1, ac)
	auth1 := NewAuth(db1, ac, encryptor1)
	usr1, err := auth1.SignUp("alice@gmail.com", "test1234")
	if err != nil {
		panic(err)
	}

	// Create fake user 2
	db2 := database.NewFake()
	encryptor2 := encryption.NewEncryptionManager(db2, ac)
	auth2 := NewAuth(db2, ac, encryptor2)
	usr2, err := auth2.SignUp("bob@gmail.com", "test1234")
	if err != nil {
		panic(err)
	}
	conversations2 := NewConversationService(db2, ac, encryptor2)

	// Create sample conversation as fake user 2
	_, err = auth2.SignIn("bob@gmail.com", "test1234")
	if err != nil {
		panic(err)
	}
	conv, err := conversations2.CreateConversation("Test Conversation", []string{usr1.ID}, models.OneOnOne)
	if err != nil {
		panic(err)
	}
	time.Sleep(time.Millisecond)

	_, err = conversations2.SendMessage(conv.ID, "Hello world!")
	if err != nil {
		panic(err)
	}

	_ = usr2

	// Create conversation service for user1 after the message from user1 was sent because we want
	//ConversationService for user 2 to attach its WS handlers to APIClient
	ac.ClearHandlers()
	_ = NewConversationService(db1, ac, encryptor1)
	_, err = auth1.SignIn("alice@gmail.com", "test1234")
	if err != nil {
		panic(err)
	}

	users := NewUserService(ac)
	app := NewApp()

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "client",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			conversations2.ConversationAdded = func(conv models.Conversation) {
				runtime.EventsEmit(ctx, "conversation_added", conv)
			}
			conversations2.ConversationUpdated = func(conv models.Conversation) {
				runtime.EventsEmit(ctx, "conversation_updated", conv)
			}
			conversations2.MessageAdded = func(msg models.Message) {
				runtime.EventsEmit(ctx, "message_added", msg)
			}
			ac.SetErrorHandler(func(err error) {
				runtime.EventsEmit(ctx, "websocket_error", err)
			})
		},
		OnShutdown: func(ctx context.Context) {
			_ = ac.Close()
		},
		Bind: []interface{}{
			app,
			auth2,
			conversations2,
			users,
			encryptor2,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
