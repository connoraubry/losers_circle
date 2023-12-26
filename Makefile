build:
	go build -o bin/server cmd/server/main.go
	go build -o bin/generator cmd/generator/main.go
	go build -o bin/scraper cmd/scraper/main.go
	sass static/sass/main.scss static/css/main.css
clean:
	rm server
test:
	go test ./...
