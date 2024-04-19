#####################################
#   STEP 1 build executable binary  #
#####################################
FROM golang:1.22.2-alpine AS builder

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

# No need to filter src as we would drop this builder image anyway
COPY . .

# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux go build -o main

#####################################
#   STEP 2 build a small image      #
#####################################
FROM scratch

LABEL org.opencontainers.image.source https://github.com/shrimpsizemoose/toggl-card-deck-api

ARG PORT=8080
ENV PORT=$PORT

# Copy our static executable.
COPY --from=builder /app/main /app/card-deck-api

# Run the hello binary.
ENTRYPOINT ["/app/card-deck-api"]
