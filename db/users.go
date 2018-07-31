package db

import (
	"time"

	"bitbucket.org/pav5000/socketbot/logger"
	"bitbucket.org/pav5000/socketbot/tg"
)

type UsersRow struct {
	Id           int64  `db:"id"`
	FirstName    string `db:"first_name"`
	LastName     string `db:"last_name"`
	Username     string `db:"username"`
	LanguageCode string `db:"language_code"`
	ChatState    string `db:"chat_state"`
	FirstMessage int64  `db:"first_message"`
	LastMessage  int64  `db:"last_message"`
}

// Update user info
// Is called every message to check if user updated it's name and other fields
func UpdateUserInfo(user tg.User) error {
	lastMessage := time.Now()
	var count int
	err := db.Get(&count, `SELECT count(*) FROM users WHERE id=?`, user.Id)
	if err != nil {
		return err
	}

	if count == 0 {
		_, err = db.Exec(
			`INSERT OR REPLACE INTO users (id, first_name, last_name, username, language_code, first_message, last_message) VALUES(?,?,?,?,?,?,?)`,
			user.Id,
			user.FirstName,
			user.LastName,
			user.Username,
			user.LanguageCode,
			lastMessage.Unix(),
			lastMessage.Unix())
	} else {
		_, err = db.Exec(
			`UPDATE users SET first_name=?, last_name=?, username=?, language_code=?, last_message=? WHERE id=?`,
			user.FirstName,
			user.LastName,
			user.Username,
			user.LanguageCode,
			lastMessage.Unix(),
			user.Id)
	}

	return err
}

func GetUserState(id uint64) (string, error) {
	row := UsersRow{}
	err := db.Get(&row, `SELECT chat_state FROM users WHERE id=?`, id)
	return row.ChatState, err
}

func SetUserState(id uint64, state string) {
	_, err := db.Exec(`UPDATE users SET chat_state=? WHERE id=?`, state, id)
	if err != nil {
		logger.Err("Setting user state failed: ", err)
	}
}
