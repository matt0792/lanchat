package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/matt0792/lanchat/internal/ui"
)

type CLI struct {
	ctx        context.Context
	cancel     context.CancelFunc
	reader     *bufio.Reader
	cmdHandler ui.CommandHandler
}

func New(ctx context.Context) *CLI {
	cliCtx, cancel := context.WithCancel(ctx)
	return &CLI{
		ctx:    cliCtx,
		cancel: cancel,
		reader: bufio.NewReader(os.Stdin),
	}
}

func (c *CLI) ShowMessage(nickname, message string) {
	fmt.Printf("\r\033[K%s: %s\n", nickname, message)
	c.ShowPrompt()
}

func (c *CLI) ShowSystemMessage(message string) {
	fmt.Printf("\r\033[K%s\n", message)
	c.ShowPrompt()
}

func (c *CLI) ShowPeerJoined(nickname string) {
	fmt.Printf("\r\033[K%s connected\n", nickname)
	c.ShowPrompt()
}

func (c *CLI) ShowPeerList(peers []string) {
	fmt.Printf("\nConnected peers (%d):\n", len(peers))
	for _, p := range peers {
		fmt.Printf("  - %s \n", p)
	}
	fmt.Println()
}

func (c *CLI) ShowRoomList(rooms []string) {
	if len(rooms) == 0 {
		fmt.Println("No active rooms found")
		return
	}
	fmt.Println("Available rooms:")
	for _, room := range rooms {
		fmt.Printf("  - %s\n", room)
	}
}

func (c *CLI) ShowError(err error) {
	fmt.Printf("Error: %v\n", err)
}

func (c *CLI) ShowPrompt() {
	fmt.Print("> ")
}

func (c *CLI) OnCommand(handler ui.CommandHandler) {
	c.cmdHandler = handler
}

func (c *CLI) Start() error {
	fmt.Println("\nCommands:")
	fmt.Println("  /join <room>  - Join a room")
	fmt.Println("  /leave        - Leave current room")
	fmt.Println("  /peers        - List connected peers")
	fmt.Println("  /rooms        - List available rooms")
	fmt.Println("  /quit         - Exit")
	fmt.Println()

	for {
		select {
		case <-c.ctx.Done():
			return nil
		default:
			c.ShowPrompt()
			input, err := c.reader.ReadString('\n')
			if err != nil {
				return err
			}

			input = strings.TrimSpace(input)
			if input == "" {
				continue
			}

			cmd := c.parseInput(input)
			if c.cmdHandler != nil {
				if err := c.cmdHandler(cmd); err != nil {
					if err.Error() == "quit" {
						return nil
					}
					c.ShowError(err)
				}
			}
		}
	}
}

func (c *CLI) Stop() {
	c.cancel()
}

func (c *CLI) parseInput(input string) ui.Command {
	if strings.HasPrefix(input, "/") {
		parts := strings.Fields(input)
		return ui.Command{
			Type: strings.TrimPrefix(parts[0], "/"),
			Args: parts[1:],
		}
	}

	// Regular message
	return ui.Command{
		Type: "send",
		Args: []string{input},
	}
}
