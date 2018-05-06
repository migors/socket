package storage

import (
	"fmt"
	"sync"

	"github.com/pav5000/socketbot/importer"
	"github.com/pav5000/socketbot/model"
)

var sockets []model.Socket
var socketLock sync.Mutex

func init() {
	var err error
	sockets, err = importer.FromKML("sockets.kml")
	if err != nil {
		panic(err)
	}
	if len(sockets) == 0 {
		panic("0 sockets in DB")
	}
	fmt.Println("Loaded", len(sockets), "sockets")
}

func GetAllSockets() []model.Socket {
	socketLock.Lock()
	defer socketLock.Unlock()
	// returning a copy of slice's header
	// all the data inside the slice is shared, but it's ok since we only read it
	// on data update we just replace the slice with another one
	res := sockets
	return res
}
