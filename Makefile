BIN_DIR := bin

.PHONY: build cycle stats test vet clean

build: cycle stats

cycle:
	go build -o $(BIN_DIR)/cycle ./cmd/cycle

stats:
	go build -o $(BIN_DIR)/stats ./cmd/stats

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -rf $(BIN_DIR)
