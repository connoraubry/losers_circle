#!/usr/bin/env bash
# Run the NFL scraper. Example: ./run.sh 2024 --postseason
set -euo pipefail
cd "$(dirname "$0")"
source .venv/bin/activate
python -m nfl_scraper "$@"
