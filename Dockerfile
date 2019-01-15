FROM golang:1.9 as build

WORKDIR /go/src/github.com/marckhouzam/custom-prometheus-exporter
COPY . .

RUN go-wrapper download github.com/prometheus/client_golang/prometheus && \
    go-wrapper download gopkg.in/yaml.v2 && \
    go-wrapper install && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o custom-prometheus-exporter .

# The below should be modified to include the tools the exporters will need.
# In our example, we start FROM a docker image because we need docker installed.
# If no exporter needs docker, choose another image to start from which will be sufficient.
FROM docker:stable

# Below we install different shells to support different syntax.  However, only the shells
# used by the exporter definition are actually required.
# Also, you may need to install other tools if your exporters need them.
RUN apk update && \
    apk --no-cache add ca-certificates && \
    apk --no-cache add bash tcsh zsh

WORKDIR /root

COPY --from=build /go/src/github.com/marckhouzam/custom-prometheus-exporter/custom-prometheus-exporter .

ENTRYPOINT ["./custom-prometheus-exporter"]
