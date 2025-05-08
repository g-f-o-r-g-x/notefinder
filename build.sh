#!/bin/sh

NAME="notefinder"
VERSION="0.1"

git archive --format=tar.gz -o ~/rpmbuild/SOURCES/$NAME-$VERSION.tar.gz --prefix=$NAME-$VERSION/ main || exit 1
output=$(rpmbuild -bb rpm/$NAME.spec)
outfile=$(echo "$output" | grep -E ^Wrote | grep -v debug | cut -d " " -f2)

rpmsign --addsign $outfile &&
rpm --checksig $outfile

sudo rpm -ivh --force $outfile &&
exit 0
