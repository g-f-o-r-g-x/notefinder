NAME=notefinder
PREFIX=/usr/local/bin
CFLAGS := $(shell perl -MExtUtils::Embed -e ccopts)
LDFLAGS := $(shell perl -MExtUtils::Embed -e ldopts)

all: compile

compile:
	CGO_CFLAGS="${CFLAGS}" CGO_LDFLAGS="${LDFLAGS}" go build .

install: compile
	cp ./${NAME} ${PREFIX}/${NAME}
