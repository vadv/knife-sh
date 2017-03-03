SOURCEDIR=./src
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

VERSION := $(shell git describe --abbrev=0 --tags)
SHA := $(shell git rev-parse --short HEAD)

GOPATH ?= /usr/local/go
GOPATH := ${CURDIR}:${GOPATH}
export GOPATH

all: submodules
	# TODO: ugly
	rm -rf $(SOURCEDIR)/github.com/vadv/knife-sh/
	mkdir -p $(SOURCEDIR)/github.com/vadv/knife-sh/
	ln -sf ${CURDIR}/config $(SOURCEDIR)/github.com/vadv/knife-sh/config
	ln -sf ${CURDIR}/ssh $(SOURCEDIR)/github.com/vadv/knife-sh/ssh
	go build -o ./bin/knife-sh -ldflags "-X main.BuildVersion=$(VERSION)-$(SHA)" knife-sh.go

.DEFAULT_GOAL: all

include Makefile.git
