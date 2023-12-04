package api

import (
	"net/http"

	"github.com/connoraubry/losers_circle/server/src/db"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type Server struct {
	DB     *gorm.DB
	Val    int
	Router *mux.Router
}

func New() *Server {
	s := &Server{}
	s.DB = db.NewDB(db.Options{})
	s.SetupRouter()
	return s
}

func (s *Server) SetupRouter() {
	r := mux.NewRouter()

	static := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	r.PathPrefix("/static").Handler(static)

	r.HandleFunc("/ping", ping).Methods(http.MethodGet)
	r.HandleFunc("/data", s.Data).Methods(http.MethodGet)
	r.HandleFunc("/teams", s.teams).Methods(http.MethodGet)
	r.HandleFunc("/test", s.Test).Methods(http.MethodGet)
	r.HandleFunc("/cycle", s.GetLargestCircle).Methods(http.MethodGet)

	r.Use(mux.CORSMethodMiddleware(r))
	s.Router = r
}

func (s *Server) Serve() {

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With"})
	originsOk := handlers.AllowedOrigins([]string{"http://localhost:5173", "http://frontend:5173"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	// http.Handle("/", s.Router)
	http.ListenAndServe(":3333", handlers.CORS(originsOk, headersOk, methodsOk)(s.Router))
}
