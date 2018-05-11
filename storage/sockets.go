package storage

import (
	"log"
	"sync"
	"time"

	"github.com/pav5000/socketbot/db"
	"github.com/pav5000/socketbot/importer"
	"github.com/pav5000/socketbot/logger"
	"github.com/pav5000/socketbot/model"
)

var sockets []model.Socket
var socketLock sync.Mutex

func init() {
	go SocketUpdater()
}

func GetAllSockets() []model.Socket {
	socketLock.Lock()
	// returning a copy of slice's header
	// all the data inside the slice is shared, but it's ok since we only read it
	// on data update we just replace the slice with another one
	res := sockets
	socketLock.Unlock()
	return res
}

func SocketUpdater() {
	for {
		UpdateSockets()
		time.Sleep(time.Minute * 10)
	}
}

func UpdateSockets() {
	logger.Debug("Update sockets")
	dbSockets, err := db.GetAllSockets()
	if err != nil {
		logger.Err("Error updating sockets: ", err)
		return
	}
	log.Println("   From DB:", len(dbSockets))
	newSockets := dbSockets

	onlineSockets, err := importer.FromKMLOnline()
	if err != nil {
		logger.Err("Error downloading new kml data: ", err)
	} else {
		log.Println("   From KML:", len(onlineSockets))
		newSockets = append(newSockets, onlineSockets...)
	}

	socketLock.Lock()
	sockets = newSockets
	socketLock.Unlock()
}
