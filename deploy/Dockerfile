FROM golang:1.13.4-alpine3.10 as builder

WORKDIR /themis 
ARG VERSION=undefined
ARG GITCOMMIT=undefined
ARG BUILDTIME=undefined

COPY . .
RUN go build -ldflags "-X 'main.BuildTime=${BUILDTIME}' -X main.GitCommit=${GITCOMMIT} -X main.Version=${VERSION}"

FROM alpine:3.10

EXPOSE 6700
EXPOSE 6701
EXPOSE 6702
EXPOSE 6703

WORKDIR /themis
COPY --from=builder /themis/themis .
COPY themis.yaml .
ENTRYPOINT ["./themis", "-f", "themis.yaml"]
