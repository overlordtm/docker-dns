
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
	go clean -r .

run:
	$(GO) run cmd/docker-dns/docker-dns.go -loglevel=debug -config=configs/docker-dns.json

install: bin/docker-dns
	cp init/docker-dns.service /etc/systemd/system/docker-dns.service
	cp configs/docker-dns /etc/NetworkManager/dnsmasq.d/docker-dns
	cp configs/docker-dns.json /etc/docker-dns.json
	cp bin/docker-dns /usr/local/bin/docker-dns

bin/docker-dns: $(GOSRC)
	$(GO) build -o $@ cmd/docker-dns/docker-dns.go
	chmod +x $@



