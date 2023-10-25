package scraper

import "fmt"

func main() {
	fmt.Println("vim-go")
	s := New(false)

	s.ScrapeYear(2023)
	fmt.Println(len(s.Games), s.Games)
}
