APP_NAME ?= k8s-s3-backup
IMAGE ?= yourrepo/$(APP_NAME)
VERSION ?= latest

.PHONY: all build clean fmt vet docker-build docker-push helm-lint

all: build

build:
	GO111MODULE=on CGO_ENABLED=0 go build -o bin/backup ./cmd/backup

clean:
	rm -rf bin

fmt:
	gofmt -s -w .

vet:
	go vet ./...

docker-build:
	docker build -t $(IMAGE):$(VERSION) .

docker-push:
	docker push $(IMAGE):$(VERSION)

helm-lint:
	helm lint helm-chart

