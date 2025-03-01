package main

import (
	"context"
	"embed"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"signal-chat/client/apiclient"
	"signal-chat/client/database"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	db := database.NewFake()
	ac := apiclient.NewFake()
	auth := NewAuth(db, ac)
	conversations := NewConversationService(db, ac)
	users := NewUserService(ac)

	usr2, _ := auth.SignUp("test@gmail.com", "test1234")
	usr1, _ := auth.SignUp("mkolesnac@gmail.com", "test1234")

	conv, _ := conversations.CreateConversation("Test Conversation", []string{usr2.ID})
	_, _ = auth.SignIn("test@gmail.com", "test1234")
	time.Sleep(time.Millisecond)
	msg, _ := conversations.SendMessage(conv.ID, "Hello World!")
	_ = usr1
	_ = usr2

	_ = msg

	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "client",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			conversations.SetContext(ctx)

			ac.SetErrorHandler(func(err error) {
				runtime.EventsEmit(ctx, "websocket_error", err)
			})
		},
		OnShutdown: func(ctx context.Context) {
			_ = ac.Close()
		},
		Bind: []interface{}{
			app,
			auth,
			conversations,
			users,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
