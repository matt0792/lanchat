package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/matt0792/lanchat/internal/app"
)

type Controller struct {
	app *app.App
	ui  UI
	ctx context.Context
}

func NewController(ctx context.Context, app *app.App, ui UI) *Controller {
	c := &Controller{
		app: app,
		ui:  ui,
		ctx: ctx,
	}

	ui.OnCommand(c.handleCommand)

	go c.handleAppEvents()

	return c
}

func (c *Controller) handleCommand(cmd Command) error {
	switch cmd.Type {
	case "join":
		if len(cmd.Args) < 1 {
			return fmt.Errorf("usage: /join <room>")
		}
		roomName := cmd.Args[0]
		password := ""
		if len(cmd.Args) > 1 {
			password = strings.Join(cmd.Args[1:], " ")
		}
		if err := c.app.JoinRoom(roomName, password); err != nil {
			return err
		}
		c.ui.ShowSystemMessage(fmt.Sprintf("Joined room: %s", cmd.Args[0]))

	case "leave":
		if err := c.app.LeaveRoom(); err != nil {
			return err
		}
		c.ui.ShowSystemMessage("Left room")

	case "peers":
		peers := c.app.GetPeerList()
		c.ui.ShowPeerList(peers)

	case "rooms":
		rooms := c.app.GetRoomList()
		c.ui.ShowRoomList(rooms)

	case "help":
		helpText := `
Available Commands:
  /join <room> [password]  	- Join a chat room
  /leave        			- Leave the current room
  /peers        			- List all connected peers
  /rooms        			- List all available rooms
  /help         			- Show this help message
  /quit         			- Exit the application`
		c.ui.ShowSystemMessage(helpText)

	case "send":
		room := c.app.GetCurrentRoom()
		if room == nil {
			return fmt.Errorf("not in a room (use /join <room>)")
		}
		if len(cmd.Args) > 0 {
			return c.app.SendMessage(cmd.Args[0])
		}

	case "quit", "exit":
		return fmt.Errorf("quit")

	default:
		return fmt.Errorf("unknown command: /%s", cmd.Type)
	}

	return nil
}

func (c *Controller) handleAppEvents() {
	for event := range c.app.GetEvents() {
		switch event.Type {
		case app.EventMessageRecv:
			msg := event.Data.(*app.ChatMessage)
			switch msg.Type {
			case app.MessageTypeText:
				c.ui.ShowMessage(msg.Nickname, msg.Content)
			}

		case app.EventPeerJoined:
			peer := event.Data.(*app.PeerInfo)
			c.ui.ShowPeerJoined(peer.Nickname)

		case app.EventPeerLeft:
			peer := event.Data.(*app.PeerInfo)
			c.ui.ShowPeerLeft(peer.Nickname)

		case app.EventRoomJoined:
			room := event.Data.(*app.Room)
			c.ui.ShowSystemMessage(fmt.Sprintf("You joined: %s", room.Name))
		}
	}
}

func (c *Controller) Start() error {
	return c.ui.Start()
}

func (c *Controller) Stop() {
	c.ui.Stop()
}
