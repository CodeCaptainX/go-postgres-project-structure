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

	return true, responses.NewErrorResponse("login_failed", smg_error)
}

// Test URL endpoint: {{ _.host }}/api/v1/admin/user?paging_options[page]=1&paging_options[per_page]=10&sorts[0][property]=u.id&sorts[0][direction]=desc&sorts[1][property]=u.user_name&sorts[1][direction]=desc&filters[0][property]=u.status_id&filters[0][value]=1
func (u *UserRepoImpl) Show(userShowRequest UserShowRequest) (*UserResponse, *responses.ErrorResponse) {
	perPage := userShowRequest.PageOptions.Perpage
	page := userShowRequest.PageOptions.Page
	offset := (page - 1) * perPage

	sqlLimit := fmt.Sprintf(" LIMIT %d OFFSET %d", perPage, offset)
	sqlOrderBy := postgres.BuildSQLSort(userShowRequest.Sorts)

	sqlFilters, argsFilters := postgres.BuildSQLFilter(userShowRequest.Filters)
	fmt.Println("🚀 ~ file: repository.go ~ line 67 ~ func ~ sqlFilters : ", sqlFilters)
	whereClause := "WHERE u.deleted_at IS NULL"

	if sqlFilters != "" {
		whereClause += " AND " + sqlFilters
	}

	// Construct SELECT query
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
			tbl_users u
		INNER JOIN 
			tbl_users_roles ur ON u.role_id = ur.id
		LEFT JOIN 
			tbl_users creator ON u.created_by = creator.id
		%s %s %s`, whereClause, sqlOrderBy, sqlLimit)

	fmt.Println("🚀 SQL Query:", query)
	fmt.Println("🚀 Args:", argsFilters)

	var users []User
	err := u.db.Select(&users, query, argsFilters...)
	if err != nil {
		custom_log.NewCustomLog("user_show_failed", err.Error(), "error")

		return nil, responses.NewErrorResponse("user_show_failed", fmt.Errorf("cannot select user: database error"))
	}

	// Count query for total records
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) as total
		FROM tbl_users u
		%s`, whereClause)

	fmt.Println("🚀 Count Query:", countQuery)

	var totalCount int
	err = u.db.Get(&totalCount, countQuery, argsFilters...)
	if err != nil {
		custom_log.NewCustomLog("user_show_failed", err.Error(), "error")

		return nil, responses.NewErrorResponse("user_show_failed", fmt.Errorf("cannot get total count: database error"))
	}

	return &UserResponse{
		Users: users,
		Total: totalCount,
	}, nil
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
			tbl_users u
		INNER JOIN 
			tbl_users_roles ur 
		ON  
			u.role_id = ur.id
		LEFT JOIN 
			tbl_users creator
		ON 
			u.created_by = creator.id
		WHERE 
			u.deleted_at IS NULL AND u.user_uuid = $1`

	fmt.Println("🚀 ~ file: repository.go ~ line 195 ~ func ~ hello : ")
	var users User
	err := u.db.Get(&users, query, user_uuid)
	if err != nil {
		custom_log.NewCustomLog("user_showone_failed", err.Error(), "error")

		return nil, responses.NewErrorResponse("user_showone_failed", fmt.Errorf("cannot select user: database error"))
	}

	return &UserResponse{Users: []User{users}, Total: 0}, nil
}

func (u *UserRepoImpl) Create(usreq UserNewRequest) (*UserResponse, *responses.ErrorResponse) {
	userAddModel := &UserAddModel{}

	// Begin transaction
	tx, err := u.db.BeginTxx(context.Background(), &sql.TxOptions{})
	if err != nil {
		custom_log.NewCustomLog("user_create_failed", err.Error(), "error")

		return nil, responses.NewErrorResponse("user_create_failed", fmt.Errorf("cannot begin transaction"))
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 	tx, err := u.db_pool.Beginx()
	// if err != nil {
	// 	logs.NewCustomLog("transaction_start_failed", err.Error(), "error")
	// 	return nil, &error_responses.ErrorResponse{
	// 		MessageID: "transaction_start_failed",
	// 		Err:       err,
	// 	}
	// }
	// defer func() {
	// 	if err != nil {
	// 		tx.Rollback()
	// 		return
	// 	}
	// 	tx.Commit()
	// }()

	// Initialize user model - modify the New function to accept sqlx.Tx instead of Tarantool stream
	err = userAddModel.New(usreq, u.userCtx, tx)
	if err != nil {
		custom_log.NewCustomLog("user_create_failed", err.Error())

		return nil, responses.NewErrorResponse("user_create_failed", err)
	}

	// Insert query
	query := `
		INSERT INTO tbl_users (
			id, user_uuid, first_name, last_name, user_name, profile_photo, user_alias, 
			password, email, role_id, status, login_session, phone_number, commission, 
			"order", created_by, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
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
		return nil, responses.NewErrorResponse("user_create_failed", err)
	}
	// Commit transaction
	err = tx.Commit()
	if err != nil {
		custom_log.NewCustomLog("user_create_failed", err.Error(), "error")
		return nil, responses.NewErrorResponse("user_create_failed", fmt.Errorf("cannot commit transaction"))
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

		return nil, responses.NewErrorResponse("user_update_failed", fmt.Errorf("cannot begin transaction"))
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

		return nil, responses.NewErrorResponse("user_update_failed", err)
	}

	// Update query - Using $1, $2, etc. for PostgreSQL
	query := `
		UPDATE tbl_users SET
			first_name = $1, 
			last_name = $2, 
			email = $3,
			role_id = $4, 
			status_id = $5, 
			phone_number = $6, 
			commission = $7, 
			updated_by = $8, 
			updated_at = $9
		WHERE user_uuid = $10`

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

		return nil, responses.NewErrorResponse("user_update_failed", fmt.Errorf("cannot execute update"))
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		custom_log.NewCustomLog("user_update_failed", err.Error(), "error")

		return nil, responses.NewErrorResponse("user_update_failed", fmt.Errorf("cannot commit transaction"))
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
	// if u.userCtx.RoleId != 1 {
	// 	var exists bool
	// 	err := u.db.Get(&exists, `
	// 		SELECT EXISTS(
	// 			SELECT 1 FROM tbl_users
	// 			WHERE role_id <= $1 AND user_uuid = $2 AND deleted_at IS NULL
	// 		)`, u.userCtx.RoleId, user_uuid)

	// 	if err != nil {
	// 		custom_log.NewCustomLog("user_delete_failed", err.Error(), "error")
	//
	// 		return nil, responses.NewErrorResponse("user_delete_failed", fmt.Errorf("failed to check permissions"))
	// 	}

	// 	if exists {
	// 		custom_log.NewCustomLog("user_delete_failed", "permission denied", "error")
	//
	// 		return nil, responses.NewErrorResponse("user_delete_failed", fmt.Errorf("permission denied: this user has the same or higher role than you"))
	// 	}
	// }

	// Get timestamp for soft delete
	app_timezone := os.Getenv("APP_TIMEZONE")
	location, err := time.LoadLocation(app_timezone)
	if err != nil {
		custom_log.NewCustomLog("user_delete_failed", fmt.Errorf("failed to load location: %w", err).Error())

		return nil, responses.NewErrorResponse("user_delete_failed", fmt.Errorf("cannot delete user"))
	}
	now := time.Now().In(location)

	// Get current user ID
	var by_id int64
	err = u.db.Get(&by_id, "SELECT id FROM tbl_users WHERE user_uuid = $1", u.userCtx.UserUuid)
	fmt.Println("🚀 ~ file: repository.go ~ line 406 ~ func ~ u.userCtx.UserUuid : ", u.userCtx.UserUuid)
	if err != nil {
		custom_log.NewCustomLog("user_delete_failed", err.Error())

		return nil, responses.NewErrorResponse("user_delete_failed", fmt.Errorf("cannot get current user ID"))
	}

	// Get target user info before deletion
	users, err_one := u.ShowOne(user_uuid)
	if err_one != nil {
		custom_log.NewCustomLog("user_delete_failed", err_one.Err.Error(), "error")

		return nil, responses.NewErrorResponse("user_delete_failed", fmt.Errorf("user to delete not found"))
	} else if len(users.Users) <= 0 {
		custom_log.NewCustomLog("user_delete_failed", "Cannot get info of user to delete", "error")

		return nil, responses.NewErrorResponse("user_delete_failed", fmt.Errorf("user to delete not found"))
	}

	// Begin transaction
	tx, err := u.db.BeginTxx(context.Background(), &sql.TxOptions{})
	if err != nil {
		custom_log.NewCustomLog("user_delete_failed", err.Error(), "error")

		return nil, responses.NewErrorResponse("user_delete_failed", fmt.Errorf("cannot begin transaction"))
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Soft delete user
	_, err = tx.Exec(`
		UPDATE tbl_users SET
			status_id = $1, 
			deleted_by = $2, 
			deleted_at = $3, 
			updated_by = $4, 
			updated_at = $5
		WHERE user_uuid = $6`,
		0, by_id, now, by_id, now, user_uuid)

	if err != nil {
		custom_log.NewCustomLog("user_delete_failed", err.Error(), "error")

		return nil, responses.NewErrorResponse("user_delete_failed", fmt.Errorf("cannot delete user"))
	}

	// Update login session
	login_session, _ := uuid.NewV7()
	_, err = tx.Exec("UPDATE tbl_users SET login_session = $1 WHERE user_uuid = $2",
		login_session.String(), user_uuid)
	if err != nil {
		log.Println("Error updating session:", err.Error())
		// Continue even if session update fails
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		custom_log.NewCustomLog("user_delete_failed", err.Error(), "error")

		return nil, responses.NewErrorResponse("user_delete_failed", fmt.Errorf("cannot commit transaction"))
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
	query := "SELECT id, user_role_name FROM tbl_users_roles WHERE deleted_at IS NULL"
	var args []interface{}

	if u.userCtx.RoleId == 1 {
		query += " AND id >= $1"
	} else {
		query += " AND id > $1"
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

		return nil, responses.NewErrorResponse("user_create_form_failed", fmt.Errorf("cannot get roles"))
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
				SELECT 1 FROM tbl_users 
				WHERE role_id <= $1 AND user_uuid = $2 AND deleted_at IS NULL
			)`, u.userCtx.RoleId, user_uuid)

		if err != nil {
			custom_log.NewCustomLog("user_update_form_failed", err.Error(), "error")

			return nil, responses.NewErrorResponse("user_update_form_failed", fmt.Errorf("failed to check permissions"))
		}

		if exists {
			custom_log.NewCustomLog("user_update_form_failed", "permission denied", "error")

			return nil, responses.NewErrorResponse("user_update_form_failed", fmt.Errorf("permission denied: this user has the same or higher role than you"))
		}
	}

	// Get user info
	users, err_one := u.ShowOne(user_uuid)
	if err_one != nil {
		custom_log.NewCustomLog("user_update_form_failed", err_one.Err.Error(), "error")

		return nil, responses.NewErrorResponse("user_update_form_failed", fmt.Errorf("failed to get user info"))
	} else if len(users.Users) <= 0 {
		custom_log.NewCustomLog("user_update_form_failed", "Cannot get user info", "error")

		return nil, responses.NewErrorResponse("user_update_form_failed", fmt.Errorf("user not found"))
	}

	status := u.GetStatus()
	roles, err := u.GetRoles()
	if err != nil {
		custom_log.NewCustomLog("user_update_form_failed", err.Error(), "error")

		return nil, responses.NewErrorResponse("user_update_form_failed", fmt.Errorf("cannot get roles"))
	}

	var userUpdateForms = []UserUpdateForm{
		{
			FirstName:   users.Users[0].FirstName,
			LastName:    users.Users[0].LastName,
			UserName:    users.Users[0].UserName,
			Email:       users.Users[0].Email,
			RoleId:      users.Users[0].RoleId,
			PhoneNumber: *users.Users[0].PhoneNumber,
			StatusId:    uint64(users.Users[0].StatusId),
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
		return nil, responses.NewErrorResponse("user_update_password_failed", fmt.Errorf("cannot begin transaction"))
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

		return nil, responses.NewErrorResponse("user_update_password_failed", err)
	}

	// Get current user ID
	var by_id int64
	err = u.db.Get(&by_id, "SELECT id FROM tbl_users WHERE user_uuid = $1", u.userCtx.UserUuid)
	if err != nil {
		custom_log.NewCustomLog("user_update_password_failed", err.Error())

		return nil, responses.NewErrorResponse("user_update_password_failed", fmt.Errorf("cannot get current user ID"))
	}

	// Get target user info
	users, err_one := u.ShowOne(user_uuid)
	if err_one != nil {
		custom_log.NewCustomLog("user_update_password_failed", err_one.Err.Error(), "error")

		return nil, responses.NewErrorResponse("user_update_password_failed", fmt.Errorf("user not found"))
	} else if len(users.Users) <= 0 {
		custom_log.NewCustomLog("user_update_password_failed", "Cannot get user info", "error")

		return nil, responses.NewErrorResponse("user_update_password_failed", fmt.Errorf("user not found"))
	}

	// Update password
	_, err = tx.Exec(`
		UPDATE tbl_users SET
			password = $1, 
			updated_by = $2, 
			updated_at = $3
		WHERE user_uuid = $4`,
		RequestChangePassword.Password,
		RequestChangePassword.UpdatedBy,
		RequestChangePassword.UpdatedAt,
		RequestChangePassword.UserUUID)

	if err != nil {
		custom_log.NewCustomLog("user_update_password_failed", err.Error(), "error")

		return nil, responses.NewErrorResponse("user_update_password_failed", fmt.Errorf("cannot update password"))
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		custom_log.NewCustomLog("user_update_password_failed", err.Error(), "error")

		return nil, responses.NewErrorResponse("user_update_password_failed", fmt.Errorf("cannot commit transaction"))
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
		FROM tbl_users u
		INNER JOIN tbl_roles ur ON u.role_id = ur.id
		WHERE u.deleted_at IS NULL AND ur.deleted_at IS NULL
		AND u.user_name = $1
	`

	err := u.db.Get(&userInfo, query, username)
	if err != nil {
		custom_log.NewCustomLog("get_userinfo_failed", err.Error(), "error")
		return nil, responses.NewErrorResponse("get_userinfo_failed", fmt.Errorf("cannot select user: %w", err))
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

		return nil, responses.NewErrorResponse("get_userinfo_failed", fmt.Errorf("cannot get user permissions: %w", err))
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
