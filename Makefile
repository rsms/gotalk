VERSION = $(shell grep 'const Version =' version.go | cut -d '"' -f 2)
GOCOV_FILE=.cache/coverage.out

test:
	go test
	bash examples/test.sh -silent
	@echo "All tests OK"

fmt:
	@echo gofmt -w -s
	@find -E . -type f -name '*.go' -not -regex '\./(_local|js|vendor)/.*' | xargs gofmt -w -s

doc:
	@echo "open http://localhost:6060/pkg/github.com/rsms/gotalk/"
	godoc -http=localhost:6060

dev:
	@mkdir -p .cache/dev
	@(serve-http -no-log -p 8148 .cache/dev &)
	autorun \
		$(firstword $(MAKEFILE_LIST)) \
		go.sum \
		*.go \
		-- "$(MAKE) dev1"

dev1: fmt
	@ echo "go test (see http://localhost:8148/ for code coverage)"
	@ go test -covermode=count "-coverprofile=$(GOCOV_FILE)"
	@ go tool cover "-html=$(GOCOV_FILE)" -o .cache/gocov.html
	@  sed 's/.cov0 { color: rgb(192, 0, 0)/.cov0 { color: rgb(255, 100, 80)/g' .cache/gocov.html \
	 | sed 's/font-weight: bold/font-weight: normal/g' \
	 | sed 's/font-family:/tab-size:2;font-family: SFMono-Regular,Consolas,Liberation Mono,Menlo,/g' \
	 | sed 's/background: black;/background: rgba(20,20,20);/g' \
	 | python -c 'import re,sys;print(re.sub(r"\n {8}", "\n\t", sys.stdin.read()))' \
	 | python -c 'import re,sys;print(re.sub(r"\n(\t+) {8}", "\n\\1\t", sys.stdin.read()))' \
	 | python -c 'import re,sys;print(re.sub(r"\n(\t+) {8}", "\n\\1\t", sys.stdin.read()))' \
	 | python -c 'import re,sys;print(re.sub(r"\n(\t+) {8}", "\n\\1\t", sys.stdin.read()))' \
	 | python -c 'import re,sys;print(re.sub(r"\n(\t+) {8}", "\n\\1\t", sys.stdin.read()))' \
	 | python -c 'import re,sys;print(re.sub(r"\n(\t+) {8}", "\n\\1\t", sys.stdin.read()))' \
	 | python -c 'import re,sys;print(re.sub(r"\n(\t+) {8}", "\n\\1\t", sys.stdin.read()))' \
	 | python -c 'import re,sys;print(re.sub(r"\n(\t+) {8}", "\n\\1\t", sys.stdin.read()))' \
	 > .cache/gocov1.html
	@ mv -f .cache/gocov1.html .cache/dev/index.html

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
	@# run code formatter and mod tidy, then check if it made changes
	$(MAKE) fmt
	go mod tidy
	@if [[ -n $$(git status --ignore-submodules=dirty --porcelain | grep -v '?? ') ]]; then \
    echo "gofmt altered some files:" >&2 ; \
    git status --ignore-submodules=dirty -s | grep -v '?? ' >&2; \
    exit 1; \
  fi
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

.PHONY: test clean release dist fmt doc dev dev1
