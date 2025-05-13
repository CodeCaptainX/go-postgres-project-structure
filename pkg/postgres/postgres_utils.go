package postgres

import (
	"fmt"
	"os"
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

	app_timezone := os.Getenv("APP_TIMEZONE")
	location, err := time.LoadLocation(app_timezone)
	if err != nil {
		return "", nil
	}

	// Group filters by property
	propertyMap := make(map[string][]interface{})
	for _, filter := range req {
		switch v := filter.Value.(type) {
		case string:
			if intVal, err := strconv.Atoi(v); err == nil {
				filter.Value = intVal
			} else if boolVal, err := strconv.ParseBool(v); err == nil {
				filter.Value = boolVal
			}
		}
		propertyMap[filter.Property] = append(propertyMap[filter.Property], filter.Value)
	}

	placeholderIndex := 1
	for property, values := range propertyMap {
		if len(values) > 1 {
			// Use IN clause
			var placeholders []string
			for _, val := range values {
				placeholders = append(placeholders, fmt.Sprintf("$%d", placeholderIndex))
				params = append(params, val)
				placeholderIndex++
			}
			sqlFilters = append(sqlFilters, fmt.Sprintf("%s IN (%s)", property, strings.Join(placeholders, ", ")))
		} else {
			value := values[0]
			paramPlaceholder := fmt.Sprintf("$%d", placeholderIndex)

			switch v := value.(type) {
			case int, bool:
				sqlFilters = append(sqlFilters, fmt.Sprintf("%s = %s", property, paramPlaceholder))
				params = append(params, v)
				placeholderIndex++

			case string:
				// LIKE
				if strings.Contains(v, "%") {
					sqlFilters = append(sqlFilters, fmt.Sprintf("%s LIKE %s", property, paramPlaceholder))
					params = append(params, v)
					placeholderIndex++

				} else if dateValue, err := time.Parse("2006-01-02", v); err == nil {
					// BETWEEN for date
					start := time.Date(dateValue.Year(), dateValue.Month(), dateValue.Day(), 0, 0, 0, 0, location)
					end := start.Add(24 * time.Hour).Add(-time.Second)

					sqlFilters = append(sqlFilters, fmt.Sprintf("%s BETWEEN $%d AND $%d", property, placeholderIndex, placeholderIndex+1))
					params = append(params, start, end)
					placeholderIndex += 2

				} else {
					// Regular string equality
					sqlFilters = append(sqlFilters, fmt.Sprintf("%s = %s", property, paramPlaceholder))
					params = append(params, v)
					placeholderIndex++
				}
			default:
				// fallback (ignore unsupported types)
			}
		}
	}

	filterClause := strings.Join(sqlFilters, " AND ")
	return filterClause, params
}

// BuildSQLFilter builds WHERE clause from Filter objects
// func BuildSQLFilter(req []types.Filter) (string, []interface{}) {
// 	var sqlFilters []string
// 	var params []interface{}

// 	// Get the current time for date handling
// 	app_timezone := os.Getenv("APP_TIMEZONE")
// 	location, err := time.LoadLocation(app_timezone)
// 	if err != nil {
// 		return "", nil
// 	}

// 	// Map to group filters by Property
// 	propertyMap := make(map[string][]interface{})

// 	// Convert filter values and group them by Property
// 	for _, filter := range req {
// 		switch v := filter.Value.(type) {
// 		case string:
// 			if intValue, err := strconv.Atoi(v); err == nil {
// 				filter.Value = intValue
// 			} else if boolValue, err := strconv.ParseBool(v); err == nil {
// 				filter.Value = boolValue
// 			}
// 		}
// 		// Group by property name
// 		propertyMap[filter.Property] = append(propertyMap[filter.Property], filter.Value)
// 	}

// 	// Process each grouped property
// 	placeholderIndex := 1
// 	for property, values := range propertyMap {
// 		if len(values) > 1 {
// 			// Use IN clause if multiple values for the same property
// 			placeholders := []string{}
// 			for _, value := range values {
// 				placeholders = append(placeholders, fmt.Sprintf("$%d", placeholderIndex))
// 				params = append(params, value)
// 				placeholderIndex++
// 			}
// 			sqlFilters = append(sqlFilters, fmt.Sprintf("%s IN (%s)", property, strings.Join(placeholders, ", ")))
// 		} else {
// 			// Default handling for a single value
// 			value := values[0]
// 			paramPlaceholder := fmt.Sprintf("$%d", placeholderIndex)
// 			switch v := value.(type) {
// 			case int, bool:
// 				sqlFilters = append(sqlFilters, fmt.Sprintf("%s = %s", property, paramPlaceholder))
// 				params = append(params, v)
// 			case string:
// 				if strings.Contains(v, "%") {
// 					sqlFilters = append(sqlFilters, fmt.Sprintf("%s LIKE %s", property, paramPlaceholder))
// 				} else if dateValue, err := time.Parse("2006-01-02", v); err == nil {
// 					// Convert date-only input to datetime range
// 					startOfDay := time.Date(dateValue.Year(), dateValue.Month(), dateValue.Day(), 0, 0, 0, 0, location)
// 					endOfDay := startOfDay.Add(24 * time.Hour).Add(-time.Second)
// 					sqlFilters = append(sqlFilters, fmt.Sprintf("%s BETWEEN %s AND %s", property, paramPlaceholder, fmt.Sprintf("$%d", placeholderIndex+1)))
// 					params = append(params, startOfDay, endOfDay)
// 					placeholderIndex += 2
// 					continue
// 				} else {
// 					sqlFilters = append(sqlFilters, fmt.Sprintf("%s = %s", property, paramPlaceholder))
// 				}
// 				params = append(params, v)
// 			}
// 			placeholderIndex++
// 		}
// 	}

// 	// Join the filters with " AND "
// 	filterClause := strings.Join(sqlFilters, " AND ")

// 	return filterClause, params
// }

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
