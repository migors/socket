package tg

import (
	"fmt"
)

func SendMdMessage(text string, chatId uint64) error {
	fmt.Println(chatId, text)
	return request("sendMessage", map[string]string{
		"chat_id":                  fmt.Sprint(chatId),
		"text":                     text,
		"parse_mode":               "Markdown",
		"disable_web_page_preview": "true",
	}, nil)
}
