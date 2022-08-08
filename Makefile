MODNAME		:= github.com/wneessen/arrgo
SPACE		:= $(null) $(null)
CURVER		:= 0.0.1-DEV
CURARCH		:= $(shell uname -m | tr 'A-Z' 'a-z')
CUROS		:= $(shell uname -s | tr 'A-Z' 'a-z')
BUILDARCH	:= $(CUROS)_$(CURARCH)
BUILDDIR	:= ./bin
TZ			:= UTC
BUILDVER    := -X github.com/wneessen/arrgo/bot.Version=$(CURVER)
CURUSER     := $(shell whoami)
BUILDUSER   := -X github.com/wneessen/arrgo/version.BuildUser=$(subst $(SPACE),_,$(CURUSER))
CURDATE     := $(shell date +'%Y-%m-%d %H:%M:%S')
TARGETS		:= clean build

all: $(TARGETS)

test:
	go test $(MODNAME)

dev:
	@/usr/bin/env CGO_ENABLED=0 go run -ldflags="-s -w $(BUILDVER)" $(MODNAME)/cmd/arrgo -c ./arrgo.toml