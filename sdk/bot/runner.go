package sdk

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/matt0792/lanchat/sdk"
)

type BotRunner struct {
	sdk.Bot
}

func NewBotRunner(bot sdk.Bot) *BotRunner {
	return &BotRunner{
		bot,
	}
}

func (b *BotRunner) Run(ctx context.Context, name, domain string) error {
	app, err := sdk.New(ctx, name, domain, nil, nil)
	if err != nil {
		app.Close()
		return err
	}
	defer app.Close()

	app.RegisterBot(b)

	go app.HandleEvents()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	return nil
}
