SHELL:=/bin/bash

TEST_REPORT_TAR:=test-report.tar.gz

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

all: build-check

build-check:
	go build ./...

testing_with_xvfb=xvfb-run --server-args="-screen 0 512x64x24" --auto-servernum
testing_env=${testing_with_xvfb}

test:
	${testing_env} go test                   ${testrunargs} ${testrunpackages}

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

lint:
	go run github.com/mgechev/revive@v1.5.1 ./...

depth:
	@go list ./... | while read PKG ; do \
		go run github.com/KyleBanks/depth/cmd/depth@v1.2.1 "$$PKG" ; \
	done

# -l for 'list files'
checkfmt:
	@gofmt -l ./

clean:
	rm -f ${TEST_REPORT_TAR}

.PHONY: build-check
.PHONY: list_rejects view_rejects clean_rejects promote_rejects
.PHONY: fmt lint depth
.PHONY: profiling/*.view
.PHONY: appveyor-test-report-and-fail
.PHONY: test test-dlv test-fresh test-nocache test-report test-spec
.PHONY: ${TEST_REPORT_TAR}
