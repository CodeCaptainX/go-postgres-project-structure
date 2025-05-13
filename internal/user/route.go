package user

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

// UserHandler struct
type UserRoute struct {
	app     *fiber.App
	db      *sqlx.DB
	handler *UserHandler
}

func NewUserRoute(app *fiber.App, db *sqlx.DB) *UserRoute {
	handler := NewHandler(db)
	return &UserRoute{
		app:     app,
		db:      db,
		handler: handler,
	}
}

func (u *UserRoute) RegisterUserRoute() *UserRoute {
	v1 := u.app.Group("/api/v1/")
	user := v1.Group("/user")
	user.Get("/getloginsession/:login_session", u.handler.GetLoginSession)
	user.Get("/info", u.handler.GetUserBasicInfo)
	user.Get("/", u.handler.Show)
	user.Get("/:id", u.handler.ShowOne)
	user.Post("/", u.handler.Create)
	user.Put("/:id", u.handler.Update)
	user.Delete("/:id", u.handler.Delete)
	user.Get("/form/create", u.handler.GetUserFormCreate)
	user.Get("/form/update/:id", u.handler.GetUserFormUpdate)
	user.Put("/change/password/:id", u.handler.Update_Password)

	return u
}
