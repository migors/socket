package importer

import (
	"encoding/xml"
	"errors"
	"io/ioutil"
	"regexp"
	"strconv"

	"github.com/golang/geo/s2"

	"github.com/pav5000/socketbot/storage"
)

func FromKML(filename string) ([]storage.Socket, error) {
	rawXml, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var parsed struct {
		XMLName  xml.Name `xml:"kml"`
		Document struct {
			Placemarks []struct {
				Name        string `xml:"name"`
				Description string `xml:"description"`
				Coordinates string `xml:"Point>coordinates"`
			} `xml:"Placemark"`
		}
	}

	err = xml.Unmarshal(rawXml, &parsed)
	if err != nil {
		return nil, err
	}

	coordsRe := regexp.MustCompile(`^\s*(\-?\d+(?:\.\d+)?)\,(\-?\d+(?:\.\d+)?)\,`)

	sockets := make([]storage.Socket, 0, len(parsed.Document.Placemarks))
	for _, placemark := range parsed.Document.Placemarks {
		var lat, lng float64
		{
			matches := coordsRe.FindStringSubmatch(placemark.Coordinates)
			if len(matches) < 3 {
				return nil, errors.New("Cannot parse coordinates: '" + placemark.Coordinates + "'")
			}
			var err error
			lat, err = strconv.ParseFloat(matches[2], 64)
			if err != nil {
				return nil, err
			}
			lng, err = strconv.ParseFloat(matches[1], 64)
			if err != nil {
				return nil, err
			}
		}
		sockets = append(sockets, storage.Socket{
			Name:        placemark.Name,
			Description: placemark.Description,
			Lat:         lat,
			Lng:         lng,
			Point:       s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lng)),
		})
	}

	return sockets, nil
}
