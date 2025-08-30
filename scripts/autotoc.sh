#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
README="README.md"

"${SCRIPT_DIR}/../bin/gh-md-toc" --no-backup --hide-footer "$README"
