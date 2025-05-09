package utils

import (
	"os"
	"strconv"
)

func GetenvInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
