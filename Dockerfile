# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/server ./main.go

FROM alpine:3.22
WORKDIR /app

RUN apk add --no-cache ca-certificates
COPY --from=builder /app/bin/server /app/server

RUN mkdir -p /app/uploads

EXPOSE 4009

CMD ["/app/server"]
