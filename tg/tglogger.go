package tg

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

var logChan = make(chan string, 50)

func init() {
	go backgroundLogger()
}

func backgroundLogger() {
	file, err := os.OpenFile("data/chat.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal("Cannot write chat log:", err)
		return
	}
	defer file.Close()

	for msg := range logChan {
		_, err = file.WriteString(msg)
		if err != nil {
			log.Println("Cannot write chat log:", err)
		} else {
			file.Sync()
		}
	}
}

func logIncomingMessage(v interface{}) {
	rawJson, err := json.Marshal(v)
	if err != nil {
		log.Println("Cannot marshal chat log json:", err)
		return
	}
	msg := fmt.Sprintln(time.Now().Unix(), "in", string(rawJson))
	logChan <- msg
}

func logOutgoingMessage(method string, params map[string]string) {
	rawJson, err := json.Marshal(params)
	if err != nil {
		log.Println("Cannot marshal chat log json:", err)
		return
	}
	msg := fmt.Sprintln(time.Now().Unix(), "out", method, string(rawJson))
	logChan <- msg
}
