FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download
COPY *.go /app/
RUN go build

FROM alpine
RUN apk add --no-cache tzdata
COPY --from=builder /app/ics /usr/local/bin/ics
ENTRYPOINT ["/usr/local/bin/ics"]
