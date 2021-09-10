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

type Photo struct {
	FileId   string `json:"file_id"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	FileSize int64  `json:"file_size"`
}

type Chat struct {
	Id   uint64 `json:"id"`
	Type string `json:"type"`
}

type Message struct {
	Id          uint64    `json:"message_id"`
	From        User      `json:"from"`
	Chat        Chat      `json:"chat"`
	ForwardFrom User      `json:"forward_from"`
	Text        string    `json:"text"`
	Location    *Location `json:"location"`
	PhotoSizes  []Photo   `json:"photo"`
}

type CallbackQuery struct {
	Id           string  `json:"id"`
	From         User    `json:"from"`
	Message      Message `json:"message"`
	ChatInstance string  `json:"chat_instance"`
	Data         string  `json:"data"`
}

func (msg *Message) GetLargestPhoto() Photo {
	if len(msg.PhotoSizes) == 0 {
		return Photo{}
	}
	largest := msg.PhotoSizes[0]
	for _, photo := range msg.PhotoSizes {
		if photo.Width > largest.Width {
			largest = photo
		}
	}
	return largest
}

type Update struct {
	Id            uint64         `json:"update_id"`
	Message       *Message       `json:"message"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}

func GetUpdates(offset uint64, timeoutSec uint) ([]Update, error) {
	updates := []Update{}
	fmt.Println("GetUpdates")
	err := request("getUpdates", map[string]string{
		"offset":  fmt.Sprint(offset),
		"timeout": fmt.Sprint(timeoutSec),
		"limit":   "1",
	}, &updates)
	fmt.Println("GetUpdates err =", err)

	return updates, err
}
