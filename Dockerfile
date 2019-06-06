FROM golang:alpine as builder
MAINTAINER Jack Murdock <jack_murdock@comcast.com>

# build the binary
WORKDIR /go/src
RUN apk add --update --repository https://dl-3.alpinelinux.org/alpine/edge/testing/ git curl
RUN curl https://glide.sh/get | sh
COPY src/ /go/src/

RUN glide -q install --strip-vendor
RUN go build -o xmidt-issuer_linux_amd64 xmidt-issuer

EXPOSE 6000 6001 6002
RUN mkdir -p /etc/xmidt-issuer
VOLUME /etc/xmidt-issuer

# the actual image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN mkdir -p /etc/xmidt-issuer
VOLUME /etc/xmidt-issuer
WORKDIR /root/
COPY --from=builder /go/src/xmidt-issuer_linux_amd64 .
ENTRYPOINT ["./xmidt-issuer_linux_amd64"]
