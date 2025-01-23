package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"syscall/js"
	"time"
)

type Config struct {
	BaseURL string
	Token   string
	Minutes int
}

type Message struct {
	Type          string `json:"type"`
	IDMessage     string `json:"idMessage"`
	Timestamp     int64  `json:"timestamp"`
	TypeMessage   string `json:"typeMessage"`
	ChatID        string `json:"chatId"`
	DownloadURL   string `json:"downloadUrl,omitempty"`
	Caption       string `json:"caption,omitempty"`
	TextMessage   string `json:"textMessage,omitempty"`
	FileName      string `json:"fileName,omitempty"`
	MimeType      string `json:"mimeType,omitempty"`
	IsAnimated    bool   `json:"isAnimated,omitempty"`
	IsForwarded   bool   `json:"isForwarded,omitempty"`
	StatusMessage string `json:"statusMessage,omitempty"`
}

type Client struct {
	config            Config
	lastProcessedTime int64
}

func NewClient(config Config) *Client {
	return &Client{
		config:            config,
		lastProcessedTime: time.Now().Add(-time.Duration(config.Minutes) * time.Minute).Unix(),
	}
}

func (c *Client) handleRequest(this js.Value, args []js.Value) interface{} {
	payload := map[string]int{"minutes": c.config.Minutes}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling payload: %v", err)
		return nil
	}

	req, err := http.NewRequest("POST", c.config.BaseURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return nil
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making request: %v", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		return nil
	}

	var messages []Message
	if err := json.Unmarshal(body, &messages); err != nil {
		log.Printf("Error unmarshaling response: %v", err)
		return nil
	}

	filteredMessages := c.filterMessages(messages)
	c.processMessages(filteredMessages)

	result, _ := json.Marshal(filteredMessages)
	return string(result)
}

func (c *Client) filterMessages(messages []Message) []Message {
	var filtered []Message
	cutoffTime := time.Now().Add(-time.Duration(c.config.Minutes) * time.Minute).Unix()

	for _, msg := range messages {
		if msg.ChatID == "917416414823@c.us" &&
			msg.TypeMessage == "imageMessage" &&
			msg.Timestamp >= cutoffTime &&
			msg.Timestamp > c.lastProcessedTime {
			filtered = append(filtered, msg)
			log.Printf("Filtered message: %+v\n", msg)
		}
	}

	if len(filtered) > 0 {
		c.lastProcessedTime = filtered[len(filtered)-1].Timestamp
	}

	return filtered
}

func (c *Client) processMessages(messages []Message) {
	for _, msg := range messages {
		messageTime := time.Unix(msg.Timestamp, 0).Format(time.RFC3339)
		log.Printf("Processing message from timestamp: %s\n", messageTime)

		jsonData, err := json.MarshalIndent(msg, "", "    ")
		if err != nil {
			log.Printf("Error marshaling message to JSON: %v\n", err)
			continue
		}
		log.Printf("Image message details:\n%s\n", string(jsonData))
	}
}

func main() {
	config := Config{
		BaseURL: "https://7105.api.greenapi.com/waInstance7105171289/lastOutgoingMessages/5c093cb8c7494f8fa75e8606e9447e01cca52bf3d6e34da9bc",
		Minutes: 2,
	}

	client := NewClient(config)
	requestHandler := js.FuncOf(client.handleRequest)
	defer requestHandler.Release()
	fmt.Println("Hello", requestHandler)
	js.Global().Set("handleRequest", requestHandler)

	select {}
}
