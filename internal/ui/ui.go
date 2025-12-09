package ui

type UI interface {
	ShowMessage(nickname, message string)
	ShowSystemMessage(message string)
	ShowPeerJoined(nickname string)
	ShowPeerList(peers []string)
	ShowRoomList(rooms []string)
	ShowError(err error)
	ShowPrompt()

	Start() error
	Stop()
	OnCommand(handler CommandHandler)
}

type CommandHandler func(cmd Command) error

type Command struct {
	Type string
	Args []string
}
