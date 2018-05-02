package tg

import (
	"fmt"
)

type User struct {
	Id           uint64 `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
}

type Location struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

type Message struct {
	Id          uint64    `json:"message_id"`
	From        User      `json:"from"`
	ForwardFrom User      `json:"forward_from"`
	Text        string    `json:"text"`
	Location    *Location `json:"location"`
}

type Update struct {
	Id      uint64   `json:"update_id"`
	Message *Message `json:"message"`
}

func GetUpdates(offset uint64, timeoutSec uint) ([]Update, error) {
	updates := []Update{}
	err := request("getUpdates", map[string]string{
		"offset":  fmt.Sprint(offset),
		"timeout": fmt.Sprint(timeoutSec),
	}, &updates)

	return updates, err
}
