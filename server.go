package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/yvasiyarov/gorelic"

	"github.com/jprobinson/go-utils/utils"
	"github.com/jprobinson/go-utils/web"
	"github.com/jprobinson/newshound/web/webserver/api"
	"github.com/jprobinson/webserver/api/subway"
)

const (
	serverLog = "/var/log/goserver/server.log"
	accessLog = "/var/log/goserver/access.log"

	houndConfig    = "/opt/newshound/etc/config.json"
	subwayConfig   = "/opt/jp/etc/keys.json"
	newRelicConfig = "/opt/jp/etc/newrelic.json"

	newshoundWeb = "/opt/newshound/www"
	wheresLWeb   = "/home/jp/www/ltrain"
	subwayWeb    = "/home/jp/www/subway"
)

func main() {

	nrconfig := NewConfig(newRelicConfig)
	agent := gorelic.NewAgent()
	agent.Verbose = true
	agent.NewrelicName = "webserver - linode"
	agent.CollectHTTPStat = true
	agent.NewrelicLicense = nrconfig.NewRelicKey
	agent.Run()

	logSetup := utils.NewDefaultLogSetup(serverLog)
	logSetup.SetupLogging()
	go utils.ListenForLogSignal(logSetup)

	router := mux.NewRouter()

	// newshound API setup
	nconfig := NewConfig(houndConfig)
	newshoundAPI := api.NewNewshoundAPI(nconfig.DBURL, nconfig.DBUser, nconfig.DBPassword, agent)
	// add newshound subdomain to webserver
	newshoundRouter := router.Host("newshound.jprbnsn.com").Subrouter()
	// add newshound's API to the subdomain
	newshoundAPIRouter := newshoundRouter.PathPrefix(newshoundAPI.UrlPrefix()).Subrouter()
	newshoundAPI.Handle(newshoundAPIRouter)
	// add newshound UI to to the subdomain
	newshoundRouter.PathPrefix("/").Handler(http.FileServer(http.Dir(newshoundWeb)))

	// add subway stuffs to server
	sconfig := NewConfig(subwayConfig)
	setupSubway(router, sconfig, subwayWeb, "subway.jprbnsn.com", agent)
	setupSubway(router, sconfig, subwayWeb, "wheresthetrain.nyc", agent)
	setupSubway(router, sconfig, subwayWeb, "www.wheresthetrain.nyc", agent)
	setupSubway(router, sconfig, subwayWeb, "wtt.nyc", agent)
	setupSubway(router, sconfig, subwayWeb, "www.wtt.nyc", agent)
	setupSubway(router, sconfig, wheresLWeb, "wheresthel.com", agent)
	setupSubway(router, sconfig, wheresLWeb, "www.wheresthel.com", agent)

	// add the countdown
	countdownRouter := router.Host("countdown.jprbnsn.com").Subrouter()
	countdownRouter.PathPrefix("/").Handler(http.FileServer(http.Dir("/home/jp/www/thecountdown")))

	// add wg4gl
	setupWG4GL(router, "wg4gl.com")
	setupWG4GL(router, "www.wg4gl.com")

	setupJP(router, "jprbnsn.com")
	setupJP(router, "www.jprbnsn.com")

	handler := web.AccessLogHandler(accessLog, router)
	handler = agent.WrapHTTPHandler(handler)

	log.Fatal(http.ListenAndServe(":80", handler))
}

func setupWG4GL(router *mux.Router, host string) {
	wgRouter := router.Host(host).Subrouter()
	wgRouter.PathPrefix("/").Handler(http.FileServer(http.Dir("/home/jp/www/wg4gl")))
}

func setupSubway(router *mux.Router, sconfig *config, www, host string, agent *gorelic.Agent) {
	subwayAPI := subway.NewSubwayAPI(sconfig.SubwayKey, agent)
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
	srouter.PathPrefix("/").Handler(http.FileServer(http.Dir("/home/jp/www/jprbnsn")))
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
