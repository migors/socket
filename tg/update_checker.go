package tg

import (
	"fmt"
	"time"
)

const (
	longPollingInterval = 18 // in seconds
	cooldownInterval    = time.Second
)

func StartCheckingUpdates() chan (interface{}) {
	updChan := make(chan interface{})
	go updateChecker(updChan)
	return updChan
}

func updateChecker(updChan chan interface{}) {
	var lastUpdateId uint64
	for {
		updates, err := GetUpdates(lastUpdateId, longPollingInterval)
		if err != nil {
			fmt.Println("Error checking updates:", err)
			time.Sleep(cooldownInterval)
		}
		if len(updates) == 0 {
			continue
		}

		for _, update := range updates {
			lastUpdateId = update.Id
			if update.Message != nil {
				logIncomingMessage(update.Message)
				updChan <- *update.Message
			}
			if update.CallbackQuery != nil {
				logIncomingMessage(update.CallbackQuery)
				updChan <- *update.CallbackQuery
			}
		}
		lastUpdateId++
	}
}
