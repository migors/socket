package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/golang/geo/s2"

	"github.com/pav5000/socketbot/db"
	"github.com/pav5000/socketbot/logger"
	"github.com/pav5000/socketbot/model"
	"github.com/pav5000/socketbot/storage"
	"github.com/pav5000/socketbot/tg"
)

const (
	maxSocketDistance         = 50000     // in meters
	earthAvgRadius    float64 = 6371000.0 // in meters
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			logger.Err("Catched panic in main:\n", r)
			time.Sleep(time.Second * 2)
		}
	}()
	tg.LoadToken("token.txt")

	{
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			sig := <-sigs
			fmt.Println()
			log.Println("Got signal", sig)
			log.Println("Closing DB")
			db.Close()
			log.Println("DB closed")
			os.Exit(0)
		}()
	}

	fmt.Println("Waiting for messages")
	updChan := tg.StartCheckingUpdates()
	for update := range updChan {
		if msg, ok := update.(tg.Message); ok {
			{
				err := db.UpdateUserInfo(msg.From.Id, msg.From.FirstName, msg.From.LastName, msg.From.Username, msg.From.LanguageCode, time.Now())
				if err != nil {
					logger.Err("Error updating user info: ", err)
				}
			}

			var chatState string
			{
				var err error
				chatState, err = db.GetUserState(msg.From.Id)
				if err != nil {
					logger.Err("Error getting chat state: ", err)
				}
			}

			if msg.Location != nil {
				if !AddCommandCheck(msg, chatState) {
					ReceivedLocation(msg)
				}
			} else {
				fmt.Println("Got message (state", chatState, "):", msg.From.Id, msg.From.Username, msg.From.FirstName, msg.From.LastName, msg.Text)

				if msg.Text == "/add" || strings.HasPrefix(msg.Text, "/add ") {
					ReceivedAddCommand(msg)
				} else {
					if !AddCommandCheck(msg, chatState) {
						SendHelp(msg)
					}
				}
			}
		}
	}
}

func SendHelp(msg tg.Message) {
	tg.SendMdMessage("Пришлите мне своё местоположение (точку на карте) и я попытаюсь найти ближайшую к вам публичную розетку.", msg.From.Id, 0)
	tg.SendVideoByUrl("https://pavl.uk/socketbot/usage.mp4", msg.From.Id, "Пример использования", 0)
	tg.SendMdMessage("Другие команды:\n/add - добавить розетку", msg.From.Id, 0)
}

func ReceivedLocation(msg tg.Message) {
	logger.Location(fmt.Sprintf("%d @%s (%s %s)", msg.From.Id, msg.From.Username, msg.From.FirstName, msg.From.LastName), msg.Location.Latitude, msg.Location.Longitude)

	userLocation := s2.PointFromLatLng(s2.LatLngFromDegrees(msg.Location.Latitude, msg.Location.Longitude))

	sockets := storage.GetAllSockets()
	// Waiting while PointIndex is beign implemented in golang/geo/s2
	// So, for now just linear scan
	closestSocket := sockets[0]
	minDist := closestSocket.Point.Distance(userLocation).Abs().Normalized()
	for _, socket := range sockets {
		dist := socket.Point.Distance(userLocation).Abs().Normalized()
		if dist < minDist {
			closestSocket = socket
			minDist = dist
		}
	}

	metersDist := int64(earthAvgRadius * float64(minDist))
	if metersDist > maxSocketDistance {
		tg.SendMdMessage("К сожалению, ближайшая розетка на расстоянии более чем "+formatDistance(maxSocketDistance)+" от этого места", msg.From.Id, msg.Id)
	} else {
		tg.SendMdMessage("Есть розетка в "+formatDistance(metersDist)+" от вас:\n"+closestSocket.Name+"\n"+closestSocket.Description, msg.From.Id, msg.Id)
		tg.SendLocation(closestSocket.Lat, closestSocket.Lng, msg.From.Id, msg.Id)
		tg.SendPhotoGroup(closestSocket.Photos, msg.From.Id, msg.Id)
	}
}

func ReceivedAddCommand(msg tg.Message) {
	db.ClearSessionValues(msg.From.Id)
	db.SetUserState(msg.From.Id, "add_waiting_location")
	tg.SendMdMessage(`Начинаем добавление новой розетки. Пришлите мне точку на карте, где находится розетка.`, msg.From.Id, msg.Id)
}

func AddCommandCheck(msg tg.Message, chatState string) bool {
	if chatState == "add_waiting_location" {
		if msg.Location == nil {
			tg.SendMdMessage(`Вы не прикрепили точку на карте. Нажмите значок скрепки, выберите "Геопозиция" и перетащите маркер в нужное место.`, msg.From.Id, msg.Id)
		} else {
			db.SetSessionValue(msg.From.Id, "lat", msg.Location.Latitude)
			db.SetSessionValue(msg.From.Id, "lng", msg.Location.Longitude)
			db.SetUserState(msg.From.Id, "add_waiting_name")
			tg.SendMdMessage(`Отлично, теперь напишите название заведения, в котором расположена розетка или какой-нибудь ориентир, кратко. Например: "Бургер Кинг", "Сушивок", "Беседка", "Платформа МЦК"`, msg.From.Id, msg.Id)
		}
		return true
	} else if chatState == "add_waiting_name" {
		if msg.Text == "" {
			tg.SendMdMessage(`Вы ничего не написали, пожалуйста, напишите название места`, msg.From.Id, msg.Id)
		} else {
			db.SetSessionValue(msg.From.Id, "name", msg.Text)
			db.SetUserState(msg.From.Id, "add_waiting_description")
			tg.SendMdMessage(`Отлично, теперь одним сообщением напишите подробнее, где расположена розетка внутри и как её найти.`, msg.From.Id, msg.Id)
		}
		return true
	} else if chatState == "add_waiting_description" {
		if msg.Text == "" {
			tg.SendMdMessage(`Вы ничего не написали, пожалуйста, опишите как найти розетку`, msg.From.Id, msg.Id)
		} else {
			db.SetSessionValue(msg.From.Id, "description", msg.Text)
			db.SetUserState(msg.From.Id, "add_waiting_near_photo")
			tg.SendMdMessage(`Замечательно, теперь прикрепите фотографию, на которой будет видна сама розетка`, msg.From.Id, msg.Id)
		}
		return true
	} else if chatState == "add_waiting_near_photo" {
		if len(msg.PhotoSizes) == 0 {
			tg.SendMdMessage(`Вы не прислали фотографию. Пожалуйста, прикрепите фотографию, на которой видна сама розетка.`, msg.From.Id, msg.Id)
		} else {
			db.SetSessionValue(msg.From.Id, "photo1", msg.GetLargestPhoto().FileId)
			db.SetUserState(msg.From.Id, "add_waiting_far_photo")
			tg.SendMdMessage(`Последний шаг, прикрепите обзорное фото входа в заведение.`, msg.From.Id, msg.Id)
		}
		return true
	} else if chatState == "add_waiting_far_photo" {
		if len(msg.PhotoSizes) == 0 {
			tg.SendMdMessage(`Вы не прислали фотографию. Пожалуйста, прикрепите обзорное фото входа в заведение.`, msg.From.Id, msg.Id)
		} else {
			socket := model.Socket{
				Lat:         db.GetSessionValueFloat64(msg.From.Id, "lat"),
				Lng:         db.GetSessionValueFloat64(msg.From.Id, "lng"),
				Name:        db.GetSessionValue(msg.From.Id, "name"),
				Description: db.GetSessionValue(msg.From.Id, "description"),
				Photos: []string{
					db.GetSessionValue(msg.From.Id, "photo1"),
					msg.GetLargestPhoto().FileId,
				},
				AddedBy: msg.From.Id,
			}

			err := db.AddSocket(socket)
			if err != nil {
				tg.SendMdMessage(`Произошла внутренняя ошибка, не могу добавить розетку в базу. Попробуйте позже.`, msg.From.Id, msg.Id)
				logger.Err("Error adding socket: ", err)
			} else {
				go storage.UpdateSockets()
				db.SetUserState(msg.From.Id, "")
				db.ClearSessionValues(msg.From.Id)
				tg.SendMdMessage(`Спасибо! Розетка добавлена в базу.`, msg.From.Id, msg.Id)
			}
		}
		return true
	}
	return false
}

func formatDistance(meters int64) string {
	if meters > 1000 {
		return fmt.Sprintf("%v км", float64(meters/100)/10)
	} else {
		return fmt.Sprintf("%d м", meters)
	}
}
