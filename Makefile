DOCKER_USERNAME=lucachot
CTRL_IMAGE_NAME=basic-sched
GO_FLAGS=

ifdef RACE
	GO_FLAGS += -race
endif

TAG ?= latest

all: build pushshed

sched: main.go
	go build ${GO_FLAGS} -o bin/sched $<

build: sched
	docker build -t ${DOCKER_USERNAME}/basic-sched:${TAG} -f docker_images/docker_basic .

pushshed:
	docker push ${DOCKER_USERNAME}/basic-sched:${TAG}

