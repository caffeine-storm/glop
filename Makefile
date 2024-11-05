SHELL:=/bin/bash

testrunpackages=./...
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
	${testing_env} go test                   ${testrunargs} ${testrunpackages}

test-spec:
	${testing_env} go test -run ".*Specs"    ${testrunargs} ${testrunpackages}

test-nocache:
	${testing_env} go test -count=1          ${testrunargs} ${testrunpackages}

cpu_profile_file=cpu-pprof.gz
profile_dir=profiling

${profile_dir}:
	mkdir -p $@

.PRECIOUS: ${profile_dir}/%.test
.PRECIOUS: ${profile_dir}/%.${cpu_profile_file}

${profile_dir}/%.test ${profile_dir}/%.${cpu_profile_file}: |${profile_dir}
	${testing_env} go test \
		-o ${profile_dir}/$*.test \
		-cpuprofile ${profile_dir}/$*.${cpu_profile_file} \
		${testrunargs} ./$*

${profile_dir}/%.view: ${profile_dir}/%.test ${profile_dir}/%.${cpu_profile_file}
	pprof -web $^

test-fresh: |clean_rejects
test-fresh: test-nocache

clean_rejects:
	rm -f testdata/text/*.rej.*

promote_rejects:
	@shopt -s nullglob ; \
	for i in testdata/text/*.rej.* ; do \
		echo mv $$i $${i/.rej} ; \
		mv $$i $${i/.rej} ; \
	done

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
.PHONY: test test-spec test-nocache test-fresh
.PHONY: profiling/*.view
.PHONY: fmt
.PHONY: clean_rejects promote_rejects
