#!/usr/bin/env bash
set -euo pipefail

BINARY="$1"
COMMAND="${2:-""}"
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

README="README.md"
TMP_FILE="$(mktemp)"

"$SCRIPT_DIR"/../"$BINARY" $COMMAND --help >"$TMP_FILE" 2>&1

sed -i 's/hdmtest/hyprdynamicmonitors/g' "$TMP_FILE"
sed -i '1s/^/```text\n/' "$TMP_FILE"
echo '```' >>"$TMP_FILE"

START="<!-- START ${COMMAND}help -->"
END="<!-- END ${COMMAND}help -->"

awk -v start="$START" -v end="$END" -v file="$TMP_FILE" '
BEGIN { inside=0 }
{
    if ($0 ~ start) {
        print
        while ((getline line < file) > 0) print line
        close(file)
        inside=1
        next
    }
    if ($0 ~ end) {
        inside=0
        print
        next
    }
    if (inside == 0) {
        print
    }
}
' "$README" >"${README}.tmp"

mv "${README}.tmp" "$README"
rm -f "$TMP_FILE"
