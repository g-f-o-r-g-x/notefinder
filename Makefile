NAME=notefinder
PREFIX=/usr/local/bin

all: compile

compile:
	go build .

install: compile
	cp ./${NAME} ${PREFIX}/${NAME}
