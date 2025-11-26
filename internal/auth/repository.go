package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	custom_log "snack-shop/pkg/logs"
	types "snack-shop/pkg/model"
	redis_util "snack-shop/pkg/redis"
	"snack-shop/pkg/responses"
	util "snack-shop/pkg/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type AuthRepository interface {
	Login(username, password string) (*AuthResponse, *responses.ErrorResponse)
	CheckSession(loginSession string) (*types.UserSession, *responses.ErrorResponse)
}

type authRepositoryImpl struct {
	dbPool *sqlx.DB
	redis  *redis.Client
}

func NewAuthRepository(dbPool *sqlx.DB, redisClient *redis.Client) AuthRepository {
	return &authRepositoryImpl{
		dbPool: dbPool,
		redis:  redisClient,
	}
}

func (a *authRepositoryImpl) Login(username, password string) (*AuthResponse, *responses.ErrorResponse) {
	var member MemberData
	msg := responses.ErrorResponse{}

	query := `
		SELECT
			id, 
			user_name,
			user_uuid,
			role_id,
			email,
			password
		FROM tbl_users 
		WHERE user_name = $1 AND password = $2 AND deleted_at IS NULL
	`

	err := a.dbPool.Get(&member, query, username, password)
	if err != nil {
		custom_log.NewCustomLog("member_not_found", err.Error(), "error")
		return nil, msg.NewErrorResponse("member_not_found", fmt.Errorf("user not found. Please check the provided information"))
	}

	var res AuthResponse

	hours := util.GetenvInt("JWT_EXP_HOUR", 7)
	expirationTime := time.Now().Add(time.Duration(hours) * time.Hour)
	loginSession, err := uuid.NewV7()

	if err != nil {
		custom_log.NewCustomLog("uuid_generate_failed", err.Error(), "error")
		return nil, msg.NewErrorResponse("uuid_generate_failed", fmt.Errorf("failed to generate UUID. Please try again later"))
	}

	claims := jwt.MapClaims{
		"user_uuid":     member.UserUuid,
		"user_id":       member.ID,
		"username":      member.Username,
		"role_id":       member.RoleId,
		"login_session": loginSession.String(),
		"exp":           expirationTime.Unix(),
	}

	// Set Redis Data
	key := fmt.Sprintf("member_info_id: %d", member.ID)
	redisUtil := redis_util.NewRedisUtil(a.redis)
	redisUtil.SetCacheKey(key, claims, context.Background())

	_ = godotenv.Load() // Ignore error if .env file not found

	secretKey := os.Getenv("JWT_SECRET_KEY")

	updateQuery := `	
		UPDATE tbl_users
		SET login_session = $1
		WHERE id = $2
	`
	_, err = a.dbPool.Exec(updateQuery, loginSession.String(), member.ID)
	if err != nil {
		custom_log.NewCustomLog("session_update_failed", err.Error(), "error")
		return nil, msg.NewErrorResponse("session_update_failed", fmt.Errorf("cannot update session"))
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		custom_log.NewCustomLog("jwt_failed", err.Error(), "error")
		return nil, msg.NewErrorResponse("jwt_failed", fmt.Errorf("failed to get jwt"))
	}

	res.Auth.Token = tokenString
	res.Auth.TokenType = "jwt"

	// auditDesc := fmt.Sprintf(`Member : %s has been login to the system`, username)
	// _, err = audit.AddMemeberAuditLog(member.ID, "Login", auditDesc, 1, "userAgent", member.Username, "ip", member.ID, a.dbPool)
	// if err != nil {
	// 	custom_log.NewCustomLog("add_audit_log_failed", err.Error(), "error")
	// 	return nil, msg.NewErrorResponse("add_audit_log_failed", fmt.Errorf("cannot insert data to audit log"))
	// }

	return &res, nil
}

func (a *authRepositoryImpl) CheckSession(loginSession string) (*types.UserSession, *responses.ErrorResponse) {
	msg := responses.ErrorResponse{}

	// key := fmt.Sprintf("member:%d", int(memberID))
	// redisUtil := redis_util.NewRedisUtil(a.redis)

	// // 1) Try Redis cache
	// keyData, err := redisUtil.GetCacheKey(key, context.Background())
	// if err == nil {
	// 	if keyData.LoginSession == loginSession {
	// 		// Redis has full session → return it
	// 		return &types.UserSession{
	// 			UserID:       int64(memberID),
	// 			LoginSession: keyData.LoginSession,
	// 			UserUUID:     keyData.UserUUID,
	// 			UserName:     keyData.UserName,
	// 			RoleID:       keyData.RoleID,
	// 		}, nil
	// 	}
	// }

	// 2) Try DB fallback
	var session types.UserSession
	query := `
        SELECT 
            id,
            user_uuid,
            user_name,
            role_id,
            login_session
        FROM tbl_users
        WHERE login_session = $1 
        LIMIT 1
    `

	err := a.dbPool.Get(&session, query, loginSession)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			custom_log.NewCustomLog("invalid_session_id", "invalid login session: "+loginSession, "warn")
			return nil, msg.NewErrorResponse("invalid_session_id", fmt.Errorf("invalid login session"))
		}
		custom_log.NewCustomLog("query_data_failed", err.Error(), "error")
		return nil, msg.NewErrorResponse("query_data_failed", fmt.Errorf("database query error"))
	}

	// 3) Save to Redis for next time
	// redisUtil.SetCacheKey(key, session, time.Hour*1)

	return &session, nil
}
