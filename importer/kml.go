package importer

import (
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/golang/geo/s2"

	"github.com/pav5000/socketbot/model"
)

const mapUrl = "https://www.google.com/maps/d/u/0/kml?mid=1z_3GfyNZp09HhOFbB5U6YSDr4PY&nl=1&lid=fHTGEqWZoeo&forcekml=1&cid=mp&cv=IDQMRld8Ryg.ru."

var client = &http.Client{
	Timeout: time.Second * 20,
}

func FromKMLOnline() ([]model.Socket, error) {
	rawKml, err := Download()
	if err != nil {
		return nil, err
	}
	return FromKML(rawKml)
}

func FromKML(rawKml []byte) ([]model.Socket, error) {
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

	err := xml.Unmarshal(rawKml, &parsed)
	if err != nil {
		return nil, err
	}

	coordsRe := regexp.MustCompile(`^\s*(\-?\d+(?:\.\d+)?)\,(\-?\d+(?:\.\d+)?)\,`)
	tagRemoveRe := regexp.MustCompile(`<[^>]+>`)
	imgRe := regexp.MustCompile(`<img[^>]+src="([^"]*)"`)

	sockets := make([]model.Socket, 0, len(parsed.Document.Placemarks))
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

		photos := []string{}
		{
			matches := imgRe.FindAllStringSubmatch(placemark.Description, -1)
			for _, match := range matches {
				photos = append(photos, match[1])
			}
		}

		sockets = append(sockets, model.Socket{
			Name:        placemark.Name,
			Description: tagRemoveRe.ReplaceAllString(placemark.Description, " "),
			Photos:      photos,
			Lat:         lat,
			Lng:         lng,
			Point:       s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lng)),
		})
	}

	return sockets, nil
}

func Get(url string) ([]byte, error) {
	res, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}

func Download() ([]byte, error) {
	rawXml, err := Get(mapUrl)
	if err != nil {
		return nil, err
	}

	linkRes := struct {
		XMLName xml.Name `xml:"kml"`
		Href    struct {
			Cdata []byte `xml:",cdata"`
		} `xml:"Document>NetworkLink>Link>href"`
	}{}

	err = xml.Unmarshal(rawXml, &linkRes)
	if err != nil {
		return nil, err
	}

	if len(linkRes.Href.Cdata) == 0 {
		return nil, errors.New("Empty link response:\n" + string(rawXml))
	}

	return Get(string(linkRes.Href.Cdata))
}
