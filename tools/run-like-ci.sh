#!/bin/bash

# Run from the project root
here=`dirname $0`
cd "$here/.."

# --rm to remove container after exit
# --volume .:/glop will mount the current glop project on /glop inside
#                    the container
# --interactive --tty for an interactive shell with a pseudo-TTY
podman run --rm --volume .:/glop --interactive --tty docker.io/caffeinestorm/haunts-custom-build-image:latest bash
