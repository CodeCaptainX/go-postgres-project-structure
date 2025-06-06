package auth

import (
	"snack-shop/pkg/responses"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// AuthService defines the service layer for authentication
type AuthService interface {
	Login(username, password string) (*AuthResponse, *responses.ErrorResponse)
	CheckSession(loginSession string, userID float64) (bool, *responses.ErrorResponse)
}

// authServiceImpl implements AuthService
type authServiceImpl struct {
	repo AuthRepository
}

func NewAuthService(dbPool *sqlx.DB, redisClient *redis.Client) AuthService {
	repo := NewAuthRepository(dbPool, redisClient)
	return &authServiceImpl{
		repo: repo,
	}
}

func (a *authServiceImpl) Login(username, password string) (*AuthResponse, *responses.ErrorResponse) {
	return a.repo.Login(username, password)
}

func (a *authServiceImpl) CheckSession(loginSession string, userID float64) (bool, *responses.ErrorResponse) {
	return a.repo.CheckSession(loginSession, userID)
}
