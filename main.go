package main

import (
	"fmt"

	"github.com/pav5000/socketbot/importer"
	"github.com/pav5000/socketbot/tg"
)

func main() {
	sockets, err := importer.FromKML("sockets.kml")
	if err != nil {
		panic(err)
	}
	fmt.Println("Loaded", len(sockets), "sockets")

	tg.LoadToken("token.txt")

	fmt.Println("Waiting for messages")
	updChan := tg.StartCheckingUpdates()
	for update := range updChan {
		if msg, ok := update.(tg.Message); ok {
			if msg.Location != nil {
				fmt.Println("Got location: ", msg.Location.Latitude, msg.Location.Longitude)
			}
		}
	}
}
