#!/usr/bin/env bash
# Scrape an NFL season's game results into data/<year>.json
# Usage: ./scrape.sh <year> [--source nflverse|espn] [--postseason] [--refresh]
set -euo pipefail
cd "$(dirname "$0")/scraper"

if [ ! -d .venv ]; then
  python3 -m venv .venv
  source .venv/bin/activate
  pip install -q -r requirements.txt
else
  source .venv/bin/activate
fi

python -m nfl_scraper "$@"
