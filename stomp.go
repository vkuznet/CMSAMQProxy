package main

import (
	"log"

	stomp "github.com/vkuznet/lb-stomp"
)

// global stomp manager
var stompMgr *stomp.StompManager

func initStompManager() {
	// init stomp manager
	c := stomp.Config{
		URI:         Config.StompURI,
		Login:       Config.StompLogin,
		Password:    Config.StompPassword,
		Iterations:  Config.StompIterations,
		SendTimeout: Config.StompSendTimeout,
		RecvTimeout: Config.StompRecvTimeout,
		Endpoint:    Config.Endpoint,
		ContentType: Config.ContentType,
		Protocol:    Config.Protocol,
		Verbose:     Config.Verbose,
	}
	stompMgr = stomp.New(c)
	log.Println(stompMgr.String())
}
