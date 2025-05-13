package auth

import (
	constants "snack-shop/pkg/constants"
	response "snack-shop/pkg/http/response"
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
		msg := utils.Translate("login_invalid", nil, c)
		return c.Status(fiber.StatusUnprocessableEntity).JSON(
			response.NewResponseError(
				msg,
				constants.Login_invalid,
				err,
			),
		)
	}

	success, err := a.authService.Login(req.Auth.Username, req.Auth.Password)

	if err != nil {
		msg := utils.Translate(err.MessageID, nil, c)
		return c.Status(fiber.StatusUnauthorized).JSON(response.NewResponseError(
			msg,
			constants.LoginFailed,
			err.Err,
		))
	}

	msg := utils.Translate("login_success", nil, c)

	return c.Status(fiber.StatusOK).JSON(response.NewResponse(
		msg,
		constants.Login_success,
		success,
	))
}
