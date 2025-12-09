package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/matt0792/lanchat/internal/logger"
)

type Topic struct {
	topic *pubsub.Topic
	sub   *pubsub.Subscription
	host  *Host
}

func (h *Host) JoinTopic(topicName string) (*Topic, error) {
	topic, err := h.pubsub.Join(topicName)
	if err != nil {
		return nil, fmt.Errorf("failed to join topic: %w", err)
	}

	sub, err := topic.Subscribe()
	if err != nil {
		topic.Close()
		return nil, fmt.Errorf("failed to subscribe to topic: %w", err)
	}

	t := &Topic{
		topic: topic,
		sub:   sub,
		host:  h,
	}

	return t, nil
}

func (t *Topic) Publish(msgType MessageType, data interface{}) error {
	now := time.Now()

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	msg := Message{
		Type:      msgType,
		From:      t.host.ID().String(),
		Timestamp: now,
		Data:      dataBytes,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal msg: %w", err)
	}

	if err := t.topic.Publish(t.host.ctx, msgBytes); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (t *Topic) ReadMessages(ctx context.Context) <-chan *Message {
	msgChan := make(chan *Message, 10)

	go func() {
		defer close(msgChan)
		for {
			msg, err := t.sub.Next(ctx)
			if err != nil {
				if err != context.Canceled {
					logger.Error("Error reading message from topic: %v", err)
				}
				return
			}

			if msg.ReceivedFrom == t.host.ID() {
				continue
			}

			var parsedMsg Message
			if err := json.Unmarshal(msg.Data, &parsedMsg); err != nil {
				logger.Warn("Failed to unmarshal message: %v", err)
				continue
			}

			select {
			case msgChan <- &parsedMsg:
			case <-ctx.Done():
				return
			}

			// call handlers
			t.host.msgHandlersMu.RLock()
			if handler, exists := t.host.msgHandlers[parsedMsg.Type]; exists {
				go func(m Message) {
					if err := handler(&m); err != nil {
						logger.Error("Error in message handler: %v", err)
					}
				}(parsedMsg)
			}
			t.host.msgHandlersMu.RUnlock()
		}
	}()

	return msgChan
}

func (t *Topic) Close() error {
	t.sub.Cancel()
	return t.topic.Close()
}
