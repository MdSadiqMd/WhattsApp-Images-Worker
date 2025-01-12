package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
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
	httpClient        *http.Client
	lastProcessedTime int64
}

func NewClient(config Config) *Client {
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		lastProcessedTime: time.Now().Add(-time.Duration(config.Minutes) * time.Minute).Unix(),
	}
}

func (c *Client) Start() {
	log.Printf("Starting client with %v interval, monitoring last %d minutes\n",
		c.config.Interval, c.config.Minutes)
	for {
		if err := c.fetchMessages(); err != nil {
			log.Printf("Error fetching messages: %v\n", err)
		}
		time.Sleep(c.config.Interval)
	}
}

func (c *Client) fetchMessages() error {
	payload := map[string]int{"minutes": c.config.Minutes}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling payload: %w", err)
	}

	req, err := http.NewRequest("GET", c.config.BaseURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %w", err)
	}

	var messages []Message
	if err := json.Unmarshal(body, &messages); err != nil {
		return fmt.Errorf("error unmarshaling response: %w", err)
	}

	filteredMessages := c.filterMessages(messages)
	if len(filteredMessages) > 0 {
		c.processMessages(filteredMessages)
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
		}
	}

	if len(filtered) > 0 {
		latestTimestamp := filtered[0].Timestamp
		for _, msg := range filtered[1:] {
			if msg.Timestamp > latestTimestamp {
				latestTimestamp = msg.Timestamp
			}
		}
		c.lastProcessedTime = latestTimestamp
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

func main() {
	config := Config{
		BaseURL:  "https://7105.api.greenapi.com/waInstance7105171289/lastOutgoingMessages/5c093cb8c7494f8fa75e8606e9447e01cca52bf3d6e34da9bc",
		Interval: 5 * time.Second,
		Minutes:  2,
	}

	client := NewClient(config)
	client.Start()
}
