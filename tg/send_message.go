package tg

import (
	"fmt"
)

func SendMdMessage(text string, chatId uint64, replyId uint64) error {
	params := map[string]string{
		"chat_id":                  fmt.Sprint(chatId),
		"text":                     text,
		"parse_mode":               "Markdown",
		"disable_web_page_preview": "true",
	}
	if replyId != 0 {
		params["reply_to_message_id"] = fmt.Sprint(replyId)
	}
	return request("sendMessage", params, nil)
}

func SendLocation(lat float64, lon float64, chatId uint64, replyId uint64) error {
	params := map[string]string{
		"chat_id":   fmt.Sprint(chatId),
		"latitude":  fmt.Sprint(lat),
		"longitude": fmt.Sprint(lon),
	}
	if replyId != 0 {
		params["reply_to_message_id"] = fmt.Sprint(replyId)
	}
	return request("sendLocation", params, nil)
}

func SendPhotoByUrl(picUrl string, chatId uint64, replyId uint64) error {
	params := map[string]string{
		"chat_id": fmt.Sprint(chatId),
		"photo":   picUrl,
	}
	if replyId != 0 {
		params["reply_to_message_id"] = fmt.Sprint(replyId)
	}
	return request("sendPhoto", params, nil)
}
