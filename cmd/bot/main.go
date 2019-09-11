package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/golang/geo/s2"

	"bitbucket.org/pav5000/socketbot/db"
	"bitbucket.org/pav5000/socketbot/exporter"
	"bitbucket.org/pav5000/socketbot/logger"
	"bitbucket.org/pav5000/socketbot/model"
	"bitbucket.org/pav5000/socketbot/storage"
	"bitbucket.org/pav5000/socketbot/tg"
)

const (
	maxSocketDistance         = 50000     // in meters
	earthAvgRadius    float64 = 6371000.0 // in meters
	MapLink                   = `https://www.google.com/maps/d/u/0/edit?mid=1z_3GfyNZp09HhOFbB5U6YSDr4PY&ll=55.64577355422915,37.757463619459486&z=11`
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

	time.Sleep(time.Second * 3)
	fmt.Println("Waiting for messages")
	updChan := tg.StartCheckingUpdates()
	for update := range updChan {
		if msg, ok := update.(tg.Message); ok {
			{
				err := db.UpdateUserInfo(msg.From)
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

				// Need msg router here, will implement later
				lowerText := strings.ToLower(msg.Text)
				if lowerText == "/add" || strings.HasPrefix(lowerText, "/add ") {
					ReceivedAddCommand(msg)
				} else if lowerText == "/start" || strings.HasPrefix(lowerText, "/start ") {
					db.ClearSessionValues(msg.From.Id)
					db.SetUserState(msg.From.Id, "")
					SendHelp(msg)
				} else if lowerText == "/help" || strings.HasPrefix(lowerText, "/help ") {
					db.ClearSessionValues(msg.From.Id)
					db.SetUserState(msg.From.Id, "")
					SendHelp(msg)
				} else if lowerText == "/map" || strings.HasPrefix(lowerText, "/map ") {
					db.ClearSessionValues(msg.From.Id)
					db.SetUserState(msg.From.Id, "")
					SendMapLink(msg)
				} else if lowerText == "/kml" {
					if msg.From.Id == logger.TgAdminId {
						db.ClearSessionValues(msg.From.Id)
						db.SetUserState(msg.From.Id, "")
						go ReceivedKMLCommand(msg)
					}
				} else {
					if !AddCommandCheck(msg, chatState) {
						tg.SendMdMessage(`Неизвестная мне команда, попробуйте почитать /help`, msg.From.Id, msg.Id)
					}
				}
			}
		}
		if query, ok := update.(tg.CallbackQuery); ok {
			{
				err := db.UpdateUserInfo(query.From)
				if err != nil {
					logger.Err("Error updating user info: ", err)
				}
			}

			var chatState string
			{
				var err error
				chatState, err = db.GetUserState(query.From.Id)
				if err != nil {
					logger.Err("Error getting chat state: ", err)
				}
			}

			if query.Data == `"cancel"` {
				if query.Message.Chat.Type == "private" {
					db.ClearSessionValues(query.Message.Chat.Id)
					db.SetUserState(query.Message.Chat.Id, "")
					tg.EditInlineKeyboard(query.Message.Id, query.Message.Chat.Id, nil)
					tg.SendMdMessage("Отменено", query.Message.Chat.Id, 0)
				}
			} else {
				AddCommandCallback(query, chatState)
			}
		}
	}
}

func SendHelp(msg tg.Message) {
	tg.SendMdMessage("Пришлите мне своё местоположение (точку на карте) и я попытаюсь найти ближайшую к вам публичную розетку.", msg.From.Id, 0)
	tg.SendVideoByUrl("https://pavl.uk/socketbot/usage.mp4", msg.From.Id, "Пример использования", 0)
	tg.SendMdMessage("Другие команды:\n/add - добавить розетку\n/map - получить ссылку на общую карту розеток от Макса\n/help - показать это сообщение", msg.From.Id, 0)
	tg.SendMdMessage("С вопросами, предложениями, критикой обращайтесь к @pav5000", msg.From.Id, 0)
}

func SendMapLink(msg tg.Message) {
	tg.SendMdMessage(`[Карта доступных розеток](`+MapLink+`)`, msg.From.Id, msg.Id)
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
		tg.SendMdMessage("К сожалению, ближайшая розетка на расстоянии более чем "+formatDistance(maxSocketDistance)+" от этого места. Вы можете помочь всем катающимся, если добавите доступные розетки в вашей местности через команду /add", msg.From.Id, msg.Id)
	} else {
		tg.SendMdMessage("Есть розетка в "+formatDistance(metersDist)+" от вас:\n"+closestSocket.Name+"\n"+closestSocket.Description, msg.From.Id, msg.Id)
		tg.SendLocation(closestSocket.Lat, closestSocket.Lng, msg.From.Id, msg.Id)
		if len(closestSocket.Photos) > 0 {
			tg.SendPhotoGroup(closestSocket.Photos, msg.From.Id, msg.Id)
		}
	}
}

func GetCancelButtonRow() []tg.InlineKeyboardButton {
	return []tg.InlineKeyboardButton{{Text: "Отмена", CallbackData: "cancel"}}
}

func ReceivedAddCommand(msg tg.Message) {
	db.ClearSessionValues(msg.From.Id)
	db.SetUserState(msg.From.Id, "add_waiting_location")
	tg.SendMdMessageWithKeyboard(
		`Начинаем добавление новой розетки. Пришлите мне точку на карте, где находится розетка.`,
		msg.From.Id, msg.Id,
		[][]tg.InlineKeyboardButton{GetCancelButtonRow()})
}

func ClearInlineKeyboard(query tg.CallbackQuery) {
	tg.EditInlineKeyboard(query.Message.Id, query.Message.Chat.Id, nil)
}

func AddCommandCallback(query tg.CallbackQuery, chatState string) bool {
	if !strings.HasPrefix(chatState, "add_") {
		return false
	}
	if chatState == "add_waiting_near_photo" {
		ClearInlineKeyboard(query)
		AddCommandFinish(query.From)
		return true
	} else if chatState == "add_waiting_far_photo" {
		ClearInlineKeyboard(query)
		AddCommandFinish(query.From)
		return true
	} else {
		ClearInlineKeyboard(query)
		tg.SendMdMessage(`Невозможно пропустить этот шаг`, query.From.Id, 0)
		return true
	}
	return false
}

func AddCommandCheck(msg tg.Message, chatState string) bool {
	if !strings.HasPrefix(chatState, "add_") {
		return false
	}
	if chatState == "add_waiting_location" {
		if msg.Location == nil {
			tg.SendMdMessage(`Вы не прикрепили точку на карте. Нажмите значок скрепки, выберите "Геопозиция" и перетащите маркер в нужное место.`, msg.From.Id, msg.Id)
		} else {
			db.SetSessionValue(msg.From.Id, "lat", msg.Location.Latitude)
			db.SetSessionValue(msg.From.Id, "lng", msg.Location.Longitude)
			db.SetUserState(msg.From.Id, "add_waiting_name")
			tg.SendMdMessageWithKeyboard(
				`Отлично, теперь напишите название заведения, в котором расположена розетка или какой-нибудь ориентир, кратко. Например: "Бургер Кинг", "Сушивок", "Беседка", "Платформа МЦК"`,
				msg.From.Id, msg.Id,
				[][]tg.InlineKeyboardButton{GetCancelButtonRow()})
		}
		return true
	} else if chatState == "add_waiting_name" {
		if msg.Text == "" {
			tg.SendMdMessage(`Вы ничего не написали, пожалуйста, напишите название места`, msg.From.Id, msg.Id)
		} else {
			db.SetSessionValue(msg.From.Id, "name", msg.Text)
			db.SetUserState(msg.From.Id, "add_waiting_description")
			tg.SendMdMessageWithKeyboard(
				`Отлично, теперь одним сообщением напишите подробнее, где расположена розетка внутри и как её найти.`,
				msg.From.Id, msg.Id,
				[][]tg.InlineKeyboardButton{GetCancelButtonRow()})
		}
		return true
	} else if chatState == "add_waiting_description" {
		if msg.Text == "" {
			tg.SendMdMessage(`Вы ничего не написали, пожалуйста, опишите как найти розетку`, msg.From.Id, msg.Id)
		} else {
			db.SetSessionValue(msg.From.Id, "description", msg.Text)
			db.SetSessionValue(msg.From.Id, "photos", []string{})
			db.SetUserState(msg.From.Id, "add_waiting_near_photo")
			tg.SendMdMessageWithKeyboard(
				`Замечательно, теперь прикрепите фотографию, на которой будет видна сама розетка`,
				msg.From.Id, msg.Id,
				[][]tg.InlineKeyboardButton{
					[]tg.InlineKeyboardButton{{Text: "Пропустить этот шаг", CallbackData: "skip"}},
					GetCancelButtonRow(),
				})
		}
		return true
	} else if chatState == "add_waiting_near_photo" {
		if len(msg.PhotoSizes) == 0 {
			tg.SendMdMessage(`Вы не прислали фотографию. Пожалуйста, прикрепите фотографию, на которой видна сама розетка.`, msg.From.Id, msg.Id)
		} else {
			var photos []string
			db.GetSessionValueJson(msg.From.Id, "photos", &photos)
			photos = append(photos, msg.GetLargestPhoto().FileId)
			db.SetSessionValue(msg.From.Id, "photos", photos)

			db.SetUserState(msg.From.Id, "add_waiting_far_photo")
			tg.SendMdMessageWithKeyboard(
				`Последний шаг, прикрепите обзорное фото входа в заведение.`,
				msg.From.Id, msg.Id,
				[][]tg.InlineKeyboardButton{
					[]tg.InlineKeyboardButton{{Text: "Пропустить этот шаг", CallbackData: "skip"}},
					GetCancelButtonRow(),
				})
		}
		return true
	} else if chatState == "add_waiting_far_photo" {
		if len(msg.PhotoSizes) == 0 {
			tg.SendMdMessage(`Вы не прислали фотографию. Пожалуйста, прикрепите обзорное фото входа в заведение.`, msg.From.Id, msg.Id)
		} else {
			var photos []string
			db.GetSessionValueJson(msg.From.Id, "photos", &photos)
			photos = append(photos, msg.GetLargestPhoto().FileId)
			db.SetSessionValue(msg.From.Id, "photos", photos)

			AddCommandFinish(msg.From)
		}
		return true
	}
	return false
}

func AddCommandFinish(user tg.User) {
	socket := model.Socket{
		Lat:         db.GetSessionValueFloat64(user.Id, "lat"),
		Lng:         db.GetSessionValueFloat64(user.Id, "lng"),
		Name:        db.GetSessionValue(user.Id, "name"),
		Description: db.GetSessionValue(user.Id, "description"),
		AddedBy:     user.Id,
	}

	db.GetSessionValueJson(user.Id, "photos", &socket.Photos)

	err := db.AddSocket(socket)
	if err != nil {
		tg.SendMdMessage(`Произошла внутренняя ошибка, не могу добавить розетку в базу. Попробуйте позже.`, user.Id, 0)
		logger.Err("Error adding socket: ", err)
	} else {
		go storage.UpdateSockets()
		db.SetUserState(user.Id, "")
		db.ClearSessionValues(user.Id)
		tg.SendMdMessage(`Спасибо! Розетка добавлена в базу. В течение суток розетка должна появиться на [карте](`+MapLink+`)`, user.Id, 0)

		tg.SendPlainMessage(
			`Пользователь `+formatUser(user)+" добавил розетку:\n"+socket.Name+"\n"+socket.Description,
			logger.TgAdminId, 0)
		tg.SendLocation(socket.Lat, socket.Lng, logger.TgAdminId, 0)
		tg.SendPhotoGroup(socket.Photos, logger.TgAdminId, 0)
	}
}

func ReceivedKMLCommand(msg tg.Message) {
	defer tg.SendMdMessage("Command finished", msg.From.Id, msg.Id)
	photoRows, err := db.GetAllPhotoRows()
	if err != nil {
		tg.SendPlainMessage("Cannot get all photos: "+err.Error(), msg.From.Id, msg.Id)
		return
	}

	err = os.MkdirAll("data/www", 0777)
	if err != nil {
		tg.SendPlainMessage("Cannot create the folder for photos: "+err.Error(), msg.From.Id, msg.Id)
		return
	}

	downloadCount := 0
	for _, photoRow := range photoRows {
		if photoRow.Url == "" {
			filename := fmt.Sprintf("%d_%d.jpg", photoRow.Socket, photoRow.Id)
			err := tg.GetFile(photoRow.MediaId, "data/www/"+filename)
			if err != nil {
				tg.SendPlainMessage("Cannot download a photo: "+err.Error(), msg.From.Id, msg.Id)
				return
			}
			err = db.SetPhotoUrl(photoRow.Id, filename)
			if err != nil {
				tg.SendPlainMessage("Cannot set photo url: "+err.Error(), msg.From.Id, msg.Id)
				return
			}
			downloadCount++
		}
	}
	tg.SendPlainMessage(fmt.Sprintf("``` Total photos: %d\nWas downloaded: %d ```", len(photoRows), downloadCount), msg.From.Id, msg.Id)
	rawKml, err := exporter.BuildKMLFile()
	if err != nil {
		tg.SendPlainMessage("Cannot build KML file: "+err.Error(), msg.From.Id, msg.Id)
		return
	}

	err = ioutil.WriteFile("data/www/sockets.kml", rawKml, 0666)
	if err != nil {
		tg.SendPlainMessage("Cannot write KML file: "+err.Error(), msg.From.Id, msg.Id)
		return
	}

	tg.SendPlainMessage(exporter.PhotosUrlBase+"sockets.kml", msg.From.Id, msg.Id)
}

func formatDistance(meters int64) string {
	if meters > 1000 {
		return fmt.Sprintf("%v км", float64(meters/100)/10)
	} else {
		return fmt.Sprintf("%d м", meters)
	}
}

func formatUser(user tg.User) string {
	if user.Username != "" {
		return "@" + user.Username
	} else {
		return fmt.Sprintf("@%d (%s %s)", user.Id, user.FirstName, user.LastName)
	}
}
