FROM golang:1.21-alpine AS builder
WORKDIR /build/ics
ADD go.mod go.sum /build/ics/
RUN go mod download
ADD . /build/ics/
RUN go build

FROM alpine
COPY --from=builder /build/ics/ics /usr/local/bin/ics
ENTRYPOINT ["/usr/local/bin/ics"]
