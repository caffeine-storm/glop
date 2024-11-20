SHELL:=/bin/bash

TEST_REPORT_TAR:=testdata/report.tar.gz
LIBGLOP:=gos/linux/lib/libglop.so

testrunpackages=./...
ifneq "${testrun}" ""
testrunargs:=-run ${testrun}
else
testrunargs:=
endif

ifeq "${debugtest}" ""
testsinglepackageargs:=--set-'debugtest'-to-a-package-dir
else
testsinglepackageargs="${debugtest}"
# Replace each -$elem of 'testrunargs' with '-test.$elem'
newtestrunargs:=$(subst -,-test.,${testrunargs})
endif

all: build-check compile_commands

build-check: ${LIBGLOP}
	go build ./...

testing_with_ld_library_path=LD_LIBRARY_PATH=`pwd -P`/gos/linux/lib
testing_with_xvfb=xvfb-run --server-args="-screen 0 512x64x24" --auto-servernum
testing_env=${testing_with_ld_library_path} ${testing_with_xvfb}

test: pre-test
	${testing_env} go test                   ${testrunargs} ${testrunpackages}

test-spec: pre-test
	${testing_env} go test -run ".*Specs"    ${testrunargs} ${testrunpackages}

test-nocache: pre-test
	${testing_env} go test -count=1          ${testrunargs} ${testrunpackages}

test-dlv: pre-test
# delve wants exactly one package at a time so 'testrunpackages' better be a
# literal directory
	[ -d ${testsinglepackageargs} ] && \
	${testing_env} dlv test ${testsinglepackageargs} -- ${newtestrunargs}

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

list_rejects:
	@find testdata/ -name '*.rej*'

clean_rejects:
	find testdata/ -name '*.rej*' -exec rm "{}" +

promote_rejects:
	@shopt -s nullglob ; \
	find testdata/ -name '*.rej*' | while read i ; do \
		echo mv $$i $${i/.rej} ; \
		mv $$i $${i/.rej} ; \
	done

# Deliberately signal failure from this recipe so that CI notices failing tests
# are red.
appveyor-test-report-and-fail: test-report
	appveyor PushArtifact ${TEST_REPORT_TAR} -DeploymentName "test report tarball"
	false

test-report: ${TEST_REPORT_TAR}

${TEST_REPORT_TAR}:
	tar \
		--auto-compress \
		--create \
		--file $@ \
		--directory testdata/ \
		--files-from <(cd testdata; find  . -name '*.rej.*' | while read fname ; do \
				echo $$fname ; \
				echo $${fname/.rej} ; \
			done \
		)

compile_commands: gos/linux/compile_commands.json

gos/linux/compile_commands.json:
	cd $(dir $@) && bear -- ${MAKE}

${LIBGLOP}:
	mkdir -p $(dir $@)
	${MAKE} -C $(dir $@)/..

pre-test: ${LIBGLOP}

fmt:
	go fmt ./...

# -l for 'list files'
checkfmt:
	@gofmt -l ./

clean:
	rm -f ${TEST_REPORT_TAR}

.PHONY: build-check
.PHONY: list_rejects clean_rejects promote_rejects
.PHONY: compile_commands
.PHONY: fmt
.PHONY: profiling/*.view
.PHONY: appveyor-test-report-and-fail
.PHONY: pre-test test test-spec test-nocache test-fresh test-report
.PHONY: ${TEST_REPORT_TAR}
