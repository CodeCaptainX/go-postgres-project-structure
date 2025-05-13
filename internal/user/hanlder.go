package user

import (
	"fmt"
	"net/http"
	"snack-shop/pkg/constants"
	custom_log "snack-shop/pkg/logs"
	types "snack-shop/pkg/model"
	"snack-shop/pkg/utils"

	response "snack-shop/pkg/http/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// UserHandler struct
type UserHandler struct {
	db          *sqlx.DB
	userService func(*fiber.Ctx) UserCreator
}

func NewHandler(db *sqlx.DB) *UserHandler {
	return &UserHandler{
		db: db,
		userService: func(c *fiber.Ctx) UserCreator {
			userContext := c.Locals("UserContext")
			// fmt.Println("xxxxx", userContext)
			var uCtx types.UserContext
			// Convert map to UserContext struct
			if contextMap, ok := userContext.(types.UserContext); ok {
				uCtx = contextMap
			} else {
				custom_log.NewCustomLog("user_context_failed", "Failed to cast UserContext to map[string]interface{}", "warn")
				uCtx = types.UserContext{}
			}

			// Pass uCtx to NewAuthService if needed
			return NewUserService(&uCtx, db)
		},
	}
}

// GetLoginSession godoc
// @Summary Get GetLoginSession
// @Description Get user login session
// @Tags Admin/User
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/admin/v1/user/profile [get]
func (h *UserHandler) GetLoginSession(c *fiber.Ctx) error {
	login_session := c.Params("login_session")
	fmt.Println("login_session", login_session)
	// Your profile logic here
	as := h.userService(c)
	_, err := as.GetLoginSession(login_session)

	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(response.NewResponseError(
			utils.Translate("getLoginSessionSuccess", nil, c),
			constants.UserGetLoginSessionFailed,
			err.Err,
		))
	} else {
		return c.Status(http.StatusOK).JSON(response.NewResponse(
			utils.Translate("getLoginSessionSuccess", nil, c),
			constants.UserGetLoginSessionSuccess,
			login_session,
		))
	}
}

func (h *UserHandler) Show(c *fiber.Ctx) error {
	var userRequest UserShowRequest

	//Bind and validate
	v := utils.NewValidator()
	if err := userRequest.bind(c, v); err != nil {
		return c.Status(http.StatusBadRequest).JSON(
			response.NewResponseError(
				utils.Translate("user_show_failed", nil, c),
				constants.UserShowFailed,
				err,
			),
		)
	}

	// Debugging output to see if the struct is populated correctly
	// fmt.Println("Parsed Request:", userRequest)

	as := h.userService(c)
	users, err := as.Show(userRequest)

	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(response.NewResponseError(
			utils.Translate(err.MessageID, nil, c),
			constants.UserShowFailed,
			err.Err,
		))
	} else {
		return c.Status(http.StatusOK).JSON(response.NewResponseWithPaging(
			utils.Translate("user_show_failed", nil, c),
			constants.UserShowSuccess,
			users,
			1,
			10,
			users.Total,
		))
	}
}

func (h *UserHandler) ShowOne(c *fiber.Ctx) error {
	// Extract the "id" parameter from the URL
	idStr := c.Params("id", "")

	// Parse the UUID string (you can use google/uuid or another library that supports UUID v7)
	id, err_uuid := uuid.Parse(idStr)
	if err_uuid != nil {
		return c.Status(http.StatusBadRequest).JSON(response.NewResponseError(
			utils.Translate("user_show_failed", nil, c),
			-2000,
			err_uuid,
		))
	}

	as := h.userService(c)
	users, err := as.ShowOne(id)

	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(response.NewResponseError(
			utils.Translate(err.MessageID, nil, c),
			-2000,
			err.Err,
		))
	} else {
		return c.Status(http.StatusOK).JSON(response.NewResponse(
			utils.Translate("user_show_failed", nil, c),
			2000,
			users,
		))
	}
}

func (h *UserHandler) Create(c *fiber.Ctx) error {
	var userNewRequest UserNewRequest

	//Bind and validate
	v := utils.NewValidator()
	if err := userNewRequest.bind(c, v); err != nil {
		return c.Status(http.StatusBadRequest).JSON(
			response.NewResponseError(
				utils.Translate("user_create_failed", nil, c),
				-1000,
				err,
			),
		)
	}

	as := h.userService(c)
	users, err := as.Create(userNewRequest)

	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(response.NewResponseError(
			utils.Translate(err.MessageID, nil, c),
			-3000,
			err.Err,
		))
	} else {
		return c.Status(http.StatusOK).JSON(response.NewResponse(
			utils.Translate("user_create_success", nil, c),
			3000,
			users,
		))
	}
}

func (h *UserHandler) Update(c *fiber.Ctx) error {
	// Extract the "id" parameter from the URL
	idStr := c.Params("id", "")

	// Parse the UUID string (you can use google/uuid or another library that supports UUID v7)
	id, err_uuid := uuid.Parse(idStr)
	if err_uuid != nil {
		return c.Status(http.StatusUnauthorized).JSON(response.NewResponseError(
			utils.Translate("user_update_failed", nil, c),
			-2000,
			err_uuid,
		))
	}

	var userUpdateRequest UserUpdateRequest

	//Bind and validate
	v := utils.NewValidator()
	if err := userUpdateRequest.bind(c, v); err != nil {
		return c.Status(http.StatusUnprocessableEntity).JSON(
			response.NewResponseError(
				utils.Translate("user_update_failed", nil, c),
				-1000,
				err,
			),
		)
	}

	as := h.userService(c)
	users, err := as.Update(id, userUpdateRequest)

	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(response.NewResponseError(
			utils.Translate(err.MessageID, nil, c),
			-3001,
			err.Err,
		))
	} else {
		return c.Status(http.StatusOK).JSON(response.NewResponse(
			utils.Translate("user_update_success", nil, c),
			3001,
			users,
		))
	}
}

func (h *UserHandler) Delete(c *fiber.Ctx) error {
	// Extract the "id" parameter from the URL
	idStr := c.Params("id", "")

	// Parse the UUID string (you can use google/uuid or another library that supports UUID v7)
	user_uuid, err_uuid := uuid.Parse(idStr)
	if err_uuid != nil {
		return c.Status(http.StatusUnauthorized).JSON(response.NewResponseError(
			utils.Translate("user_delete_failed", nil, c),
			-2000,
			err_uuid,
		))
	}

	as := h.userService(c)
	users, err := as.Delete(user_uuid)

	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(response.NewResponseError(
			utils.Translate(err.MessageID, nil, c),
			-4001,
			err.Err,
		))
	} else {
		return c.Status(http.StatusOK).JSON(response.NewResponse(
			utils.Translate("user_delete_success", nil, c),
			4001,
			users,
		))
	}

}

func (h *UserHandler) GetUserFormCreate(c *fiber.Ctx) error {

	as := h.userService(c)
	users, err := as.GetUserFormCreate()

	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(response.NewResponseError(
			utils.Translate(err.MessageID, nil, c),
			-5000,
			err.Err,
		))
	} else {
		return c.Status(http.StatusOK).JSON(response.NewResponse(
			utils.Translate("user_create_form_success", nil, c),
			5000,
			users,
		))
	}
}

func (h *UserHandler) GetUserFormUpdate(c *fiber.Ctx) error {
	// Extract the "id" parameter from the URL
	idStr := c.Params("id", "")

	// Parse the UUID string (you can use google/uuid or another library that supports UUID v7)
	user_uuid, err_uuid := uuid.Parse(idStr)
	if err_uuid != nil {
		return c.Status(http.StatusBadRequest).JSON(response.NewResponseError(
			utils.Translate("user_update_form_failed", nil, c),
			-2000,
			err_uuid,
		))
	}
	as := h.userService(c)
	users, err := as.GetUserFormUpdate(user_uuid)

	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(response.NewResponseError(
			utils.Translate(err.MessageID, nil, c),
			-5000,
			err.Err,
		))
	} else {
		return c.Status(http.StatusOK).JSON(response.NewResponse(
			utils.Translate("user_update_form_success", nil, c),
			5000,
			users,
		))
	}
}

func (h *UserHandler) Update_Password(c *fiber.Ctx) error {
	// Extract the "id" parameter from the URL
	idStr := c.Params("id", "")

	// Parse the UUID string (you can use google/uuid or another library that supports UUID v7)
	id, err_uuid := uuid.Parse(idStr)
	fmt.Println(id)
	if err_uuid != nil {
		return c.Status(http.StatusBadRequest).JSON(response.NewResponseError(
			utils.Translate("user_update_password_failed", nil, c),
			-2000,
			err_uuid,
		))
	}
	var userUpdatePasswordRequest UserUpdatePasswordRequest

	//Bind and validate
	v := utils.NewValidator()
	if err := userUpdatePasswordRequest.bind(c, v); err != nil {
		return c.Status(http.StatusBadRequest).JSON(
			response.NewResponseError(
				utils.Translate("user_update_password_failed", nil, c),
				-1000,
				err,
			),
		)
	}

	as := h.userService(c)
	users, err := as.Update_Password(id, userUpdatePasswordRequest)

	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(response.NewResponseError(
			utils.Translate(err.MessageID, nil, c),
			-3001,
			err.Err,
		))
	} else {
		return c.Status(http.StatusOK).JSON(response.NewResponse(
			utils.Translate("user_update_password_success", nil, c),
			3001,
			users,
		))
	}
}

func (h *UserHandler) GetUserBasicInfo(c *fiber.Ctx) error {
	user_resp, err := h.userService(c).GetUserBasicInfo()
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(
			response.NewResponseError(
				utils.Translate(err.MessageID, nil, c),
				-2001,
				err.Err,
			))
	}

	return c.Status(http.StatusOK).JSON(
		response.NewResponse(
			utils.Translate("get_userinfo_success", nil, c),
			2001,
			user_resp,
		))
}
