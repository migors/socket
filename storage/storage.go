package storage

import (
	"time"

	"github.com/golang/geo/s2"
)

type Socket struct {
	Name             string
	Description      string
	Photos           []string
	Lat              float64
	Lng              float64
	Point            s2.Point
	AddedBy          string // link to tg user: @username
	LastConfirmation time.Time
}
