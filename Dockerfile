FROM golang:alpine AS builder

LABEL image_author="Michael Prunchak"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY ./config/local.yaml ./config/local.yaml

RUN go build -o read_adviser_bot ./cmd/bot

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/read_adviser_bot .
COPY --from=builder /app/config/local.yaml ./config/local.yaml

ENTRYPOINT ["./read_adviser_bot"]