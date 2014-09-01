package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/jprobinson/go-utils/utils"
	"github.com/jprobinson/go-utils/web"
	"github.com/jprobinson/newshound/web/webserver/api"
	"github.com/jprobinson/webserver/api/subway"
)

const (
	serverLog = "/var/log/goserver/server.log"
	accessLog = "/var/log/goserver/access.log"

	houndConfig  = "/opt/newshound/etc/config.json"
	subwayConfig = "/opt/jp/etc/keys.json"

	newshoundWeb = "/opt/newshound/www"
	subwayWeb    = "/home/jp/www/subway"
)

func main() {
	logSetup := utils.NewDefaultLogSetup(serverLog)
	logSetup.SetupLogging()
	go utils.ListenForLogSignal(logSetup)

	router := mux.NewRouter()

	// newshound API setup
	nconfig := NewConfig(houndConfig)
	newshoundAPI := api.NewNewshoundAPI(nconfig.DBURL, nconfig.DBUser, nconfig.DBPassword)
	// add newshound subdomain to webserver
	newshoundRouter := router.Host("newshound.jprbnsn.com").Subrouter()
	// add newshound's API to the subdomain
	newshoundAPIRouter := newshoundRouter.PathPrefix(newshoundAPI.UrlPrefix()).Subrouter()
	newshoundAPI.Handle(newshoundAPIRouter)
	// add newshound UI to to the subdomain
	newshoundRouter.PathPrefix("/").Handler(http.FileServer(http.Dir(newshoundWeb)))

	// add subway stuffs to server
	sconfig := NewConfig(subwayConfig)
	setupSubway(router, sconfig, "subway.jprbnsn.com")
	setupSubway(router, sconfig, "wheresthel.com")
	setupSubway(router, sconfig, "www.wheresthel.com")

	// add the countdown
	countdownRouter := router.Host("countdown.jprbnsn.com").Subrouter()
	countdownRouter.PathPrefix("/").Handler(http.FileServer(http.Dir("/home/jp/www/thecountdown")))

	jpRouter := router.Host("jprbnsn.com").Subrouter()
	jpRouter.PathPrefix("/").Handler(http.FileServer(http.Dir("/home/jp/www/jprbnsn")))
	wwwJPRouter := router.Host("www.jprbnsn.com").Subrouter()

	wwwJPRouter.PathPrefix("/").Handler(http.FileServer(http.Dir("/home/jp/www/jprbnsn")))
	handler := web.AccessLogHandler(accessLog, router)

	log.Fatal(http.ListenAndServe(":80", handler))
}

func setupSubway(router *mux.Router, sconfig *config, host string) {
	subwayAPI := subway.NewSubwayAPI(sconfig.SubwayKey)
	// add subway subdomain to webserver
	subwayRouter := router.Host(host).Subrouter()
	// add subways's API to the subdomain
	subwayAPIRouter := subwayRouter.PathPrefix(subwayAPI.UrlPrefix()).Subrouter()
	subwayAPI.Handle(subwayAPIRouter)
	// add subway UI to to the subdomain...web we have one
	subwayRouter.PathPrefix("/").Handler(http.FileServer(http.Dir(subwayWeb)))
}

type config struct {
	DBURL      string `json:"db-url"`
	DBUser     string `json:"db-user"`
	DBPassword string `json:"db-pw"`

	SubwayKey string
}

func NewConfig(filename string) *config {
	config := config{}

	readBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Cannot read config file: %s %s", filename, err)
	}

	err = json.Unmarshal(readBytes, &config)
	if err != nil {
		log.Fatalf("Cannot parse JSON in config file: %s %s", filename, err)
	}

	return &config
}
