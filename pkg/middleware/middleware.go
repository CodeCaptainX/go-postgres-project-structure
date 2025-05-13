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

	jwtware "github.com/gofiber/contrib/jwt"
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
	app.Use(func(c *fiber.Ctx) error {
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
					"error": "Invalid webSocket protocal authenticaion format",
				})
			}

			tokenString := strings.TrimSpace(parts[1])

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret_key), nil
			})
			if err != nil || !token.Valid {
				return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
					"error": "Invalid or expired JWT token",
				})
			}

			c.Locals("jwt_data", token)
			c.Set("Sec-WebSocket-Protocol", "Bearer")
			return c.Next()
		}

		return jwtware.New(jwtware.Config{
			SigningKey: jwtware.SigningKey{Key: []byte(secret_key)},
			ContextKey: "jwt_data",
		})(c)
	})

	app.Use(func(c *fiber.Ctx) error {
		user_token := c.Locals("jwt_data").(*jwt.Token)
		uclaim := user_token.Claims.(jwt.MapClaims)

		if websocketUpgrade := c.Get("Upgrade"); websocketUpgrade == "websocket" {
			return handleUserContext(c, uclaim, db_pool, redis)
		}
		return handleUserContext(c, uclaim, db_pool, redis)
	})
}

func handleUserContext(c *fiber.Ctx, uclaim jwt.MapClaims, db *sqlx.DB, redis *redis.Client) error {
	// Verify login_session claim
	loginSession, ok := uclaim["login_session"].(string)
	if !ok || loginSession == "" {
		errMsg := utils.Translate("login_session_missing", nil, c)
		return c.Status(http.StatusUnprocessableEntity).JSON(response.NewResponseError(
			errMsg,
			-500,
			fmt.Errorf("missing or invalid 'login_session' in claims"),
		))
	}

	// Verify player_id claim
	userID, ok := uclaim["user_id"].(float64)
	if !ok {
		errMsg := utils.Translate("player_id_missing", nil, c)
		return c.Status(http.StatusUnprocessableEntity).JSON(response.NewResponseError(
			errMsg,
			-500,
			fmt.Errorf("missing or invalid 'player_id' in claims"),
		))
	}

	uuid, ok := uclaim["user_uuid"].(string)
	if !ok {
		return c.Status(http.StatusUnprocessableEntity).JSON(response.NewResponseError(
			"Invalid or missing 'user_uuid' in claims", -500, fmt.Errorf("missing or invalid 'user_uuid'"),
		))
	}

	// Verify username claim
	username, ok := uclaim["username"].(string)
	if !ok || username == "" {
		errMsg := utils.Translate("username_missing", nil, c)
		return c.Status(http.StatusUnprocessableEntity).JSON(response.NewResponseError(
			errMsg,
			-500,
			fmt.Errorf("missing or invalid 'username' in claims"),
		))
	}

	// Verify exp (expiration time) claim
	exp, ok := uclaim["exp"].(float64)
	if !ok {
		errMsg := utils.Translate("exp_missing", nil, c)
		return c.Status(http.StatusUnprocessableEntity).JSON(response.NewResponseError(
			errMsg,
			-500,
			fmt.Errorf("missing or invalid 'exp' in claims"),
		))
	}

	// Create PlayerContext struct for storing session details
	uCtx := types.UserContext{
		UserID:       userID,
		UserUuid:     uuid,
		UserName:     username,
		LoginSession: loginSession,
		Exp:          time.Unix(int64(exp), 0),
		UserAgent:    c.Get("User-Agent", "unknown"),
		Ip:           c.IP(),
	}
	c.Locals("UserContext", uCtx)

	// Validate session with Redis and database
	sv := auth.NewAuthService(db, redis)
	success, err := sv.CheckSession(loginSession, uCtx.UserID)
	if err != nil || !success {
		errMsg := utils.Translate("login_session_invalid", nil, c)
		return c.Status(http.StatusUnprocessableEntity).JSON(response.NewResponseError(
			errMsg,
			-500,
			fmt.Errorf("invalid session or unable to verify session"),
		))
	}

	// Proceed with next handler if the session is valid
	return c.Next()
}
