#####################################
#   STEP 1 build executable binary  #
#####################################
FROM golang:1.22.2-alpine AS builder

# git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

# no need to filter src as we would drop this builder image anyway
COPY . .

# build the binary.
RUN CGO_ENABLED=0 GOOS=linux go build -o main

#####################################
#   STEP 2 build a small image      #
#####################################
FROM scratch

LABEL org.opencontainers.image.source https://github.com/shrimpsizemoose/card-deck-api

ARG PORT=8080
ENV PORT=$PORT

COPY --from=builder /app/main /app/card-deck-api

ENTRYPOINT ["/app/card-deck-api"]
