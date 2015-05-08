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

type SubwayAPI struct {
	key   string
	stops map[string][][]string
}

func NewSubwayAPI(key string) *SubwayAPI {
	return &SubwayAPI{key: key}
}

func (s *SubwayAPI) UrlPrefix() string {
	return "/svc/subway-api/v1"
}

func (s *SubwayAPI) Handle(subRtr *mux.Router) {
	subRtr.HandleFunc("/next-trains/{feed}/{stopId}", s.nextTrains).Methods("GET")

	subRtr.HandleFunc("/stops/{train}", s.getStops).Methods("GET")
}

var ErrFeed = errors.New("sorry! we had problems with the MTA feed.")

const (
	ltrain = "L"
	other  = "123456S"
)

func (s *SubwayAPI) nextTrains(w http.ResponseWriter, r *http.Request) {
	setCommonHeaders(w, r)
	vars := mux.Vars(r)
	stop := vars["stopId"]
	feedType := vars["feed"]
	feed, err := gosubway.GetFeed(s.key, (feedType == ltrain))
	if err != nil {
		web.ErrorResponse(w, ErrFeed, http.StatusBadRequest)
		return
	}

	north, south := feed.NextTrainTimes(stop)
	resp := nextTrainResp{north, south}

	fmt.Fprint(w, web.JsonResponseWrapper{resp})
}

func (s *SubwayAPI) getStops(w http.ResponseWriter, r *http.Request) {
	setCommonHeaders(w, r)
	vars := mux.Vars(r)
	train := vars["tain"]

	fmt.Fprint(w, web.JsonResponseWrapper{s.stops[train]})
}

type nextTrainResp struct {
	Northbound []time.Time `json:"northbound"`
	Southbound []time.Time `json:"southbound"`
}

func setCommonHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", web.JsonContentType)
}
