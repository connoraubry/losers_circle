#!/usr/bin/env bash
# Run the men's college basketball scraper. Example: ./run_cbb.sh 2025 --postseason
set -euo pipefail
cd "$(dirname "$0")"
source .venv/bin/activate
python -m cbb_scraper "$@"
