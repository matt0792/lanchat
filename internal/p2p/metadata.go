package p2p

import (
	"encoding/json"
	"fmt"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/matt0792/lanchat/internal/logger"
)

func (h *Host) SetMetadata(md MetadataResponse) {
	h.metadataMu.Lock()
	defer h.metadataMu.Unlock()
	h.metadata = md
	if h.metadata.Custom == nil {
		h.metadata.Custom = make(map[string]string)
	}
}

func (h *Host) GetMetadata() MetadataResponse {
	h.metadataMu.RLock()
	defer h.metadataMu.RUnlock()
	return h.metadata
}

func (h *Host) RequestPeerMetadata(peerID peer.ID) (*MetadataResponse, error) {
	stream, err := h.NewStream(h.ctx, peerID, ProtocolMetadata)
	if err != nil {
		return nil, fmt.Errorf("failed to open stream %w", err)
	}
	defer stream.Close()

	req := MetadataRequest{Type: "get"}
	if err := json.NewEncoder(stream).Encode(req); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// read response
	var resp MetadataResponse
	if err := json.NewDecoder(stream).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return &resp, nil
}

func (h *Host) handleMetadataStream(stream network.Stream) {
	defer stream.Close()

	var req MetadataRequest
	if err := json.NewDecoder(stream).Decode(&req); err != nil {
		logger.Warn("Failed to decode metadata request: %v", err)
		return
	}

	h.metadataMu.RLock()
	resp := h.metadata
	h.metadataMu.RUnlock()

	if err := json.NewEncoder(stream).Encode(resp); err != nil {
		logger.Warn("Failed to encode metadata response: %v", err)
	}
}
