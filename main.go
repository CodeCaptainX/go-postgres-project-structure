package main

import (
	"fmt"
	configs "snack-shop/config"
	database "snack-shop/config/database"
	redis "snack-shop/config/redis"
	"snack-shop/handler"
	routers "snack-shop/routers"
)

func main() {

	// Initial configuration
	app_configs := configs.NewConfig()

	// Initial database
	db_pool := database.GetDB()

	// Initialize router
	app := routers.New()

	// Initialize redis client
	rdb := redis.NewRedisClient()

	handler.NewFrontService(app, db_pool, rdb)

	app.Listen(fmt.Sprintf("%s:%d", app_configs.AppHost, app_configs.AppPort))
}
