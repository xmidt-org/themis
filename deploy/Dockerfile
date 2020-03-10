FROM docker.io/library/golang:1.14-alpine as builder

MAINTAINER Jack Murdock <jack_murdock@comcast.com>

WORKDIR /go/src/github.com/xmidt-org/themis

ARG VERSION=unknown
ARG GITCOMMIT=unknown
ARG BUILDTIME=unknown

ADD https://github.com/geofffranks/spruce/releases/download/v1.25.2/spruce-linux-amd64 /usr/local/bin/spruce
RUN chmod +x /usr/local/bin/spruce

RUN apk add --no-cache --no-progress \
    ca-certificates \
    make \
    git \
    openssh \
    gcc \
    libc-dev \
    upx

COPY . .
RUN make build

FROM alpine:latest

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/xmidt-org/themis/themis.yaml /_themis.yaml
COPY --from=builder /go/src/github.com/xmidt-org/themis/themis /themis
COPY --from=builder /go/src/github.com/xmidt-org/themis/deploy/Dockerfile /go/src/github.com/xmidt-org/themis/NOTICE /go/src/github.com/xmidt-org/themis/LICENSE /go/src/github.com/xmidt-org/themis/CHANGELOG.md /
COPY --from=builder /go/src/github.com/xmidt-org/themis/deploy/packaging/entrypoint.sh /entrypoint.sh
COPY --from=builder /go/src/github.com/xmidt-org/themis/deploy/packaging/themis_spruce.yaml /tmp/themis_spruce.yaml
COPY --from=builder /usr/local/bin/spruce /spruce

RUN mkdir /etc/themis/ && touch /etc/themis/themis.yaml && chmod 666 /etc/themis/themis.yaml

USER nobody

ENTRYPOINT ["/entrypoint.sh"]

EXPOSE 6500
EXPOSE 6501
EXPOSE 6502
EXPOSE 6503
EXPOSE 6504

CMD ["/themis"]