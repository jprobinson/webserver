package main

import (
	"encoding/json"
	"fmt"
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

	configFile   = "/opt/newshound/etc/config.json"
	newshoundWeb = "/opt/newshound/www"
)

func main() {
	logSetup := utils.NewDefaultLogSetup(serverLog)
	logSetup.SetupLogging()
	go utils.ListenForLogSignal(logSetup)

	router := mux.NewRouter()

	// newshound API setup
	config := NewConfig()
	newshoundAPI := api.NewNewshoundAPI(config.DBURL, config.DBUser, config.DBPassword)
	// add newshound subdomain to webserver
	newshoundRouter := router.Host("newshound.jprbnsn.com").Subrouter()
	// add newshound's API to the subdomain
	newshoundAPIRouter := newshoundRouter.PathPrefix(newshoundAPI.UrlPrefix()).Subrouter()
	newshoundAPI.Handle(newshoundAPIRouter)
	// add newshound UI to to the subdomain
	newshoundRouter.PathPrefix("/").Handler(http.FileServer(http.Dir(newshoundWeb)))

	// add subway stuffs to server
	subwayAPI := &subway.SubwayAPI{}
	// add subway subdomain to webserver
	subwayRouter := router.Host("subway.jprbnsn.com").Subrouter()
	// add subways's API to the subdomain
	subwayAPIRouter := subwayRouter.PathPrefix(subwayAPI.UrlPrefix()).Subrouter()
	subwayAPI.Handle(subwayAPIRouter)
	// add subway UI to to the subdomain...web we have one
	//subwayRouter.PathPrefix("/").Handler(http.FileServer(http.Dir(subwayWeb)))

	// add the countdown
	countdownRouter := router.Host("countdown.jprbnsn.com").Subrouter()
	countdownRouter.PathPrefix("/").Handler(http.FileServer(http.Dir("/home/jp/www/thecountdown")))

	jpRouter := router.Host("jprbnsn.com").Subrouter()
	subwayAPIRouter1 := jpRouter.PathPrefix(subwayAPI.UrlPrefix()).Subrouter()
	subwayAPI.Handle(subwayAPIRouter1)

	jpRouter.PathPrefix("/").Handler(http.FileServer(http.Dir("/home/jp/www/jprbnsn")))
	wwwJPRouter := router.Host("www.jprbnsn.com").Subrouter()
	subwayAPIRouter2 := wwwJPRouter.PathPrefix(subwayAPI.UrlPrefix()).Subrouter()
	subwayAPI.Handle(subwayAPIRouter2)

	wwwJPRouter.PathPrefix("/").Handler(http.FileServer(http.Dir("/home/jp/www/jprbnsn")))
	handler := web.AccessLogHandler(accessLog, router)

	log.Fatal(http.ListenAndServe(":80", handler))
}

type config struct {
	DBURL      string `json:"db-url"`
	DBUser     string `json:"db-user"`
	DBPassword string `json:"db-pw"`
}

func NewConfig() *config {
	config := config{}

	readBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		panic(fmt.Sprintf("Cannot read config file: %s %s", config, err))
	}

	err = json.Unmarshal(readBytes, &config)
	if err != nil {
		panic(fmt.Sprintf("Cannot parse JSON in config file: %s %s", config, err))
	}

	return &config
}
