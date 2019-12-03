FROM golang:1.13.4-alpine3.10 as builder

WORKDIR /themis 
ARG VERSION=undefined
ARG GITCOMMIT=undefined
ARG BUILDTIME=undefined

COPY . .
RUN go build -ldflags "-X 'main.BuildTime=${BUILDTIME}' -X main.GitCommit=${GITCOMMIT} -X main.Version=${VERSION}"

FROM alpine:3.10

EXPOSE 6500
EXPOSE 6501
EXPOSE 6502
EXPOSE 6503

WORKDIR /themis
COPY --from=builder /themis/themis .
COPY deploy/docker/modes/cpe/cpe_themis.yaml .
ENTRYPOINT ["./themis", "-f", "cpe_themis.yaml"]
