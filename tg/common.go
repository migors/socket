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

const (
	retryTimeout = time.Millisecond * 500
)

var token string
var longClient = &http.Client{
	Timeout: time.Second * 60,
}
var shortClient = &http.Client{
	Timeout: time.Second * 5,
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

func requestWithRetry(cmdName string, params map[string]string, v interface{}, retryCount int) error {
	var err error
	for i := 0; i < retryCount; i++ {
		err = request(cmdName, params, v)
		if err == nil {
			return nil
		}
		time.Sleep(retryTimeout * time.Duration(i))
	}
	if err != nil {
		criticalLogChan <- err.Error()
	}
	return err
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

	var res *http.Response
	if cmdName == "getUpdates" {
		res, err = longClient.Do(req)
	} else {
		res, err = shortClient.Do(req)
	}
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

	if _, ok := v.(*Dummy); ok {
	} else {
		err = json.Unmarshal(parsed.Result, v)
		if err != nil {
			return err
		}
	}
	return nil
}
