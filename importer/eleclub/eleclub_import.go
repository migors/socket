package eleclub

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"strconv"

	"bitbucket.org/pav5000/socketbot/model"
)

const (
	apiUrl = "https://electro.club/api.php?get=sockets"
)

var client = &http.Client{
	Timeout: time.Second * 20,
}

type Socket struct {
	ID     string   `json:"id"`
	Text   string   `json:"text"`
	Lat    string   `json:"lat"`
	Lng    string   `json:"lng"`
	Images []string `json:"images"`
}

func Import() ([]model.Socket, error) {
	res, err := client.Get(apiUrl)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Wrong status code: %d", res.StatusCode)
	}
	rawJson, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	eleclubSockets := []Socket{}
	err = json.Unmarshal(rawJson, &eleclubSockets)
	if err != nil {
		return nil, err
	}

	sockets := make([]model.Socket, 0, len(eleclubSockets))
	for _, eleclubSocket := range eleclubSockets {
		var socket model.Socket
		var err error
		socket.Id, err = strconv.ParseUint(eleclubSocket.ID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("Cannot parse eleclub socket ID '%s': %s", eleclubSocket.ID, err.Error())
		}
		socket.Lat, err = strconv.ParseFloat(eleclubSocket.Lat, 64)
		if err != nil {
			return nil, fmt.Errorf("Cannot parse eleclub socket Lat '%s': %s", eleclubSocket.Lat, err.Error())
		}
		socket.Lng, err = strconv.ParseFloat(eleclubSocket.Lng, 64)
		if err != nil {
			return nil, fmt.Errorf("Cannot parse eleclub socket Lng '%s': %s", eleclubSocket.Lng, err.Error())
		}
		socket.Description = eleclubSocket.Text
		socket.Photos = eleclubSocket.Images
		sockets = append(sockets, socket)
	}
	return sockets, nil
}
