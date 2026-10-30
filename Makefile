BIN_DIR := bin
SEASON ?= $(shell date +%Y)

.PHONY: build cycle stats server test vet clean scrape-cfb

build: cycle stats server

cycle:
	go build -o $(BIN_DIR)/cycle ./cmd/cycle

stats:
	go build -o $(BIN_DIR)/stats ./cmd/stats

server:
	go build -o $(BIN_DIR)/server ./cmd/server

scrape-cfb:
	PYTHONPATH=scraper python -m cfb_scraper $(SEASON) $(ARGS)

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -rf $(BIN_DIR)
