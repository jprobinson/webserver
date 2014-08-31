package subway

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jprobinson/go-utils/web"
	"github.com/jprobinson/gosubway"
)

type SubwayAPI struct{}

func (s *SubwayAPI) UrlPrefix() string {
	return "/svc/subway-api/v1"
}

func (s *SubwayAPI) Handle(subRtr *mux.Router) {
	subRtr.HandleFunc("/next-trains/{stopId}/{key}", nextTrains).Methods("GET")
}

var ErrFeed = errors.New("sorry! we had problems with the MTA feed.")

func nextTrains(w http.ResponseWriter, r *http.Request) {
	setCommonHeaders(w, r)
	vars := mux.Vars(r)
	stop := vars["stopId"]
	key := vars["key"]

	feed, err := gosubway.GetFeed(key)
	if err != nil {
		web.ErrorResponse(w, ErrFeed, http.StatusBadRequest)
		return
	}

	north, south := feed.NextTrains(stop)
	resp := nextTrainResp{north, south}

	fmt.Fprint(w, web.JsonResponseWrapper{resp})

}

type nextTrainResp struct {
	Northbound time.Duration `json:"northbound"`
	Southbound time.Duration `json:"southbound"`
}

func setCommonHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if len(origin) == 0 {
		origin = "*"
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, *")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", web.JsonContentType)
}
