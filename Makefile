VERSION = $(shell grep 'const Version =' version.go | cut -d '"' -f 2)

test:
	go test
	@
	@echo "go build examples/*"
	@for d in examples/*; do (cd $$d && go build) & done ; wait
	@
	(cd examples/tcp          && ./tcp >/dev/null)
	(cd examples/tls          && ./tls >/dev/null)
	(cd examples/pipe         && ./pipe >/dev/null)
	(cd examples/limits       && ./limits >/dev/null)
	(cd examples/read-timeout && ./read-timeout >/dev/null)
	@echo "All tests OK"


fmt:
	find -E . -type f -name '*.go' -not -regex '\./(_local|js|vendor)/.*' | xargs gofmt -w

doc:
	@echo "open http://localhost:6060/pkg/github.com/rsms/gotalk/"
	godoc -http=localhost:6060

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
	@if (git rev-list v${VERSION}.. >/dev/null 2>&1); then \
		echo "git tag v${VERSION} already exists:" >&2; \
		git log "v${VERSION}" -n 1 --format="%H%d%n%ad %an%n%s" | cat 1>&2; \
		echo "--------------------------------------------------" >&2; \
		echo "Did you forget to update version.go?" >&2; \
		echo "--------------------------------------------------" >&2; \
		exit 1; \
	fi
	@# make sure git status is clean
	@if [[ -n $$(git status --ignore-submodules=dirty --porcelain | grep -v '?? ') ]]; then \
    echo "uncommitted changes:" >&2 ; \
    git status --ignore-submodules=dirty -s | grep -v '?? ' >&2; \
    exit 1; \
  fi
	@# run code formatter and check if it made changes
	$(MAKE) fmt
	@if [[ -n $$(git status --ignore-submodules=dirty --porcelain | grep -v '?? ') ]]; then \
    echo "gofmt altered some files:" >&2 ; \
    git status --ignore-submodules=dirty -s | grep -v '?? ' >&2; \
    exit 1; \
  fi
	go mod tidy
	$(MAKE) test
	@echo "Finally, run the following to publish v${VERSION}:"
	@echo "  git tag v${VERSION}"
	@echo "  git push origin v${VERSION} master"
	@echo "  open https://github.com/rsms/gotalk/releases/new?tag=v${VERSION}&title=v${VERSION}"
	@echo ""
	@echo "Optional extras:"
	@echo "  open https://pkg.go.dev/github.com/rsms/gotalk@v${VERSION}"
	@echo ""



dist: release

clean:
	@true

.PHONY: test clean release dist fmt doc
