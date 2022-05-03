#!/bin/sh
rm charts/helm-common-*
helm package ../
mkdir -p charts && mv -v helm-common-* charts/
go mod download
gotestsum --format pkgname-and-test-fails
