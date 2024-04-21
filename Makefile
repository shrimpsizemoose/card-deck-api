REGISTRY=ghcr.io
NAMESPACE=shrimpsizemoose
APP=toggl-card-service
VERSION=0.1.0
PORT=8088

SOURCES = *.go */*.go
EXE = card-deck-api

SERVICE_TAG=${REGISTRY}/${NAMESPACE}/${APP}:${VERSION}


@docker-build: *.go Dockerfile
	docker build --tag ${SERVICE_TAG} .

.PHONY: @docker-run
@docker-run:
	docker run -e PORT=${PORT} -p ${PORT}:${PORT} ${SERVICE_TAG}

@test: $(SOURCES)
	go test ./...

@test-verbose: $(SOURCES)
	go test -v ./...

@test-race: $(SOURCES)
	go test -race -v ./...

@run: $(SOURCES)
	PORT=${PORT} go run .

@build: $(SOURCES)
	go build -o ${EXE} .

local-debug-run: $(SOURCES)
	DEBUG=1 go run .

lint:
	golangci-lint run ./...

local-http-create-shuffled-deck:
	curl -X POST http://localhost:${PORT}/decks/?shuffle=true

local-http-open-deck:
	@echo specify ID in env DECK_ID or directly hardcode like:
	@echo 'curl -X GET http://localhost:${PORT}/decks/1b4a8074-3c3e-4d0b-bfd5-85ff38ea9d00'
	curl -X GET http://localhost:${PORT}/decks/${DECK_ID}

local-http-open-deck-jq:
	@echo specify ID in env DECK_ID or directly hardcode like:
	@echo 'curl -X GET http://localhost:${PORT}/decks/1b4a8074-3c3e-4d0b-bfd5-85ff38ea9d00'
	curl -X GET http://localhost:${PORT}/decks/${DECK_ID} | jq .

local-http-draw-deck:
	@echo specify ID in env DECK_ID or directly hardcode like:
	@echo 'curl -X GET http://localhost:${PORT}/decks/1b4a8074-3c3e-4d0b-bfd5-85ff38ea9d00'
	curl -X POST http://localhost:${PORT}/decks/${DECK_ID}/draw?count=5

local-http-draw-deck-jq:
	@echo specify ID in env DECK_ID or directly hardcode like:
	@echo 'curl -X GET http://localhost:${PORT}/decks/1b4a8074-3c3e-4d0b-bfd5-85ff38ea9d00'
	curl -X POST http://localhost:${PORT}/decks/${DECK_ID}/draw?count=5 | jq .
