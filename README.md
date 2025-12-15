
# lanchat

A tiny peer-to-peer CLI chat app that works over a local network. Uses libp2p with mDNS for local peer discovery and GossipSub for pub/sub messaging. Supports stream handlers for direct communication (metadata). 

Rooms can be password-protected with AES-256-GCM encryption, and are logically separated & hidden if a password is set. 


## Installation

**Build from source:**
```bash
git clone https://github.com/matt0792/lanchat.git

go install cmd/lanchat/main.go
```

## Usage

**Start the app:**
```bash
lanchat 
```

**Basic commands:**
```
/join <room> [password]  - Join a room
/leave                   - Leave current room
/peers                   - List connected peers
/rooms                   - List available rooms
/help                    - Show help
/quit                    - Exit
```

## Bot Support

Bots can join rooms and respond to messages/events programmatically using the SDK.

**Bot interface:**
```go
type Bot interface {
    Initialize(lc *Lanchat) error
    OnMessage(msg ChatMessage, lc *Lanchat) error
    OnPeerJoined(peer PeerInfo, lc *Lanchat) error
    OnRoomJoined(room Room, lc *Lanchat) error
}
```

**Usage**
- Implement the `Bot` interface
- Use `BotRunner` to connect to the network
- Bots receive events (messages, peer joins, room joins)
- Bots can send messages and interact with the Lanchat instance

See `bots/templatebot.go` for a starting point.

## Safety

This is a toy for trusted networks, not a secure messenger.

**Protections:**
- Message encryption (in password protected rooms)
- Rate limiting 
- Input & output sanitation 
- Transport-level encryption (libp2p)

**Limitations:**
- No authentication
- No forward secrecy
- No protection against malicious peers on your LAN
