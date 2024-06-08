FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY ./ ./

RUN go mod download 
RUN go build -o ./app ./cmd/url-shortener/main.go

FROM alpine

WORKDIR /app

COPY --from=builder /app/app ./cmd/url-shortener/
COPY --from=builder /app/.env ./.env

CMD ["./cmd/url-shortener/app"]
