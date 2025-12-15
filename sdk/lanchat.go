package sdk

import (
	"context"
	"fmt"

	"github.com/matt0792/lanchat/internal/app"
)

type Lanchat struct {
	app     *app.App
	handler EventHandler
	logger  Logger
	bots    []Bot
}

func New(ctx context.Context, nickname, domain string, handler EventHandler, logger Logger) (*Lanchat, error) {
	app, err := app.NewApp(ctx, nickname, domain)
	if err != nil {
		return nil, err
	}

	if handler == nil {
		handler = &BaseEventHandler{}
	}

	if logger == nil {
		logger = &defaultLogger{}
	}

	lc := &Lanchat{
		app:     app,
		handler: handler,
		logger:  logger,
	}

	return lc, nil
}

func (l *Lanchat) RegisterBot(bot Bot) error {
	l.bots = append(l.bots, bot)
	return bot.Initialize(l)
}

func (l *Lanchat) JoinRoom(roomName, password string) error {
	return l.app.JoinRoom(roomName, password)
}

func (l *Lanchat) LeaveRoom() error {
	return l.app.LeaveRoom()
}

func (l *Lanchat) GetRoomList() []string {
	return l.app.GetRoomList()
}

func (l *Lanchat) GetPeerList() []string {
	return l.app.GetPeerList()
}

func (l *Lanchat) SendMessage(text string) error {
	return l.app.SendMessage(text)
}

func (l *Lanchat) HandleEvents() {
	for event := range l.app.GetEvents() {
		switch event.Type {
		case app.EventMessageRecv:
			msg := convertChatMessage(event.Data.(*app.ChatMessage))
			l.handler.HandleMessageRecv(msg)

			for _, bot := range l.bots {
				if err := bot.OnMessage(*msg, l); err != nil {
					l.logger.LogError(fmt.Sprintf("bot error: %s", err.Error()))
				}
			}

		case app.EventPeerJoined:
			peerInfo := convertPeerInfo(event.Data.(*app.PeerInfo))
			l.handler.HandlePeerJoined(peerInfo)

			for _, bot := range l.bots {
				if err := bot.OnPeerJoined(*peerInfo, l); err != nil {
					l.logger.LogError(fmt.Sprintf("bot error: %s", err.Error()))
				}
			}

		case app.EventRoomJoined:
			room := convertRoom(event.Data.(*app.Room))
			l.handler.HandleRoomJoined(room)

			for _, bot := range l.bots {
				if err := bot.OnRoomJoined(*room, l); err != nil {
					l.logger.LogError(fmt.Sprintf("bot error: %s", err.Error()))
				}
			}
		}
	}
}

func (l *Lanchat) Close() error {
	return l.app.Close()
}
