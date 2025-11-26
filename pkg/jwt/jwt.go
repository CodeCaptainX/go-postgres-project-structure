package jwt

import (
	"fmt"

	types "snack-shop/pkg/model"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	jtoken "github.com/golang-jwt/jwt/v4"
)

type JWT struct {
	userContext *jtoken.Token
}

func NewJWT() *JWT {
	return &JWT{
		userContext: &jtoken.Token{},
	}
}

// ! For now i handle this for websocket connection, for player from Bet project only

func ExtractUserFromWebSocketToken(c *fiber.Ctx) error {
	fmt.Println("🚀 Step 1: Starting ExtractUserFromWebSocketToken")

	// Get the subprotocols from the WebSocket header
	subprotocols := c.Get("Sec-Websocket-Protocol")
	if subprotocols == "" {
		fmt.Println("❌ Step 2 Error: No Sec-WebSocket-Protocol header found")
		return fiber.ErrUnauthorized
	}

	// Token format: "Bearer <token>"
	tokenParts := strings.Split(subprotocols, " ")
	if len(tokenParts) < 2 {
		fmt.Println("❌ Step 3 Error: Token not in expected format")
		return fiber.ErrUnauthorized
	}

	tokenStr := tokenParts[1]
	fmt.Println("🚀 Step 4: Token extracted:", tokenStr)

	// Parse token
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			fmt.Println("❌ Step 5 Error: Unexpected signing method")
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("SECRET"), nil
	})
	if err != nil {
		fmt.Println("❌ Step 6 Error: Failed to parse token:", err)
		return fiber.ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		fmt.Println("❌ Step 7 Error: Invalid token claims")
		return fiber.ErrUnauthorized
	}

	fmt.Println("🚀 Step 7: Claims extracted:", claims)

	// Extract User ID
	userId := 0
	if v, ok := claims["user_id"].(float64); ok {
		userId = int(v)
	}

	// Build the WebSocket key alias
	keyAliasForWebsocket := "user"

	// Store context
	userContext := &types.UserContext{
		UserID:               float64(userId),
		KeyAliasForWebsocket: keyAliasForWebsocket,
		UserAgent:            c.Get("User-Agent"),
		Ip:                   c.IP(),
	}

	c.Locals("userContext", userContext)

	fmt.Println("🚀 Step 9: Final userContext:", userContext)

	return nil
}
