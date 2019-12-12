FROM golang:1.13.4-alpine3.10 as builder

WORKDIR /themis
ARG VERSION=undefined
ARG GITCOMMIT=undefined
ARG BUILDTIME=undefined

COPY . .
RUN go build -ldflags "-X 'main.BuildTime=${BUILDTIME}' -X main.GitCommit=${GITCOMMIT} -X main.Version=${VERSION}"

FROM alpine:3.10

EXPOSE 6600
EXPOSE 6601
EXPOSE 6602

WORKDIR /themis
COPY --from=builder /themis/themis .
COPY deploy/remote-claims/rbl_themis.yaml .
ENTRYPOINT ["./themis", "-f", "rbl_themis.yaml"]
