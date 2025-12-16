package bots

import (
	"context"
	"fmt"
	"log"

	"github.com/matt0792/lanchat/sdk"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type OpenaiBot struct {
	client       *openai.Client
	systemPrompt string
}

func NewOpenaiBot(key, systemPrompt string) *OpenaiBot {
	client := openai.NewClient(
		option.WithAPIKey(key),
	)
	return &OpenaiBot{
		client:       &client,
		systemPrompt: systemPrompt,
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
		resp, err := b.invoke(msg.Content)
		if err != nil {
			lc.SendMessage(fmt.Sprintf("[Error] %v", err))
			return err
		}
		lc.SendMessage(resp)
	}
	return nil
}

func (b *OpenaiBot) OnRoomJoined(room sdk.Room, lc *sdk.Lanchat) error {
	return nil
}

func (b *OpenaiBot) invoke(prompt string) (string, error) {
	log.Printf("Received: %s\n", prompt)

	resp, err := b.client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(b.systemPrompt),
			openai.UserMessage(prompt),
		},
		Model: openai.ChatModelGPT5Mini,
	})
	if err != nil {
		log.Println(err.Error())
		return "", err
	}

	content := resp.Choices[0].Message.Content
	log.Printf("Response: %s\n", content)
	return content, nil
}
