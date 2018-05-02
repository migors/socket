package main

import (
	"fmt"
	"github.com/golang/geo/s2"

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

	fmt.Println("Waiting for messages")
	updChan := tg.StartCheckingUpdates()
	for update := range updChan {
		if msg, ok := update.(tg.Message); ok {
			if msg.Location != nil {
				fmt.Println("Got location: ", msg.Location.Latitude, msg.Location.Longitude)

				userLocation := s2.PointFromLatLng(s2.LatLngFromDegrees(msg.Location.Latitude, msg.Location.Longitude))

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
					tg.SendLocation(closestSocket.Lat, closestSocket.Lng, msg.From.Id, msg.Id)
					tg.SendMdMessage("Есть розетка в "+formatDistance(metersDist)+" от вас:\n"+closestSocket.Name+"\n"+closestSocket.Description, msg.From.Id, msg.Id)
				}

			} else {
				tg.SendMdMessage("Пришлите мне своё местоположение и я попытаюсь найти ближайшую к вам публичную розетку", msg.From.Id, 0)
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
