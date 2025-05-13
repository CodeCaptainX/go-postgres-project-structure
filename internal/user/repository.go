package user

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	custom_log "snack-shop/pkg/logs"
	types "snack-shop/pkg/model"
	"snack-shop/pkg/postgres"
	"snack-shop/pkg/responses"
	utils "snack-shop/pkg/utils"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UserRepo interface {
	GetLoginSession(login_session string) (bool, *responses.ErrorResponse)
	Show(userShowRequest UserShowRequest) (*UserResponse, *responses.ErrorResponse)
	ShowOne(user_uuid uuid.UUID) (*UserResponse, *responses.ErrorResponse)
	Create(usreq UserNewRequest) (*UserResponse, *responses.ErrorResponse)
	Update(user_uuid uuid.UUID, usreq UserUpdateRequest) (*UserResponse, *responses.ErrorResponse)
	Delete(user_uuid uuid.UUID) (*UserDeleteResponse, *responses.ErrorResponse)
	GetUserFormCreate() (*UserFormCreateResponse, *responses.ErrorResponse)
	GetUserFormUpdate(user_uuid uuid.UUID) (*UserFormUpdateResponse, *responses.ErrorResponse)
	Update_Password(user_uuid uuid.UUID, usreq UserUpdatePasswordRequest) (*UserUpdatePasswordReponse, *responses.ErrorResponse)
	GetUserBasicInfo(username string) (*UserBasicInfoResponse, *responses.ErrorResponse)
}

type UserRepoImpl struct {
	userCtx *types.UserContext
	db      *sqlx.DB
}

func NewUserRepoImpl(u *types.UserContext, db *sqlx.DB) *UserRepoImpl {
	return &UserRepoImpl{
		userCtx: u,
		db:      db,
	}
}

func (u *UserRepoImpl) GetLoginSession(login_session string) (bool, *responses.ErrorResponse) {
	// TODO: Implement subscription/notification pattern with a standard SQL database
	// This would likely involve polling, websockets, or a pub/sub mechanism

	smg_error := fmt.Errorf("invalid login session")
	custom_log.NewCustomLog("login_failed", smg_error.Error())
	err_resp := &responses.ErrorResponse{}
	return true, err_resp.NewErrorResponse("login_failed", smg_error)
}

// Test URL endpoint: {{ _.host }}/api/v1/admin/user?paging_options[page]=1&paging_options[per_page]=10&sorts[0][property]=u.id&sorts[0][direction]=desc&sorts[1][property]=u.user_name&sorts[1][direction]=desc&filters[0][property]=u.status_id&filters[0][value]=1
func (u *UserRepoImpl) Show(userShowRequest UserShowRequest) (*UserResponse, *responses.ErrorResponse) {
	var per_page = userShowRequest.PageOptions.Perpage
	var page = userShowRequest.PageOptions.Page
	var offset = (page - 1) * per_page
	var sql_limit = fmt.Sprintf(" LIMIT %d OFFSET %d", per_page, offset)

	// Order By output will be: `ORDER BY u.id asc, u.user_name desc`
	var sql_orderby = postgres.BuildSQLSort(userShowRequest.Sorts)

	// Filters output of BuildSQLFilter() will be e.g. tarantool.NewExecuteRequest("WHERE.. AND u.status_id=$1").Args([1])
	sql_filters, args_filters := postgres.BuildSQLFilter(userShowRequest.Filters)

	if len(args_filters) > 0 {
		sql_filters = " AND " + sql_filters
	}

	// SQL query with parameters
	query := fmt.Sprintf(`
		SELECT 
			u.id, 
			u.user_uuid, 
			u.first_name, 
			u.last_name, 
			u.user_name, 
			u.email, 
			u.role_id, 
			ur.user_role_name AS role_name, 
			u.status, 
			u.login_session, 
			u.profile_photo, 
			u.user_alias, 
			u.phone_number, 
			u.user_avatar_id, 
			u.commission, 
			u.status_id, 
			u.order,            
			u.created_by, 
			creator.user_name AS creator, 
			u.created_at, 
			u.updated_by, 
			u.updated_at, 
			u.deleted_by, 
			u.deleted_at
		FROM 
			users_space u
		INNER JOIN 
			users_roles_space ur 
		ON  
			u.role_id = ur.id
		LEFT JOIN 
			users_space creator
		ON 
			u.created_by = creator.id
		WHERE 
			%s%s%s`, sql_filters, sql_orderby, sql_limit)

	// Count query for pagination
	countQuery := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total
		FROM 
			users_space u
		WHERE %s`, sql_filters)

	var users []User
	err := u.db.Select(&users, query, args_filters...)
	if err != nil {
		custom_log.NewCustomLog("user_show_failed", err.Error(), "error")
		err_msg := &responses.ErrorResponse{}
		return nil, err_msg.NewErrorResponse("user_show_failed", fmt.Errorf("cannot select user: database error"))
	}

	var totalCount int
	err = u.db.Get(&totalCount, countQuery, args_filters[:len(args_filters)-len(userShowRequest.Filters)]...)
	if err != nil {
		custom_log.NewCustomLog("user_show_failed", err.Error(), "error")
		err_msg := &responses.ErrorResponse{}
		return nil, err_msg.NewErrorResponse("user_show_failed", fmt.Errorf("cannot get total count: database error"))
	}

	return &UserResponse{Users: users, Total: totalCount}, nil
}

func (u *UserRepoImpl) ShowOne(user_uuid uuid.UUID) (*UserResponse, *responses.ErrorResponse) {
	query := `
		SELECT 
			u.id, 
			u.user_uuid, 
			u.first_name, 
			u.last_name, 
			u.user_name, 
			u.email, 
			u.role_id, 
			ur.user_role_name AS role_name, 
			u.status, 
			u.login_session, 
			u.profile_photo, 
			u.user_alias, 
			u.phone_number, 
			u.user_avatar_id, 
			u.commission, 
			u.status_id, 
			u.order,            
			u.created_by, 
			creator.user_name AS creator, 
			u.created_at, 
			u.updated_by, 
			u.updated_at, 
			u.deleted_by, 
			u.deleted_at
		FROM 
			users_space u
		INNER JOIN 
			users_roles_space ur 
		ON  
			u.role_id = ur.id
		LEFT JOIN 
			users_space creator
		ON 
			u.created_by = creator.id
		WHERE 
			u.deleted_at IS NULL AND u.user_uuid = ?`

	var users []User
	err := u.db.Select(&users, query, user_uuid)
	if err != nil {
		custom_log.NewCustomLog("user_showone_failed", err.Error(), "error")
		err_msg := &responses.ErrorResponse{}
		return nil, err_msg.NewErrorResponse("user_showone_failed", fmt.Errorf("cannot select user: database error"))
	}

	return &UserResponse{Users: users, Total: 0}, nil
}

func (u *UserRepoImpl) Create(usreq UserNewRequest) (*UserResponse, *responses.ErrorResponse) {
	userAddModel := &UserAddModel{}

	// Begin transaction
	tx, err := u.db.BeginTxx(context.Background(), &sql.TxOptions{})
	if err != nil {
		custom_log.NewCustomLog("user_create_failed", err.Error(), "error")
		err_msg := &responses.ErrorResponse{}
		return nil, err_msg.NewErrorResponse("user_create_failed", fmt.Errorf("cannot begin transaction"))
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Initialize user model - modify the New function to accept sqlx.Tx instead of Tarantool stream
	err = userAddModel.New(usreq, u.userCtx, tx)
	if err != nil {
		custom_log.NewCustomLog("user_create_failed", err.Error())
		err_resp := &responses.ErrorResponse{}
		return nil, err_resp.NewErrorResponse("user_create_failed", err)
	}

	// Insert query
	query := `
		INSERT INTO users_space (
			id, user_uuid, first_name, last_name, user_name, profile_photo, user_alias, 
			password, email, role_id, status, login_session, phone_number, commission, 
			"order", created_by, created_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)`

	_, err = tx.Exec(query,
		userAddModel.ID,
		userAddModel.UserUUID,
		userAddModel.FirstName,
		userAddModel.LastName,
		userAddModel.UserName,
		userAddModel.ProfilePhoto,
		userAddModel.UserAlias,
		userAddModel.Password,
		userAddModel.Email,
		userAddModel.RoleId,
		userAddModel.Status,
		userAddModel.LoginSession,
		userAddModel.PhoneNumber,
		userAddModel.Commission,
		userAddModel.Order,
		userAddModel.CreatedBy,
		userAddModel.CreatedAt,
	)

	if err != nil {
		custom_log.NewCustomLog("user_create_failed", err.Error(), "error")
		err_resp := &responses.ErrorResponse{}
		return nil, err_resp.NewErrorResponse("user_create_failed", err)
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		custom_log.NewCustomLog("user_create_failed", err.Error(), "error")
		err_resp := &responses.ErrorResponse{}
		return nil, err_resp.NewErrorResponse("user_create_failed", fmt.Errorf("cannot commit transaction"))
	}

	// Add Audit - Update this function to use sqlx
	var audit_des = fmt.Sprintf("New user `%s` has been created", userAddModel.UserName)
	_, err = utils.AddUserAuditLog(
		int(userAddModel.ID), "New User", audit_des, 1, u.userCtx.UserAgent,
		u.userCtx.UserName, u.userCtx.Ip, int(userAddModel.CreatedBy), u.db)
	if err != nil {
		custom_log.NewCustomLog("user_create_failed", err.Error(), "warn")
		// Audit failures are not critical, so we don't return an error
	}

	return u.ShowOne(userAddModel.UserUUID)
}

func (u *UserRepoImpl) Update(user_uuid uuid.UUID, usreq UserUpdateRequest) (*UserResponse, *responses.ErrorResponse) {
	userUpdateModel := &UserUpdateModel{}

	// Begin transaction
	tx, err := u.db.BeginTxx(context.Background(), &sql.TxOptions{})
	if err != nil {
		custom_log.NewCustomLog("user_update_failed", err.Error(), "error")
		err_msg := &responses.ErrorResponse{}
		return nil, err_msg.NewErrorResponse("user_update_failed", fmt.Errorf("cannot begin transaction"))
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Initialize user model - modify the New function to accept sqlx.Tx instead of Tarantool stream
	err = userUpdateModel.New(user_uuid, usreq, u.userCtx, tx)
	if err != nil {
		custom_log.NewCustomLog("user_update_failed", err.Error())
		err_resp := &responses.ErrorResponse{}
		return nil, err_resp.NewErrorResponse("user_update_failed", err)
	}

	// Update query
	query := `
		UPDATE users_space SET
			first_name = ?, 
			last_name = ?, 
			email = ?,
			role_id = ?, 
			status_id = ?, 
			phone_number = ?, 
			commission = ?, 
			updated_by = ?, 
			updated_at = ?
		WHERE user_uuid = ?`

	_, err = tx.Exec(query,
		userUpdateModel.FirstName,
		userUpdateModel.LastName,
		userUpdateModel.Email,
		userUpdateModel.RoleId,
		userUpdateModel.StatusId,
		userUpdateModel.PhoneNumber,
		userUpdateModel.Commission,
		userUpdateModel.UpdatedBy,
		userUpdateModel.UpdatedAt,
		userUpdateModel.UserUUID,
	)

	if err != nil {
		custom_log.NewCustomLog("user_update_failed", err.Error(), "error")
		err_resp := &responses.ErrorResponse{}
		return nil, err_resp.NewErrorResponse("user_update_failed", fmt.Errorf("cannot execute update"))
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		custom_log.NewCustomLog("user_update_failed", err.Error(), "error")
		err_resp := &responses.ErrorResponse{}
		return nil, err_resp.NewErrorResponse("user_update_failed", fmt.Errorf("cannot commit transaction"))
	}

	// Add Audit
	var audit_des = fmt.Sprintf("Updating user `%s %s` has been successful", userUpdateModel.FirstName, userUpdateModel.LastName)
	_, err = utils.AddUserAuditLog(
		int(userUpdateModel.ID), "Update User", audit_des, 1, u.userCtx.UserAgent,
		u.userCtx.UserName, u.userCtx.Ip, int(userUpdateModel.UpdatedBy), u.db)
	if err != nil {
		custom_log.NewCustomLog("user_update_failed", err.Error(), "warn")
		// Audit failures are not critical, so we don't return an error
	}

	return u.ShowOne(userUpdateModel.UserUUID)
}

func (u *UserRepoImpl) Delete(user_uuid uuid.UUID) (*UserDeleteResponse, *responses.ErrorResponse) {
	// Check permission (admin can't delete users with equal or higher roles)
	if u.userCtx.RoleId != 1 {
		var exists bool
		err := u.db.Get(&exists, `
			SELECT EXISTS(
				SELECT 1 FROM users_space 
				WHERE role_id <= ? AND user_uuid = ? AND deleted_at IS NULL
			)`, u.userCtx.RoleId, user_uuid)

		if err != nil {
			custom_log.NewCustomLog("user_delete_failed", err.Error(), "error")
			err_resp := &responses.ErrorResponse{}
			return nil, err_resp.NewErrorResponse("user_delete_failed", fmt.Errorf("failed to check permissions"))
		}

		if exists {
			custom_log.NewCustomLog("user_delete_failed", "permission denied", "error")
			err_resp := &responses.ErrorResponse{}
			return nil, err_resp.NewErrorResponse("user_delete_failed", fmt.Errorf("permission denied: this user has the same or higher role than you"))
		}
	}

	// Get timestamp for soft delete
	app_timezone := os.Getenv("APP_TIMEZONE")
	location, err := time.LoadLocation(app_timezone)
	if err != nil {
		custom_log.NewCustomLog("user_delete_failed", fmt.Errorf("failed to load location: %w", err).Error())
		err_resp := &responses.ErrorResponse{}
		return nil, err_resp.NewErrorResponse("user_delete_failed", fmt.Errorf("cannot delete user"))
	}
	now := time.Now().In(location)

	// Get current user ID
	var by_id int64
	err = u.db.Get(&by_id, "SELECT id FROM users_space WHERE user_uuid = ?", u.userCtx.UserUuid)
	if err != nil {
		custom_log.NewCustomLog("user_delete_failed", err.Error())
		err_resp := &responses.ErrorResponse{}
		return nil, err_resp.NewErrorResponse("user_delete_failed", fmt.Errorf("cannot get current user ID"))
	}

	// Get target user info before deletion
	users, err_one := u.ShowOne(user_uuid)
	if err_one != nil {
		custom_log.NewCustomLog("user_delete_failed", err_one.Err.Error(), "error")
		err_resp := &responses.ErrorResponse{}
		return nil, err_resp.NewErrorResponse("user_delete_failed", fmt.Errorf("user to delete not found"))
	} else if len(users.Users) <= 0 {
		custom_log.NewCustomLog("user_delete_failed", "Cannot get info of user to delete", "error")
		err_resp := &responses.ErrorResponse{}
		return nil, err_resp.NewErrorResponse("user_delete_failed", fmt.Errorf("user to delete not found"))
	}

	// Begin transaction
	tx, err := u.db.BeginTxx(context.Background(), &sql.TxOptions{})
	if err != nil {
		custom_log.NewCustomLog("user_delete_failed", err.Error(), "error")
		err_resp := &responses.ErrorResponse{}
		return nil, err_resp.NewErrorResponse("user_delete_failed", fmt.Errorf("cannot begin transaction"))
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Soft delete user
	_, err = tx.Exec(`
		UPDATE users_space SET
			status_id = ?, 
			deleted_by = ?, 
			deleted_at = ?, 
			updated_by = ?, 
			updated_at = ?
		WHERE user_uuid = ?`,
		0, by_id, now, by_id, now, user_uuid)

	if err != nil {
		custom_log.NewCustomLog("user_delete_failed", err.Error(), "error")
		err_resp := &responses.ErrorResponse{}
		return nil, err_resp.NewErrorResponse("user_delete_failed", fmt.Errorf("cannot delete user"))
	}

	// Update login session
	login_session, _ := uuid.NewV7()
	_, err = tx.Exec("UPDATE users_space SET login_session = ? WHERE user_uuid = ?",
		login_session.String(), user_uuid)
	if err != nil {
		log.Println("Error updating session:", err.Error())
		// Continue even if session update fails
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		custom_log.NewCustomLog("user_delete_failed", err.Error(), "error")
		err_resp := &responses.ErrorResponse{}
		return nil, err_resp.NewErrorResponse("user_delete_failed", fmt.Errorf("cannot commit transaction"))
	}

	// Add Audit
	var audit_des = fmt.Sprintf("Deleting user `%s %s` has been successful", users.Users[0].FirstName, users.Users[0].LastName)
	_, err = utils.AddUserAuditLog(
		int(users.Users[0].ID), "Delete User", audit_des, 1, u.userCtx.UserAgent,
		u.userCtx.UserName, u.userCtx.Ip, int(by_id), u.db)
	if err != nil {
		custom_log.NewCustomLog("user_delete_failed", err.Error(), "warn")
		// Non-critical error, continue
	}

	return &UserDeleteResponse{Success: true}, nil
}

func (u *UserRepoImpl) GetStatus() *[]types.Status {
	return &types.StatusData
}

func (u *UserRepoImpl) GetRoles() (*[]Role, error) {
	query := "SELECT id, user_role_name FROM users_roles_space WHERE deleted_at IS NULL"
	var args []interface{}

	if u.userCtx.RoleId == 1 {
		query += " AND id >= ?"
	} else {
		query += " AND id > ?"
	}
	query += " ORDER BY user_role_name ASC"
	args = append(args, u.userCtx.RoleId)

	var roles []Role
	err := u.db.Select(&roles, query, args...)
	if err != nil {
		return nil, err
	}
	return &roles, nil
}

func (u *UserRepoImpl) GetUserFormCreate() (*UserFormCreateResponse, *responses.ErrorResponse) {
	status := u.GetStatus()
	roles, err := u.GetRoles()
	if err != nil {
		custom_log.NewCustomLog("user_create_form_failed", err.Error(), "error")
		err_resp := &responses.ErrorResponse{}
		return nil, err_resp.NewErrorResponse("user_create_form_failed", fmt.Errorf("cannot get roles"))
	}

	var userCreateForms = []UserCreateForm{
		{
			FirstName:       "",
			LastName:        "",
			UserName:        "",
			Password:        "",
			PasswordComfirm: "",
			Email:           "",
			RoleId:          1,
			PhoneNumber:     "",
			StatusId:        1,
			Status:          *status,
			Roles:           *roles,
		},
	}
	return &UserFormCreateResponse{Users: userCreateForms}, nil
}

func (u *UserRepoImpl) GetUserFormUpdate(user_uuid uuid.UUID) (*UserFormUpdateResponse, *responses.ErrorResponse) {
	// Check permission (admin can't update users with equal or higher roles)
	if u.userCtx.RoleId != 1 {
		var exists bool
		err := u.db.Get(&exists, `
			SELECT EXISTS(
				SELECT 1 FROM users_space 
				WHERE role_id <= ? AND user_uuid = ? AND deleted_at IS NULL
			)`, u.userCtx.RoleId, user_uuid)

		if err != nil {
			custom_log.NewCustomLog("user_update_form_failed", err.Error(), "error")
			err_resp := &responses.ErrorResponse{}
			return nil, err_resp.NewErrorResponse("user_update_form_failed", fmt.Errorf("failed to check permissions"))
		}

		if exists {
			custom_log.NewCustomLog("user_update_form_failed", "permission denied", "error")
			err_resp := &responses.ErrorResponse{}
			return nil, err_resp.NewErrorResponse("user_update_form_failed", fmt.Errorf("permission denied: this user has the same or higher role than you"))
		}
	}

	// Get user info
	users, err_one := u.ShowOne(user_uuid)
	if err_one != nil {
		custom_log.NewCustomLog("user_update_form_failed", err_one.Err.Error(), "error")
		err_resp := &responses.ErrorResponse{}
		return nil, err_resp.NewErrorResponse("user_update_form_failed", fmt.Errorf("failed to get user info"))
	} else if len(users.Users) <= 0 {
		custom_log.NewCustomLog("user_update_form_failed", "Cannot get user info", "error")
		err_resp := &responses.ErrorResponse{}
		return nil, err_resp.NewErrorResponse("user_update_form_failed", fmt.Errorf("user not found"))
	}

	status := u.GetStatus()
	roles, err := u.GetRoles()
	if err != nil {
		custom_log.NewCustomLog("user_update_form_failed", err.Error(), "error")
		err_resp := &responses.ErrorResponse{}
		return nil, err_resp.NewErrorResponse("user_update_form_failed", fmt.Errorf("cannot get roles"))
	}

	var userUpdateForms = []UserUpdateForm{
		{
			FirstName:   users.Users[0].FirstName,
			LastName:    users.Users[0].LastName,
			UserName:    users.Users[0].UserName,
			Email:       users.Users[0].Email,
			RoleId:      users.Users[0].RoleId,
			PhoneNumber: *users.Users[0].PhoneNumber,
			StatusId:    users.Users[0].StatusId,
			Commission:  users.Users[0].Commission,
			Status:      *status,
			Roles:       *roles,
		},
	}
	return &UserFormUpdateResponse{Users: userUpdateForms}, nil
}

func (u *UserRepoImpl) Update_Password(user_uuid uuid.UUID, usreq UserUpdatePasswordRequest) (*UserUpdatePasswordReponse, *responses.ErrorResponse) {
	var RequestChangePassword = &UserUpdatePasswordModel{}

	// Begin transaction
	tx, err := u.db.BeginTxx(context.Background(), &sql.TxOptions{})
	if err != nil {
		custom_log.NewCustomLog("user_update_password_failed", err.Error(), "error")
		err_msg := &responses.ErrorResponse{}
		return nil, err_msg.NewErrorResponse("user_update_password_failed", fmt.Errorf("cannot begin transaction"))
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Initialize password model
	err = RequestChangePassword.New(user_uuid, usreq, u.userCtx, tx)
	if err != nil {
		custom_log.NewCustomLog("user_update_password_failed", err.Error())
		err_resp := &responses.ErrorResponse{}
		return nil, err_resp.NewErrorResponse("user_update_password_failed", err)
	}

	// Get current user ID
	var by_id int64
	err = u.db.Get(&by_id, "SELECT id FROM users_space WHERE user_uuid = ?", u.userCtx.UserUuid)
	if err != nil {
		custom_log.NewCustomLog("user_update_password_failed", err.Error())
		err_resp := &responses.ErrorResponse{}
		return nil, err_resp.NewErrorResponse("user_update_password_failed", fmt.Errorf("cannot get current user ID"))
	}

	// Get target user info
	users, err_one := u.ShowOne(user_uuid)
	if err_one != nil {
		custom_log.NewCustomLog("user_update_password_failed", err_one.Err.Error(), "error")
		err_resp := &responses.ErrorResponse{}
		return nil, err_resp.NewErrorResponse("user_update_password_failed", fmt.Errorf("user not found"))
	} else if len(users.Users) <= 0 {
		custom_log.NewCustomLog("user_update_password_failed", "Cannot get user info", "error")
		err_resp := &responses.ErrorResponse{}
		return nil, err_resp.NewErrorResponse("user_update_password_failed", fmt.Errorf("user not found"))
	}

	// Update password
	_, err = tx.Exec(`
		UPDATE users_space SET
			password = ?, 
			updated_by = ?, 
			updated_at = ?
		WHERE user_uuid = ?`,
		RequestChangePassword.Password,
		RequestChangePassword.UpdatedBy,
		RequestChangePassword.UpdatedAt,
		RequestChangePassword.UserUUID)

	if err != nil {
		custom_log.NewCustomLog("user_update_password_failed", err.Error(), "error")
		err_resp := &responses.ErrorResponse{}
		return nil, err_resp.NewErrorResponse("user_update_password_failed", fmt.Errorf("cannot update password"))
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		custom_log.NewCustomLog("user_update_password_failed", err.Error(), "error")
		err_resp := &responses.ErrorResponse{}
		return nil, err_resp.NewErrorResponse("user_update_password_failed", fmt.Errorf("cannot commit transaction"))
	}

	// Add Audit
	var audit_des = fmt.Sprintf("Updating `%s %s`'s password has been successful", users.Users[0].FirstName, users.Users[0].LastName)
	_, err = utils.AddUserAuditLog(
		int(users.Users[0].ID), "Update User's password", audit_des, 1, u.userCtx.UserAgent,
		u.userCtx.UserName, u.userCtx.Ip, int(by_id), u.db)
	if err != nil {
		custom_log.NewCustomLog("user_update_password_failed", err.Error(), "warn")
		// Non-critical error, continue
	}

	// // Add notification
	// var notificationContext = "Password Update"
	// var notificationSubject = "Password Changed"
	// var notificationDescription = "Your account password has been successfully updated."

	// // Add the notification to the user's account
	// err = utils.AddNotification("users_notifications_space", "user", int(users.Users[0].ID),
	// 	notificationContext, notificationSubject, notificationDescription, 1, 1, by_id, u.db)

	// if err != nil {
	// 	fmt.Println("failed to add password update notification")
	// 	// Non-critical error, continue
	// }

	return &UserUpdatePasswordReponse{Success: true}, nil
}
func (u *UserRepoImpl) GetUserBasicInfo(username string) (*UserBasicInfoResponse, *responses.ErrorResponse) {
	var userInfo UserInfo

	query := `
		SELECT 
			u.id, u.user_uuid, u.first_name, u.last_name, u.user_name, u.email, 
			u.role_id, ur.user_role_name AS role_name, u.status, u.login_session, u.profile_photo, 
			u.user_alias, u.phone_number, u.user_avatar_id, u.commission, u.status_id
		FROM users_space u
		INNER JOIN users_roles_space ur ON u.role_id = ur.id
		WHERE u.deleted_at IS NULL AND ur.deleted_at IS NULL
		AND u.user_name = $1
	`

	err := u.db.Get(&userInfo, query, username)
	if err != nil {
		custom_log.NewCustomLog("get_userinfo_failed", err.Error(), "error")
		errResp := &responses.ErrorResponse{}
		return nil, errResp.NewErrorResponse("get_userinfo_failed", fmt.Errorf("cannot select user: %w", err))
	}

	// Get permissions
	var permissions []Permission
	permQuery := `
		SELECT
			m.module_name,
			rm.function_ids
		FROM rel_roles_modules_space rm
		INNER JOIN modules_space m ON rm.module_id = m.id
		WHERE rm.deleted_at IS NULL AND rm.role_id = $1
	`

	err = u.db.Select(&permissions, permQuery, userInfo.RoleId)
	if err != nil {
		custom_log.NewCustomLog("get_userinfo_failed", err.Error(), "warn")
		errMsg := &responses.ErrorResponse{}
		return nil, errMsg.NewErrorResponse("get_userinfo_failed", fmt.Errorf("cannot get user permissions: %w", err))
	}

	// Process permissions
	userPermission := UserPermission{
		Modules: make(map[string][]string),
	}

	for _, perm := range permissions {
		functionIDs := strings.Split(perm.FunctionIDs, ",")
		userPermission.Modules[perm.ModuleName] = functionIDs
	}

	return &UserBasicInfoResponse{
		UserBasicInfo: UserBasicInfo{
			UserInfo:       userInfo,
			UserPermission: userPermission,
		},
	}, nil
}
