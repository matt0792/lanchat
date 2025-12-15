package bots

import (
	"github.com/matt0792/lanchat/sdk"
)

type TemplateBot struct{}

func (b *TemplateBot) Initialize(lc *sdk.Lanchat) error {
	return nil
}

func (b *TemplateBot) OnPeerJoined(peer sdk.PeerInfo, lc *sdk.Lanchat) error {
	return nil
}

func (b *TemplateBot) OnMessage(msg sdk.ChatMessage, lc *sdk.Lanchat) error {
	switch msg.Type {
	case sdk.MessageTypeJoin:
	case sdk.MessageTypeLeave:
	case sdk.MessageTypeText:
	}
	return nil
}

func (b *TemplateBot) OnRoomJoined(room sdk.Room, lc *sdk.Lanchat) error {
	return nil
}
