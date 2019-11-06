#
# Copyright 2017 Nippon Telegraph and Telephone Corporation.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

RM		:= rm -f
PWDIR		:= `pwd`

VERSION		:=	1.0.0
REVISION	:=	$(shell git rev-parse --short HEAD)
BUILDDATE	:=	$(shell date '+%Y-%m-%d %H:%M:%S %Z')
GOVERSION	:=	$(shell go version)
LDFLAGS		:=	-X 'main.version=$(VERSION)' \
			-X 'main.revision=$(REVISION)' \
			-X 'main.builddate=$(BUILDDATE)' \
			-X 'main.goversion=$(GOVERSION)'

CONFIG_INSTALL_DIR	:= /usr/local/etc
CONFIG_DIR	:= $(PWDIR)/conf
CONFIG_FILE	:= vsw_vrrpd.yml

all:	vendor build

setup:
	@echo "Get dep..."
	go get -u github.com/golang/dep/cmd/dep

vendor:	setup
	@echo "Exec dep..."
	dep ensure

update:	setup
	dep ensure -update

setup-dev:
	go get github.com/alecthomas/gometalinter
	gometalinter --install

build:
	go build -ldflags "$(LDFLAGS)"

install:
	go install -ldflags "$(LDFLAGS)"

test:
	go test -v --cover $$(go list ./...)

bench:
	go test -bench . -benchmem $$(go list ./...)

lint:	setup-dev
	-gometalinter $$(go list ./...)

copy-config:
	@echo "Copy $(CONFIG_DIR)/$(CONFIG_FILE) => $(CONFIG_INSTALL_DIR)/$(CONFIG_FILE)"
	cp $(CONFIG_DIR)/$(CONFIG_FILE) $(CONFIG_INSTALL_DIR)/$(CONFIG_FILE)

clean:
	go clean

distclean:	clean
	$(RM) -r ./vendor

.PHONY: all setup vendor update setup-dev build install test lint clean distclean
