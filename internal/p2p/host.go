package p2p

import (
	"context"
	"fmt"
	"sync"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

// Host wraps the libp2p host & discovery mechanisms
type Host struct {
	host.Host
	pubsub   *pubsub.PubSub
	ctx      context.Context
	cancel   context.CancelFunc
	peerChan chan peer.AddrInfo
	mu       sync.RWMutex
	peers    map[peer.ID]peer.AddrInfo

	msgHandlers   map[MessageType]MessageHandler
	msgHandlersMu sync.RWMutex

	metadata   MetadataResponse
	metadataMu sync.RWMutex
}

func NewHost(ctx context.Context) (*Host, error) {
	hostCtx, cancel := context.WithCancel(ctx)

	h, err := libp2p.New()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}

	// create pubsub (gossipsub)
	ps, err := pubsub.NewGossipSub(hostCtx, h)
	if err != nil {
		h.Close()
		cancel()
		return nil, fmt.Errorf("failed to create pubsub: %w", err)
	}

	p2pHost := &Host{
		Host:        h,
		pubsub:      ps,
		ctx:         hostCtx,
		cancel:      cancel,
		peerChan:    make(chan peer.AddrInfo, 10),
		peers:       make(map[peer.ID]peer.AddrInfo),
		msgHandlers: make(map[MessageType]MessageHandler),
		metadata: MetadataResponse{
			Version: "1.0.0",
			Custom:  make(map[string]string),
		},
	}

	h.SetStreamHandler(ProtocolMetadata, p2pHost.handleMetadataStream)

	return p2pHost, nil
}

// StartDiscovery starts mDNS peer discovery for local network
func (h *Host) StartDiscovery(rendezvous string) error {
	// local network discovery
	mdnsService := mdns.NewMdnsService(h, rendezvous, &discoveryNotifee{h: h})
	if err := mdnsService.Start(); err != nil {
		return fmt.Errorf("failed to start mDNS: %w", err)
	}

	return nil
}

// GetPeerChan returns a channel that recieves newly discovered peers
func (h *Host) GetPeerChan() <-chan peer.AddrInfo {
	return h.peerChan
}

// GetPeers returns all currently known peers
func (h *Host) GetPeers() []peer.AddrInfo {
	h.mu.RLock()
	defer h.mu.Unlock()

	peers := make([]peer.AddrInfo, 0, len(h.peers))
	for _, p := range h.peers {
		peers = append(peers, p)
	}
	return peers
}

func (h *Host) RegisterMessageHandler(msgType MessageType, handler MessageHandler) {
	h.msgHandlersMu.Lock()
	defer h.msgHandlersMu.Unlock()
	h.msgHandlers[msgType] = handler
}

// Close shuts down host
func (h *Host) Close() error {
	h.cancel()
	return h.Host.Close()
}
