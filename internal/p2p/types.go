package p2p

import (
	"encoding/json"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/matt0792/lanchat/internal/logger"
)

const (
	ProtocolMetadata protocol.ID = "/chat/metadata/1.0.0"
)

type MessageType string

const (
	MessageTypeChat     MessageType = "chat"
	MessageTypeMetadata MessageType = "metadata"
	MessageTypeStatus   MessageType = "status"
)

// Message is the structure for pubsub messages
type Message struct {
	Type      MessageType     `json:"type"`
	From      string          `json:"from"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

// MessageHandler is a callback for handling messages
type MessageHandler func(msg *Message) error

// MetadataRequest for direct peer communication
type MetadataRequest struct {
	Type string `json:"type"`
}

type MetadataResponse struct {
	Nickname    string            `json:"nickname,omitempty"`
	Version     string            `json:"version,omitempty"`
	CurrentRoom string            `json:"current_room,omitempty"`
	Custom      map[string]string `json:"custom,omitempty"`
}

type discoveryNotifee struct {
	h *Host
}

func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	if pi.ID == n.h.ID() {
		return
	}

	n.h.mu.Lock()
	_, exists := n.h.peers[pi.ID]
	n.h.mu.Unlock()

	if !exists {
		logger.Debug("mDNS discovered peer: %s", pi.ID.String()[:8])
		if err := n.h.Connect(n.h.ctx, pi); err != nil {
			logger.Warn("Failed to connect to mDNS peer %s: %v", pi.ID.String()[:8], err)
		} else {
			logger.Info("Connected to mDNS peer: %s", pi.ID.String()[:8])
			n.h.mu.Lock()
			n.h.peers[pi.ID] = pi
			n.h.mu.Unlock()
			n.h.peerChan <- pi
		}
	}
}

type PeerEvent struct {
	PeerId peer.ID
	Type   PeerEventType
}

type PeerEventType string

const (
	PeerEventConnected    PeerEventType = "connected"
	PeerEventDisconnected PeerEventType = "disconnected"
)
