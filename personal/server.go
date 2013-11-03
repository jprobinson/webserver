package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/jprobinson/go-utils/utils"
	"github.com/jprobinson/go-utils/web"
)

const (
	serverLog = "/var/log/goserver/server.log"
	accessLog = "/var/log/goserver/access.log"
)

func main() {
	logSetup := &utils.DefaultLogSetup{LogFile: serverLog}
	logSetup.SetupLogging()
	go utils.ListenForLogSignal(logSetup)

	router := mux.NewRouter()

	jpRouter := router.Host("jprbnsn.com").Subrouter()
	jpRouter.PathPrefix("/").Handler(http.FileServer(http.Dir("/home/jp/www")))

	wwwJPRouter := router.Host("www.jprbnsn.com").Subrouter()
	wwwJPRouter.PathPrefix("/").Handler(http.FileServer(http.Dir("/home/jp/www")))

	apiHandler := web.AccessLogHandler(accessLog, router)

	http.Handle("/", apiHandler)

	log.Fatal(http.ListenAndServe(":80", apiHandler))
}
