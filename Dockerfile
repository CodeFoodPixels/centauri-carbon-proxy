FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . ./
RUN go build -o centauri-carbon-proxy

FROM alpine:latest

COPY --from=builder /app/centauri-carbon-proxy .
EXPOSE 3000
CMD ["/centauri-carbon-proxy"]