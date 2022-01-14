package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

// Configuration stores server configuration parameters
type Configuration struct {

	// HTTP server configuration options
	Port      int    `json:"port"`      // server port number
	Hmac      string `json:"hmac"`      // hmac string
	Base      string `json:"base"`      // base URL
	Verbose   int    `json:"verbose"`   // verbose output
	ServerCrt string `json:"serverCrt"` // path to server crt file
	ServerKey string `json:"serverKey"` // path to server key file
	LogFile   string `json:"logFile"`   // log file name

	// Stomp configuration options
	StompURI         string `json:"stompURI"`         // StompAMQ URI
	StompLogin       string `json:"stompLogin"`       // StompAQM login name
	StompPassword    string `json:"stompPassword"`    // StompAQM password
	StompIterations  int    `json:"stompIterations"`  // Stomp iterations
	StompSendTimeout int    `json:"stompSendTimeout"` // heartbeat send timeout
	StompRecvTimeout int    `json:"stompRecvTimeout"` // heartbeat recv timeout
	Endpoint         string `json:"endpoint"`         // StompAMQ endpoint
	ContentType      string `json:"contentType"`      // ContentType of UDP packet
	Protocol         string `json:"protocol"`         // protocol to use in stomp, e.g. tcp, tcp4 or tcp6
	Producer         string `json:"producer"`         // producer name

	// CMS options
	CMSRole  string `json:"cms_role"`  // CMS role
	CMSGroup string `json:"cms_group"` // CMS group
	CMSSite  string `json:"cms_site"`  // CMS site
}

func (c *Configuration) String() string {
	msg := fmt.Sprintf("<Config: role=%s group=%s site=%s port=%d stormURI=%s>", c.CMSRole, c.CMSGroup, c.CMSSite, c.Port, c.StompURI)
	return msg
}

// Config variable represents configuration object
var Config Configuration

// helper function to parse configuration
func parseConfig(configFile string) error {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Println("Unable to read", err)
		return err
	}
	err = json.Unmarshal(data, &Config)
	if err != nil {
		log.Println("Unable to parse", err)
		return err
	}
	if Config.StompIterations == 0 {
		Config.StompIterations = 3 // number of Stomp attempts
	}
	if Config.ContentType == "" {
		Config.ContentType = "application/json"
	}
	if Config.Protocol == "" {
		Config.Protocol = "tcp4"
	}
	if Config.StompSendTimeout == 0 {
		Config.StompSendTimeout = 1000 // miliseconds
	}
	if Config.StompRecvTimeout == 0 {
		Config.StompRecvTimeout = 1000 // miliseconds
	}
	if Config.Producer == "" {
		log.Fatal("Wrong configuration, producer is missing")
	}
	if Config.StompURI == "" {
		log.Fatal("Wrong configuration, StompURI is missing")
	}
	if Config.StompLogin == "" {
		log.Fatal("Wrong configuration, StompLogin is missing")
	}
	if Config.StompPassword == "" {
		log.Fatal("Wrong configuration, StompPassword is missing")
	}
	return nil
}
