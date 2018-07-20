
GO?=$(shell which go)

OS:=$(shell lsb_release -si)
ARCH:=$(shell uname -m | sed 's/x86_//;s/i[3-6]86/32/')
VER:=$(shell lsb_release -sr)

GOSRC:=$(shell find ./cmd ./pkg -type f -print)

.DEFAULT_GOAL: all


all: clean build package

build: bin/docker-dns

package:

clean:

bin/docker-dns: $(GOSRC)
	$(GO) build -o $@ cmd/docker-dns/docker-dns.go



