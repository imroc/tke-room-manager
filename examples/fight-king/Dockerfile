FROM alpine:latest as base

RUN apk add curl

FROM golang:1.22 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -a -o fight-king

FROM base
WORKDIR /
COPY --from=builder /workspace/fight-king .

ENTRYPOINT ["/fight-king"]
