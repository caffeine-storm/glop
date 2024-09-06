#!/bin/bash

# TODO(tmckee): make this a Makefile
set -e

g++ -o glop.o -Wall -fPIC -c -Iinclude glop.cpp
g++ -shared -dynamiclib -Wall -o libglop.so glop.o
rm -f glop.o
mkdir -p lib
mv libglop.so lib
