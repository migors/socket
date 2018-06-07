package tg

import (
	"fmt"
)

func EditInlineKeyboard(msgId uint64, chatId uint64, newKeyboard [][]InlineKeyboardButton) error {
	keyboardString, err := InlineKeyboard(newKeyboard).marshal()
	if err != nil {
		return err
	}
	params := map[string]string{
		"chat_id":      fmt.Sprint(chatId),
		"message_id":   fmt.Sprint(msgId),
		"reply_markup": keyboardString,
	}
	logOutgoingMessage("editMessageReplyMarkup", params)
	return requestWithRetry("editMessageReplyMarkup", params, &Dummy{}, messageRetryCount)
}
