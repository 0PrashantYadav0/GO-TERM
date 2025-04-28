# Stage 1: Build the GO-TERM application
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git for dependency management
RUN apk add --no-cache git

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN go build -o /go/bin/goterm ./cmd/goterm

# Stage 2: Create the runtime container
FROM alpine:latest

RUN apk add --no-cache ca-certificates bash

# Create a non-root user to run the application
RUN adduser -D goterm
USER goterm
WORKDIR /home/goterm

# Copy the binary from the builder stage
COPY --from=builder /go/bin/goterm /usr/local/bin/goterm

# Set up environment variables
ENV HOME=/home/goterm
ENV TERM=xterm-256color
ENV NO_COLOR=

# Create volume mount points for persisting history and configuration
VOLUME ["/home/goterm/.goterm_history", "/home/goterm/.goterm_error", "/home/goterm/.goterm.json"]

# Add entrypoint script
COPY --chown=goterm:goterm docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["goterm"]