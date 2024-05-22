FROM golang:1.22 as BUILD

ENV GOOS "linux"
ENV GOVCS "*:all"
ENV CGO_ENABLED "0"

ENV GOPRIVATE "gitlab.com/ethrgg/*,github.com/iac-factory/*"

WORKDIR /

COPY . ./

RUN go build --mod vendor --ldflags="-a -s -w -X 'main.VERSION=$(cat VERSION)'" -o /usr/local/bin/manager cmd/main.go

CMD ["manager"]
