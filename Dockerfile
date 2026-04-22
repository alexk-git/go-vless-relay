FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o vless-server ./cmd/vless-server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/vless-server .
COPY configs/env.properties ./data/.env

RUN mkdir -p /app/data

EXPOSE 443/tcp
EXPOSE 443/udp

CMD ["./vless-server"]
