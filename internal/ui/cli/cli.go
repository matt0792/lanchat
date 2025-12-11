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

// clearLine clears the current terminal line and returns cursor to start
func clearLine() {
	fmt.Print("\r\033[K")
}

func (c *CLI) ShowMessage(nickname, message string) {
	clearLine()
	fmt.Printf("<%s> %s\n", nickname, message)
	c.ShowPrompt()
}

func (c *CLI) ShowSystemMessage(message string) {
	clearLine()
	fmt.Printf("* %s\n", message)
	c.ShowPrompt()
}

func (c *CLI) ShowPeerJoined(nickname string) {
	clearLine()
	fmt.Printf("→ %s joined\n", nickname)
	c.ShowPrompt()
}

func (c *CLI) ShowPeerLeft(nickname string) {
	clearLine()
	fmt.Printf("← %s left\n", nickname)
	c.ShowPrompt()
}

func (c *CLI) ShowPeerList(peers []string) {
	clearLine()
	if len(peers) == 0 {
		fmt.Println("No peers connected")
	} else {
		fmt.Printf("Connected peers (%d):\n", len(peers))
		for _, p := range peers {
			fmt.Printf("  %s\n", p)
		}
	}
	fmt.Println()
	c.ShowPrompt()
}

func (c *CLI) ShowRoomList(rooms []string) {
	clearLine()
	if len(rooms) == 0 {
		fmt.Println("No active rooms")
	} else {
		fmt.Printf("Available rooms (%d):\n", len(rooms))
		for _, room := range rooms {
			fmt.Printf("  %s\n", room)
		}
	}
	fmt.Println()
	c.ShowPrompt()
}

func (c *CLI) ShowError(err error) {
	clearLine()
	fmt.Printf("! %v\n", err)
	c.ShowPrompt()
}

func (c *CLI) ShowPrompt() {
	fmt.Print("> ")
}

func (c *CLI) OnCommand(handler ui.CommandHandler) {
	c.cmdHandler = handler
}

func (c *CLI) Start() error {
	c.showWelcome()

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

func (c *CLI) showWelcome() {
	fmt.Println("lanchat v1.0")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  /join <room>    Join a room")
	fmt.Println("  /leave          Leave current room")
	fmt.Println("  /peers          List connected peers")
	fmt.Println("  /rooms          List available rooms")
	fmt.Println("  /quit           Exit")
	fmt.Println()
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

	return ui.Command{
		Type: "send",
		Args: []string{input},
	}
}
