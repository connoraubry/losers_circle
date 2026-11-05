#!/usr/bin/env bash
# Pull up-to-date NFL and CFB data for a season and regenerate the static
# site. Usage: scripts/update-site.sh [season]
# Defaults season to the current year.
set -euo pipefail
cd "$(dirname "$0")/.."

SEASON="${1:-$(date +%Y)}"
VENV_DIR="$PWD/venv"
PYTHON="$VENV_DIR/bin/python"

# venv/ is tied to the machine it was created on (its interpreter symlinks
# are absolute paths), so build it fresh here rather than assuming one
# already exists.
if [ ! -x "$PYTHON" ]; then
    echo "==> No venv found, creating one"
    python3 -m venv "$VENV_DIR"
    "$PYTHON" -m pip install --quiet -r requirements.txt
fi

echo "==> Fetching ${SEASON} NFL data"
PYTHONPATH=scraper "$PYTHON" -m nfl_scraper "$SEASON" --postseason

echo "==> Fetching ${SEASON} CFB data"
PYTHONPATH=scraper "$PYTHON" -m cfb_scraper "$SEASON" --postseason

echo "==> Rebuilding static site"
go run ./cmd/buildsite

echo "==> Done. Site written to site/"
