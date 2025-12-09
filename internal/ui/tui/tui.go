package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/matt0792/lanchat/internal/app"
	"github.com/matt0792/lanchat/internal/ui"
)

type messageReceivedMsg struct {
	nickname string
	message  string
	isSystem bool
	msgType  app.MessageType
}

type peerJoinedMsg string

type statusUpdateMsg struct {
	room  string
	peers int
}

type TUI struct {
	ctx        context.Context
	cancel     context.CancelFunc
	program    *tea.Program
	cmdHandler ui.CommandHandler

	msgChan    chan messageReceivedMsg
	peerChan   chan peerJoinedMsg
	statusChan chan statusUpdateMsg
}

type model struct {
	tui         *TUI
	viewport    viewport.Model
	textInput   textinput.Model
	messages    []string
	ready       bool
	width       int
	height      int
	currentRoom string
	peerCount   int
	err         error
}

var (
	accentColor = lipgloss.Color("#ff1f1fff")
	whiteColor  = lipgloss.Color("#FFFFFF")

	appStyle = lipgloss.NewStyle().Padding(1, 2)

	helpStyle = lipgloss.NewStyle().
			Foreground(whiteColor).
			Padding(0, 1)

	messageStyle = lipgloss.NewStyle().
			Foreground(whiteColor)

	nicknameStyle = lipgloss.NewStyle().
			Foreground(whiteColor).
			Bold(true)

	systemMessageStyle = lipgloss.NewStyle().
				Foreground(whiteColor).
				Italic(true)

	joinMessageStyle = lipgloss.NewStyle().
				Foreground(whiteColor).
				Italic(true)

	leaveMessageStyle = lipgloss.NewStyle().
				Foreground(whiteColor).
				Italic(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(whiteColor).
			Bold(true)

	inputStyle = lipgloss.NewStyle().
			Foreground(whiteColor).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(accentColor)
)

func New(ctx context.Context) *TUI {
	tuiCtx, cancel := context.WithCancel(ctx)
	return &TUI{
		ctx:        tuiCtx,
		cancel:     cancel,
		msgChan:    make(chan messageReceivedMsg, 100),
		peerChan:   make(chan peerJoinedMsg, 100),
		statusChan: make(chan statusUpdateMsg, 10),
	}
}

func (t *TUI) Start() error {
	m := model{
		tui:      t,
		messages: []string{},
	}

	m.textInput = textinput.New()
	m.textInput.Placeholder = "Type a message or /help for commands..."
	m.textInput.Focus()
	m.textInput.CharLimit = 500
	m.textInput.Width = 80

	t.program = tea.NewProgram(m, tea.WithAltScreen())

	go t.listenForUpdates()

	_, err := t.program.Run()
	return err
}

func (t *TUI) Stop() {
	if t.program != nil {
		t.program.Quit()
	}
	t.cancel()
}

func (t *TUI) OnCommand(handler ui.CommandHandler) {
	t.cmdHandler = handler
}

func (t *TUI) ShowMessage(nickname, message string) {
	t.msgChan <- messageReceivedMsg{
		nickname: nickname,
		message:  message,
		isSystem: false,
		msgType:  app.MessageTypeText,
	}
}

func (t *TUI) ShowSystemMessage(message string) {
	t.msgChan <- messageReceivedMsg{
		message:  message,
		isSystem: true,
		msgType:  app.MessageTypeText,
	}
}

func (t *TUI) ShowPeerJoined(nickname string) {
	t.peerChan <- peerJoinedMsg(nickname)
}

func (t *TUI) ShowPeerList(peers []string) {
	var sb strings.Builder
	if len(peers) == 0 {
		sb.WriteString("No peers connected\n")
	} else {
		sb.WriteString("Connected Peers:\n")
		for _, p := range peers {
			sb.WriteString(fmt.Sprintf("  • %s\n", p))
		}
	}

	t.msgChan <- messageReceivedMsg{
		message:  sb.String(),
		isSystem: true,
	}
}

func (t *TUI) ShowRoomList(rooms []string) {
	var sb strings.Builder
	if len(rooms) == 0 {
		sb.WriteString("No active rooms found\n")
	} else {
		sb.WriteString("Available Rooms:\n")
		for _, room := range rooms {
			sb.WriteString(fmt.Sprintf("  • %s\n", room))
		}
	}

	t.msgChan <- messageReceivedMsg{
		message:  sb.String(),
		isSystem: true,
	}
}

func (t *TUI) ShowError(err error) {
	t.msgChan <- messageReceivedMsg{
		message:  "Error: " + err.Error(),
		isSystem: true,
	}
}

func (t *TUI) ShowPrompt() {
}

func (t *TUI) UpdateStatus(room string, peerCount int) {
	t.statusChan <- statusUpdateMsg{
		room:  room,
		peers: peerCount,
	}
}

func (t *TUI) listenForUpdates() {
	for {
		select {
		case <-t.ctx.Done():
			return
		case msg := <-t.msgChan:
			if t.program != nil {
				t.program.Send(msg)
			}
		case peer := <-t.peerChan:
			if t.program != nil {
				t.program.Send(peer)
			}
		case status := <-t.statusChan:
			if t.program != nil {
				t.program.Send(status)
			}
		}
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width-4, msg.Height-8)
			m.viewport.YPosition = 2
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = msg.Height - 8
		}
		m.width = msg.Width
		m.height = msg.Height
		m.textInput.Width = msg.Width - 6
		m.updateViewport()

	case messageReceivedMsg:
		timestamp := time.Now().Format("15:04:05")
		var formatted string

		if msg.isSystem {
			formatted = systemMessageStyle.Render(fmt.Sprintf("[%s] %s", timestamp, msg.message))
		} else {
			switch msg.msgType {
			case app.MessageTypeJoin:
				formatted = joinMessageStyle.Render(fmt.Sprintf("[%s] %s", timestamp, msg.message))
			case app.MessageTypeLeave:
				formatted = leaveMessageStyle.Render(fmt.Sprintf("[%s] %s", timestamp, msg.message))
			default:
				nickname := nicknameStyle.Render(msg.nickname)
				formatted = fmt.Sprintf("[%s] %s: %s", timestamp, nickname, messageStyle.Render(msg.message))
			}
		}

		m.messages = append(m.messages, formatted)
		m.updateViewport()
		m.viewport.GotoBottom()

	case peerJoinedMsg:
		timestamp := time.Now().Format("15:04:05")
		formatted := joinMessageStyle.Render(fmt.Sprintf("[%s] ○ %s connected", timestamp, string(msg)))
		m.messages = append(m.messages, formatted)
		m.updateViewport()
		m.viewport.GotoBottom()

	case statusUpdateMsg:
		m.currentRoom = msg.room
		m.peerCount = msg.peers

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.tui.Stop()
			return m, tea.Quit

		case tea.KeyEnter:
			input := strings.TrimSpace(m.textInput.Value())
			if input != "" {
				cmd := m.parseInput(input)
				if m.tui.cmdHandler != nil {
					if err := m.tui.cmdHandler(cmd); err != nil {
						if err.Error() == "quit" {
							return m, tea.Quit
						}
						m.err = err
						m.messages = append(m.messages, errorStyle.Render("Error: "+err.Error()))
						m.updateViewport()
					}
				}
				m.textInput.SetValue("")
			}
			return m, nil
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	help := helpStyle.Render("Enter: Send | Esc: Quit | /help: Commands")

	inputBox := inputStyle.Width(m.width - 6).Render(m.textInput.View())

	return appStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			"",
			m.viewport.View(),
			"",
			inputBox,
			help,
		),
	)
}

func (m *model) updateViewport() {
	content := strings.Join(m.messages, "\n")
	m.viewport.SetContent(content)
}

func (m model) parseInput(input string) ui.Command {
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
