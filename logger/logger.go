package logger

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pav5000/socketbot/tg"
)

const (
	tgAdminId     uint64 = 117576436
	adminMsgPause        = time.Millisecond * 200
)

var adminMsgChan = make(chan string, 30)

func init() {
	go adminNotifier()
}

func adminNotifier() {
	adminMsgChan <- "Logger started"
	time.Sleep(time.Millisecond * 500)
	tg.WaitForReady()
	for msg := range adminMsgChan {
		tg.SendMdMessage("``` "+msg+" ```", tgAdminId, 0)
		time.Sleep(adminMsgPause)
	}
}

func Debug(v ...interface{}) {
	log.Println(v...)
}

func Err(v ...interface{}) {
	ErrStr(fmt.Sprint(v...))
}

func Errf(format string, v ...interface{}) {
	ErrStr(fmt.Sprintf(format, v...))
}

func ErrStr(text string) {
	log.Println("Err:", text)
	select {
	case adminMsgChan <- text:
	default:
	}
}

func Location(user string, lat float64, lng float64) {
	// Thread unsafe because we process incoming messages in single thread
	file, err := os.OpenFile("data/locations.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		Err("Cannot write location log:", err)
	}
	defer file.Close()
	msg := fmt.Sprintln(time.Now().Unix(), lat, lng, user)
	Debug("Received location ", user, lat, lng)
	_, err = file.WriteString(msg)
	if err != nil {
		Err("Cannot write location log:", err)
	}
}
