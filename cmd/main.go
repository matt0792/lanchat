package main

import (
	"context"
	"flag"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/matt0792/lanchat/internal/app"
	"github.com/matt0792/lanchat/internal/logger"
	"github.com/matt0792/lanchat/internal/ui"
	"github.com/matt0792/lanchat/internal/ui/cli"
	"github.com/matt0792/lanchat/internal/ui/tui"
)

func main() {
	useTUI := flag.Bool("tui", false, "Start app with TUI")
	flag.Parse()

	logger.SetLevel(logger.LevelNone)

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer cancel()

	nickname := getNickname()

	chatApp, err := app.NewApp(ctx, nickname)
	if err != nil {
		fmt.Printf("Failed to start: %v\n", err)
		return
	}
	defer chatApp.Close()

	var userInterface ui.UI
	if *useTUI {
		userInterface = tui.New(ctx)
	} else {
		userInterface = cli.New(ctx)
	}

	controller := ui.NewController(ctx, chatApp, userInterface)

	controller.Start()
}

func getNickname() string {
	var name string
	fmt.Print("Name: ")
	fmt.Scan(&name)
	return name
}
