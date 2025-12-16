package p2p

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/multiformats/go-multiaddr"
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

	peerEventChan chan PeerEvent

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
		Host:          h,
		pubsub:        ps,
		ctx:           hostCtx,
		cancel:        cancel,
		peerChan:      make(chan peer.AddrInfo, 10),
		peerEventChan: make(chan PeerEvent, 10),
		peers:         make(map[peer.ID]peer.AddrInfo),
		msgHandlers:   make(map[MessageType]MessageHandler),
		metadata: MetadataResponse{
			Version: "1.0.0",
			Custom:  make(map[string]string),
		},
	}

	h.SetStreamHandler(ProtocolMetadata, p2pHost.handleMetadataStream)

	p2pHost.setupNetworkNotifications()

	go p2pHost.cleanupStalePeers()

	return p2pHost, nil
}

func (h *Host) StartDiscovery(rendezvous string) error {
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

func (h *Host) GetPeerEventChan() <-chan PeerEvent {
	return h.peerEventChan
}

func (h *Host) IsConnected(peerId peer.ID) bool {
	return h.Network().Connectedness(peerId) == network.Connected
}

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

func (h *Host) setupNetworkNotifications() {
	notifee := &network.NotifyBundle{
		ConnectedF: func(_ network.Network, conn network.Conn) {
			peerId := conn.RemotePeer()

			if peerId == h.ID() {
				return
			}

			h.mu.Lock()
			h.peers[peerId] = peer.AddrInfo{
				ID:    peerId,
				Addrs: []multiaddr.Multiaddr{conn.RemoteMultiaddr()},
			}
			h.mu.Unlock()

			select {
			case h.peerEventChan <- PeerEvent{PeerId: peerId, Type: PeerEventConnected}:
			case <-h.ctx.Done():
			}
		},
		DisconnectedF: func(_ network.Network, conn network.Conn) {
			peerId := conn.RemotePeer()

			if len(h.Network().ConnsToPeer(peerId)) > 0 {
				return
			}

			h.mu.Lock()
			delete(h.peers, peerId)
			h.mu.Unlock()

			select {
			case h.peerEventChan <- PeerEvent{PeerId: peerId, Type: PeerEventDisconnected}:
			case <-h.ctx.Done():
			}
		},
	}

	h.Network().Notify(notifee)
}

func (h *Host) cleanupStalePeers() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			h.mu.Lock()
			for peerId := range h.peers {
				if h.Network().Connectedness(peerId) != network.Connected {
					delete(h.peers, peerId)

					select {
					case h.peerEventChan <- PeerEvent{PeerId: peerId, Type: PeerEventDisconnected}:
					default:
					}
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Host) Close() error {
	h.cancel()
	return h.Host.Close()
}
