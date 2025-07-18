SHELL:=/bin/bash

TEST_REPORT_TAR:=test-report.tar.gz

UNAME:=$(shell uname)
ifeq (${UNAME},Linux)
PLATFORM:=linux
else
$(error unknown uname value '${UNAME}')
endif
NATIVE_SRCS:=$(shell find gos/${PLATFORM}/ \
     -name '*.cpp' \
  -o -name '*.hpp' \
  -o -name '*.c' \
  -o -name '*.h' \
)

testrunpackages=./...
ifneq "${testrun}" ""
testrunargs:=-run ${testrun}
else
testrunargs:=
endif

ifeq "${pkg}" ""
testsinglepackageargs:=--set-'pkg'-to-a-package-dir
else
testsinglepackageargs="${pkg}"
# Replace each -$elem of 'testrunargs' with '-test.$elem'
newtestrunargs:=$(subst -,-test.,${testrunargs})
endif

all: build-check compile-commands

build-check:
	go build ./...

# By default, the Xvfb instance will create a new Xauthority file in
# /tmp/xvfb-run.PID/Xauthority for access control.
# To interact with the Xvfb instance, you can set your XAUTHORITY and DISPLAY
# environment vars accordingly.
testing_with_xvfb=xvfb-run --server-args="-screen 0 512x64x24" --auto-servernum
testing_env=${testing_with_xvfb}

test:
	${testing_env} go test                   ${testrunargs} ${testrunpackages}

test-verbose:
	${testing_env} go test -v                ${testrunargs} ${testrunpackages}

test-racy:
	${testing_env} go test -count=1 -race    ${testrunargs} ${testrunpackages}

test-racy-with-cache:
	${testing_env} go test          -race    ${testrunargs} ${testrunpackages}

test-spec:
	${testing_env} go test -run ".*Specs"    ${testrunargs} ${testrunpackages}

test-nocache:
	${testing_env} go test -count=1          ${testrunargs} ${testrunpackages}

test-dlv:
# delve wants exactly one package at a time so "testrunpackages" isn't what we
# want here. We use a var specifically for pointing at a single directory.
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
	@find . -name testdata -type d | while read testdatadir ; do \
		find "$$testdatadir" -name '*.rej.*' ; \
	done

# opens expected and rejected files in 'feh'
view_rejects:
	@find . -name testdata -type d | while read testdatadir ; do \
		find "$$testdatadir" -name '*.rej.*' | while read rejfile ; do \
			echo -e >&2 "$${rejfile/.rej}\n$$rejfile" ; \
			echo "$${rejfile/.rej}" "$$rejfile" ; \
		done ; \
	done | xargs -r feh

clean_rejects:
	find . -name testdata -type d | while read testdatadir ; do \
		find "$$testdatadir" -name '*.rej.*' -exec rm "{}" + ; \
	done

promote_rejects:
	@find . -name testdata -type d | while read testdatadir ; do \
		find "$$testdatadir" -name '*.rej.*' | while read rejfile ; do \
			echo mv "$$rejfile" "$${rejfile/.rej}" ; \
			mv "$$rejfile" "$${rejfile/.rej}" ; \
		done \
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
		--files-from <(find  . -name '*.rej.*' | while read fname ; do \
				echo "$$fname" ; \
				echo "$${fname/.rej}" ; \
			done \
		)

fmt:
	go fmt ./...
	clang-format -i ${NATIVE_SRCS}

lint:
	go run github.com/mgechev/revive@v1.5.1 ./...
	clang-tidy ${NATIVE_SRCS}

native-lints.txt: ${NATIVE_SRCS}
	clang-tidy $^ &> $@

count-native-lints: ${NATIVE_SRCS}
	clang-tidy --quiet $^ 2>/dev/null | wc --lines

count-native-lint-groups: native-lints.txt
	grep '^[^ ]' native-lints.txt | grep -v 'Processing file' | grep -v 'note:' | sed 's,.*\[,,g' | grep ']' | sed 's,],,' | sort | uniq --count | sort --reverse --numeric

depth:
	@go list ./... | while read PKG ; do \
		go run github.com/KyleBanks/depth/cmd/depth@v1.2.1 "$$PKG" ; \
	done

# -l for 'list files'
checkfmt:
	@gofmt -l ./
	@clang-format -n -Werror ${NATIVE_SRCS}

# Rebuild gos/$PLATFORM/compile_commands.json if any native code changes.
compile-commands: gos/${PLATFORM}/compile_commands.json

gos/${PLATFORM}/compile_commands.json: ${NATIVE_SRCS}
	bear --output `pwd -P`/gos/${PLATFORM}/compile_commands.json --force-wrapper -- go build -a ./gos/${PLATFORM}/

clean:
	rm -f ${TEST_REPORT_TAR}

.PHONY: build-check compile-commands
.PHONY: list_rejects view_rejects clean_rejects promote_rejects
.PHONY: fmt lint depth count-native-lints
.PHONY: profiling/*.view
.PHONY: appveyor-test-report-and-fail
.PHONY: test test-dlv test-fresh test-nocache test-report test-spec test-verbose
.PHONY: ${TEST_REPORT_TAR}
