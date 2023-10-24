package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"

	"gorm.io/gorm"
)

type TestTable struct {
	gorm.Model
	Number int
}

type Server struct {
	DB  *gorm.DB
	Val int
}

func New() *Server {
	s := &Server{}
	s.DB = NewDB()
	s.Val = rand.Intn(100000)
	return s
}

func (s *Server) Serve() error {
	http.HandleFunc("/", s.getIndex)

	return http.ListenAndServe(":3333", nil)
}

func (s *Server) getIndex(w http.ResponseWriter, r *http.Request) {

	fmt.Println("http response requested from /")
	s.Val += 1

	t := TestTable{Number: s.Val}
	s.DB.Create(&t)

	io.WriteString(w, fmt.Sprintf("test - %v", s.Val))
}
