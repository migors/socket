package tg

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

var chatLogChan = make(chan string, 50)
var criticalLogChan = make(chan string)

func init() {
	go backgroundLogger("data/chat.log", chatLogChan)
	go backgroundLogger("data/critical.log", criticalLogChan)
}

func backgroundLogger(filename string, c chan string) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal("Cannot write log to '"+filename+"':", err)
		return
	}
	defer file.Close()

	for msg := range c {
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
	chatLogChan <- msg
}

func logOutgoingMessage(method string, params map[string]string) {
	rawJson, err := json.Marshal(params)
	if err != nil {
		log.Println("Cannot marshal chat log json:", err)
		return
	}
	msg := fmt.Sprintln(time.Now().Unix(), "out", method, string(rawJson))
	chatLogChan <- msg
}
