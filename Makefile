build:
	go build -o bin/scraper cmd/scraper/main.go
	go build -o bin/newalg cmd/newalg/main.go
	sass static/sass/main.scss static/css/main.css
clean:
	rm server
	rm profile.out
	rm bin/*
test:
	go test ./... -coverprofile=profile.out -coverpkg=./...
profile:
	go tool cover -html=profile.out
