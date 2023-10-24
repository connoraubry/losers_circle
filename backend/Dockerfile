FROM golang:1.21

WORKDIR /app

COPY ./backend/go.mod ./backend/go.sum ./

RUN go mod download

COPY ./backend/*.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-ping

CMD ["/docker-ping"]
