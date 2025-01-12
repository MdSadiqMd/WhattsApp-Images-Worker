package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/syumai/workers"
)

type Config struct {
	BaseURL  string
	Token    string
	Interval time.Duration
	Minutes  int
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

func (c *Client) fetchMessages() error {
	payload := map[string]int{"minutes": c.config.Minutes}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling payload: %w", err)
	}

	req, err := http.NewRequest("POST", c.config.BaseURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %w", err)
	}

	/* log.Printf("Response Body: %s\n", body) // Log the response body */

	var messages []Message
	if err := json.Unmarshal(body, &messages); err != nil {
		return fmt.Errorf("error unmarshaling response: %w", err)
	}

	filteredMessages := c.filterMessages(messages)
	if len(filteredMessages) > 0 {
		c.processMessages(filteredMessages)
	} else {
		log.Println("No new messages found.")
	}

	return nil
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
		log.Printf("Filtered message data:\n%s\n", string(jsonData))
	}
}

func (client *Client) startFetching() {
	ticker := time.NewTicker(client.config.Interval)
	defer ticker.Stop()

	for range ticker.C {
		if err := client.fetchMessages(); err != nil {
			log.Printf("Error fetching messages: %v\n", err)
		}
	}
}

func main() {
	config := Config{
		BaseURL:  "https://7105.api.greenapi.com/waInstance7105171289/lastOutgoingMessages/5c093cb8c7494f8fa75e8606e9447e01cca52bf3d6e34da9bc",
		Interval: 5 * time.Second,
		Minutes:  2,
	}

	client := NewClient(config)

	go client.startFetching()
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Fetching messages in the background..."))
	})

	workers.Serve(r)
}
