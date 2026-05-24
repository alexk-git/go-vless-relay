FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

# Copy all source files first
COPY . .

# Initialize Go module if go.mod doesn't exist (e.g. fresh clone),
# then download dependencies
RUN if [ ! -f go.mod ]; then \
        go mod init go-vless-server; \
    fi && \
    go mod tidy && \
    go mod download

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o vless-server ./cmd/vless-server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/vless-server .
COPY configs/env.properties.example ./data/.env

RUN mkdir -p /app/data

EXPOSE 443/tcp
EXPOSE 443/udp

CMD ["./vless-server"]
