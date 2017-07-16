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
	router := mux.NewRouter()

	// add subway redirects to server
	subway(router, "subway.jprbnsn.com")
	subway(router, "wheresthel.com")
	subway(router, "www.wheresthel.com")

	// add the countdown
	static(router, "countdown.jprbnsn.com", "/opt/jp/www/thecountdown")

	// add colin
	static(router, "colinjhiggins.com", "/opt/jp/www/colinjhiggins")
	static(router, "www.colinjhiggins.com", "/opt/jp/www/colinjhiggins")

	// add personal site
	static(router, "jprbnsn.com", "/opt/jp/www/jprbnsn")
	static(router, "www.jprbnsn.com", "/opt/jp/www/jprbnsn")

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
	logSetup := utils.NewDefaultLogSetup(serverLog)
	logSetup.SetupLogging()
	go utils.ListenForLogSignal(logSetup)

	// https
	certManager := autocert.Manager{
		Prompt: autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(
			"jprbnsn.com", "www.jprbnsn.com",
			"newshound.jprbnsn.com",
			"colinjhiggins.com", "www.colinjhiggins.com",
			"countdown.jprbnsn.com",
			"wheresthel.com", "www.wheresthel.com",
			"subway.jprbnsn.com",
		),
		Cache: autocert.DirCache("certs"),
	}
	server := &http.Server{
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
		Handler: handler,
		Addr:    ":https",
	}
	go func() {
		log.Println("starting https...")
		log.Fatal(server.ListenAndServeTLS("", ""))
	}()

	// http
	log.Println("starting http...")
	log.Fatal(http.ListenAndServe(":http", http.HandlerFunc(https)))
}

func subway(router *mux.Router, host string) {
	router.Host(host).Handler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "http://wheresthetrain.nyc", http.StatusMovedPermanently)
		}),
	)
}

func static(router *mux.Router, host, dir string) {
	srouter := router.Host(host).Subrouter()
	srouter.PathPrefix("/").Handler(http.FileServer(http.Dir(dir)))
}

func https(w http.ResponseWriter, r *http.Request) {
	target := "https://" + r.Host + r.URL.Path
	if len(r.URL.RawQuery) > 0 {
		target += "?" + r.URL.RawQuery
	}
	http.Redirect(w, r, target, http.StatusMovedPermanently)
}

type config struct {
	DBURL      string `json:"db-url"`
	DBUser     string `json:"db-user"`
	DBPassword string `json:"db-pw"`
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
