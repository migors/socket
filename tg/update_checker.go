package tg

import (
	"fmt"
	"time"
)

const (
	longPollingInterval = 30 // in seconds
	cooldownInterval    = time.Second
)

func StartCheckingUpdates() chan (interface{}) {
	updChan := make(chan interface{}, 30)
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

		for _, update := range updates {
			lastUpdateId = update.Id
			if update.Message != nil {
				updChan <- *update.Message
			}
		}
		lastUpdateId++
	}
}
