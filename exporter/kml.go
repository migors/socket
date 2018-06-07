package exporter

import (
	"bytes"
	"encoding/xml"
	"fmt"

	"bitbucket.org/pav5000/socketbot/db"
)

const PhotosUrlBase = "https://pavl.uk/sockets/"

type Placemark struct {
	Name        string `xml:"name"`
	Description string `xml:"description"`
	Coordinates string `xml:"Point>coordinates"`
}

type KML struct {
	XMLName    xml.Name    `xml:"Document"`
	Placemarks []Placemark `xml:"Placemark"`
}

func BuildKMLFile() ([]byte, error) {
	sockets, err := db.GetAllSockets()
	if err != nil {
		return nil, err
	}

	kml := KML{}
	kml.Placemarks = make([]Placemark, 0, len(sockets))
	for _, socket := range sockets {
		photos, err := db.GetSocketPhotoUrls(socket.Id)
		if err != nil {
			return nil, err
		}

		imgStr := bytes.NewBuffer([]byte{})
		for _, photoUrl := range photos {
			imgStr.WriteString(`<br><img src="`)
			imgStr.WriteString(PhotosUrlBase)
			imgStr.WriteString(photoUrl)
			imgStr.WriteString(`" height="200" width="auto" />`)
		}

		kml.Placemarks = append(kml.Placemarks, Placemark{
			Name:        socket.Name,
			Description: socket.Description + string(imgStr.Bytes()),
			Coordinates: fmt.Sprintf("%v,%v", socket.Lng, socket.Lat),
		})

	}

	rawKml, err := xml.Marshal(kml)
	if err != nil {
		return nil, err
	}
	return []byte(`<?xml version="1.0" encoding="UTF-8"?><kml xmlns="http://www.opengis.net/kml/2.2">` + string(rawKml) + `</kml>`), nil
}
