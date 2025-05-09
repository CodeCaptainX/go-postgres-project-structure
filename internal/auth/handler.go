package auth

import (
	constants "snack-shop/pkg/constants"
	utils "snack-shop/pkg/utils"
	custom_validator "snack-shop/pkg/validator"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// AuthHandler handles HTTP requests related to authentication
type AuthHandler struct {
	authService AuthService
}

// NewAuthHandler creates a new instance of AuthHandler
func NewAuthHandler(dbPool *sqlx.DB, redisClient *redis.Client) *AuthHandler {
	return &AuthHandler{
		authService: NewAuthService(dbPool, redisClient),
	}
}

// Login handles user login request
func (a *AuthHandler) Login(c *fiber.Ctx) error {
	v := custom_validator.NewValidator()
	req := &AuthLoginRequest{}

	if err := req.bind(c, v); err != nil {
		msg := utils.Translate(c, nil, "login_invalid")
		return c.Status(fiber.StatusUnprocessableEntity).JSON(
			utils.NewResponseError(
				msg,
				constants.Login_invalid,
				err,
			),
		)
	}

	success, err := a.authService.Login(req.Auth.Username, req.Auth.Password)

	if err != nil {
		msg := utils.Translate(c, nil, err.MessageID)
		return c.Status(fiber.StatusUnauthorized).JSON(utils.NewResponseError(
			msg,
			constants.LoginFailed,
			err.Err,
		))
	}

	msg := utils.Translate(c, nil, "login_success")

	return c.Status(fiber.StatusOK).JSON(utils.NewResponse(
		msg,
		constants.Login_success,
		success,
	))
}
