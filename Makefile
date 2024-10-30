ifneq "${testrun}" ""
testrunargs:=-run ${testrun}
else
testrunargs:=
endif

all: build-check compile_commands

build-check:
	go build ./...

testing_with_ld_library_path=LD_LIBRARY_PATH=`pwd -P`/gos/linux/lib
testing_with_xvfb=xvfb-run --server-args="-screen 0 512x64x24" --auto-servernum
testing_env=${testing_with_ld_library_path} ${testing_with_xvfb}

test:
	${testing_env} go test                ${testrunargs} ./...

test-spec:
	${testing_env} go test -run ".*Specs" ${testrunargs} ./...

test-nocache:
	${testing_env} go test -count=1       ${testrunargs} ./...

compile_commands: gos/linux/compile_commands.json

gos/linux/compile_commands.json:
	cd $(dir $@) && bear -- ${MAKE}

gos/linux/lib/libglop.so:
	mkdir -p $(dir $@)
	${MAKE} -C $(dir $@)

fmt:
	go fmt ./...

# -l for 'list files'
checkfmt:
	@gofmt -l ./

.PHONY: build-check
.PHONY: compile_commands
.PHONY: test
.PHONY: fmt
