FROM golang:alpine AS builder

WORKDIR /build

COPY . .

RUN go build -o gateway gateway.go

FROM alpine

WORKDIR /app

COPY --from=builder /build/gateway /app/gateway

CMD ["./gateway"]