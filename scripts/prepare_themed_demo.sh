#!/bin/bash

set -o pipefail
set -e

THEME_HDM="$1"
THEME_VHS="$2"
DEMO_BASE_TAPE="preview/tapes/base.tape"
DEMO_TAPE="preview/tapes/demo.tape"
DEMO_CFG="preview/tapes/configs/demo.toml"

# Validate arguments
if [ -z "$THEME_HDM" ] || [ -z "$THEME_VHS" ]; then
  echo "Usage: $0 <theme_hdm> <theme_vhs>"
  echo "Example: $0 tokyo-night dracula"
  exit 1
fi

# Validate theme exists
if [ ! -f "themes/static/$THEME_HDM/theme.toml" ]; then
  echo "Error: HDM theme not found at themes/static/$THEME_HDM/theme.toml"
  exit 1
fi

# Add VHS theme at the start of DEMO_TAPE (after Require lines)
{
  # Keep the Require lines
  grep "^Require" "$DEMO_BASE_TAPE"
  echo ""
  echo "Set Theme \"$THEME_VHS\""
  # Add the rest of the file (skip Require lines)
  grep -v "^Require" "$DEMO_BASE_TAPE"
} >"${DEMO_BASE_TAPE}.tmp"
mv "${DEMO_BASE_TAPE}.tmp" "$DEMO_BASE_TAPE"

# Add TUI colors section at the end of DEMO_CFG
# First, remove any existing [tui.colors] section
sed -i '/^\[tui\.colors\]/,/^source = /d' "$DEMO_CFG"

# Then add the new section
{
  cat "$DEMO_CFG"
  echo ""
  echo "[tui.colors]"
  echo "source = \"../../../themes/static/$THEME_HDM/theme.toml\""
} >"${DEMO_CFG}.tmp"
mv "${DEMO_CFG}.tmp" "$DEMO_CFG"

# Replace Output line in DEMO_TAPE to include theme name
sed -i "s|Output preview/output/demo\.gif|Output preview/output/demo-${THEME_HDM}.gif|g" "$DEMO_TAPE"

echo "Successfully configured themes:"
echo "  HDM theme: $THEME_HDM"
echo "  VHS theme: $THEME_VHS"
echo "  Output: preview/output/demo-${THEME_HDM}.gif"
