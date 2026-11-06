#!/usr/bin/env bash
# Pull up-to-date NFL, CFB, and CBB data for a season and regenerate the
# static site. Usage: scripts/update-site.sh [season]
# Defaults season to the current year (NFL/CFB convention; the matching
# CBB season is derived separately below).
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

# CBB seasons are named by their end year (the season that starts Nov
# 2026 is "2027"), unlike NFL/CFB's single-calendar-year seasons, so this
# can't reuse $SEASON directly: Jan-Oct falls within (or just after) the
# season that started last Nov, named this calendar year; Nov-Dec means
# a new season just started, named next calendar year.
CBB_MONTH="$(date +%m)"
if [ "$((10#$CBB_MONTH))" -ge 11 ]; then
    CBB_SEASON="$((SEASON + 1))"
else
    CBB_SEASON="$SEASON"
fi

echo "==> Fetching ${CBB_SEASON} CBB data"
PYTHONPATH=scraper "$PYTHON" -m cbb_scraper "$CBB_SEASON" --postseason

echo "==> Rebuilding static site"
go run ./cmd/buildsite

echo "==> Done. Site written to site/"
