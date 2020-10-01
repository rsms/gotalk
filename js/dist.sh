#!/bin/bash -e
OUTER_PWD=$PWD
cd "$(dirname "$0")"

GO_VERSION=$(grep 'const Version =' ../version.go | cut -d '"' -f 2)
JS_VERSION=$(node -e 'process.stdout.write(require("./package.json").version)')

BUMP_VERSION=
if [ "$1" == "" ]; then true
elif [ "$1" == "-major" ]; then BUMP_VERSION=major
elif [ "$1" == "-minor" ]; then BUMP_VERSION=minor
elif [ "$1" == "-patch" ]; then BUMP_VERSION=patch
else
  cat << _MESSAGE_ >&2
$0: Unexpected option $1
Usage: $0 [-major | -minor | -patch]
  -major    Bump major version. e.g. 1.2.3 => 2.0.0
  -minor    Bump minor version. e.g. 1.2.3 => 1.3.0
  -patch    Bump patch version. e.g. 1.2.3 => 1.2.4
  (nothing) Leave version in package.json unchanged ($JS_VERSION)
_MESSAGE_
  exit 1
fi

# checkout products so that npm version doesn't fail. These are regenerated anyways.
git checkout -- gotalk.js gotalk.js.map ../jslib.go

# Make sure there are no uncommitted changes.
# Note that we look in the entire project, not just the js directory since a new js build affects
# the gotalk go library (embed.)
if [ -n "$(git status --untracked-files=no --ignore-submodules=dirty --porcelain)" ]; then
  echo "There are uncommitted changes:" >&2
  git status -s --untracked-files=no --ignore-submodules=dirty
  exit 1
fi

# Bump version in package.js. This Will fail and stop the script if git is not clean
JS_VERSION_PREV=$JS_VERSION
if [ "$BUMP_VERSION" != "" ]; then
  npm --no-git-tag-version version "$BUMP_VERSION"
  JS_VERSION=$(node -e 'process.stdout.write(require("./package.json").version)')
fi

# make sure that the major version matches between gotalk and gotalk.js
GO_VERSION_MAJOR=$(echo $GO_VERSION | cut -d . -f 1)
JS_VERSION_MAJOR=$(echo $JS_VERSION | cut -d . -f 1)
if [[ "$GO_VERSION_MAJOR" != "$JS_VERSION_MAJOR" ]]; then
  # restore package.json (in case version field was updated by bump)
  echo "Major version mismatch between go and js libraries: go ${GO_VERSION}, js ${JS_VERSION}" >&2
  echo "First update version.go then re-run this script."
  echo "Reverting change to package.json"
  git checkout -- package.json package-lock.json
  exit 1
fi

# build gotalk.{js,js.map} & ../jslib.go
echo "" ; echo "node build.js"
node build.js

# commit, tag and push git
echo "Ready to commit, publish & push:"
echo ""
if [[ "$PWD" != "$OUTER_PWD" ]]; then
  echo "  cd '$PWD'"
fi
cat << _MESSAGE_
  git commit -m "js: release v${JS_VERSION}" . ../jslib.go
  npm publish
  git push

Since jslib.go changed, you may want to publish a new version of gotalk as well:

  cd '$(dirname "$PWD")' && make release

_MESSAGE_
