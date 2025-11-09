package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type APIClient struct {
	BaseURL string
	Client  *http.Client
}

type ClientInfo struct {
	ID       string `json:"id"`
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	LastSeen string `json:"last_seen"`
	Status   string `json:"status"`
}

type Command struct {
	ID      string `json:"id"`
	Command string `json:"command"`
	Status  string `json:"status"`
	Result  string `json:"result,omitempty"`
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		BaseURL: baseURL,
		Client:  &http.Client{},
	}
}

func (c *APIClient) ListClients() ([]ClientInfo, error) {
	resp, err := c.Client.Get(c.BaseURL + "/clients")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %d", resp.StatusCode)
	}

	var clients []ClientInfo
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &clients); err != nil {
		return nil, err
	}

	return clients, nil
}

func (c *APIClient) SendCommand(clientID, command string) (*Command, error) {
	cmd := Command{
		ClientID: clientID,
		Command:  command,
	}

	jsonData, err := json.Marshal(cmd)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Post(c.BaseURL+"/command", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %d", resp.StatusCode)
	}

	var sentCmd Command
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &sentCmd); err != nil {
		return nil, err
	}

	return &sentCmd, nil
}

func (c *APIClient) GetCommands(clientID string) ([]Command, error) {
	resp, err := c.Client.Get(c.BaseURL + "/commands/" + clientID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %d", resp.StatusCode)
	}

	var commands []Command
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &commands); err != nil {
		return nil, err
	}

	return commands, nil
}