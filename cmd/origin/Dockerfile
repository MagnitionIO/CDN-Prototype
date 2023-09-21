# Environment
FROM alpine:3.18 as environment
RUN apk add --no-cache bash curl busybox-extras vim tcpdump iproute2 mtr

# Build
FROM golang:1.21 as builder
COPY . /usr/local/cdn
WORKDIR /usr/local/cdn
RUN go version && go env -w CGO_ENABLED=0 && \
        mkdir ./bin && \
        cd cmd/origin && go build -o ../../bin/cdn_origin

# Run
FROM environment
COPY --from=builder /usr/local/cdn/bin /usr/local/cdn/bin
WORKDIR /usr/local/cdn
ENV PATH="${PATH}:/usr/local/cdn/bin"

EXPOSE 8080

# CMD ["cdn_origin"]