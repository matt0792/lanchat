package bots

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/matt0792/lanchat/sdk"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
)

type Model string

const (
	ModelGPT5Mini Model = "gpt-5-mini"
)

type PastMessage struct {
	Name    string
	Content string
}

type OpenaiBot struct {
	client       *openai.Client
	systemPrompt string
	model        Model

	mu          sync.Mutex
	history     []PastMessage
	historySize int
}

func NewOpenaiBot(key, systemPrompt, model string, historySize int) *OpenaiBot {
	client := openai.NewClient(
		option.WithAPIKey(key),
	)

	if model == "" {
		model = string(ModelGPT5Mini)
	}

	return &OpenaiBot{
		client:       &client,
		systemPrompt: systemPrompt,
		model:        Model(model),
		history:      make([]PastMessage, 0, historySize),
		historySize:  historySize,
	}
}

func (b *OpenaiBot) Initialize(lc *sdk.Lanchat) error {
	return lc.JoinRoom("general", "")
}

func (b *OpenaiBot) OnPeerJoined(peer sdk.PeerInfo, lc *sdk.Lanchat) error {
	return nil
}

func (b *OpenaiBot) OnMessage(msg sdk.ChatMessage, lc *sdk.Lanchat) error {
	switch msg.Type {
	case sdk.MessageTypeJoin:
	case sdk.MessageTypeLeave:
	case sdk.MessageTypeText:
		b.addToHistory(msg.Nickname, msg.Content)

		resp, err := b.invoke(msg.Nickname, msg.Content)
		if err != nil {
			lc.SendMessage(fmt.Sprintf("[Error] %v", err))
			return err
		}
		if strings.ToLower(resp) == "skip" {
			return nil
		}

		b.addToHistory("You", resp)
		lc.SendMessage(resp)
	}
	return nil
}

func (b *OpenaiBot) OnRoomJoined(room sdk.Room, lc *sdk.Lanchat) error {
	return nil
}

func (b *OpenaiBot) invoke(nickname, prompt string) (string, error) {
	log.Printf("Received: %s: %s\n", nickname, prompt)

	resp, err := b.client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: b.buildMessages(nickname, prompt),
		Model:    shared.ChatModel(b.model),
	})
	if err != nil {
		log.Println(err.Error())
		return "", err
	}

	content := resp.Choices[0].Message.Content
	log.Printf("Response: %s\n", content)
	return content, nil
}

func (b *OpenaiBot) addToHistory(name, content string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.history = append(b.history, PastMessage{
		Name:    name,
		Content: content,
	})

	if len(b.history) > b.historySize {
		b.history = b.history[len(b.history)-b.historySize:]
	}
}

func (b *OpenaiBot) buildMessages(currentUser, currentPrompt string) []openai.ChatCompletionMessageParamUnion {
	b.mu.Lock()
	defer b.mu.Unlock()

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(b.systemPrompt),
	}

	for _, msg := range b.history[:len(b.history)-1] {
		formatted := fmt.Sprintf("%s: %s", msg.Name, msg.Content)
		messages = append(messages, openai.UserMessage(formatted))
	}

	messages = append(messages, openai.UserMessage(fmt.Sprintf("%s: %s", currentUser, currentPrompt)))

	return messages
}
