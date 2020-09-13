VERSION = $(shell grep 'const Version =' version.go | cut -d '"' -f 2)

test: *.go
	go test
	(cd examples/tcp          && go build && ./tcp >/dev/null)
	(cd examples/pipe         && go build && ./pipe >/dev/null)
	(cd examples/limits       && go build && ./limits >/dev/null)
	(cd examples/read-timeout && go build && ./read-timeout >/dev/null)
	@echo "All tests OK"


# release prepares the project for a new release:
#
#   1. Compare existing git tags with the value in version.go to ensure
#      the version declared in version.go is not already released.
#
#   2. "go mod tidy" to clean up go modules
#
#   3. run all tests
#
#   4. Print commands for publishing
#
release:
	@if (git rev-list v${VERSION}.. 2>/dev/null); then \
		echo "--------------------------------------------------" >&2; \
		echo "git tag v${VERSION} already exists:" >&2; \
		git log v1.0.0 -n 1 | cat 1>&2; \
		echo "--------------------------------------------------" >&2; \
		echo "Did you forget to update version.go?" >&2; \
		echo "--------------------------------------------------" >&2; \
		exit 1; \
	fi
	go mod tidy
	$(MAKE) test
	@echo "Finally, run the following to publish v${VERSION}:"
	@echo "  git tag v${VERSION}"
	@echo "  git push origin v${VERSION} master"


dist: release

clean:
	rm -f bin/*

.PHONY: test clean release dist
