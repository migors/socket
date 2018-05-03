package tg

import (
	"encoding/json"
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

func SendVideoByUrl(vidUrl string, chatId uint64, caption string, replyId uint64) error {
	params := map[string]string{
		"chat_id": fmt.Sprint(chatId),
		"video":   vidUrl,
		"caption": caption,
	}
	if replyId != 0 {
		params["reply_to_message_id"] = fmt.Sprint(replyId)
	}
	return request("sendVideo", params, nil)
}

type InputMediaPhoto struct {
	Type  string `json:"type"`
	Media string `json:"media"`
}

func SendPhotoGroup(photoUrls []string, chatId uint64, replyId uint64) error {
	media := make([]InputMediaPhoto, 0, len(photoUrls))

	for _, photoUrl := range photoUrls {
		media = append(media, InputMediaPhoto{
			Type:  "photo",
			Media: photoUrl,
		})
	}

	rawJson, err := json.Marshal(media)
	if err != nil {
		return err
	}

	params := map[string]string{
		"chat_id": fmt.Sprint(chatId),
		"media":   string(rawJson),
	}
	if replyId != 0 {
		params["reply_to_message_id"] = fmt.Sprint(replyId)
	}
	return request("sendMediaGroup", params, nil)
}
