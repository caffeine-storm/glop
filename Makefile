ifneq "${testrun}" ""
testrunargs:=-run ${testrun}
else
testrunargs:=
endif

all: build-check compile_commands

build-check:
	go build ./...

test:
	LD_LIBRARY_PATH=`pwd -P`/gos/linux/lib xvfb-run --server-args="-fbdir ./test -screen 0 512x64x24" --auto-servernum go test ${testrunargs} ./...

compile_commands: gos/linux/compile_commands.json

gos/linux/compile_commands.json:
	cd $(dir $@) && bear -- ${MAKE}

gos/linux/lib/libglop.so:
	mkdir -p $(dir $@)
	${MAKE} -C $(dir $@)

fmt:
	go fmt ./...
	go -C test/drawstring fmt ./...

# -l for 'list files'
checkfmt:
	@gofmt -l ./

.PHONY: build-check
.PHONY: compile_commands
.PHONY: test
.PHONY: fmt
