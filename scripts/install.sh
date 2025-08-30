#!/bin/bash

set -o pipefail
set -eu

ARCH=$(uname -m)

case $ARCH in
x86_64) RELEASE_ARCH="x86_64" ;;
aarch64 | arm64) RELEASE_ARCH="arm64" ;;
i686 | i386) RELEASE_ARCH="i386" ;;
*) echo "Unsupported architecture: $ARCH" && exit 1 ;;
esac

OS=$(uname -s)
OWNER="fiffeek"
PROJ="hyprdynamicmonitors"
GITHUB_API_ENDPOINT="api.github.com"
GITHUB_ENDPOINT="github.com"
TOKEN="${1:-$GITHUB_TOKEN}"
TAR="hyprdynamicmonitors_${OS}_${RELEASE_ARCH}.tar.gz"
SUMS="checksums.txt"
DESTINATION="${DESTDIR:-"$HOME/.local/bin"}"
BINARY="hyprdynamicmonitors"

function download_asset {
  local output file latest_tag
  output="$1"
  file="$2"
  latest_tag="$3"

  if [ -z "$TOKEN" ]; then
    curl -L -o "$output" \
      "https://$GITHUB_ENDPOINT/$OWNER/$PROJ/releases/download/$latest_tag/$file"
  else
    local asset_id

    asset_id=$(curl -sL -H "Authorization: token $TOKEN" \
      -H "Accept: application/vnd.github.v3.raw" \
      "https://$GITHUB_API_ENDPOINT/repos/$OWNER/$PROJ/releases" |
      jq ". | map(select(.tag_name == \"$latest_tag\"))[0].assets | map(select(.name == \"$file\"))[0].id")

    if [ "$asset_id" = "null" ]; then
      echo "ERROR: version not found $latest_tag"
      exit 1
    fi

    curl -sL \
      -H "Authorization: token $TOKEN" \
      -H 'Accept: application/octet-stream' \
      "https://$TOKEN:@$GITHUB_API_ENDPOINT/repos/$OWNER/$PROJ/releases/assets/$asset_id" >"$output"
  fi
}

function download_assets {
  local latest_tag
  latest_tag="$1"
  download_asset "$TAR" "hyprdynamicmonitors_${OS}_${RELEASE_ARCH}.tar.gz" "$latest_tag"
  download_asset "$SUMS" "hyprdynamicmonitors_${latest_tag#v}_checksums.txt" "$latest_tag"
}

function fetch_latest_tag {
  local tag
  local auth_header=()

  if [ -n "$TOKEN" ]; then
    auth_header=("-H" "Authorization: token $TOKEN")
  fi

  tag=$(curl -s "${auth_header[@]}" \
    "https://api.github.com/repos/$OWNER/$PROJ/releases/latest" | jq '.tag_name' | sed 's/"//g')

  if [ -z "$tag" ] || [ "$tag" == "null" ]; then
    echo "Failed while fetching the last tag"
    exit 1
  fi

  echo "$tag"
}

function checksums {
  if command -v sha256sum >/dev/null 2>&1; then
    grep "$TAR" "$SUMS" | sha256sum -c -
  elif command -v shasum >/dev/null 2>&1; then
    grep "$TAR" "$SUMS" | shasum -a 256 -c -
  fi
}

function install_hdm {
  tar -xzf "$TAR"
  if [ ! -f "$BINARY" ]; then
    echo "Binary not found in the tar archive"
    exit 1
  fi

  mkdir -p "$DESTINATION"
  mv "$BINARY" "$DESTINATION/"
  rm "$TAR" "$SUMS"

}

tag=$(fetch_latest_tag)
download_assets "$tag"
checksums
install_hdm
