package router

import (
	"github.com/gofiber/contrib/fiberi18n/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"golang.org/x/text/language"
)

func New() *fiber.App {
	f := fiber.New(fiber.Config{})

	f.Use(logger.New())

	f.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, HEAD, PUT, PATCH, POST, DELETE",
	})).Use(
		fiberi18n.New(&fiberi18n.Config{
			RootPath: "pkg/translates/localize/i18n",
			AcceptLanguages: []language.Tag{
				language.Chinese,
				language.MustParse("km"),
				language.English,
			},
			DefaultLanguage: language.Khmer,
		}),
	)
	return f
}
