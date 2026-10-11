#!/usr/bin/env bash
# Run the college football scraper. Example: ./run_cfb.sh 2024 --postseason
set -euo pipefail
cd "$(dirname "$0")"
source .venv/bin/activate
python -m cfb_scraper "$@"
