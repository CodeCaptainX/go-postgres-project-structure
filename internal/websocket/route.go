package websocket

import (
	"log"

	jwt "snack-shop/pkg/jwt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/jmoiron/sqlx"
)

func WebSocketRoute(root fiber.Router, db *sqlx.DB, jwtMiddleware fiber.Handler) {
	// Handler instance
	wsHandler := NewHandler(db)

	ws := root.Group("/websocket")

	ws.Get("/ws", func(c *fiber.Ctx) error {

		log.Println("➡️ Incoming request to /websocket/ws")

		// Must be websocket upgrade
		if !websocket.IsWebSocketUpgrade(c) {
			log.Println("❌ Not a WebSocket upgrade request")
			return fiber.ErrUpgradeRequired
		}

		log.Println("🔄 WebSocket upgrade requested")

		// Validate JWT if "Upgrade: websocket" header exists
		if websocketUpgrade := c.Get("Upgrade"); websocketUpgrade == "websocket" {
			log.Println("🔐 Checking JWT token in WebSocket request")

			if err := jwt.ExtractUserFromWebSocketToken(c); err != nil {
				log.Println("❌ JWT validation failed:", err)
				return err
			}

			log.Println("✅ JWT validated successfully")
		}

		// Mark connection as allowed
		c.Locals("allowed", true)

		// Upgrade to websocket
		return websocket.New(func(conn *websocket.Conn) {

			log.Println("🌐 WebSocket connected!")

			// Handle disconnect
			defer func() {
				log.Println("🔌 WebSocket disconnected")
			}()

			// Actual handler
			wsHandler.HandleWebSocket(conn)

		}, websocket.Config{
			Subprotocols: []string{"Bearer"},
		})(c)
	})
}
