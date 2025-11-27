package postgres

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	types "snack-shop/pkg/model"
)

// CountRow counts the number of rows returned by a query
func CountRow(rows *sqlx.Rows) (int, error) {
	// Count the rows
	var count int
	for rows.Next() {
		count++
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	return count, nil
}

// ToTimeFormat converts datetime values in a struct to formatted time strings
func ToTimeFormat(v interface{}) (map[string]interface{}, error) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	result := make(map[string]interface{})

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		tag := val.Type().Field(i).Tag.Get("json")
		if tag != "" {
			tagParts := strings.Split(tag, ",")
			fieldName := tagParts[0]
			// Skip fields with `json:"-"`
			if fieldName == "-" {
				continue
			}

			// Handle time.Time fields
			if field.Type() == reflect.TypeOf(time.Time{}) {
				if !field.IsZero() {
					t := field.Interface().(time.Time)
					result[fieldName] = t.Format(time.RFC3339)
				} else {
					result[fieldName] = nil
				}
			} else if field.Type() == reflect.TypeOf(&time.Time{}) {
				if !field.IsNil() {
					t := field.Interface().(*time.Time)
					result[fieldName] = t.Format(time.RFC3339)
				} else {
					result[fieldName] = nil
				}
			} else {
				result[fieldName] = field.Interface()
			}
		}
	}

	return result, nil
}

// BuildSQLSort builds ORDER BY clause from Sort objects
func BuildSQLSort(sorts []types.Sort) string {
	if len(sorts) == 0 {
		return " ORDER BY id" // Default order
	}

	var orderClauses []string
	for _, sort := range sorts {
		orderClauses = append(orderClauses, fmt.Sprintf("%s %s", sort.Property, sort.Direction))
	}

	return " ORDER BY " + strings.Join(orderClauses, ", ")
}

func BuildSQLFilter(req []types.Filter) (string, []interface{}) {
	var sqlFilters []string
	var params []interface{}
	paramIndex := 1

	for _, filter := range req {
		if filter.Property == "" || filter.Value == nil {
			continue
		}

		property := filter.Property
		operator := "="

		// ✅ Support operators: __gte, __lte, __gt, __lt, __ne
		if strings.Contains(property, "__") {
			parts := strings.Split(property, "__")
			property = parts[0]
			switch parts[1] {
			case "gte":
				operator = ">="
			case "lte":
				operator = "<="
			case "gt":
				operator = ">"
			case "lt":
				operator = "<"
			case "ne":
				operator = "!="
			}
		}

		// ✅ Handle BETWEEN for slice values
		switch v := filter.Value.(type) {
		case []interface{}:
			if len(v) == 2 {
				// Try to convert date strings to time.Time
				startStr, ok1 := v[0].(string)
				endStr, ok2 := v[1].(string)
				if ok1 && ok2 {
					if startTime, err := time.Parse("2006-01-02", startStr); err == nil {
						if endTime, err := time.Parse("2006-01-02", endStr); err == nil {
							v[0] = startTime
							v[1] = endTime.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
						}
					}
				}

				sqlFilters = append(sqlFilters, fmt.Sprintf("%s BETWEEN $%d AND $%d", property, paramIndex, paramIndex+1))
				params = append(params, v[0], v[1])
				paramIndex += 2
				continue
			}
		}

		// ✅ Auto-convert types for single values
		switch val := filter.Value.(type) {
		case string:
			// Try parse date
			if t, err := time.Parse("2006-01-02", val); err == nil {
				filter.Value = t
			} else if intValue, err := strconv.Atoi(val); err == nil {
				filter.Value = intValue
			} else if boolValue, err := strconv.ParseBool(val); err == nil {
				filter.Value = boolValue
			}
		}

		// Build SQL by final type
		switch val := filter.Value.(type) {
		case int, bool, time.Time:
			sqlFilters = append(sqlFilters, fmt.Sprintf("%s %s $%d", property, operator, paramIndex))
			params = append(params, val)
		case string:
			if strings.Contains(val, "%") {
				sqlFilters = append(sqlFilters, fmt.Sprintf("%s LIKE $%d", property, paramIndex))
			} else {
				sqlFilters = append(sqlFilters, fmt.Sprintf("%s %s $%d", property, operator, paramIndex))
			}
			params = append(params, val)
		}

		paramIndex++
	}

	return strings.Join(sqlFilters, " AND "), params
}

// GetIdByUuid returns the ID for a given UUID
func GetIdByUuid(space_name string, uuid_field_name string, uuid_str string, db *sqlx.Tx) (*int, error) {
	var id int

	// Parse the UUID
	uid, err := uuid.Parse(uuid_str)
	if err != nil {
		return nil, err
	}

	// Define the SQL query
	sql := fmt.Sprintf(`SELECT id FROM %s WHERE %s=$1`, space_name, uuid_field_name)

	// Execute the query
	err = db.Get(&id, sql, uid)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

// GetSeqNextVal returns the next value from a sequence
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

// SetSeqNextVal sets and returns the next sequence value
func SetSeqNextVal(seq_name string, db *sqlx.Tx) (*int, error) {
	var id int

	// Define the SQL query - adjust to PostgreSQL sequence operations
	sql := fmt.Sprintf(`SELECT nextval('%s') as id`, seq_name)

	// Execute the query
	err := db.Get(&id, sql)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

// IsExists checks if a record exists with the given field value
func IsExists(space_name string, field_name string, value interface{}, db *sqlx.Tx) (bool, error) {
	var exists bool

	// Define the SQL query
	sql := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE %s=$1 AND deleted_at IS NULL)`, space_name, field_name)

	// Execute the query
	err := db.Get(&exists, sql, value)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// IsExistsWhere checks if a record exists with custom WHERE conditions
func IsExistsWhere(space_name string, where_sqlstr string, args []interface{}, db *sqlx.Tx) (bool, error) {
	var exists bool

	// Define the SQL query
	sql := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE %s AND deleted_at IS NULL)`, space_name, where_sqlstr)

	// Execute the query
	err := db.Get(&exists, sql, args...)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func BuildPaging(page int, perPage int) string {
	// var params []interface{}

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}

	offset := (page - 1) * perPage
	limit := perPage

	// params = append(params, offset, limit)

	return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
}
