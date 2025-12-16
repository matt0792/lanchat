package bots

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/matt0792/lanchat/sdk"
)

type MarkovBot struct {
	mu         sync.RWMutex
	chain      map[string]map[string]int
	starters   map[string]int
	messageLog []string
	active     bool
}

func (b *MarkovBot) Initialize(lc *sdk.Lanchat) error {
	b.chain = make(map[string]map[string]int)
	b.starters = make(map[string]int)
	b.messageLog = make([]string, 0)
	b.active = true

	return lc.JoinRoom("general", "")
}

func (b *MarkovBot) OnPeerJoined(peer sdk.PeerInfo, lc *sdk.Lanchat) error {
	return nil
}

func (b *MarkovBot) OnMessage(msg sdk.ChatMessage, lc *sdk.Lanchat) error {
	if msg.Type != sdk.MessageTypeText {
		return nil
	}

	text := strings.TrimSpace(msg.Content)
	parts := strings.Fields(strings.ToLower(text))

	if len(parts) == 0 {
		return nil
	}

	if parts[0] == "markov" {
		if len(parts) < 2 {
			return b.showHelp(lc)
		}

		switch parts[1] {
		case "story":
			length := 30
			if len(parts) > 2 {
				fmt.Sscanf(parts[2], "%d", &length)
			}
			return b.generateStory(lc, length)
		case "stats":
			return b.showStats(lc)
		case "reset":
			return b.resetChain(lc)
		case "help":
			return b.showHelp(lc)
		}
		return nil
	}

	// Learn from all non-command messages
	b.learn(text)

	// Randomly respond
	if rand.Float64() < 0.10 && b.active {
		go func() {
			time.Sleep(time.Duration(1+rand.Intn(3)) * time.Second)
			response := b.generateContextual(text, 10, 20)
			if response != "" && !strings.Contains(strings.ToLower(response), "markov") {
				lc.SendMessage(fmt.Sprintf("💭 %s", response))
			}
		}()
	}

	return nil
}

func (b *MarkovBot) OnRoomJoined(room sdk.Room, lc *sdk.Lanchat) error {
	return nil
}

func (b *MarkovBot) showHelp(lc *sdk.Lanchat) error {
	help := `Commands:
- markov generate [length] 
- markov personality <mode> 
- markov stats 
- markov reset `
	return lc.SendMessage(help)
}

func (b *MarkovBot) learn(text string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Store message
	b.messageLog = append(b.messageLog, text)
	if len(b.messageLog) > 1000 {
		b.messageLog = b.messageLog[1:]
	}

	// Tokenize and build chain
	words := b.tokenize(text)
	if len(words) < 2 {
		return
	}

	// Record starter
	b.starters[words[0]]++

	// Build bigram chain
	for i := 0; i < len(words)-1; i++ {
		word := words[i]
		next := words[i+1]

		if b.chain[word] == nil {
			b.chain[word] = make(map[string]int)
		}
		b.chain[word][next]++
	}
}

func (b *MarkovBot) tokenize(text string) []string {
	text = strings.ToLower(text)
	replacer := strings.NewReplacer(
		".", " .",
		"!", " !",
		"?", " ?",
		",", " ,",
	)
	text = replacer.Replace(text)

	words := strings.Fields(text)
	filtered := make([]string, 0)

	for _, word := range words {
		word = strings.TrimSpace(word)
		if word != "" && !strings.HasPrefix(word, "http") {
			filtered = append(filtered, word)
		}
	}

	return filtered
}

func (b *MarkovBot) generate(minWords, maxWords int) string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if len(b.starters) == 0 {
		return ""
	}

	// Weighted starter
	current := b.weightedChoice(b.starters)

	result := []string{current}
	targetLength := minWords + rand.Intn(maxWords-minWords+1)

	for i := 0; i < targetLength-1; i++ {
		if b.chain[current] == nil || len(b.chain[current]) == 0 {
			break
		}

		current = b.weightedChoice(b.chain[current])

		result = append(result, current)
	}

	return b.formatSentence(result)
}

func (b *MarkovBot) generateContextual(context string, minWords, maxWords int) string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	contextWords := b.tokenize(context)
	if len(contextWords) == 0 {
		return b.generate(minWords, maxWords)
	}

	var starter string
	for _, word := range contextWords {
		if len(b.chain[word]) > 0 {
			starter = word
			break
		}
	}

	if starter == "" {
		return b.generate(minWords, maxWords)
	}

	result := []string{starter}
	current := starter
	targetLength := minWords + rand.Intn(maxWords-minWords+1)

	for i := 0; i < targetLength-1; i++ {
		if b.chain[current] == nil || len(b.chain[current]) == 0 {
			break
		}
		current = b.weightedChoice(b.chain[current])
		result = append(result, current)
	}

	return b.formatSentence(result)
}

func (b *MarkovBot) weightedChoice(choices map[string]int) string {
	total := 0
	for _, count := range choices {
		total += count
	}

	if total == 0 {
		// Fallback to random choice
		for k := range choices {
			return k
		}
	}

	r := rand.Intn(total)
	cumulative := 0

	for word, count := range choices {
		cumulative += count
		if r < cumulative {
			return word
		}
	}

	// Fallback
	for k := range choices {
		return k
	}
	return ""
}

func (b *MarkovBot) formatSentence(words []string) string {
	if len(words) == 0 {
		return ""
	}

	sentence := strings.Join(words, " ")

	// Clean up spacing around punctuation
	sentence = strings.ReplaceAll(sentence, " .", ".")
	sentence = strings.ReplaceAll(sentence, " !", "!")
	sentence = strings.ReplaceAll(sentence, " ?", "?")
	sentence = strings.ReplaceAll(sentence, " ,", ",")

	if len(sentence) > 0 {
		sentence = strings.ToUpper(string(sentence[0])) + sentence[1:]
	}

	return sentence
}

func (b *MarkovBot) generateStory(lc *sdk.Lanchat, length int) error {
	if length > 200 {
		length = 200
	}
	if length < 10 {
		length = 10
	}

	story := b.generate(length-10, length+10)

	if story == "" {
		return lc.SendMessage("Need to learn more words")
	}

	return lc.SendMessage(story)
}

func (b *MarkovBot) showStats(lc *sdk.Lanchat) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	uniqueWords := len(b.chain)
	totalConnections := 0
	for _, next := range b.chain {
		for _, count := range next {
			totalConnections += count
		}
	}

	stats := []string{
		fmt.Sprintf("- Unique words learned: %d", uniqueWords),
		fmt.Sprintf("- Total word connections: %d", totalConnections),
		fmt.Sprintf("- Total word connections: %d", len(b.messageLog)),
		fmt.Sprintf("- Starter phrases: %d", len(b.starters)),
	}

	for _, v := range stats {
		lc.SendMessage(v)
	}
	return nil
}

func (b *MarkovBot) resetChain(lc *sdk.Lanchat) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.chain = make(map[string]map[string]int)
	b.starters = make(map[string]int)
	b.messageLog = make([]string, 0)

	return lc.SendMessage("Chain reset successfully")
}
