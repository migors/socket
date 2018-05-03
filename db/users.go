package db

import (
	"time"
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
func UpdateUserInfo(id uint64, firstName string, lastName string, username string, languageCode string, lastMessage time.Time) error {
	var count int
	err := db.Get(&count, `SELECT count(*) FROM users WHERE id=?`, id)
	if err != nil {
		return err
	}

	if count == 0 {
		_, err = db.Exec(
			`INSERT OR REPLACE INTO users (id, first_name, last_name, username, language_code, first_message, last_message) VALUES(?,?,?,?,?,?,?)`,
			id,
			firstName,
			lastName,
			username,
			languageCode,
			lastMessage.Unix(),
			lastMessage.Unix())
	} else {
		_, err = db.Exec(
			`UPDATE users SET first_name=?, last_name=?, username=?, language_code=?, last_message=? WHERE id=?`,
			firstName,
			lastName,
			username,
			languageCode,
			lastMessage.Unix(),
			id)
	}

	return err
}
