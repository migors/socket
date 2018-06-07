package db

import (
	"encoding/json"
	"reflect"
	"strconv"

	"bitbucket.org/pav5000/socketbot/logger"
)

type SessionDataRow struct {
	User  uint64 `db:"user"`
	Key   string `db:"key"`
	Value string `db:"value"`
}

func SetSessionValue(user uint64, key string, value interface{}) {
	t := reflect.TypeOf(value)
	if t.Kind() == reflect.Slice ||
		t.Kind() == reflect.Array ||
		t.Kind() == reflect.Map ||
		t.Kind() == reflect.Struct {

		rawJson, err := json.Marshal(value)
		if err != nil {
			logger.Err("Error marshaling session value into json: ", err)
			return
		}
		value = string(rawJson)
	}

	_, err := db.Exec(`INSERT INTO session_data (user, key, value) VALUES(?,?,?)`, user, key, value)
	if err != nil {
		logger.Err("Error setting session value: ", err)
	}
}

func GetSessionValue(user uint64, key string) string {
	row := SessionDataRow{}
	err := db.Get(&row, `SELECT value FROM session_data WHERE user=? AND key=?`, user, key)
	if err != nil {
		logger.Err("Error getting session value: ", err)
	}
	return row.Value
}

func GetSessionValueFloat64(user uint64, key string) float64 {
	str := GetSessionValue(user, key)
	num, err := strconv.ParseFloat(str, 64)
	if err != nil {
		logger.Err("Error converting session value '", str, "' to float64: ", err)
		return 0
	}
	return num
}

func GetSessionValueInt64(user uint64, key string) int64 {
	str := GetSessionValue(user, key)
	num, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		logger.Err("Error converting session value '", str, "' to int64: ", err)
		return 0
	}
	return num
}

func GetSessionValueUint64(user uint64, key string) uint64 {
	str := GetSessionValue(user, key)
	num, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		logger.Err("Error converting session value '", str, "' to uint64: ", err)
		return 0
	}
	return num
}

func GetSessionValueJson(user uint64, key string, v interface{}) bool {
	row := SessionDataRow{}
	err := db.Get(&row, `SELECT value FROM session_data WHERE user=? AND key=?`, user, key)
	if err != nil {
		logger.Err("Error getting session value: ", err)
		return false
	}
	err = json.Unmarshal([]byte(row.Value), v)
	if err != nil {
		logger.Err("Error unmarshaling session value: ", err)
		return false
	}
	return true
}

func ClearSessionValues(user uint64) {
	_, err := db.Exec(`DELETE FROM session_data WHERE user=?`, user)
	if err != nil {
		logger.Err("Error clearing session values: ", err)
	}
}
