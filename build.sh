#!/bin/sh

NAME="notefinder"
VERSION="0.1"

git archive --format=tar.gz -o ~/rpmbuild/SOURCES/$NAME-$VERSION.tar.gz --prefix=$NAME-$VERSION/ main
rpmbuild -bb rpm/$NAME.spec

exit 0
