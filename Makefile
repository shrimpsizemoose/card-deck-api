REGISTRY=ghcr.io
NAMESPACE=shrimpsizemoose
APP=toggl-card-service
VERSION=0.1.0
PORT=8088

SOURCES = *.go
EXE = card-deck-api

SERVICE_TAG=${REGISTRY}/${NAMESPACE}/${APP}:${VERSION}


@docker-build: *.go Dockerfile
	docker build --tag ${SERVICE_TAG} .

.PHONY: @docker-run
@docker-run:
	docker run -e PORT=${PORT} ${SERVICE_TAG}

@test: $(SOURCES)
	go test -v .

@run: $(SOURCES)
	PORT=${PORT} go run .

@build: $(SOURCES)
	go build -o ${EXE} .

@test-utils: $(SOURCES)
	go test -v . 

local-debug-run: $(SOURCES)
	DEBUG=1 go run .

local-http-create-deck:
	curl -X POST http://localhost:${PORT}/decks/?shuffle=true

local-http-open-deck-jq:
	@echo specify ID in env DECK_ID or directly hardcode like:
	@echo 'curl -X GET http://localhost:${PORT}/decks/1b4a8074-3c3e-4d0b-bfd5-85ff38ea9d00'
	curl -X GET http://localhost:${PORT}/decks/${DECK_ID} | jq .


local-http-draw-deck-jq:
	@echo specify ID in env DECK_ID or directly hardcode like:
	@echo 'curl -X GET http://localhost:${PORT}/decks/1b4a8074-3c3e-4d0b-bfd5-85ff38ea9d00'
	curl -X POST http://localhost:${PORT}/decks/${DECK_ID}/draw?count=5 | jq .
