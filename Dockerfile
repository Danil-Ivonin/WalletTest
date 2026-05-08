FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/wallet-api ./cmd

FROM alpine:3.22

WORKDIR /app

RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /out/wallet-api /app/wallet-api
COPY configs /app/configs
COPY migrations /app/migrations

USER app

EXPOSE 8081

ENTRYPOINT ["/app/wallet-api"]
