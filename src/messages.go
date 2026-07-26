package main

import (
	"embed"
	"encoding/json"
	"math/rand"
)

//go:embed messages.json
var messagesFS embed.FS

type scoreMessages struct {
	Messages []tierMessages `json:"messages"`
}

type tierMessages struct {
	MinScore int      `json:"min_score"`
	Messages []string `json:"messages"`
}

var (
	loadedMessages []tierMessages
)

func initMessages() {
	data, err := messagesFS.ReadFile("messages.json")
	if err != nil {
		loadedMessages = nil
		return
	}
	var sm scoreMessages
	if err := json.Unmarshal(data, &sm); err != nil {
		loadedMessages = nil
		return
	}
	loadedMessages = sm.Messages
}

func randomScoreMessage(score int) string {
	if loadedMessages == nil {
		initMessages()
		if loadedMessages == nil {
			return ""
		}
	}
	for _, t := range loadedMessages {
		if score >= t.MinScore {
			return t.Messages[rand.Intn(len(t.Messages))]
		}
	}
	return ""
}
