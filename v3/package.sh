#!/usr/bin/env bash

echo "package nemo for macos,linux and windows..."
echo "MAKS SURE \"RunMode = Release\" in pkg/conf/config.go"
rm -rf release/*
rm -rf web/static/download/*
echo > log/runtime.log
echo > log/access.log

CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o server_darwin_amd64 cmd/server/main.go
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o worker_darwin_amd64 cmd/worker/main.go
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o daemon_worker_darwin_amd64 cmd/daemon_worker/main.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o server_linux_amd64 cmd/server/main.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o worker_linux_amd64 cmd/worker/main.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o daemon_worker_linux_amd64 cmd/daemon_worker/main.go
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o server_windows_amd64.exe cmd/server/main.go
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o worker_windows_amd64.exe cmd/worker/main.go
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o daemon_worker_windows_amd64.exe cmd/daemon_worker/main.go

tar -cvzf release/nemo_darwin_amd64.tar \
  --exclude=thirdparty/nuclei/nuclei_linux_amd64 \
  --exclude=thirdparty/nuclei/nuclei_windows_amd64.exe \
  --exclude=thirdparty/subfinder/subfinder_windows_amd64.exe \
  --exclude=thirdparty/subfinder/subfinder_linux_amd64 \
  --exclude=thirdparty/httpx/httpx_windows_amd64.exe \
  --exclude=thirdparty/httpx/httpx_linux_amd64 \
  --exclude=thirdparty/massdns/massdns_windows_amd64.exe \
  --exclude=thirdparty/massdns/cygwin1.dll \
  --exclude=thirdparty/massdns/massdns_linux_amd64 \
  --exclude=thirdparty/gogo/gogo_linux_amd64 \
  --exclude=thirdparty/gogo/gogo_windows_amd64.exe \
  --exclude=thirdparty/fingerprintx/fingerprintx_linux_amd64 \
  --exclude=thirdparty/fingerprintx/fingerprintx_windows_amd64.exe \
  server_darwin_amd64 worker_darwin_amd64 daemon_worker_darwin_amd64 \
  conf log thirdparty web docs version.txt

tar -cvzf release/nemo_linux_amd64.tar \
  --exclude=thirdparty/nuclei/nuclei_darwin_amd64 \
  --exclude=thirdparty/nuclei/nuclei_windows_amd64.exe \
  --exclude=thirdparty/subfinder/subfinder_windows_amd64.exe \
  --exclude=thirdparty/subfinder/subfinder_darwin_amd64 \
  --exclude=thirdparty/httpx/httpx_windows_amd64.exe \
  --exclude=thirdparty/httpx/httpx_darwin_amd64 \
  --exclude=thirdparty/massdns/massdns_windows_amd64.exe \
  --exclude=thirdparty/massdns/cygwin1.dll \
  --exclude=thirdparty/massdns/massdns_darwin_amd64 \
  --exclude=thirdparty/gogo/gogo_darwin_amd64 \
  --exclude=thirdparty/gogo/gogo_windows_amd64.exe \
  --exclude=thirdparty/fingerprintx/fingerprintx_darwin_amd64 \
  --exclude=thirdparty/fingerprintx/fingerprintx_windows_amd64.exe \
  server_linux_amd64 worker_linux_amd64 daemon_worker_linux_amd64 \
  conf log thirdparty web docs version.txt \
  Dockerfile docker docker-compose.yaml docker_start.sh server_install.sh worker_install.sh

tar -cvzf release/nemo_windows_amd64.tar \
  --exclude=thirdparty/nuclei/nuclei_darwin_amd64 \
  --exclude=thirdparty/nuclei/nuclei_linux_amd64 \
  --exclude=thirdparty/subfinder/subfinder_darwin_amd64 \
  --exclude=thirdparty/subfinder/subfinder_linux_amd64 \
  --exclude=thirdparty/httpx/httpx_darwin_amd64 \
  --exclude=thirdparty/httpx/httpx_linux_amd64 \
  --exclude=thirdparty/massdns/massdns_darwin_amd64 \
  --exclude=thirdparty/massdns/massdns_linux_amd64 \
  --exclude=thirdparty/gogo/gogo_darwin_amd64 \
  --exclude=thirdparty/gogo/gogo_linux_amd64 \
  --exclude=thirdparty/fingerprintx/fingerprintx_darwin_amd64 \
  --exclude=thirdparty/fingerprintx/fingerprintx_linux_amd64 \
  server_windows_amd64.exe worker_windows_amd64.exe daemon_worker_windows_amd64.exe \
  conf log thirdparty web docs version.txt

  echo "package done..."