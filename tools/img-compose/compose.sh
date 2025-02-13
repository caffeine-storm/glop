#!/bin/bash

: ${bgcolour:=black}

if [[ $# -ne 3 ]]; then
	echo 1>&2 "usage: $0 in1.png in2.png out.png"
	exit 1
fi

tmp1=`mktemp`
tmp2=`mktemp`
tmp3=`mktemp`
bgtmp=`mktemp`

convert -transparent "${bgcolour}" "$1" "$tmp1"
convert -transparent "${bgcolour}" "$2" "$tmp2"
convert "$tmp1" -alpha opaque +level-colors "${bgcolour}" "$bgtmp"

composite \
	-compose over \
	-gravity center \
	"$tmp1" "$tmp2" \
	"$tmp3"

composite \
	-compose over \
	-gravity center \
	"$tmp3" "$bgtmp" \
	"$3"

rm "$tmp1"
rm "$tmp2"
rm "$tmp3"
rm "$bgtmp"
