package p2p

import (
	"encoding/json"
	"fmt"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

func (h *Host) SendDirectMessage(peerID peer.ID, protocolID protocol.ID, data interface{}) error {
	stream, err := h.NewStream(h.ctx, peerID, protocolID)
	if err != nil {
		return fmt.Errorf("failed to open stream: %w", err)
	}
	defer stream.Close()

	if err := json.NewEncoder(stream).Encode(data); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (h *Host) SetStreamHandler(protocolID protocol.ID, handler network.StreamHandler) {
	h.Host.SetStreamHandler(protocolID, handler)
}
