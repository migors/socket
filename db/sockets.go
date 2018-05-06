package db

import (
	"errors"
	"fmt"
	"time"

	"github.com/pav5000/socketbot/model"
)

func AddSocket(socket model.Socket) (err error) {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			err = errors.New(fmt.Sprint("Got panic: ", p))
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	_, err = tx.Exec(`INSERT INTO sockets (lat, lng, name, description, added_by, last_confirmation) VALUES(?,?,?,?,?,?)`,
		socket.Lat,
		socket.Lng,
		socket.Name,
		socket.Description,
		socket.AddedBy,
		time.Now().Unix())
	if err != nil {
		return err
	}

	var socketId int64
	err = tx.Get(&socketId, `SELECT last_insert_rowid()`)
	if err != nil {
		return err
	}

	for _, mediaId := range socket.Photos {
		_, err := tx.Exec(`INSERT INTO photos (socket,user,added,media_id) VALUES(?,?,?,?)`,
			socketId,
			socket.AddedBy,
			time.Now().Unix(),
			mediaId)
		if err != nil {
			return err
		}
	}

	return nil
}
