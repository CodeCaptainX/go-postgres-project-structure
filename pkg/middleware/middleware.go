package middleware

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	auth "snack-shop/internal/auth"
	response "snack-shop/pkg/http/response"
	types "snack-shop/pkg/model"
	utils "snack-shop/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func NewJwtMinddleWare(app *fiber.App, db_pool *sqlx.DB, redis *redis.Client) {
	errs := godotenv.Load()
	if errs != nil {
		log.Fatalf("Error loading .env file")
	}
	secret_key := os.Getenv("JWT_SECRET_KEY")

	// First middleware handles JWT extraction and validation
	app.Use(func(c *fiber.Ctx) error {
		// Check if this is a WebSocket upgrade request
		if websocketUpgrade := c.Get("Upgrade"); websocketUpgrade == "websocket" {
			webSocketProtocol := c.Get("Sec-webSocket-Protocol")
			if webSocketProtocol == "" {
				return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
					"error": "Missing WebSocket protocol for authentication",
				})
			}

			parts := strings.Split(webSocketProtocol, ",")
			if len(parts) != 2 || strings.TrimSpace(parts[0]) != "Bearer" {
				return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
					"error": "Invalid webSocket protocol authentication format",
				})
			}

			tokenString := strings.TrimSpace(parts[1])
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret_key), nil
			})
			if err != nil || !token.Valid {
				log.Printf("❌ WebSocket JWT validation failed: %v", err)
				return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
					"error": "Invalid or expired JWT token",
				})
			}

			c.Locals("jwt_data", token)
			c.Set("Sec-WebSocket-Protocol", "Bearer")
			return c.Next()
		}

		// Handle standard HTTP requests with Authorization header
		authHeader := c.Get("Authorization")
		// log.Println("🔍 Authorization Header:", authHeader)

		// Manual JWT extraction for HTTP requests
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			// Manually extract the token, trim any extra spaces
			tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer"))

			// Parse and validate the token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret_key), nil
			})

			if err != nil {
				log.Printf("❌ JWT parsing error: %v", err)
				return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
					"error": fmt.Sprintf("Invalid token: %v", err),
				})
			}

			if !token.Valid {
				log.Println("❌ Token is invalid")
				return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
					"error": "Invalid token",
				})
			}

			// Store the parsed token for the next middleware
			c.Locals("jwt_data", token)
			return c.Next()
		}

		// If no Authorization header or no valid Bearer token, return unauthorized
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Missing or invalid Authorization header",
		})
	})

	// Second middleware handles user context validation
	app.Use(func(c *fiber.Ctx) error {
		// Check if jwt_data exists and is valid
		tokenData := c.Locals("jwt_data")
		if tokenData == nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing JWT data",
			})
		}

		userToken, ok := tokenData.(*jwt.Token)
		if !ok {
			log.Println("❌ Failed to cast JWT token")
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid JWT token format",
			})
		}

		// Cast claims to jwt.MapClaims
		uclaim, ok := userToken.Claims.(jwt.MapClaims)
		if !ok {
			log.Println("❌ Failed to cast JWT claims")
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid JWT claims",
			})
		}

		// Debug log to see JWT contents
		// log.Println("✅ JWT Claims:")
		// for k, v := range uclaim {
		// 	log.Printf("  %s: %v\n", k, v)
		// }

		return handleUserContext(c, uclaim, db_pool, redis)
	})
}

func handleUserContext(c *fiber.Ctx, uclaim jwt.MapClaims, db *sqlx.DB, redis *redis.Client) error {

	// --- Get ONLY login_session from JWT ---
	loginSession, ok := uclaim["login_session"].(string)
	if !ok || loginSession == "" {
		errMsg := utils.Translate("login_session_missing", nil, c)
		return c.Status(http.StatusUnprocessableEntity).JSON(response.NewResponseError(
			errMsg, -500, fmt.Errorf("missing or invalid 'login_session' in claims"),
		))
	}

	// --- (Optional) Verify token exp ---
	exp, ok := uclaim["exp"].(float64)
	if !ok {
		errMsg := utils.Translate("exp_missing", nil, c)
		return c.Status(http.StatusUnprocessableEntity).JSON(response.NewResponseError(
			errMsg, -500, fmt.Errorf("missing or invalid 'exp' in claims"),
		))
	}

	// -------- Validate Session (NEW WAY) --------
	sv := auth.NewAuthService(db, redis)
	sessionData, errResp := sv.CheckSession(loginSession)

	if errResp != nil {
		errMsg := utils.Translate("login_session_invalid", nil, c)
		return c.Status(http.StatusUnprocessableEntity).JSON(
			response.NewResponseError(errMsg, -500, errResp.Err),
		)
	}

	// -------- Build UserContext USING sessionData --------
	uCtx := types.UserContext{
		UserID:       float64(sessionData.UserID),
		UserUuid:     sessionData.UserUUID,
		UserName:     sessionData.UserName,
		RoleId:       uint64(sessionData.RoleID),
		LoginSession: sessionData.LoginSession,
		Exp:          time.Unix(int64(exp), 0),
		UserAgent:    c.Get("User-Agent", "unknown"),
		Ip:           c.IP(),
	}

	// Save to Fiber context for controllers to use
	c.Locals("UserContext", uCtx)
	c.Locals("UserSessionData", sessionData)

	return c.Next()
}
