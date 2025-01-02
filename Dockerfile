# SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
# SPDX-License-Identifier: Apache-2.0
FROM docker.io/library/golang:1.23-alpine as builder

WORKDIR /src

ARG VERSION
ARG GITCOMMIT
ARG BUILDTIME

RUN apk add --no-cache --no-progress \
    ca-certificates \
    make \
    curl \
    git \
    openssh \
    gcc \
    libc-dev \
    upx

# Download spruce here to eliminate the need for curl in the final image
RUN mkdir -p /go/bin && \
    curl -L -o /go/bin/spruce https://github.com/geofffranks/spruce/releases/download/v1.29.0/spruce-linux-amd64 && \
    chmod +x /go/bin/spruce

COPY . .

RUN go build -o themis .

RUN make test release

##########################
# Build the final image.
##########################

FROM alpine:latest

# Copy over the standard things you'd expect.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt  /etc/ssl/certs/
COPY --from=builder /src/themis                         /
COPY --from=builder /src/.release/docker/entrypoint.sh  /

# Copy over spruce and the spruce template file used to make the actual configuration file.
COPY --from=builder /src/.release/docker/themis_spruce.yaml /tmp/themis_spruce.yaml
COPY --from=builder /go/bin/spruce                          /bin/

# Include compliance details about the container and what it contains.
COPY --from=builder /src/Dockerfile \
                    /src/NOTICE \
                    /src/LICENSE \
                    /src/CHANGELOG.md   /

# Make the location for the configuration file that will be used.
RUN     mkdir /etc/themis/ \
    &&  touch /etc/themis/themis.yaml \
    &&  chmod 666 /etc/themis/themis.yaml

USER nobody

#ENTRYPOINT ["/entrypoint.sh"]

EXPOSE 6500
EXPOSE 6501
EXPOSE 6502
EXPOSE 6503

ENTRYPOINT ["/themis"]
