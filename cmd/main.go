package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/matt0792/lanchat/internal/app"
	"github.com/matt0792/lanchat/internal/logger"
	"github.com/matt0792/lanchat/internal/ui"
	"github.com/matt0792/lanchat/internal/ui/cli"
)

func main() {
	logger.SetLevel(logger.LevelNone)

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer cancel()

	scanner := bufio.NewScanner(os.Stdin)

	nickname := getInput(scanner, "Name: ")
	domain := getInput(scanner, "Domain (empty for default): ")
	domain = strings.ReplaceAll(domain, " ", "")
	if domain == "" {
		domain = "lanchat"
	}

	chatApp, err := app.NewApp(ctx, nickname, domain)
	if err != nil {
		fmt.Printf("Failed to start: %v\n", err)
		return
	}
	defer chatApp.Close()

	var userInterface ui.UI
	userInterface = cli.New(ctx)

	controller := ui.NewController(ctx, chatApp, userInterface)

	controller.Start()
}

func getInput(scanner *bufio.Scanner, prompt string) string {
	fmt.Print(prompt)
	if !scanner.Scan() {
		return ""
	}
	return strings.TrimSpace(scanner.Text())
}
