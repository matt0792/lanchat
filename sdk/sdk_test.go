package sdk

import (
	"context"
	"testing"
	"time"
)

func TestCreateDefault(t *testing.T) {
	app, err := New(context.Background(), "test", "test", nil, nil)
	if err != nil {
		t.Errorf("failed to create app: %s", err)
	}
	defer app.Close()
}

type testHandler struct {
	BaseEventHandler
	messages    chan *ChatMessage
	peersJoined chan *PeerInfo
	roomsJoined chan *Room
}

func newTestHandler() *testHandler {
	return &testHandler{
		messages:    make(chan *ChatMessage, 10),
		peersJoined: make(chan *PeerInfo, 10),
		roomsJoined: make(chan *Room, 10),
	}
}

func (h *testHandler) HandleMessageRecv(msg *ChatMessage) {
	h.messages <- msg
}

func (h *testHandler) HandlePeerJoined(peer *PeerInfo) {
	h.peersJoined <- peer
}

func (h *testHandler) HandleRoomJoined(room *Room) {
	h.roomsJoined <- room
}

func TestMessageRoundTrip(t *testing.T) {
	ctx := context.Background()

	handler1 := newTestHandler()
	app1, err := New(ctx, "testUser1", "test", handler1, nil)
	if err != nil {
		t.Fatalf("failed to create app1: %v", err)
	}
	defer app1.Close()

	go app1.HandleEvents()

	handler2 := newTestHandler()
	app2, err := New(ctx, "testUser2", "test", handler2, nil)
	if err != nil {
		t.Fatalf("failed to create app2: %v", err)
	}
	defer app2.Close()

	go app2.HandleEvents()

	roomName := "test-room"
	if err := app1.JoinRoom(roomName, ""); err != nil {
		t.Fatalf("app1 failed to join room: %v", err)
	}

	select {
	case <-handler1.roomsJoined:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for app1 to join room")
	}

	if err := app2.JoinRoom(roomName, ""); err != nil {
		t.Fatalf("app2 failed to join room: %v", err)
	}

	select {
	case <-handler2.roomsJoined:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for app2 to join room")
	}

	select {
	case <-handler1.peersJoined:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for peer discovery")
	}

	time.Sleep(500 * time.Millisecond)

	testMessage := "Hello World!"
	if err := app1.SendMessage(testMessage); err != nil {
		t.Fatalf("failed to send message: %v", err)
	}

	select {
	case msg := <-handler2.messages:
		if msg.Content != testMessage {
			t.Errorf("received wrong message: got %q, want %q", msg.Content, testMessage)
		}
		if msg.Nickname != "testUser1" {
			t.Errorf("wrong sender: got %q, want %q", msg.Nickname, "testUser1")
		}
		if msg.Type != MessageTypeText {
			t.Errorf("wrong message type: got %q, want %q", msg.Type, MessageTypeText)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for message")
	}

	replyMessage := "Hello there"
	if err := app2.SendMessage(replyMessage); err != nil {
		t.Fatalf("failed to send reply: %v", err)
	}

	select {
	case msg := <-handler1.messages:
		if msg.Content != replyMessage {
			t.Errorf("received wrong reply: got %q, want %q", msg.Content, replyMessage)
		}
		if msg.Nickname != "testUser2" {
			t.Errorf("wrong sender: got %q, want %q", msg.Nickname, "testUser2")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for reply")
	}
}
