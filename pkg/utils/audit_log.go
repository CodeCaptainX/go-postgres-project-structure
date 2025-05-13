package utils

import (
	"fmt"
	"os"
	"time"

	custom_log "snack-shop/pkg/custom_log"
	sql "snack-shop/pkg/postgres"

	"github.com/jmoiron/sqlx"
)

func AddUserAuditLog(userID int, auditContext string, auditDesc string, auditTypeID int, userAgent string, userName string, ip string, createdBy int, dbPool *sqlx.DB) (*bool, error) {
	// Get next sequence value
	seqName := "tbl_users_audits_id_seq"
	seqVal, err := sql.GetSeqNextVal(seqName, dbPool)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch next sequence value: %w", err)
	}

	// Build insert query
	query := `INSERT INTO tbl_users_audits (
		id, user_id, audit_context, audit_desc, audit_type_id, user_agent, operator, ip, status_id, "order", created_by, created_at
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
	)`

	// Get local time
	appTimezone := os.Getenv("APP_TIMEZONE")
	location, err := time.LoadLocation(appTimezone)
	if err != nil {
		return nil, fmt.Errorf("failed to load location: %w", err)
	}
	localNow := time.Now().In(location)

	// Execute insert
	_, err = dbPool.Exec(
		query,
		*seqVal,
		userID,
		auditContext,
		auditDesc,
		auditTypeID,
		userAgent,
		userName,
		ip,
		1,
		*seqVal,
		createdBy,
		localNow,
	)
	if err != nil {
		custom_log.NewCustomLog("user_audit_create_failed", err.Error(), "error")
		return nil, err
	}

	success := true
	return &success, nil
}
