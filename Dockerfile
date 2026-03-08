FROM golang:1.26.1-alpine AS builder
WORKDIR /bot
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN GOOS=linux go build -o onlineSteamGE .

FROM alpine:latest
WORKDIR /bot/
COPY --from=builder /bot/onlineSteamGE .
CMD ["./onlineSteamGE"]