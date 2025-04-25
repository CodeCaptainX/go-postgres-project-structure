package main

import (
	"fmt"
	configs "snack-shop/config"
	database "snack-shop/config/database"
	redis "snack-shop/config/redis"
	"snack-shop/handler"
	custom_log "snack-shop/pkg/custom_log"
	translate "snack-shop/pkg/utils/translate"
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

	// Initialize the translate
	if err := translate.Init(); err != nil {
		custom_log.NewCustomLog("Failed_initialize_i18n", err.Err.Error(), "error")
	}

	handler.NewFrontService(app, db_pool, rdb)

	app.Listen(fmt.Sprintf("%s:%d", app_configs.AppHost, app_configs.AppPort))
}
