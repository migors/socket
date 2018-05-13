package db

import (
	"errors"
	"fmt"
	"time"

	"github.com/pav5000/socketbot/model"
)

type SocketsRow struct {
	Id               uint64  `db:"id"`
	Lat              float64 `db:"lat"`
	Lng              float64 `db:"lng"`
	Name             string  `db:"name"`
	Description      string  `db:"description"`
	AddedBy          uint64  `db:"added_by"`
	LastConfirmation int64   `db:"last_confirmation"`
	Layer            string  `db:"layer"`
}

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

	_, err = tx.Exec(`INSERT INTO sockets (lat, lng, name, description, added_by, last_confirmation, layer) VALUES(?,?,?,?,?,?,?)`,
		socket.Lat,
		socket.Lng,
		socket.Name,
		socket.Description,
		socket.AddedBy,
		time.Now().Unix(),
		"bot")
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

func GetAllSockets() ([]model.Socket, error) {
	rows := []SocketsRow{}
	err := db.Select(&rows, `SELECT * FROM sockets`)
	if err != nil {
		return nil, err
	}

	sockets := make([]model.Socket, 0, len(rows))
	for _, row := range rows {
		photoStrings, err := GetSocketPhotoStrings(row.Id)
		if err != nil {
			return nil, err
		}
		socket := model.Socket{
			Id:               row.Id,
			Name:             row.Name,
			Description:      row.Description,
			Photos:           photoStrings,
			Lat:              row.Lat,
			Lng:              row.Lng,
			AddedBy:          row.AddedBy,
			LastConfirmation: time.Unix(row.LastConfirmation, 0),
			Layer:            row.Layer,
		}
		socket.Init()
		sockets = append(sockets, socket)
	}

	return sockets, nil
}
