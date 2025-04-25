package auth

import (
	custom_validator "snack-shop/pkg/validator"

	"github.com/gofiber/fiber/v2"
)

// AuthLoginRequest represents the login request payload
type AuthLoginRequest struct {
	Auth struct {
		Username string `json:"username" validate:"required"`
		Password string `json:"password" validate:"required"`
	} `json:"auth"`
}

// bind validates and parses the login request
func (r *AuthLoginRequest) bind(c *fiber.Ctx, v *custom_validator.Validator) error {
	if err := c.BodyParser(r); err != nil {
		return err
	}
	if err := v.Validate(r); err != nil {
		return err
	}
	return nil
}

type AuthResponse struct {
	Auth struct {
		Token     string `json:"token"`
		TokenType string `json:"token_type"`
	} `json:"auths"`
}

type MemberData struct {
	ID       int    `db:"id"`
	Username string `db:"user_name"`
	Email    string `db:"email"`
	Password string `db:"password"`
}

type RedisSession struct {
	LoginSession string `json:"login_session"`
}
