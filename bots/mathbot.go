package bots

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/matt0792/lanchat/sdk"
)

type MathBot struct{}

func (b *MathBot) Initialize(lc *sdk.Lanchat) error {
	return lc.JoinRoom("general", "")
}

func (b *MathBot) OnPeerJoined(peer sdk.PeerInfo, lc *sdk.Lanchat) error {
	return nil
}

func (b *MathBot) OnMessage(msg sdk.ChatMessage, lc *sdk.Lanchat) error {
	if msg.Type != sdk.MessageTypeText {
		return nil
	}

	parts := strings.Fields(msg.Content)
	if len(parts) < 3 {
		return nil
	}
	if parts[0] != "mathbot" {
		return nil
	}

	operator := parts[1]

	var nums []int
	for _, part := range parts[2:] {
		num, err := strconv.Atoi(part)
		if err != nil {
			return lc.SendMessage(fmt.Sprintf("invalid number: %s", part))
		}
		nums = append(nums, num)
	}

	if len(nums) == 0 {
		return lc.SendMessage("no numbers provided")
	}

	var result int
	var err error

	switch operator {
	case "add", "+":
		result = b.add(nums)
	case "subtract", "-":
		result = b.subtract(nums)
	case "multiply", "*", "x":
		result = b.multiply(nums)
	case "divide", "/":
		result, err = b.divide(nums)
		if err != nil {
			return lc.SendMessage(err.Error())
		}
	default:
		return nil
	}

	return lc.SendMessage(fmt.Sprintf("%d", result))
}

func (b *MathBot) OnRoomJoined(room sdk.Room, lc *sdk.Lanchat) error {
	return lc.SendMessage("MathBot online. Use: mathbot add/subtract/multiply/divide <numbers>")
}

func (b *MathBot) add(args []int) int {
	total := 0
	for _, v := range args {
		total += v
	}
	return total
}

func (b *MathBot) subtract(args []int) int {
	if len(args) == 0 {
		return 0
	}
	result := args[0]
	for _, v := range args[1:] {
		result -= v
	}
	return result
}

func (b *MathBot) multiply(args []int) int {
	if len(args) == 0 {
		return 0
	}
	result := args[0]
	for _, v := range args[1:] {
		result *= v
	}
	return result
}

func (b *MathBot) divide(args []int) (int, error) {
	if len(args) == 0 {
		return 0, fmt.Errorf("no numbers to divide")
	}
	result := args[0]
	for _, v := range args[1:] {
		if v == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		result /= v
	}
	return result, nil
}
