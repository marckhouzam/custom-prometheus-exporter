#!/usr/bin/env bash


set -u -e -o pipefail

TAG="${1:-}"

if [[ -z "$TAG" ]] ; then
  echo "ERROR: No release tag specified" >&2
  exit 1
fi

if [[ -z "$GITHUB_TOKEN" ]] ; then
  echo "ERROR: Env variable GITHUB_TOKEN not set. Get one in GH account Settings -> Developer settings -> Personal access tokens (use 'public_repo' scope)" >&2
  exit 1
fi

if [[ ! -x "custom-prometheus-exporter" ]] ; then
  echo "ERROR: Missing binary, please build it" >&2
  exit 1
fi

echo "Getting ghr"

go get github.com/tcnksm/ghr

echo "Creating archive"
rm -rf dist/
mkdir dist
tar -czf "dist/custom-prometheus-exporter-$TAG-amd64.tar.gz" custom-prometheus-exporter LICENSE README.md
sha256sum dist/*.tar.gz > dist/sha256sums.txt

echo "Upload to Github"

ghr -parallel 1 -u marckhouzam -r custom-prometheus-exporter --replace "$TAG" dist/

echo "Done"
