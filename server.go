package main

import (
	"fmt"
	"log"
	"net/http"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

// custom rotate logger
type rotateLogWriter struct {
	RotateLogs *rotatelogs.RotateLogs
}

func (w rotateLogWriter) Write(data []byte) (int, error) {
	return w.RotateLogs.Write([]byte(utcMsg(data)))
}

// http server implementation
func server(serverCrt, serverKey string) {
	// define server handlers
	//     base := Config.Base
	//     http.Handle(base+"/css/", http.StripPrefix(base+"/css/", http.FileServer(http.Dir(Config.Styles))))
	//     http.Handle(base+"/js/", http.StripPrefix(base+"/js/", http.FileServer(http.Dir(Config.Jscripts))))
	//     http.Handle(base+"/images/", http.StripPrefix(base+"/images/", http.FileServer(http.Dir(Config.Images))))
	// the request handler
	http.HandleFunc(fmt.Sprintf("%s/status", Config.Base), StatusHandler)
	http.HandleFunc(fmt.Sprintf("%s", Config.Base), DataHandler)

	// start HTTP or HTTPs server based on provided configuration
	addr := fmt.Sprintf(":%d", Config.Port)
	if serverCrt != "" && serverKey != "" {
		//start HTTPS server which require user certificates
		server := &http.Server{Addr: addr}
		log.Printf("Starting HTTPs server on %s%s\n", addr, Config.Base)
		log.Fatal(server.ListenAndServeTLS(serverCrt, serverKey))
	} else {
		// Start server without user certificates
		log.Printf("Starting HTTP server on %s%s\n", addr, Config.Base)
		log.Fatal(http.ListenAndServe(addr, nil))
	}
}
