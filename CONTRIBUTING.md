# Contributing to the Gotalk project

First off, thank you for considering contributing to Gotalk. It's people like
you that make the Internet such a great place.

Following these guidelines helps to communicate that you respect the time of
the people managing and developing this open source project. In return, they
should reciprocate that respect in addressing your issue or suggestion.

By contributing work to the Gotalk project you agree to have all work
contributed becoming the intellectual property of the Gotalk project as
described by [The MIT License](LICENSE.txt).


## Development

Requirements:

- Go version `grep '^go' go.mod` or later (Gotalk uses go modules)
- Make version 3 or later (only required to use the Makefile setup)
- Bash version 3 or later (only required to use the Makefile setup)

Overview of `make` targets:

- `make`     -- builds gotalk and runs go tests & tests in the examples dir.
- `make dev` -- start iterative development mode. See below for details.
- `make fmt` -- run `gofmt` on all source code
- `make doc` -- generate & serve documentation at
  `http://localhost:6060/pkg/github.com/rsms/gotalk/`


### Iterative development mode

`make dev` starts an iterative development mode which watches source files
for changes and performs the following tasks when started and whenever source
files change:

1. `gofmt` to format source code if needed
2. `go test` to run all go tests (but not tests in `examples/`)
3. `go tool cover` to generate code coverage report

When `make dev` is running, live code coverage report can be viewed at
`http://localhost:8148/`


### Building & testing without make and bash

    go test


## Making a contribution

Contributions this project is looking for:

- Implementations of Gotalk in languages other than Go and JavaScript
- Bug fixes (please first file an issue)
- Security vulnerabilities
- Performance improvements (with some reservations for complexity)

For small things like spelling mistakes, readme changes or code comments,
please open an issue instead of submitting a pull request.


### Contribution checklist

1. Check [issues](https://github.com/rsms/gotalk/issues?q=is%3Aissue) for an
   existing change.
   If there is one, chime in the conversation.
   If not, please file a new issue.

2. If this is a bug or security vulnerability, make sure to have a clear and
   minimmal reproduction. Ideally something that is as focused, small and
   specific as possible that proves the bug or vulnerability. Include this in
   the issue (or add it to an existing issue.)

3. Optional but very much appreciated: Add a test (e.g. `bugname_test.go`)
   which triggers the bug before your fix and passes with your fix. This can
   be the same thing as checklist-item 2.

4. Make sure that the test suite passes by running `make test`.
   It is also a good idea to run one of the complex example programs in the
   `examples/` directory and verify that it works. E.g.
   `cd examples/websocket-chat && go build && ./websocket-chat`

5. [Submit a PR](https://github.com/rsms/gotalk/compare)
   referencing the issue


Thank you!
