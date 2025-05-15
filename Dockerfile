FROM golang:1.24 AS builder

WORKDIR /app

COPY src ./
RUN go mod download

RUN go build -o go_rest_wallets ./

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/go_rest_wallets .

EXPOSE 8080

CMD ["./go_rest_wallets"]
