package app

import (
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

type User struct {
	Nickname string
	Status   string
}

type PeerInfo struct {
	ID       peer.ID
	Nickname string
	Status   string
	LastSeen time.Time
	Metadata map[string]string
}

type Room struct {
	Name          string
	Topic         string // pubsub topic name
	Peers         map[peer.ID]*PeerInfo
	Messages      []*ChatMessage
	Password      string
	EncryptionKey []byte
	mu            sync.RWMutex
}

type ChatMessage struct {
	ID        string
	From      peer.ID
	Identity  string
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

type Event struct {
	Type EventType
	Data interface{}
}

type EventType string

const (
	EventPeerJoined    EventType = "peer_joined"
	EventPeerLeft      EventType = "peer_left"
	EventMessageRecv   EventType = "message_received"
	EventRoomJoined    EventType = "room_joined"
	EventStatusChange  EventType = "status_change"
	EventSystemMessage EventType = "system_message"
)
