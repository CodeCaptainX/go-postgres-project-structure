package sql

import (
	"fmt"
	"github.com/jmoiron/sqlx"
)

// SeqResult struct to store sequence result
type SeqResult struct {
	ID int `db:"id"`
}

// Supports both normal DB connection and transactions
func GetSeqNextVal(seqName string, exec sqlx.Ext) (*int, error) {
	var result SeqResult
	sql := `SELECT nextval($1) AS id`

	// Execute query using either DB or transaction
	err := sqlx.Get(exec, &result, sql, seqName)
	if err != nil {
		return nil, fmt.Errorf("failed to get sequence value: %w", err)
	}
	return &result.ID, nil
}
