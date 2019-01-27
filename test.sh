#!/bin/bash
set -x
./custom-prometheus-exporter -f example-configurations/test-exporter.yaml -f example-configurations/docker-exporter.yaml&
PID=$!
curl localhost:9530
curl localhost:9530/validate
curl localhost:9550/metrics
curl localhost:12345/test
curl localhost:9530/test
kill $PID
