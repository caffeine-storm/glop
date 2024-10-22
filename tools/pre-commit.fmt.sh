#!/bin/bash
#
# Called by "git commit" with no arguments.  The hook should
# exit with non-zero status after issuing an appropriate message if
# it wants to stop the commit.
#
# To enable this hook, copy this file to ".git/hooks/pre-commit".
output=`make checkfmt`
if [ -z "$output" ]; then
	exit 0
fi

echo 1>&2 "please run 'make fmt' first"
echo 1>&2 "(or pass '--no-verify' to 'git commit')"
echo 1>&2
echo 1>&2 "these files need formatting:"
echo 1>&2 "$output"

exit 1
