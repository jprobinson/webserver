package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

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
	wheresLWeb   = "/opt/jp/www/ltrain"
	subwayWeb    = "/opt/jp/www/subway"
)

func main() {
	logSetup := utils.NewDefaultLogSetup(serverLog)
	logSetup.SetupLogging()
	go utils.ListenForLogSignal(logSetup)

	router := mux.NewRouter()

	// add subway stuffs to server
	sconfig := NewConfig(subwayConfig)
	setupSubway(router, sconfig, subwayWeb, "subway.jprbnsn.com")
	setupSubway(router, sconfig, subwayWeb, "wheresthetrain.nyc")
	setupSubway(router, sconfig, subwayWeb, "www.wheresthetrain.nyc")
	setupSubway(router, sconfig, subwayWeb, "wtt.nyc")
	setupSubway(router, sconfig, subwayWeb, "www.wtt.nyc")
	setupSubway(router, sconfig, wheresLWeb, "wheresthel.com")
	setupSubway(router, sconfig, wheresLWeb, "www.wheresthel.com")

	// add subway stuffs to server
	// add the countdown
	countdownRouter := router.Host("countdown.jprbnsn.com").Subrouter()
	countdownRouter.PathPrefix("/").Handler(http.FileServer(http.Dir("/opt/jp/www/thecountdown")))

	// add wg4gl
	setupWG4GL(router, "wg4gl.com")
	setupWG4GL(router, "www.wg4gl.com")

	setupColin(router, "colinjhiggins.com")
	setupColin(router, "www.colinjhiggins.com")

	setupJP(router, "jprbnsn.com")
	setupJP(router, "www.jprbnsn.com")

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

	// add newshound barkd websockets
	barkdRouter := router.Host("newshound.jprbnsn.com:8888").Subrouter()
	barkdURL, _ := url.Parse("http://127.0.0.1:8888")
	barkdRouter.PathPrefix("/").Handler(httputil.NewSingleHostReverseProxy(barkdURL))

	handler := web.AccessLogHandler(accessLog, router)

	log.Fatal(http.ListenAndServe(":80", handler))
}

func setupWG4GL(router *mux.Router, host string) {
	wgRouter := router.Host(host).Subrouter()
	wgRouter.PathPrefix("/").Handler(http.FileServer(http.Dir("/opt/jp/www/wg4gl")))
}

func setupSubway(router *mux.Router, sconfig *config, www, host string) {
	subwayAPI := subway.NewSubwayAPI(sconfig.SubwayKey)
	// add subway subdomain to webserver
	subwayRouter := router.Host(host).Subrouter()
	// add subways's API to the subdomain
	subwayAPIRouter := subwayRouter.PathPrefix(subwayAPI.UrlPrefix()).Subrouter()
	subwayAPI.Handle(subwayAPIRouter)
	// add subway UI to to the subdomain...web we have one
	subwayRouter.PathPrefix("/").Handler(http.FileServer(http.Dir(www)))
}

func setupJP(router *mux.Router, host string) {
	srouter := router.Host(host).Subrouter()
	srouter.PathPrefix("/").Handler(http.FileServer(http.Dir("/opt/jp/www/jprbnsn")))
}

func setupColin(router *mux.Router, host string) {
	wgRouter := router.Host(host).Subrouter()
	wgRouter.PathPrefix("/").Handler(http.FileServer(http.Dir("/opt/jp/www/colinjhiggins")))
}

type config struct {
	DBURL      string `json:"db-url"`
	DBUser     string `json:"db-user"`
	DBPassword string `json:"db-pw"`

	SubwayKey string

	NewRelicKey string
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
