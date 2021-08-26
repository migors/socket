package model

import (
	"time"

	"github.com/golang/geo/s2"
)

type Socket struct {
	Id               uint64
	EleclubID        uint64
	Name             string
	Description      string
	Photos           []string
	Lat              float64
	Lng              float64
	Point            s2.Point
	AddedBy          uint64
	LastConfirmation time.Time
	Layer            string
}

func (s *Socket) Init() {
	s.Point = s2.PointFromLatLng(s2.LatLngFromDegrees(s.Lat, s.Lng))
}
