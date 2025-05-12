package utils

import (
	"fmt"
	"os"
	"time"

	custom_log "snack-shop/pkg/custom_log"
	sql "snack-shop/pkg/sql"

	"github.com/jmoiron/sqlx"
)

func AddMemeberAuditLog(user_id int, audit_context string, audit_desc string, audit_type_id int, member_agent string, user_name string, ip string, by_id int, db_pool *sqlx.DB) (*bool, error) {

	orderSeqName := "tbl_members_audit_id_seq"
	orderVal, err := sql.GetSeqNextVal(orderSeqName, db_pool)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch next order value: %w", err)
	}

	var query = `INSERT INTO tbl_members_audit (
		id, member_id, member_audit_context, member_audit_desc, audit_type_id, member_agent, operator, ip, status_id, "order", created_by, created_at
		) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)`
	app_timezone := os.Getenv("APP_TIMEZONE")
	location, err := time.LoadLocation(app_timezone)
	if err != nil {
		return nil, fmt.Errorf("failed to load location: %w", err)
	}
	local_now := time.Now().In(location)
	_, err = db_pool.Exec(
		query,
		*orderVal,
		user_id,
		audit_context,
		audit_desc,
		audit_type_id,
		member_agent,
		user_name,
		ip,
		1,
		*orderVal,
		by_id,
		local_now,
	)
	if err != nil {
		custom_log.NewCustomLog("user_create_failed", err.Error(), "error")
		return nil, err
	}
	state := true
	return &state, nil
}
