package tg

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var token string
var client = &http.Client{
	Timeout: time.Second * 60,
}

var ready sync.WaitGroup

func init() {
	ready.Add(1)
}

func WaitForReady() {
	ready.Wait()
}

func LoadToken(filename string) {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Cannot read token file", err)
	}
	token = strings.TrimSpace(string(raw))
	ready.Done()
}

type Response struct {
	Ok     bool            `json:"ok"`
	Result json.RawMessage `json:"result"`
}

func request(cmdName string, params map[string]string, v interface{}) error {
	urlValues := make(url.Values, len(params))
	for key, value := range params {
		urlValues[key] = []string{value}
	}

	req, err := http.NewRequest("GET", "https://api.telegram.org/bot"+token+"/"+cmdName+"?"+urlValues.Encode(), nil)
	if err != nil {
		return err
	}
	// log.Println(req.URL.String())

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("Wrong status code: %d", res.StatusCode))
	}

	rawJson, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	parsed := Response{}
	err = json.Unmarshal(rawJson, &parsed)
	if err != nil {
		return err
	}

	if !parsed.Ok {
		return errors.New("Result returned ok:false")
	}

	err = json.Unmarshal(parsed.Result, v)
	if err != nil {
		return err
	}

	return nil
}
