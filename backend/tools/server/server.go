package server

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"

	"github.com/connoraubry/losers_circle/backend/tools/db"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Server struct {
	DB  *gorm.DB
	Val int
}

func New() *Server {
	s := &Server{}
	s.DB = db.NewDB(db.Options{})
	s.Val = rand.Intn(100000000)
	return s
}

func (s *Server) Serve() error {
	http.HandleFunc("/", s.getIndex)

	return http.ListenAndServe(":3333", nil)
}

func (s *Server) getIndex(w http.ResponseWriter, r *http.Request) {

	log.Info("http response requested from /")
	s.Val += 1
	io.WriteString(w, fmt.Sprintf("test - %v", s.Val))
}
