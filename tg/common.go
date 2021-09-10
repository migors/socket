package tg

import (
	"bytes"
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

	"bitbucket.org/pav5000/socketbot/config"
)

const (
	retryTimeout = time.Millisecond * 500
	requestRate  = time.Millisecond * 500
)

var requestRateLimiter = time.NewTicker(requestRate)
var token string
var longClient = &http.Client{
	Timeout: time.Second * 20,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		// Proxy:               http.ProxyFromEnvironment,
	},
}
var shortClient = &http.Client{
	Timeout: time.Second * 20,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		// Proxy:               http.ProxyFromEnvironment,
	},
}

var ready sync.WaitGroup

func init() {
	ready.Add(1)
}

func WaitForReady() {
	ready.Wait()
}

func LoadToken() {
	token = strings.TrimSpace(config.Config.TelegramToken)
	if token == "" {
		log.Fatal("telegram_token is empty in config.yml")
	}
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
		criticalLogChan <- fmt.Sprintf("requestWithRetry(%s,...): %s", cmdName, err.Error())
	}
	return err
}

func request(cmdName string, params map[string]string, v interface{}) error {
	urlValues := make(url.Values, len(params))
	for key, value := range params {
		urlValues[key] = []string{value}
	}

	requestBody := bytes.NewBuffer([]byte(urlValues.Encode()))
	req, err := http.NewRequest("GET", "https://api.telegram.org/bot"+token+"/"+cmdName, requestBody)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// log.Println(req.URL.String() + "    " + urlValues.Encode())

	var res *http.Response
	if cmdName == "getUpdates" {
		res, err = longClient.Do(req)
	} else {
		<-requestRateLimiter.C
		res, err = shortClient.Do(req)
	}
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		bodyData, _ := ioutil.ReadAll(res.Body)
		return errors.New(fmt.Sprintf("Wrong status code: %d; Body: %s", res.StatusCode, bodyData))
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
