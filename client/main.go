package main

import (
	"context"
	"embed"
	"signal-chat/client/apiclient"
	"signal-chat/client/database"

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

	usr1, _ := auth.SignUp("mkolesnac@gmail.com", "test1234")
	usr2, _ := auth.SignUp("test@gmail.com", "test1234")
	conv, _ := conversations.CreateConversation("Hello world!", usr2.ID)
	_, _ = conversations.SendMessage(conv.ID, "Followup message")
	_ = usr1
	_ = usr2

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
		OnStartup:        app.startup,
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
