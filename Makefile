NAME=notefinder
PREFIX=/usr/local/bin
CFLAGS := $(shell perl -MExtUtils::Embed -e ccopts | sed -e 's/-D_GNU_SOURCE//g')
LDFLAGS := $(shell perl -MExtUtils::Embed -e ldopts)

all: compile

compile:
	CGO_CFLAGS="${CFLAGS} -Wno-builtin-macro-redefined" CGO_LDFLAGS="${LDFLAGS}" go build .

install: compile
	cp ./${NAME} ${PREFIX}/${NAME}
