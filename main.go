package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang/geo/s2"

	"github.com/pav5000/socketbot/db"
	"github.com/pav5000/socketbot/importer"
	"github.com/pav5000/socketbot/tg"
)

const (
	maxSocketDistance         = 50000     // in meters
	earthAvgRadius    float64 = 6371000.0 // in meters
)

func main() {
	sockets, err := importer.FromKML("sockets.kml")
	if err != nil {
		panic(err)
	}
	if len(sockets) == 0 {
		panic("0 sockets in DB")
	}
	fmt.Println("Loaded", len(sockets), "sockets")

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
					log.Println("Error updating user info: ", err)
				}
			}

			if msg.Location != nil {
				fmt.Println("Got location: ", msg.From.Id, msg.From.Username, msg.Location.Latitude, msg.Location.Longitude)

				userLocation := s2.PointFromLatLng(s2.LatLngFromDegrees(msg.Location.Latitude, msg.Location.Longitude))

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
			} else {
				fmt.Println("Got message: ", msg.From.Id, msg.From.Username, msg.Text)
				tg.SendMdMessage("Пришлите мне своё местоположение (точку на карте) и я попытаюсь найти ближайшую к вам публичную розетку. Это пока что первая, тестовая версия бота. Позже будет возможность добавить розетки в базу самому.", msg.From.Id, 0)
				tg.SendVideoByUrl("https://pavl.uk/socketbot/usage.mp4", msg.From.Id, "Пример использования", 0)
			}
		}
	}
}

func formatDistance(meters int64) string {
	if meters > 1000 {
		return fmt.Sprintf("%v км", float64(meters/100)/10)
	} else {
		return fmt.Sprintf("%d м", meters)
	}
}
