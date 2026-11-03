.PHONY: build web test test-go test-js e2e run clean generate images help

BINARY_NAME=bin/seedpoet

help:
	@echo "Usage:"
	@echo "  make build      Build the binary (website embedded)"
	@echo "  make web        Rebuild the website into web/dist"
	@echo "  make test       Go + JavaScript tests"
	@echo "  make e2e        Browser tests against a running server (needs uv)"
	@echo "  make run        Build and serve on 127.0.0.1:8080"
	@echo "  make generate   Regenerate data/seedpoet.json from scripts/"
	@echo "  make images     Re-render icons and the social image (needs uv)"

web:
	go run ./cmd/buildweb

build: web
	go build -o $(BINARY_NAME) ./cmd/seedpoet

test: test-go test-js

test-go:
	go test -race ./...

test-js:
	node --test js/

e2e:
	uv run --with playwright --with pypdf --with mnemonic python web/tools/e2e.py http://localhost:8080

run: build
	./$(BINARY_NAME) serve

clean:
	rm -f $(BINARY_NAME)

generate:
	cd scripts && python3 categorize.py && python3 phonetic_rhymes.py

images:
	uv run --with playwright --with pillow python web/tools/render_images.py
