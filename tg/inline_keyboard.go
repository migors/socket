package tg

import (
	"encoding/json"
	"errors"
	"fmt"
)

type InlineKeyboard [][]InlineKeyboardButton

type InlineKeyboardButton struct {
	Text         string
	Url          string
	CallbackData interface{}
}

func (btn *InlineKeyboardButton) marshal() (inlineKeyboardButton, error) {
	rawJson, err := json.Marshal(btn.CallbackData)
	if err != nil {
		return inlineKeyboardButton{}, err
	}
	if len(rawJson) > 64 {
		return inlineKeyboardButton{}, errors.New("callback_data is too long, should be max 64 bytes: " + string(rawJson))
	}
	return inlineKeyboardButton{
		Text:         btn.Text,
		Url:          btn.Url,
		CallbackData: string(rawJson),
	}, nil
}

func (keyboard InlineKeyboard) marshal() (string, error) {
	processedKeyboard := make([][]inlineKeyboardButton, 0, len(keyboard))
	for _, row := range keyboard {
		processedRow := make([]inlineKeyboardButton, 0, len(row))
		for _, btn := range row {
			marshaled, err := btn.marshal()
			if err != nil {
				return "", err
			}
			processedRow = append(processedRow, marshaled)
		}
		processedKeyboard = append(processedKeyboard, processedRow)
	}
	rawJson, err := json.Marshal(ReplyMarkup{InlineKeyboard: processedKeyboard})
	if err != nil {
		return "", err
	}
	return string(rawJson), nil
}

type inlineKeyboardButton struct {
	Text         string `json:"text,omitempty"`
	Url          string `json:"url,omitempty"`
	CallbackData string `json:"callback_data,omitempty"`
}

type ReplyMarkup struct {
	InlineKeyboard [][]inlineKeyboardButton `json:"inline_keyboard"`
}

func SendMdMessageWithKeyboard(text string, chatId uint64, replyId uint64, keyboard [][]InlineKeyboardButton) error {
	params := map[string]string{
		"chat_id":                  fmt.Sprint(chatId),
		"text":                     text,
		"parse_mode":               "Markdown",
		"disable_web_page_preview": "true",
	}
	{
		keyboardString, err := InlineKeyboard(keyboard).marshal()
		if err != nil {
			return err
		}
		params["reply_markup"] = keyboardString
	}
	if replyId != 0 {
		params["reply_to_message_id"] = fmt.Sprint(replyId)
	}
	logOutgoingMessage("sendMessage", params)
	return requestWithRetry("sendMessage", params, &Dummy{}, messageRetryCount)
}
