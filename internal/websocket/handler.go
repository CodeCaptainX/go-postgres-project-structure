package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	response "snack-shop/pkg/http/response"
	types "snack-shop/pkg/model"

	// "scratch_card_admin/pkg/constant"
	// statuscode "scratch_card_admin/pkg/constant/statuscode/bet_rmq"
	// pkg_model "scratch_card_admin/pkg/models"
	// "scratch_card_admin/pkg/utils/apiresponse"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Client represents a single WebSocket connection with metadata
type Client struct {
	ID          string          // Unique connection ID
	Conn        *websocket.Conn // The actual WebSocket connection
	UserID      string          // User identifier from JWT
	ConnectedAt time.Time       // When this connection was established
	LastActive  time.Time       // Last activity timestamp
	SendChan    chan []byte     // Channel for sending messages to this client
	mu          sync.Mutex      // Mutex for thread-safe operations on this client
	done        chan struct{}   // Signal channel for shutdown
	closeOnce   sync.Once       // Ensure cleanup only happens once
}

// Global storage for clients - now supports multiple connections per user
var (
	// Map: UserID -> Array of Client connections
	Clients      = make(map[string][]*Client)
	ClientsMutex = &sync.RWMutex{}
)

// WebSocketHandler struct to handle database and WebSocket connections
type WebSocketHandler struct {
	db *sqlx.DB
}

func NewHandler(db *sqlx.DB) *WebSocketHandler {
	return &WebSocketHandler{
		db: db,
	}
}

// HandleWebSocket handles WebSocket communication with support for multiple connections
func (h *WebSocketHandler) HandleWebSocket(c *websocket.Conn) {
	// Extract user context from JWT
	Context, ok := c.Locals("userContext").(*types.UserContext)
	if !ok || Context == nil {
		log.Println("Invalid user context")
		return
	}

	// Build final WebSocket key: "user1" or "member10" etc.
	userID := Context.KeyAliasForWebsocket

	// Create a unique client instance for this connection
	client := &Client{
		ID:          uuid.New().String(), // Unique ID for this specific connection
		Conn:        c,
		UserID:      userID,
		ConnectedAt: time.Now(),
		LastActive:  time.Now(),
		SendChan:    make(chan []byte, 256), // Buffered channel for messages
		done:        make(chan struct{}),    // Shutdown signal
	}

	// Add this client to the map (supports multiple connections per user)
	h.AddClient(client)

	log.Printf("✅ User %s connected (ConnectionID: %s, Total connections for this user: %d)\n",
		userID, client.ID, h.GetUserConnectionCount(userID))

	// Cleanup when connection closes
	defer func() {
		client.cleanup()
		h.DropClient(client)
		log.Printf("🔌 User %s disconnected (ConnectionID: %s, Remaining connections: %d)\n",
			userID, client.ID, h.GetUserConnectionCount(userID))
	}()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start goroutines for reading and writing
	var wg sync.WaitGroup
	wg.Add(2)

	// Writer goroutine - sends messages from SendChan to client
	go func() {
		defer wg.Done()
		h.writePump(ctx, client)
	}()

	// Reader goroutine - receives messages from client
	go func() {
		defer wg.Done()
		h.readPump(ctx, client, cancel)
	}()

	// Wait for both goroutines to finish
	wg.Wait()
}

// cleanup handles safe shutdown of client resources
func (c *Client) cleanup() {
	c.closeOnce.Do(func() {
		close(c.done)
		close(c.SendChan)
		c.Conn.Close()
	})
}

// readPump handles incoming messages from the WebSocket client
func (h *WebSocketHandler) readPump(ctx context.Context, client *Client, cancel context.CancelFunc) {
	defer cancel()

	// Set initial read deadline
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	// Handle pong messages to keep connection alive
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		client.mu.Lock()
		client.LastActive = time.Now()
		client.mu.Unlock()
		return nil
	})

	for {
		select {
		case <-ctx.Done():
			return
		case <-client.done:
			return
		default:
			_, msg, err := client.Conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("❌ WebSocket error for user %s (conn: %s): %v", client.UserID, client.ID, err)
				}
				return
			}

			// Update last active time
			client.mu.Lock()
			client.LastActive = time.Now()
			client.mu.Unlock()

			log.Printf("📨 Received from user %s (conn: %s): %s", client.UserID, client.ID, msg)

			// Handle the message (you can add your business logic here)
			// Example: h.handleClientMessage(client, msg)
		}
	}
}

// writePump handles outgoing messages to the WebSocket client
func (h *WebSocketHandler) writePump(ctx context.Context, client *Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Log connection stats safely
	h.logConnectionStats()

	for {
		select {
		case <-ctx.Done():
			log.Printf("🛑 Context done for user %s (conn: %s)", client.UserID, client.ID)
			return

		case <-client.done:
			log.Printf("🛑 Client shutdown for user %s (conn: %s)", client.UserID, client.ID)
			return

		case message, ok := <-client.SendChan:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

			log.Println("-----------------------------------------------------")
			log.Printf("📨 Received SendChan event for user %s (conn: %s)", client.UserID, client.ID)
			log.Printf("📦 ok = %v", ok)
			log.Printf("📦 message raw = %s", string(message))
			log.Println("-----------------------------------------------------")

			if !ok {
				log.Printf("❌ SendChan CLOSED for user %s (conn: %s)", client.UserID, client.ID)
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("❌ Error writing message to user %s (conn: %s): %v", client.UserID, client.ID, err)
				return
			}

			log.Printf("✅ Message delivered to user %s (conn: %s)\n", client.UserID, client.ID)

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("❌ Ping failed for user %s (conn: %s): %v", client.UserID, client.ID, err)
				return
			}
			log.Printf("🏓 Ping sent to user %s (conn: %s)", client.UserID, client.ID)
		}
	}
}

// logConnectionStats safely logs connection statistics
func (h *WebSocketHandler) logConnectionStats() {
	ClientsMutex.RLock()
	defer ClientsMutex.RUnlock()

	for uid, conns := range Clients {
		log.Printf("👥 User %s has %d connections", uid, len(conns))
	}
}

// AddClient adds a new client connection (supports multiple connections per user)
func (h *WebSocketHandler) AddClient(client *Client) {
	ClientsMutex.Lock()
	defer ClientsMutex.Unlock()

	// Append to the array of connections for this user
	Clients[client.UserID] = append(Clients[client.UserID], client)

	log.Printf("📊 Client added - User: %s, ConnID: %s, Total connections for user: %d",
		client.UserID, client.ID, len(Clients[client.UserID]))
}

// DropClient removes a specific client connection
func (h *WebSocketHandler) DropClient(client *Client) {
	ClientsMutex.Lock()
	defer ClientsMutex.Unlock()

	connections := Clients[client.UserID]

	// Find and remove this specific connection
	for i, conn := range connections {
		if conn.ID == client.ID {
			// Remove this connection from the slice
			Clients[client.UserID] = append(connections[:i], connections[i+1:]...)

			// If no more connections for this user, delete the key
			if len(Clients[client.UserID]) == 0 {
				delete(Clients, client.UserID)
				log.Printf("🗑️  User %s has no more connections, removed from map", client.UserID)
			}

			log.Printf("📊 Client removed - User: %s, ConnID: %s, Remaining connections: %d",
				client.UserID, client.ID, len(Clients[client.UserID]))
			break
		}
	}
}

// GetUserConnectionCount returns the number of active connections for a user
func (h *WebSocketHandler) GetUserConnectionCount(userID string) int {
	ClientsMutex.RLock()
	defer ClientsMutex.RUnlock()
	return len(Clients[userID])
}

// DropClients closes all WebSocket connections (for graceful shutdown)
func DropClients() {
	ClientsMutex.Lock()
	defer ClientsMutex.Unlock()

	for userID, connections := range Clients {
		for _, client := range connections {
			// Use cleanup to safely close everything
			client.cleanup()
			log.Printf("🛑 Closed connection for user %s (conn: %s)", userID, client.ID)
		}
		// Remove all connections for this user
		delete(Clients, userID)
	}

	log.Println("🛑 All WebSocket connections closed")
}

// GetAllClients returns all connected user IDs
func GetAllClients() []string {
	ClientsMutex.RLock()
	defer ClientsMutex.RUnlock()

	clientIDs := make([]string, 0, len(Clients))
	for userID := range Clients {
		clientIDs = append(clientIDs, userID)
	}
	return clientIDs
}

// GetAllConnections returns detailed info about all connections
func GetAllConnections() map[string][]string {
	ClientsMutex.RLock()
	defer ClientsMutex.RUnlock()

	result := make(map[string][]string)
	for userID, connections := range Clients {
		connIDs := make([]string, len(connections))
		for i, client := range connections {
			connIDs[i] = client.ID
		}
		result[userID] = connIDs
	}
	return result
}

// BroadcastToUser sends a message to ALL connections of a specific user
func (h *WebSocketHandler) BroadcastToUser(c *fiber.Ctx) error {
	userContext := c.Locals("userContext").(*types.UserContext)
	keyWebsocket := userContext.KeyAliasForWebsocket
	fmt.Println("🚀 ~ file: handler.go ~ line 304 ~ func ~ keyWebsocket : ", keyWebsocket)
	message := []byte(c.FormValue("message") + " - " + keyWebsocket)

	sent := h.SendToUser(keyWebsocket, message)

	if sent == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fmt.Sprintf("No active connections for user %s", keyWebsocket),
		})
	}

	return c.JSON(fiber.Map{
		"message":             fmt.Sprintf("Message sent to %d connection(s) for user %s", sent, keyWebsocket),
		"user_id":             keyWebsocket,
		"connections_reached": sent,
	})
}

// SendToUser sends message to all connections of a specific user
func (h *WebSocketHandler) SendToUser(userID string, message []byte) int {
	fmt.Println("🚀 ~ file: handler.go ~ line 324 ~ func ~ userID : ", userID)

	// Get connections with read lock
	ClientsMutex.RLock()
	connections := make([]*Client, len(Clients[userID]))
	copy(connections, Clients[userID])
	ClientsMutex.RUnlock()

	fmt.Println("🚀 ~ file: handler.go ~ line 327 ~ func ~ connections : ", connections)

	if len(connections) == 0 {
		log.Printf("⚠️  No connections found for user %s", userID)
		return 0
	}

	sentCount := 0
	for _, client := range connections {
		// Check if client is still active before sending
		select {
		case <-client.done:
			log.Printf("⚠️  Connection closed for user %s (conn: %s)", userID, client.ID)
			continue
		default:
		}

		select {
		case client.SendChan <- message:
			sentCount++
			log.Printf("✅ Message queued for user %s (conn: %s)", userID, client.ID)
		case <-time.After(100 * time.Millisecond):
			log.Printf("⚠️  Send channel timeout for user %s (conn: %s)", userID, client.ID)
		}
	}

	return sentCount
}

// BroadcastBalance broadcasts balance updates to admins and the specific member
func (h *WebSocketHandler) BroadcastBalance(memberID int64, currencyID int64, balance float64) {
	update := BalanceUpdate{
		MemberID:  memberID,
		CurrentID: currencyID,
		Balance:   balance,
		Action:    "update",
		Topic:     "update_member_balance",
	}

	data := response.NewResponse(
		"hello Im under water please help me !",
		2000,
		update,
	)
	// Marshal message OUTSIDE the lock
	message, err := json.Marshal(data)
	if err != nil {
		log.Printf("❌ Error marshaling balance update: %v", err)
		return
	}

	targetMember := fmt.Sprintf("member%d", memberID)

	// Get snapshot of clients to avoid holding lock during send
	ClientsMutex.RLock()
	clientsToSend := make(map[string][]*Client)
	for userKey, connections := range Clients {
		// Send only to admin users ("userX") or the member who owns this balance
		if strings.HasPrefix(userKey, "user") || userKey == targetMember {
			clientsToSend[userKey] = make([]*Client, len(connections))
			copy(clientsToSend[userKey], connections)
		}
	}
	ClientsMutex.RUnlock()

	sentCount := 0
	for userKey, connections := range clientsToSend {
		fmt.Println("🚀 ~ file: handler.go ~ line 387 ~ broadcasting to userKey : ", userKey)

		for _, client := range connections {
			// Check if client is still active
			select {
			case <-client.done:
				log.Printf("⚠️ Connection closed for %s (conn %s)", userKey, client.ID)
				continue
			default:
			}

			select {
			case client.SendChan <- message:
				sentCount++
				log.Printf("📤 Sent balance update to %s (conn: %s)", userKey, client.ID)
			case <-time.After(100 * time.Millisecond):
				log.Printf("⚠️ Send timeout for %s (conn %s)", userKey, client.ID)
			}
		}
	}

	log.Printf("📤 BroadcastBalance: sent to %d connections (admins + %s)",
		sentCount, targetMember)
}

// BroadcastBetSettlement broadcasts bet settlement to ALL connected clients
// func BroadcastBetSettlement(message string, status string, StatusCode int) {
// 	payload := BetSattlement{
// 		Time:    time.Now().Format(time.RFC3339),
// 		Message: message,
// 		Topic:   "bet_settlement",
// 	}

// 	var success bool
// 	switch status {
// 	case "success":
// 		success = true
// 	case "pending":
// 		success = false
// 	}

// 	betSettlement := apiresponse.APIResponseData(
// 		success,
// 		message,
// 		StatusCode,
// 		payload,
// 	)

// 	// Marshal OUTSIDE the lock
// 	betSettlementResponse, err := json.Marshal(betSettlement)
// 	if err != nil {
// 		log.Printf("❌ Error marshaling bet settlement: %v", err)
// 		return
// 	}

// 	// Get snapshot of all clients
// 	ClientsMutex.RLock()
// 	allClients := make(map[string][]*Client)
// 	for userID, connections := range Clients {
// 		allClients[userID] = make([]*Client, len(connections))
// 		copy(allClients[userID], connections)
// 	}
// 	ClientsMutex.RUnlock()

// 	totalSent := 0
// 	for userID, connections := range allClients {
// 		for _, client := range connections {
// 			// Check if client is still active
// 			select {
// 			case <-client.done:
// 				log.Printf("⚠️ Connection closed for user %s (conn: %s)", userID, client.ID)
// 				continue
// 			default:
// 			}

// 			select {
// 			case client.SendChan <- betSettlementResponse:
// 				totalSent++
// 			case <-time.After(100 * time.Millisecond):
// 				log.Printf("⚠️  Timeout sending bet settlement to user %s (conn: %s)",
// 					userID, client.ID)
// 			}
// 		}
// 	}

// 	log.Printf("📤 Bet settlement broadcast sent to %d total connections across %d users",
// 		totalSent, len(allClients))
// }

// SendToUsers sends a message to multiple specific users
func (h *WebSocketHandler) SendToUsers(userIDs []string, message []byte) {
	for _, uid := range userIDs {
		h.SendToUser(uid, message)
	}
}

// BalanceUpdate represents a balance update message
