package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type ConfigStruct struct {
	// TelegramToken is a bot token for Telegram API
	TelegramToken string `yaml:"telegram_token"`

	// HttpDomain should be filled when you want the bot to
	// host a https site with sockets map
	// if this field is empty, https server won't run
	HttpDomain string `yaml:"http_domain"`

	// HttpListen should contain [address]:<port>
	// example: ":80" or "127.0.0.1:80"
	HttpListen string `yaml:"http_listen"`

	// HttpsListen should contain [address]:<port>
	// example: ":443" or "127.0.0.1:443"
	HttpsListen string `yaml:"https_listen"`

	// GoogleToken is a token for Google maps API
	GoogleToken string `yaml:"google_token"`

	// KMLDownload enables feature of fetching sockets from custom google map
	// it's a legacy feature
	KMLDownload bool `yaml:"kml_download"`
}

var Config ConfigStruct

// MapLink link to a map hosted by https server
var MapLink string

func init() {
	rawYML, err := ioutil.ReadFile("./data/config.yml")
	if err != nil {
		log.Fatal("error reading config.yml:", err)
	}

	err = yaml.Unmarshal(rawYML, &Config)
	if err != nil {
		log.Fatal("error unmarshaling config:", err)
	}

	if Config.HttpListen != "" {
		MapLink = "https://" + Config.HttpsListen
	}
}
