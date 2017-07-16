package main

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/jprobinson/go-utils/utils"
	"github.com/jprobinson/go-utils/web"
	"golang.org/x/crypto/acme/autocert"

	"github.com/jprobinson/newshound/web/webserver/api"
)

const (
	serverLog = "/var/log/goserver/server.log"
	accessLog = "/var/log/goserver/access.log"

	houndConfig = "/opt/newshound/etc/config.json"

	newshoundWeb = "/opt/newshound/www"
)

func main() {
	logSetup := utils.NewDefaultLogSetup(serverLog)
	logSetup.SetupLogging()
	go utils.ListenForLogSignal(logSetup)

	router := mux.NewRouter()

	// add subway redirects to server
	setupSubway(router, "subway.jprbnsn.com")
	setupSubway(router, "wheresthel.com")
	setupSubway(router, "www.wheresthel.com")

	// add the countdown
	countdownRouter := router.Host("countdown.jprbnsn.com").Subrouter()
	countdownRouter.PathPrefix("/").Handler(http.FileServer(http.Dir("/opt/jp/www/thecountdown")))

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

	certManager := autocert.Manager{
		Prompt: autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(
			"jprbnsn.com", "www.jprbnsn.com",
			"newshound.jprbnsn.com",
			"colinjhiggins.com", "www.colinjhiggins.com",
		),
		Cache: autocert.DirCache("certs"),
	}

	// http
	go log.Fatal(http.ListenAndServe(":80", handler))

	// https
	server := &http.Server{
		Addr: ":443",
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
	}
	log.Fatal(server.ListenAndServeTLS("", ""))
}

func setupSubway(router *mux.Router, host string) {
	router.Host(host).Handler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "http://wheresthetrain.nyc", http.StatusMovedPermanently)
		}),
	)
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
