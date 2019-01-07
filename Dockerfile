FROM golang:1.9 as build

WORKDIR /go/src/github.com/marckhouzam/custom-prometheus-exporter
COPY . .

RUN go-wrapper download github.com/prometheus/client_golang/prometheus && \
    go-wrapper download gopkg.in/yaml.v2 && \
    go-wrapper install && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o custom-prometheus-exporter .

FROM docker:stable

RUN apk update && \
    apk --no-cache add ca-certificates && \
    apk --no-cache add bash tcsh zsh

WORKDIR /root

COPY --from=build /go/src/github.com/marckhouzam/custom-prometheus-exporter/custom-prometheus-exporter .

ENTRYPOINT ["./custom-prometheus-exporter"]
