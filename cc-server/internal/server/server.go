package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/supabase/postgrest-go"
)

type Server struct {
	router   *gin.Engine
	upgrader websocket.Upgrader
	db       *postgrest.Client
	addr     string
	clients  map[string]*websocket.Conn
}

type ClientInfo struct {
	ID       string    `json:"id"`
	Hostname string    `json:"hostname"`
	IP       string    `json:"ip"`
	LastSeen time.Time `json:"last_seen"`
	Status   string    `json:"status"`
}

// Command represents a command to be executed on a client
type Command struct {
	ID          string    `json:"id"`
	ClientID    string    `json:"client_id"`
	Command     string    `json:"command"`
	Status      string    `json:"status"` // pending, executing, completed, failed
	Result      string    `json:"result,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
}

// RegistrationRequest represents the initial client registration
type RegistrationRequest struct {
	Token    string `json:"token"`
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
}

// NewServer creates a new C&C server instance
func NewServer(addr, supabaseURL, supabaseKey string) *Server {
	// Initialize Supabase client
	db := postgrest.NewClient(supabaseURL, &postgrest.ClientOptions{
		Headers: map[string]string{
			"apikey": supabaseKey,
			"Authorization": "Bearer " + supabaseKey,
		},
	})

	router := gin.Default()
	
	// Setup secure websocket upgrader
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// In production, check origin properly
			return true
		},
	}

	s := &Server{
		router:   router,
		upgrader: upgrader,
		db:       db,
		addr:     addr,
		clients:  make(map[string]*websocket.Conn),
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// Public registration endpoint
	s.router.POST("/register", s.handleRegistration)

	// Authenticated endpoints for CLI
	protected := s.router.Group("/")
	protected.Use(s.authMiddleware)
	{
		protected.GET("/clients", s.handleListClients)
		protected.POST("/command", s.handleSendCommand)
		protected.GET("/commands/:client_id", s.handleGetCommands)
		protected.GET("/ws", s.handleWebSocket)
	}
}

// Register new client
func (s *Server) handleRegistration(c *gin.Context) {
	var req RegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Validate registration token against database
	tokenValid, err := s.validateRegistrationToken(req.Token)
	if err != nil || !tokenValid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid registration token"})
		return
	}

	// Create client entry in database
	clientInfo := ClientInfo{
		ID:       generateClientID(), // This would be a proper ID generation function
		Hostname: req.Hostname,
		IP:       req.IP,
		LastSeen: time.Now(),
		Status:   "connected",
	}

	// Insert client into Supabase
	resp, err := s.db.From("clients").Insert(clientInfo, false, "", "", "").Execute()
	if err != nil || resp.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register client"})
		return
	}

	// Generate JWT for ongoing authentication
	token, err := s.generateClientJWT(clientInfo.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate auth token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"client_id": clientInfo.ID,
		"token":     token,
	})
}

// Validate registration token against Supabase
func (s *Server) validateRegistrationToken(token string) (bool, error) {
	var tokens []map[string]interface{}
	resp, err := s.db.From("registration_tokens").Select("*", false, "", "", "").Eq("token", token).Execute()
	if err != nil {
		return false, err
	}

	if err := s.db.ParseJSON(resp.Body, &tokens); err != nil {
		return false, err
	}

	if len(tokens) == 0 {
		return false, nil
	}

	// Check if token is still valid (not expired)
	tokenData := tokens[0]
	if expiresAt, exists := tokenData["expires_at"]; exists && expiresAt != nil {
		// This would require proper time parsing in a real implementation
		// For now, we'll assume the DB handles expiration checks
	}

	// Mark token as used to prevent reuse
	if id, exists := tokenData["id"]; exists {
		_, updateErr := s.db.From("registration_tokens").Update(map[string]interface{}{"is_used": true}, "", "").Eq("id", id).Execute()
		if updateErr != nil {
			log.Printf("Failed to mark token as used: %v", updateErr)
		}
	}

	return true, nil
}

// Generate JWT for client authentication
func (s *Server) generateClientJWT(clientID string) (string, error) {
	// In a real implementation, we'd use proper JWT signing
	// For now, returning a placeholder
	return fmt.Sprintf("jwt_for_%s", clientID), nil
}

// WebSocket handler for client communication
func (s *Server) handleWebSocket(c *gin.Context) {
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Authenticate client using JWT from query param or header
	clientID := c.Query("client_id")
	authToken := c.GetHeader("Authorization")
	if clientID == "" || authToken == "" {
		conn.WriteMessage(websocket.CloseMessage, []byte("Authentication required"))
		return
	}

	// Validate the token against Supabase
	valid, err := s.validateClientToken(clientID, authToken)
	if err != nil || !valid {
		conn.WriteMessage(websocket.CloseMessage, []byte("Invalid token"))
		return
	}

	// Add client to active connections
	s.clients[clientID] = conn
	defer func() {
		delete(s.clients, clientID)
	}()

	// Update client status in database
	s.updateClientStatus(clientID, "connected")

	// Handle incoming messages from client
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			s.updateClientStatus(clientID, "disconnected")
			break
		}

		// Process client heartbeat or command result
		if messageType == websocket.TextMessage {
			s.handleClientMessage(clientID, message)
		}
	}
}

// Validate client's auth token against Supabase
func (s *Server) validateClientToken(clientID, token string) (bool, error) {
	// In a real implementation, verify the JWT and check against Supabase
	// This is a simplified version for demonstration
	return true, nil
}

// Update client status in Supabase
func (s *Server) updateClientStatus(clientID, status string) {
	clientInfo := ClientInfo{
		ID:       clientID,
		Status:   status,
		LastSeen: time.Now(),
	}

	_, err := s.db.From("clients").Update(clientInfo, "", "").Eq("id", clientID).Execute()
	if err != nil {
		log.Printf("Failed to update client status: %v", err)
	}
}

// Handle messages from client (heartbeats, command results, etc.)
func (s *Server) handleClientMessage(clientID string, message []byte) {
	// Process different types of messages from the client
	// For example, command execution results, heartbeats, etc.
	log.Printf("Received message from client %s: %s", clientID, string(message))
}

// Middleware to authenticate CLI requests
func (s *Server) authMiddleware(c *gin.Context) {
	// In a real implementation, validate the admin token
	// For now, we'll skip this for development
	c.Next()
}

// List all registered clients
func (s *Server) handleListClients(c *gin.Context) {
	var clients []ClientInfo
	resp, err := s.db.From("clients").Select("*", false, "", "", "").Execute()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
		return
	}

	if err := s.db.ParseJSON(resp.Body, &clients); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
		return
	}

	c.JSON(http.StatusOK, clients)
}

// Send command to a client
func (s *Server) handleSendCommand(c *gin.Context) {
	var cmd Command
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid command"})
		return
	}

	// Insert command into database
	cmd.ID = generateCommandID() // This would be a proper ID generation function
	cmd.Status = "pending"
	cmd.CreatedAt = time.Now()

	resp, err := s.db.From("commands").Insert(cmd, false, "", "", "").Execute()
	if err != nil || resp.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store command"})
		return
	}

	// Send command to client if connected
	if clientConn, ok := s.clients[cmd.ClientID]; ok {
		if err := clientConn.WriteJSON(cmd); err != nil {
			log.Printf("Failed to send command to client %s: %v", cmd.ClientID, err)
		}
	} else {
		log.Printf("Client %s is not connected, command queued", cmd.ClientID)
	}

	c.JSON(http.StatusOK, cmd)
}

// Get command history for a client
func (s *Server) handleGetCommands(c *gin.Context) {
	clientID := c.Param("client_id")
	var commands []Command
	
	resp, err := s.db.From("commands").Select("*", false, "", "", "").Eq("client_id", clientID).Execute()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
		return
	}

	if err := s.db.ParseJSON(resp.Body, &commands); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
		return
	}

	c.JSON(http.StatusOK, commands)
}

// Start the C&C server
func (s *Server) Start() error {
	return s.router.Run(s.addr)
}

// generateClientID is a placeholder for client ID generation
func generateClientID() string {
	return fmt.Sprintf("client_%d", time.Now().Unix())
}

// generateCommandID is a placeholder for command ID generation
func generateCommandID() string {
	return fmt.Sprintf("cmd_%d", time.Now().Unix())
}