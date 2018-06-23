.PHONY: deps

# This gets all of the third party dependencies that pkg is known to work well
# with. We don't want to vendor these here, because they should be vendored by
# `package main`.
deps:
	cat deps | ./get-deps.sh
