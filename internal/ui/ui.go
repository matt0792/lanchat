package ui

type UI interface {
	ShowMessage(nickname, identity, message string)
	ShowSystemMessage(message string)
	ShowPeerJoined(nickname, identity string)
	ShowPeerLeft(nickname, identity string)
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
