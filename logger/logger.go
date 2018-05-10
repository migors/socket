package logger

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pav5000/socketbot/tg"
)

const tgAdmin uint64 = 117576436

func Debug(v ...interface{}) {
	log.Println(v...)
}

func Err(v ...interface{}) {
	ErrStr(fmt.Sprintln(v...))
}

func Errf(format string, v ...interface{}) {
	ErrStr(fmt.Sprintf(format, v...))
}

func ErrStr(text string) {
	log.Println("Err:", text)
	tg.SendMdMessage("Err:```"+text+"```", tgAdmin, 0)
}

func Location(user string, lat float64, lng float64) {
	file, err := os.OpenFile("data/locations.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		Err("Cannot write location log:", err)
	}
	defer file.Close()
	msg := fmt.Sprintln(time.Now().Unix(), lat, lng, user)
	Debug("Received location ", user, lat, lng)
	file.WriteString(msg)
}
