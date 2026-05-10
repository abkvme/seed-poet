.PHONY: build run clean generate help

BINARY_NAME=seedpoet

help:
	@echo "Usage:"
	@echo "  make build      Build the binary"
	@echo "  make run        Build and run the web server"
	@echo "  make clean      Remove binaries"
	@echo "  make generate   Update Go data files from Python scripts"

build:
	go build -o $(BINARY_NAME) ./cmd/bip39poem

run: build
	./$(BINARY_NAME) serve

clean:
	rm -f $(BINARY_NAME)

generate:
	cd scripts && python3 categorize.py
	cd scripts && python3 phonetic_rhymes.py
