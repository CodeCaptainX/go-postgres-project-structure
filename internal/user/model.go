package user

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	types "snack-shop/pkg/model"
	postgres "snack-shop/pkg/postgres"
	"snack-shop/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

type UserContext struct {
	UserUuid     string    `json:"user_uuid"`
	UserName     string    `json:"user_name"`
	LoginSession string    `json:"login_session"`
	Exp          time.Time `json:"exp"`
}

func NewUserContext(userUuid string, userName string, loginSession string, exp time.Time) *UserContext {
	return &UserContext{
		UserUuid:     userUuid,
		UserName:     userName,
		LoginSession: loginSession,
		Exp:          exp,
	}
}

type User struct {
	ID           uint64          `db:"id"`
	UserUUID     uuid.UUID       `db:"user_uuid"`
	FirstName    string          `db:"first_name"`
	LastName     string          `db:"last_name"`
	UserName     string          `db:"user_name"`
	Email        string          `db:"email"`
	RoleId       int             `db:"role_id"`
	RoleName     string          `db:"role_name"` // from ur.user_role_name (was incorrect as int)
	Status       string          `db:"status"`    // assuming PostgreSQL BOOLEAN column
	LoginSession *string         `db:"login_session"`
	ProfilePhoto *string         `db:"profile_photo"`
	UserAlias    *string         `db:"user_alias"`
	PhoneNumber  *string         `db:"phone_number"`
	UserAvatarID *int            `db:"user_avatar_id"`
	Commission   decimal.Decimal `db:"commission"`
	StatusId     int             `db:"status_id"`
	Order        int             `db:"order"`
	CreatedBy    int             `db:"created_by"`
	Creator      string          `db:"creator"`
	CreatedAt    time.Time       `db:"created_at"`
	UpdatedBy    *int            `db:"updated_by"`
	UpdatedAt    *time.Time      `db:"updated_at"`
	DeletedBy    *int            `db:"deleted_by"`
	DeletedAt    *time.Time      `db:"deleted_at"`
}

type UserAddModel struct {
	ID           uint64          `db:"id"`
	UserUUID     uuid.UUID       `db:"user_uuid"`
	FirstName    string          `db:"first_name"`
	LastName     string          `db:"last_name"`
	UserName     string          `db:"user_name"`
	Password     string          `db:"password"`
	Email        string          `db:"email"`
	RoleId       int             `db:"role_id"`
	Status       bool            `db:"status"`
	LoginSession *string         `db:"login_session"`
	ProfilePhoto *string         `db:"profile_photo"`
	UserAlias    *string         `db:"user_alias"`
	PhoneNumber  *string         `db:"phone_number"`
	UserAvatarID *float64        `db:"user_avatar_id"`
	Commission   decimal.Decimal `db:"commission"`
	StatusId     uint64          `db:"status_id"`
	Order        uint64          `db:"order_num"`
	CreatedBy    uint64          `db:"created_by"`
	CreatedAt    time.Time       `db:"created_at"`
}

type UserNewRequest struct {
	FirstName       string          `json:"first_name" validate:"required"`
	LastName        string          `json:"last_name" validate:"required"`
	UserName        string          `json:"user_name" validate:"required"`
	Password        string          `json:"password" validate:"required,min=6"`
	PasswordConfirm string          `json:"password_confirm" validate:"required,min=6"`
	Email           string          `json:"email" validate:"required,email"`
	RoleId          int             `json:"role_id" validate:"required"`
	PhoneNumber     *string         `json:"phone_number" validate:"required"`
	Commission      decimal.Decimal `json:"commission"`
}

func (r *UserNewRequest) bind(c *fiber.Ctx, v *utils.Validator) error {

	if err := c.BodyParser(r); err != nil {
		return err
	}
	// Trim spaces from FirstName and LastName and check other neccessary
	r.FirstName = strings.TrimSpace(r.FirstName)
	r.LastName = strings.TrimSpace(r.LastName)
	r.Email = strings.TrimSpace(r.Email)
	r.UserName = strings.TrimSpace(r.UserName)

	if err := v.Validate(r); err != nil {
		return err
	}

	//Check confirm password
	if r.Password != r.PasswordConfirm {
		return fmt.Errorf("confirm password not match")
	}

	return nil
}

// user.cTx
func (u *UserAddModel) New(usreq UserNewRequest, usctx *types.UserContext, dbtx *sqlx.Tx) error {
	if usctx.RoleId > uint64(usreq.RoleId) {
		return fmt.Errorf("permission denied: cannot create a user with a higher role")
	}

	uid, err := uuid.NewV7()
	if err != nil {
		return err
	}

	uidSession, err := uuid.NewV7()
	if err != nil {
		return err
	}
	sessionString := uidSession.String()

	byID, err := postgres.GetIdByUuid("tbl_users", "user_uuid", usctx.UserUuid, dbtx)
	if err != nil {
		return err
	}

	appTimezone := os.Getenv("APP_TIMEZONE")
	location, err := time.LoadLocation(appTimezone)
	if err != nil {
		return fmt.Errorf("failed to load location: %w", err)
	}
	localNow := time.Now().In(location)

	id, errSeq := postgres.GetSeqNextVal("tbl_users_id_seq", dbtx)
	if errSeq != nil {
		return errSeq
	}

	isUsername, err := postgres.IsExists("tbl_users", "user_name", usreq.UserName, dbtx)
	if err != nil {
		return err
	}
	if isUsername {
		return fmt.Errorf("username `%s` already exists", usreq.UserName)
	}

	photo := "user2.png"
	u.ID = uint64(*id)
	u.UserUUID = uid
	u.FirstName = usreq.FirstName
	u.LastName = usreq.LastName
	u.UserName = usreq.UserName
	u.Password = usreq.Password
	u.Email = usreq.Email
	u.RoleId = usreq.RoleId
	u.Status = true
	u.LoginSession = &sessionString
	u.ProfilePhoto = &photo
	u.UserAlias = &usreq.UserName
	u.PhoneNumber = usreq.PhoneNumber
	u.UserAvatarID = nil
	u.Commission = usreq.Commission
	u.StatusId = 1
	u.Order = u.ID
	u.CreatedBy = uint64(*byID)
	u.CreatedAt = localNow

	return nil
}

type UserUpdateRequest struct {
	FirstName   string          `json:"first_name" validate:"required"`
	LastName    string          `json:"last_name" validate:"required"`
	Email       string          `json:"email"  validate:"required,email"`
	RoleId      int             `json:"role_id"  validate:"required"`
	PhoneNumber *string         `json:"phone_number"  validate:"required"`
	Commission  decimal.Decimal `json:"commission"`
	StatusId    int             `json:"status_id"  validate:"required"`
}

func (r *UserUpdateRequest) bind(c *fiber.Ctx, v *utils.Validator) error {

	if err := c.BodyParser(r); err != nil {
		return err
	}
	// Trim spaces from FirstName and LastName and check other neccessary
	r.FirstName = strings.TrimSpace(r.FirstName)
	r.LastName = strings.TrimSpace(r.LastName)
	r.Email = strings.TrimSpace(r.Email)

	if err := v.Validate(r); err != nil {
		return err
	}

	return nil
}

type UserUpdateModel struct {
	ID           uint64
	UserUUID     uuid.UUID
	FirstName    string
	LastName     string
	UserName     string
	Password     string
	Email        string
	RoleId       int
	Status       bool
	LoginSession *string
	ProfilePhoto *string
	UserAlias    *string
	PhoneNumber  *string
	UserAvatarID *float64
	Commission   decimal.Decimal
	StatusId     uint64
	Order        uint64
	UpdatedBy    uint64
	UpdatedAt    time.Time
}

func (u *UserUpdateModel) New(user_uuid uuid.UUID, usreq UserUpdateRequest, usctx *types.UserContext, dbstream *sqlx.Tx) error {
	// check permission
	isYours, err := postgres.IsExistsWhere(
		"tbl_users",
		"user_uuid = $1 AND user_name = $2",
		[]interface{}{user_uuid, usctx.UserName},
		dbstream,
	)
	if err != nil {
		return fmt.Errorf("cannot check permission on user")
	} else {
		if isYours {
			return fmt.Errorf("you cannot update your own information")
		}
	}
	if usctx.RoleId > uint64(usreq.RoleId) {
		return fmt.Errorf("permission denied : you can't update a user that have bigger or equal role to you")
	}

	//Check if user uuid exits
	is_useruuid, err_seq := postgres.IsExists("tbl_users", "user_uuid", user_uuid, dbstream)
	if err_seq != nil {
		return err_seq
	} else {
		if !is_useruuid {
			return fmt.Errorf(fmt.Sprintf("user uuid:`%s` not found dddd", user_uuid))
		}
	}

	//Get user logined id
	by_id, err := postgres.GetIdByUuid("tbl_users", "user_uuid", usctx.UserUuid, dbstream)
	if err != nil {
		return err
	}

	//Get user logined id
	id, err := postgres.GetIdByUuid("tbl_users", "user_uuid", user_uuid.String(), dbstream)
	if err != nil {
		return err
	}

	//Get current OS time
	app_timezone := os.Getenv("APP_TIMEZONE")
	location, err := time.LoadLocation(app_timezone)
	if err != nil {
		return fmt.Errorf("failed to load location: %w", err)
	}
	local_now := time.Now().In(location)

	u.ID = uint64(*id)
	u.UserUUID = user_uuid
	u.FirstName = usreq.FirstName
	u.LastName = usreq.LastName
	u.Email = usreq.Email
	u.RoleId = usreq.RoleId
	u.PhoneNumber = usreq.PhoneNumber
	u.Commission = usreq.Commission
	u.StatusId = uint64(usreq.StatusId)
	u.UpdatedBy = uint64(*by_id)
	u.UpdatedAt = local_now
	return nil
}

type UserResponse struct {
	Users []User `json:"users"`
	Total int    `json:"-"`
}

type TotalRecord struct {
	Total int
}
type UserShowRequest struct {
	PageOptions types.Paging   `json:"paging_options" query:"paging_options" validate:"required"`
	Sorts       []types.Sort   `json:"sorts,omitempty" query:"sorts"`
	Filters     []types.Filter `json:"filters,omitempty" query:"filters"`
}

func (r *UserShowRequest) bind(c *fiber.Ctx, v *utils.Validator) error {

	if err := c.QueryParser(r); err != nil {
		return err
	}

	//Fix bug `Filter.Value` nil when http query params failed parse to json type `interface{}`
	for i := range r.Filters {
		value := c.Query(fmt.Sprintf("filters[%d][value]", i))
		if intValue, err := strconv.Atoi(value); err == nil {
			r.Filters[i].Value = intValue
		} else if boolValue, err := strconv.ParseBool(value); err == nil {
			r.Filters[i].Value = boolValue
		} else {
			r.Filters[i].Value = value
		}
	}

	if err := v.Validate(r); err != nil {
		return err
	}
	return nil
}

type UserDeleteResponse struct {
	Success bool `json:"success"`
}
type Role struct {
	Id           uint64 `json:"id"`
	UserRoleName string `json:"user_role_name"`
}

type UserCreateForm struct {
	FirstName       string         `json:"first_name"`
	LastName        string         `json:"last_name"`
	UserName        string         `json:"user_name"`
	Password        string         `json:"password"`
	PasswordComfirm string         `json:"password_confirm"`
	Email           string         `json:"email"`
	RoleId          int            `json:"role_id"`
	PhoneNumber     string         `json:"phone_number"`
	StatusId        uint64         `json:"status_id"`
	Status          []types.Status `json:"status"`
	Roles           []Role         `json:"roles"`
}
type UserFormCreateResponse struct {
	Users []UserCreateForm `json:"users"`
}

type UserUpdateForm struct {
	FirstName   string          `json:"first_name"`
	LastName    string          `json:"last_name"`
	UserName    string          `json:"user_name"`
	Email       string          `json:"email"`
	RoleId      int             `json:"role_id"`
	PhoneNumber string          `json:"phone_number"`
	StatusId    uint64          `json:"status_id"`
	Commission  decimal.Decimal `json:"commission"`
	Status      []types.Status  `json:"status"`
	Roles       []Role          `json:"roles"`
}
type UserFormUpdateResponse struct {
	Users []UserUpdateForm `json:"users"`
}

type UserUpdatePasswordReponse struct {
	Success bool `json:"success"`
}
type UserUpdatePasswordModel struct {
	UserUUID  uuid.UUID
	Password  string `json:"password" validate:"required,min=6"`
	UpdatedBy uint64
	UpdatedAt time.Time
}
type UserUpdatePasswordRequest struct {
	OldPassword     string `json:"old_password" validate:"required"`
	Password        string `json:"password" validate:"required,min=6"`
	PasswordConfirm string `json:"password_confirm" validate:"required,min=6"`
}

func (r *UserUpdatePasswordRequest) bind(c *fiber.Ctx, v *utils.Validator) error {

	// Parse the request body into the UserUpdatePasswordRequest struct
	if err := c.BodyParser(r); err != nil {
		return err
	}

	// Trim spaces from the password fields
	r.OldPassword = strings.TrimSpace(r.OldPassword)
	r.Password = strings.TrimSpace(r.Password)
	r.PasswordConfirm = strings.TrimSpace(r.PasswordConfirm)

	// Validate the struct fields
	if err := v.Validate(r); err != nil {
		return err
	}

	// Check if the new password matches the confirmation password
	if r.Password != r.PasswordConfirm {
		return fmt.Errorf("confirm password does not match")
	}

	return nil
}

func (u *UserUpdatePasswordModel) New(user_uuid uuid.UUID, usreq UserUpdatePasswordRequest, usctx *types.UserContext, db *sqlx.Tx) error {

	// Check if user_uuid exists
	is_useruuid, err := postgres.IsExists("tbl_users", "user_uuid", user_uuid, db)
	if err != nil {
		return err
	}
	if !is_useruuid {
		return fmt.Errorf("user update by uuid:`%s` not found", user_uuid)
	}

	// Get the ID of the user performing the update
	by_id, err := postgres.GetIdByUuid("tbl_users", "user_uuid", usctx.UserUuid, db)
	if err != nil {
		return err
	}

	// Fetch the existing password for the target user
	var oldPassword string
	query := `SELECT password FROM tbl_users WHERE user_uuid = $1`
	err = db.Get(&oldPassword, query, user_uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("can't find the old password")
		}
		return fmt.Errorf("failed to query old password: %w", err)
	}

	// Verify the old password matches
	if usreq.OldPassword != oldPassword {
		return fmt.Errorf("old password does not match")
	}

	// Get current time in configured timezone
	app_timezone := os.Getenv("APP_TIMEZONE")
	location, err := time.LoadLocation(app_timezone)
	if err != nil {
		return fmt.Errorf("failed to load location: %w", err)
	}
	local_now := time.Now().In(location)

	// Update struct values (presumably for later use)
	u.Password = usreq.Password
	u.UserUUID = user_uuid
	u.UpdatedBy = uint64(*by_id)
	u.UpdatedAt = local_now

	return nil
}

type UserInfo struct {
	ID           uint64          `json:"-"`
	UserUUID     uuid.UUID       `json:"user_uuid"`
	FirstName    string          `json:"first_name"`
	LastName     string          `json:"last_name"`
	UserName     string          `json:"user_name"`
	Email        string          `json:"email"`
	RoleId       int             `json:"role_id"`
	RoleName     string          `json:"role_name"`
	Status       bool            `json:"status"`
	LoginSession *string         `json:"login_session"`
	ProfilePhoto *string         `json:"profile_photo"`
	UserAlias    *string         `json:"user_alias"`
	PhoneNumber  *string         `json:"phone_number"`
	UserAvatarID *float64        `json:"user_avatar_id"`
	Commission   decimal.Decimal `json:"commission"`
	StatusId     uint64          `json:"status_id"`
}

type UserBasicInfo struct {
	UserInfo       UserInfo       `json:"user_info"`
	UserPermission UserPermission `json:"user_permission"`
}

type UserPermission struct {
	Modules map[string][]string `json:"modules"`
}

type UserBasicInfoResponse struct {
	UserBasicInfo UserBasicInfo `json:"user_basic_info"`
}

type Permission struct {
	ModuleName  string `json:"module_name"`
	FunctionIDs string `json:"function_ids"`
}
