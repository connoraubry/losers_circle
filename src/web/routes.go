package web

import (
	"net/http"

	"github.com/connoraubry/losers_circle/src/db"
	"github.com/connoraubry/losers_circle/src/tools"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Server struct {
	DB     *gorm.DB
	Router *mux.Router
}

func New(useDB bool) *Server {
	log.WithField("useDB", useDB).Info("Starting new server")
	s := &Server{}

	if useDB {
		s.DB = db.NewDB(db.Options{})
	}
	s.SetupRouter()
	return s
}
func (s *Server) SetupRouter() {

	r := mux.NewRouter()

	static := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	r.PathPrefix("/static").Handler(static)

	//Static flies
	r.HandleFunc("/", handler)
	r.HandleFunc("/test", s.testAPI)
	//r.HandleFunc("/{LEAGUE}", handleLeague)
	nfl := r.PathPrefix("/nfl").Subrouter()
	nfl.HandleFunc("/", handler)
	nfl.HandleFunc("", handler)
	nfl.HandleFunc("/{year}", nflHandler)

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/testMatchup", testReplaceMatchup)

	s.Router = r
}

func (s *Server) Run() {

	http.Handle("/", s.Router)

	log.Info("Starting server on port 3030")
	log.Fatal(http.ListenAndServe(":3030", nil))
}

func (s *Server) testAPI(w http.ResponseWriter, r *http.Request) {
	log.Info("Handler called", "testAPI")
	htmlTest := "<span>You clicked it!</span>"
	w.Write([]byte(htmlTest))
}

func testReplaceMatchup(w http.ResponseWriter, r *http.Request) {
	log.Info("Handler called", "testReplaceMatchup")
	w.Write(tools.GenerateMatchups())
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Info("Calling handler", "handler")
	w.Write(tools.GenerateMain())
}

func nflHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("Calling handler", "nflHandler")

	vars := mux.Vars(r)
	games := r.URL.Query().Get("games")

	val := vars["year"]
	log.WithField("year", val).Info("Got Year")
	if games != "" {
		log.WithField("games", games).Info("Got games string")
	}
	w.Write(tools.GenerateMain())
}
