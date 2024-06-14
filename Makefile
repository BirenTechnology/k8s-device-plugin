REPO ?= ghcr.io/BirenTechnology
PROJECT ?= k8s-device-plugin
BUILD_ENV?=GOPROXY=direct
tag=$(shell git describe --abbrev=0 --tags)
VERSION=$(shell git describe --tags --always)

image-build:
	docker build --build-arg build_arch=amd64 -t $(REPO)/$(PROJECT):$(VERSION) -f deploy/Dockerfile . 

image-build-arm:
	docker build --build-arg build_arch=arm64 -t $(REPO)/$(PROJECT):$(VERSION)-arm64 -f deploy/Dockerfile .

push:
	docker push $(REPO)/$(PROJECT):$(VERSION)


build:
	${BUILD_ENV} GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-X 'main.Version=$(VERSION)' -X 'main.Time=$(shell LC_TIME=en_US.UTF-8 date)' -X 'main.Commit=$(shell git rev-parse --short HEAD)'" -o k8s-device-plugin cmd/manager.go

build-arm:
	${BUILD_ENV} GOOS=linux GOARCH=arm64 CGO_ENABLED=1 go build -ldflags="-X 'main.Version=$(VERSION)' -X 'main.Time=$(shell LC_TIME=en_US.UTF-8 date)' -X 'main.Commit=$(shell git rev-parse --short HEAD)'" -o k8s-device-plugin cmd/manager.go