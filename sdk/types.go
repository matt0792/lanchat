package sdk

import (
	"time"

	"github.com/matt0792/lanchat/internal/app"
)

type User struct {
	Nickname string
	Status   string
}

type PeerInfo struct {
	ID       string
	Nickname string
	Status   string
	LastSeen time.Time
	Metadata map[string]string
}

type Room struct {
	Name     string
	Peers    []*PeerInfo
	Messages []*ChatMessage
}

type ChatMessage struct {
	ID        string
	From      string
	Nickname  string
	Content   string
	Timestamp time.Time
	Type      MessageType
}

type MessageType string

const (
	MessageTypeText  MessageType = "text"
	MessageTypeJoin  MessageType = "join"
	MessageTypeLeave MessageType = "leave"
)

type EventType string

const (
	EventPeerJoined   EventType = "peer_joined"
	EventPeerLeft     EventType = "peer_left"
	EventMessageRecv  EventType = "message_received"
	EventRoomJoined   EventType = "room_joined"
	EventStatusChange EventType = "status_change"
)

func convertUser(u *app.User) *User {
	if u == nil {
		return nil
	}
	return &User{
		Nickname: u.Nickname,
		Status:   u.Status,
	}
}

func convertPeerInfo(p *app.PeerInfo) *PeerInfo {
	if p == nil {
		return nil
	}
	return &PeerInfo{
		ID:       p.ID.String(),
		Nickname: p.Nickname,
		Status:   p.Status,
		LastSeen: p.LastSeen,
		Metadata: p.Metadata,
	}
}

func convertRoom(r *app.Room) *Room {
	if r == nil {
		return nil
	}

	peers := make([]*PeerInfo, 0, len(r.Peers))
	for _, p := range r.Peers {
		peers = append(peers, convertPeerInfo(p))
	}

	messages := make([]*ChatMessage, len(r.Messages))
	for i, m := range r.Messages {
		messages[i] = convertChatMessage(m)
	}

	return &Room{
		Name:     r.Name,
		Peers:    peers,
		Messages: messages,
	}
}

func convertChatMessage(msg *app.ChatMessage) *ChatMessage {
	if msg == nil {
		return nil
	}
	return &ChatMessage{
		ID:        msg.ID,
		From:      msg.From.String(),
		Nickname:  msg.Nickname,
		Content:   msg.Content,
		Timestamp: msg.Timestamp,
		Type:      MessageType(msg.Type),
	}
}

func convertMessageType(mt app.MessageType) MessageType {
	return MessageType(mt)
}

func convertEventType(et app.EventType) EventType {
	return EventType(et)
}

type Logger interface {
	LogInfo(message string)
	LogError(message string)
}

type defaultLogger struct{}

func (dl *defaultLogger) LogInfo(message string)  {}
func (dl *defaultLogger) LogError(message string) {}

type EventHandler interface {
	HandleMessageRecv(*ChatMessage)
	HandlePeerJoined(*PeerInfo)
	HandleRoomJoined(*Room)
}

type BaseEventHandler struct{}

func (h *BaseEventHandler) HandleMessageRecv(msg *ChatMessage) {}

func (h *BaseEventHandler) HandlePeerJoined(peerInfo *PeerInfo) {}

func (h *BaseEventHandler) HandleRoomJoined(room *Room) {}

type Bot interface {
	Initialize(lc *Lanchat) error
	OnMessage(msg ChatMessage, lc *Lanchat) error
	OnPeerJoined(peer PeerInfo, lc *Lanchat) error
	OnRoomJoined(room Room, lc *Lanchat) error
}
