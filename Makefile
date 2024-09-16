all: build-check compile_commands

build-check:
	go build ./...

test:
	LD_LIBRARY_PATH=`pwd -P`/gos/linux/lib go test ./...

compile_commands: gos/linux/compile_commands.json

gos/linux/compile_commands.json:
	cd $(dir $@) && bear -- bash make.bash

fmt:
	go fmt ./...

.PHONY: build-check
.PHONY: compile_commands
.PHONY: test
.PHONY: fmt