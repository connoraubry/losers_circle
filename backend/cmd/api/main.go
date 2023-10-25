package main

import (
	"fmt"

	"github.com/connoraubry/losers_circle/backend/tools/server"
)

func main() {
	fmt.Println("vim-go")

	s := server.New()
	s.Serve()
}
