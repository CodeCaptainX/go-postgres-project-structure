package middleware

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	auth "snack-shop/internal/auth"
	custom_models "snack-shop/pkg/model"
	custom_translate "snack-shop/pkg/utils"

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

	login_session, ok := uclaim["login_session"].(string)
	if !ok || login_session == "" {
		smg_error := custom_translate.NewResponseError(
			custom_translate.Translate(c, "login_session_missing"),
			-500,
			fmt.Errorf("%s", custom_translate.Translate(c, "login_session_missing")),
		)
		return c.Status(http.StatusUnprocessableEntity).JSON(smg_error)
	}

	uCtx := custom_models.PlayerContext{
		PlayerID:     uclaim["player_id"].(float64),
		UserName:     uclaim["username"].(string),
		LoginSession: uclaim["login_session"].(string),
		Exp:          time.Unix(int64(uclaim["exp"].(float64)), 0),
		UserAgent:    c.Get("User-Agent", "unknown"),
		Ip:           c.IP(),
	}
	c.Locals("PlayerContext", uCtx)

	sv := auth.NewAuthService(db, redis)
	success, err := sv.CheckSession(login_session, uCtx.PlayerID)
	if err != nil || !success {
		smg_error := custom_translate.NewResponseError(
			custom_translate.Translate(c, "login_session_invalid"),
			-500,
			fmt.Errorf("%s", custom_translate.Translate(c, "login_session_invalid")),
		)
		return c.Status(http.StatusUnprocessableEntity).JSON(smg_error)
	}

	return c.Next()
}
