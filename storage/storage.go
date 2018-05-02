package storage

import (
	"github.com/golang/geo/s2"
)

type Socket struct {
	Name        string
	Description string
	Photos      []string
	Point       s2.Point
}
