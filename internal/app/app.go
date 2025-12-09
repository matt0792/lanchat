package app

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/matt0792/lanchat/internal/logger"
	"github.com/matt0792/lanchat/internal/p2p"
)

const (
	maxMessageLength  = 100
	maxNicknameLength = 30
	maxRoomNameLength = 30

	rateLimitAmount = 20
	rateLimitWindow = 10 * time.Second
)

const (
	letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits  = "0123456789"
	symbols = "-_/!:?() "
	allowed = letters + digits + symbols
)

type App struct {
	ctx    context.Context
	cancel context.CancelFunc

	host *p2p.Host
	user *User

	currentRoom     *Room
	currentRoomName string
	topic           *p2p.Topic

	// all known peers across all rooms
	peers   map[peer.ID]*PeerInfo
	peersMu sync.RWMutex

	rateLimiter *RateLimiter

	events chan Event
}

func NewApp(ctx context.Context, nickname string) (*App, error) {
	appCtx, cancel := context.WithCancel(ctx)

	nickname = sanitize(nickname)
	if len(nickname) == 0 {
		cancel()
		return nil, fmt.Errorf("invalid nickname")
	}
	if len(nickname) > maxNicknameLength {
		nickname = nickname[:maxNicknameLength]
	}

	host, err := p2p.NewHost(appCtx)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	// set metadata
	user := &User{
		Nickname: nickname,
		Status:   "online",
	}

	host.SetMetadata(p2p.MetadataResponse{
		Nickname: nickname,
		Version:  "1.0.0",
		Custom: map[string]string{
			"status": user.Status,
		},
	})

	app := &App{
		ctx:         appCtx,
		cancel:      cancel,
		host:        host,
		user:        user,
		peers:       make(map[peer.ID]*PeerInfo),
		events:      make(chan Event, 100),
		rateLimiter: NewRateLimiter(rateLimitAmount, rateLimitWindow),
	}

	// start discovery
	if err := host.StartDiscovery("chatv1"); err != nil {
		cancel()
		host.Close()
		return nil, fmt.Errorf("failed to start discovery: %w", err)
	}

	go app.handlePeerDiscovery()
	go app.startRateLimiterCleanup()

	logger.Info("app initialized for user: %s (ID: %s)", nickname, host.ID().String()[:8])

	return app, nil
}

func (a *App) handlePeerDiscovery() {
	for {
		select {
		case <-a.ctx.Done():
			return
		case peerInfo := <-a.host.GetPeerChan():
			go a.onPeerDiscovered(peerInfo)
		}
	}
}

func (a *App) onPeerDiscovered(peerInfo peer.AddrInfo) {
	logger.Debug("Peer discovered: %s", peerInfo.ID.String()[:8])

	time.Sleep(500 * time.Millisecond)

	md, err := a.host.RequestPeerMetadata(peerInfo.ID)
	if err != nil {
		logger.Warn("Failed to get metadata from peer %s: %v", peerInfo.ID.String()[:8], err)
		return
	}

	nickname := sanitize(md.Nickname)
	if len(nickname) == 0 {
		nickname = "Unknown"
	}
	if len(nickname) > maxNicknameLength {
		nickname = nickname[:maxNicknameLength]
	}

	status := md.Custom["status"]
	if len(status) > 50 {
		status = status[:50]
	}

	a.peersMu.Lock()
	a.peers[peerInfo.ID] = &PeerInfo{
		ID:       peerInfo.ID,
		Nickname: nickname,
		Status:   status,
		LastSeen: time.Now(),
		Metadata: md.Custom,
	}
	a.peersMu.Unlock()

	logger.Info("Peer %s connected: %s", peerInfo.ID.String()[:8], md.Nickname)

	a.events <- Event{
		Type: EventPeerJoined,
		Data: a.peers[peerInfo.ID],
	}
}

func (a *App) JoinRoom(roomName, password string) error {
	roomName = sanitize(roomName)
	if len(roomName) == 0 {
		return fmt.Errorf("invalid room name")
	}
	if len(roomName) > maxRoomNameLength {
		roomName = roomName[:maxRoomNameLength]
	}

	if a.currentRoom != nil {
		if err := a.LeaveRoom(); err != nil {
			logger.Warn("Error leaving current room: %v", err)
		}
	}

	topicName := fmt.Sprintf("chat/rooms/%s", roomName)
	if password != "" {
		hash := sha256.Sum256([]byte(password))
		topicName = fmt.Sprintf("chat/rooms/%s/%x", roomName, hash[:8])
	}

	topic, err := a.host.JoinTopic(topicName)
	if err != nil {
		return fmt.Errorf("failed to join topic: %w", err)
	}

	room := &Room{
		Name:     roomName,
		Topic:    topicName,
		Peers:    make(map[peer.ID]*PeerInfo),
		Messages: make([]*ChatMessage, 0),
		Password: password,
	}

	if password != "" {
		room.EncryptionKey = DeriveKey(password, roomName)
		logger.Info("Room encryption enabled")
	}

	a.currentRoom = room
	a.currentRoomName = roomName
	a.topic = topic

	a.host.RegisterMessageHandler(p2p.MessageTypeChat, a.handleChatMessage)

	go a.readMessages()

	joinMsg := map[string]string{
		"type":     string(MessageTypeJoin),
		"nickname": a.user.Nickname,
	}
	if err := topic.Publish(p2p.MessageTypeChat, joinMsg); err != nil {
		logger.Warn("Failed to announce join: %v", err)
	}

	md := a.host.GetMetadata()
	md.CurrentRoom = roomName
	if password != "" {
		md.Custom["room_encrypted"] = "true"
	}
	a.host.SetMetadata(md)

	logger.Info("Joined room: %s", roomName)
	a.events <- Event{
		Type: EventRoomJoined,
		Data: room,
	}

	return nil
}

func (a *App) LeaveRoom() error {
	md := a.host.GetMetadata()
	md.CurrentRoom = ""
	delete(md.Custom, "room_encrypted")
	a.host.SetMetadata(md)

	if a.currentRoom == nil {
		return nil
	}

	leaveMsg := map[string]string{
		"type":     string(MessageTypeLeave),
		"nickname": a.user.Nickname,
	}
	if err := a.topic.Publish(p2p.MessageTypeChat, leaveMsg); err != nil {
		logger.Warn("Failed to announce leave: %v", err)
	}

	if err := a.topic.Close(); err != nil {
		logger.Warn("Error closing topic: %v", err)
	}

	logger.Info("Left room: %s", a.currentRoom.Name)

	a.currentRoom = nil
	a.currentRoomName = ""
	a.topic = nil

	return nil
}

func (a *App) GetRoomList() []string {
	roomSet := make(map[string]bool)

	if a.currentRoomName != "" {
		roomName := a.currentRoomName
		if a.currentRoom.Password != "" {
			roomName = fmt.Sprintf("%s (encrypted)", roomName)
		}
		roomSet[roomName] = true
	}

	for peerID := range a.peers {
		metadata, err := a.host.RequestPeerMetadata(peerID)
		if err == nil && metadata.CurrentRoom != "" {
			if metadata.Custom["room_encrypted"] == "true" {
				continue
			}
			roomName := sanitize(metadata.CurrentRoom)
			if len(roomName) > 0 && len(roomName) <= maxRoomNameLength {
				roomSet[roomName] = true
			}
		}
	}

	rooms := make([]string, 0, len(roomSet))
	for room := range roomSet {
		rooms = append(rooms, room)
	}
	return rooms
}

func (a *App) GetPeerList() []string {
	peers := a.GetPeers()
	peerList := make([]string, 0)

	for _, peer := range peers {
		nickname := sanitize(peer.Nickname)
		metadata, err := a.host.RequestPeerMetadata(peer.ID)
		if err == nil && metadata.CurrentRoom != "" && metadata.Custom["room_encrypted"] != "true" {
			roomName := sanitize(metadata.CurrentRoom)
			peerList = append(peerList, fmt.Sprintf("%s (In room: %s)", nickname, roomName))
		} else {
			peerList = append(peerList, nickname)
		}
	}
	return peerList
}

func (a *App) readMessages() {
	msgChan := a.topic.ReadMessages(a.ctx)
	for msg := range msgChan {
		// messages dealt with by handler
		_ = msg
	}
}

func (a *App) handleChatMessage(msg *p2p.Message) error {
	if a.currentRoom == nil {
		return nil
	}

	peerID, err := peer.Decode(msg.From)
	if err != nil {
		logger.Warn("Invalid peer ID in message: %v", err)
		return err
	}

	if !a.rateLimiter.Allow(peerID) {
		logger.Warn("Rate limit exceeded for peer %s", peerID.String()[:8])
		return nil
	}

	var content map[string]interface{}
	if err := json.Unmarshal(msg.Data, &content); err != nil {
		logger.Warn("Failed to parse message: %v", err)
		return err
	}

	msgType := MessageTypeText
	if t, ok := content["type"].(string); ok {
		msgType = MessageType(t)
	}

	// Get peer info
	a.peersMu.RLock()
	peerInfo := a.peers[peerID]
	a.peersMu.RUnlock()

	nickname := "Unknown"
	if peerInfo != nil {
		nickname = peerInfo.Nickname
	} else if n, ok := content["nickname"].(string); ok {
		nickname = sanitize(n)
		if len(nickname) == 0 {
			nickname = "Unknown"
		}
		if len(nickname) > maxNicknameLength {
			nickname = nickname[:maxNicknameLength]
		}
	}

	switch msgType {
	case MessageTypeJoin:
		logger.Debug("Peer %s joined room", nickname)
		if peerInfo != nil {
			a.currentRoom.Peers[peerID] = peerInfo
		}

		chatMsg := &ChatMessage{
			ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
			From:      peerID,
			Nickname:  nickname,
			Content:   fmt.Sprintf("%s joined the room", nickname),
			Timestamp: msg.Timestamp,
			Type:      MessageTypeJoin,
		}
		a.currentRoom.Messages = append(a.currentRoom.Messages, chatMsg)
		a.events <- Event{Type: EventMessageRecv, Data: chatMsg}

	case MessageTypeLeave:
		logger.Debug("Peer %s left room", nickname)
		delete(a.currentRoom.Peers, peerID)

		chatMsg := &ChatMessage{
			ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
			From:      peerID,
			Nickname:  nickname,
			Content:   fmt.Sprintf("%s left the room", nickname),
			Timestamp: msg.Timestamp,
			Type:      MessageTypeLeave,
		}
		a.currentRoom.Messages = append(a.currentRoom.Messages, chatMsg)
		a.events <- Event{Type: EventMessageRecv, Data: chatMsg}

	case MessageTypeText:
		text, ok := content["text"].(string)
		if !ok {
			return nil
		}

		if a.currentRoom.EncryptionKey != nil {
			decrypted, err := Decrypt(text, a.currentRoom.EncryptionKey)
			if err != nil {
				logger.Warn("Failed to decrypt message from %s (wrong password?): %w", nickname, err)
				text = ""
			} else {
				text = decrypted
			}

		}

		text = sanitize(text)
		if len(text) == 0 {
			logger.Debug("Dropped empty message after sanitization from %s", nickname)
			return nil
		}
		if len(text) > maxMessageLength {
			text = text[:maxMessageLength]
			logger.Debug("Truncated oversized message from %s", nickname)
		}

		chatMsg := &ChatMessage{
			ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
			From:      peerID,
			Nickname:  nickname,
			Content:   text,
			Timestamp: msg.Timestamp,
			Type:      MessageTypeText,
		}
		a.currentRoom.Messages = append(a.currentRoom.Messages, chatMsg)
		a.events <- Event{Type: EventMessageRecv, Data: chatMsg}
	}

	return nil
}

func (a *App) SendMessage(text string) error {
	if a.currentRoom == nil {
		return fmt.Errorf("not in a room")
	}

	text = sanitize(text)
	if len(text) == 0 {
		return fmt.Errorf("message cannot be empty after sanitization")
	}
	if len(text) > maxMessageLength {
		return fmt.Errorf("message too long (max %d characters)", maxMessageLength)
	}

	messageText := text

	if a.currentRoom.EncryptionKey != nil {
		encrypted, err := Encrypt(text, a.currentRoom.EncryptionKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt message: %w", err)
		}
		messageText = encrypted
	}

	msg := map[string]string{
		"type":     string(MessageTypeText),
		"text":     messageText,
		"nickname": a.user.Nickname,
	}

	if err := a.topic.Publish(p2p.MessageTypeChat, msg); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	chatMsg := &ChatMessage{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		From:      a.host.ID(),
		Nickname:  a.user.Nickname,
		Content:   text,
		Timestamp: time.Now(),
		Type:      MessageTypeText,
	}
	a.currentRoom.Messages = append(a.currentRoom.Messages, chatMsg)
	a.events <- Event{Type: EventMessageRecv, Data: chatMsg}

	return nil
}

func (a *App) GetEvents() <-chan Event {
	return a.events
}

func (a *App) GetCurrentRoom() *Room {
	return a.currentRoom
}

func (a *App) GetPeers() []*PeerInfo {
	a.peersMu.RLock()
	defer a.peersMu.RUnlock()

	peers := make([]*PeerInfo, 0, len(a.peers))
	for _, p := range a.peers {
		peers = append(peers, p)
	}
	return peers
}

func (a *App) Close() error {
	logger.Info("Closing app")

	if a.currentRoom != nil {
		a.LeaveRoom()
	}

	a.cancel()
	close(a.events)
	return a.host.Close()
}

func sanitize(text string) string {

	return strings.Map(func(r rune) rune {
		if strings.ContainsRune(allowed, r) {
			return r
		}
		return -1
	}, text)
}

func (a *App) startRateLimiterCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for {
			select {
			case <-a.ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				a.cleanupRateLimiter()
			}
		}
	}()
}

func (a *App) cleanupRateLimiter() {
	a.rateLimiter.mu.Lock()
	defer a.rateLimiter.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-a.rateLimiter.window * 2)

	for peerID, times := range a.rateLimiter.messages {
		if len(times) == 0 || times[len(times)-1].Before(cutoff) {
			delete(a.rateLimiter.messages, peerID)
		}
	}
}
