package sdk

import (
	"context"

	"github.com/matt0792/lanchat/internal/app"
)

type Lanchat struct {
	app     *app.App
	handler EventHandler
	logger  Logger
}

func New(ctx context.Context, nickname string, handler EventHandler, logger Logger) (*Lanchat, error) {
	app, err := app.NewApp(ctx, nickname)
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
			l.handler.HandleMessageRecv(convertChatMessage(event.Data.(*app.ChatMessage)))

		case app.EventPeerJoined:
			l.handler.HandlePeerJoined(convertPeerInfo(event.Data.(*app.PeerInfo)))

		case app.EventRoomJoined:
			l.handler.HandleRoomJoined(convertRoom(event.Data.(*app.Room)))
		}
	}
}

func (l *Lanchat) Close() error {
	return l.app.Close()
}
