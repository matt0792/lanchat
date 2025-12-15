package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/matt0792/lanchat/internal/ui"
)

const (
	colorReset = "\033[0m"
	colorGray  = "\033[90m"
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

func clearLine() {
	fmt.Print("\r\033[K")
}

func (c *CLI) ShowMessage(nickname, identity, message string) {
	clearLine()
	fmt.Printf("\n%s %s%s\t%s%s\n", nickname, colorGray, identity, time.Now().Format("15:04"), colorReset)
	fmt.Printf("%s\n", message)
	c.ShowPrompt()
}

func (c *CLI) ShowSystemMessage(message string) {
	clearLine()
	fmt.Printf("\n%s%s%s\n", colorGray, message, colorReset)
	c.ShowPrompt()
}

func (c *CLI) ShowPeerJoined(nickname, identity string) {
	clearLine()
	fmt.Printf("\n%s%s %s joined%s\n", colorGray, nickname, identity, colorReset)
	c.ShowPrompt()
}

func (c *CLI) ShowPeerLeft(nickname, identity string) {
	clearLine()
	fmt.Printf("\n%s%s %s left%s\n", colorGray, nickname, identity, colorReset)
	c.ShowPrompt()
}

func (c *CLI) ShowPeerList(peers []string) {
	clearLine()
	if len(peers) == 0 {
		fmt.Println("\nNo peers connected")
	} else {
		fmt.Printf("\nConnected peers (%d):\n", len(peers))
		for _, p := range peers {
			fmt.Printf("  %s\n", p)
		}
	}
	c.ShowPrompt()
}

func (c *CLI) ShowRoomList(rooms []string) {
	clearLine()
	if len(rooms) == 0 {
		fmt.Println("\nNo active rooms")
	} else {
		fmt.Printf("\nAvailable rooms (%d):\n", len(rooms))
		for _, room := range rooms {
			fmt.Printf("  %s\n", room)
		}
	}
	c.ShowPrompt()
}

func (c *CLI) ShowError(err error) {
	clearLine()
	fmt.Printf("%v\n", err)
	c.ShowPrompt()
}

func (c *CLI) ShowPrompt() {
	clearLine()
	fmt.Print("> ")
}

func (c *CLI) OnCommand(handler ui.CommandHandler) {
	c.cmdHandler = handler
}

func (c *CLI) Start() error {
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

	return ui.Command{
		Type: "send",
		Args: []string{input},
	}
}
