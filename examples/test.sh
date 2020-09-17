#!/bin/bash -e
set -e
cd "$(dirname "$0")"

TEST_EXAMPLES=( \
  tcp \
  tls \
  pipe \
  limits \
  read-timeout \
)

SILENT=false
if [[ "$1" == "-silent" ]]; then
  SILENT=true
fi

echo "building examples/{$(IFS=, ; echo "${TEST_EXAMPLES[*]}")}"
for name in "${TEST_EXAMPLES[@]}"; do
  pushd "$name" >/dev/null
  go build &
  popd >/dev/null
done

wait

for name in "${TEST_EXAMPLES[@]}"; do
  echo "run $name/$name"
  pushd "$name" >/dev/null
  if $SILENT; then
    ./"$name" >/dev/null
  else
    ./"$name"
  fi
  popd >/dev/null
done
