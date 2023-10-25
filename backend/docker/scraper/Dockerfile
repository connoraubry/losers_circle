FROM golang:1.21

WORKDIR /app

COPY ./backend/ ./

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /scraper cmd/scraper/main.go

CMD ["/scraper"]
