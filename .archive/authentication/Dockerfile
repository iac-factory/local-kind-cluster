# syntax = docker/dockerfile:1.0-experimental

FROM golang:1.22-alpine as BUILD

ENV GOOS "linux"
ENV GOVCS "*:all"
ENV CGO_ENABLED "0"

ENV GOPRIVATE "github.com/iac-factory/*"

WORKDIR /

COPY . ./

RUN apk add --no-cache git

RUN go build --mod vendor --ldflags="-s -w" -o /service

# --> Prevents shell access
RUN adduser -h "/dev/null" -g "" -s "/sbin/nologin" -D -H -u 10000 api-service-user

FROM scratch as RELEASE

WORKDIR /

COPY --from=BUILD /etc/passwd /etc/passwd
COPY --from=BUILD /service /usr/local/bin/service
COPY --from=BUILD /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

USER api-service-user

CMD ["service"]
