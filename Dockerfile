FROM golang:1.24 AS builder

WORKDIR /app
COPY . .
RUN go build -o plugin ./cmd/plugin

FROM ubuntu:22.04

# Install required packages
RUN apt-get update && apt-get install -y \
    restic \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Copy the plugin binary
COPY --from=builder /app/plugin /usr/local/bin/

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/plugin"]
