package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	auth "snack-shop/internal/auth"
	middleware "snack-shop/pkg/middleware"
)

type ServiceHandlers struct {
	Fronted *FrontService
}

type FrontService struct {
	AuthHandler *auth.AuthRoute
}

func NewFrontService(app *fiber.App, db_pool *sqlx.DB, redis *redis.Client) *FrontService {

	// Authentication
	auth := auth.NewAuthRoute(app, db_pool, redis).RegisterAuthRoute()

	// Middleware
	middleware.NewJwtMinddleWare(app, db_pool, redis)

	return &FrontService{
		AuthHandler: auth,
	}
}

func NewServiceHandlers(app *fiber.App, db_pool *sqlx.DB, redis *redis.Client) *ServiceHandlers {

	return &ServiceHandlers{
		Fronted: NewFrontService(app, db_pool, redis),
	}
}
