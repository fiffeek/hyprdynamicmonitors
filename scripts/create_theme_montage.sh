#!/bin/bash

set -o pipefail
set -ex

OUTPUT_DIR="preview/output"
SPLICE_OUTPUT="${OUTPUT_DIR}/themes-showcase.mp4"
SECONDS_PER_THEME="${1:-3}" # Default 3 seconds per theme

# Check if ffmpeg is installed
if ! command -v ffmpeg &>/dev/null; then
  echo "Error: ffmpeg is not installed. Please install it first."
  echo "  Arch: sudo pacman -S ffmpeg"
  echo "  Ubuntu: sudo apt install ffmpeg"
  exit 1
fi

# Find all demo GIFs
mapfile -t DEMO_GIFS < <(find "${OUTPUT_DIR}" -name "demo-*.gif" | sort)

if [ ${#DEMO_GIFS[@]} -eq 0 ]; then
  echo "Error: No demo GIFs found in ${OUTPUT_DIR}"
  echo "Generate theme GIFs first using prepare_themed_demo.sh"
  exit 1
fi

echo "Found ${#DEMO_GIFS[@]} theme GIFs:"
for gif in "${DEMO_GIFS[@]}"; do
  echo "  - $(basename "$gif")"
done

# Get duration and fps of first GIF (assuming all are the same)
DURATION=$(ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 "${DEMO_GIFS[0]}")
DURATION_INT=$(printf "%.0f" "$DURATION")
SOURCE_FPS=$(ffprobe -v error -select_streams v -of default=noprint_wrappers=1:nokey=1 -show_entries stream=r_frame_rate "${DEMO_GIFS[0]}")

echo ""
echo "GIF duration: ${DURATION_INT}s, FPS: ${SOURCE_FPS}"
echo "Creating spliced showcase (switching themes every ${SECONDS_PER_THEME}s)..."

# Create temporary directory for processing
TEMP_DIR=$(mktemp -d)
trap 'rm -rf "${TEMP_DIR}"' EXIT

# Calculate number of segments
NUM_SEGMENTS=$((DURATION_INT / SECONDS_PER_THEME))

echo "Will create ${NUM_SEGMENTS} segments"

# Convert each GIF segment directly to MP4
CONCAT_LIST="${TEMP_DIR}/concat.txt"

for ((seg = 0; seg < NUM_SEGMENTS; seg++)); do
  # Calculate which theme to use (cycle through themes)
  theme_idx=$((seg % ${#DEMO_GIFS[@]}))
  gif="${DEMO_GIFS[$theme_idx]}"
  theme_name=$(basename "$gif" .gif | sed 's/demo-//')

  # Calculate start time for this segment
  start_time=$((seg * SECONDS_PER_THEME))

  echo "Segment $((seg + 1))/${NUM_SEGMENTS}: ${theme_name} (${start_time}s-$((start_time + SECONDS_PER_THEME))s)"

  # Convert segment directly to MP4
  segment_file="${TEMP_DIR}/segment_${seg}.mp4"
  ffmpeg -ss ${start_time} -t ${SECONDS_PER_THEME} -i "$gif" \
    -vf "drawtext=text='${theme_name}':fontcolor=white:fontsize=24:box=1:boxcolor=black@0.6:boxborderw=8:x=(w-text_w)/2:y=h-th-20" \
    -c:v libx264 -preset slow -crf 23 -pix_fmt yuv420p \
    -y "$segment_file" 2>&1 | grep -v "frame=" || true

  echo "file '${segment_file}'" >>"${CONCAT_LIST}"
done

# Concatenate all MP4 segments
echo ""
echo "Concatenating ${NUM_SEGMENTS} segments..."
ffmpeg -f concat -safe 0 -i "${CONCAT_LIST}" \
  -c copy -movflags +faststart \
  -y "${SPLICE_OUTPUT}" 2>&1 | grep -v "frame=" || true

echo ""
echo "✓ Theme showcase created successfully!"
echo "  File: ${SPLICE_OUTPUT}"
echo "  Themes cycle every ${SECONDS_PER_THEME}s through ${DURATION_INT}s timeline"
echo "  Size: $(du -h "${SPLICE_OUTPUT}" | cut -f1)"
