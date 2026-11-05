BIN_DIR := bin
SEASON ?= $(shell date +%Y)

.PHONY: build cycle stats buildsite test test-go test-js vet clean scrape-cfb scrape-cbb

build: cycle stats buildsite

cycle:
	go build -o $(BIN_DIR)/cycle ./cmd/cycle

stats:
	go build -o $(BIN_DIR)/stats ./cmd/stats

buildsite:
	go build -o $(BIN_DIR)/buildsite ./cmd/buildsite

scrape-cfb:
	PYTHONPATH=scraper python -m cfb_scraper $(SEASON) $(ARGS)

scrape-cbb:
	PYTHONPATH=scraper python -m cbb_scraper $(SEASON) $(ARGS)

test: test-go test-js

test-go:
	go test ./...

test-js:
	node --test internal/webgraph/web/*.test.js

vet:
	go vet ./...

clean:
	rm -rf $(BIN_DIR)
