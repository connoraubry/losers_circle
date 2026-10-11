BIN_DIR := bin
SEASON ?= $(shell date +%Y)

.PHONY: build cycle stats test vet clean scrape-cfb

build: cycle stats

cycle:
	go build -o $(BIN_DIR)/cycle ./cmd/cycle

stats:
	go build -o $(BIN_DIR)/stats ./cmd/stats

scrape-cfb:
	PYTHONPATH=scraper python -m cfb_scraper $(SEASON) $(ARGS)

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -rf $(BIN_DIR)
