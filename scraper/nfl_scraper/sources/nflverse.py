import csv
import io
from pathlib import Path

import requests

from nfl_scraper.models import Game

GAMES_CSV_URL = "https://github.com/nflverse/nflverse-data/releases/download/schedules/games.csv"
CACHE_PATH = Path(__file__).resolve().parent.parent.parent / "data" / ".cache" / "games.csv"


def _load_csv_text(refresh: bool = False) -> str:
    if not refresh and CACHE_PATH.exists():
        return CACHE_PATH.read_text()

    resp = requests.get(GAMES_CSV_URL, timeout=30)
    resp.raise_for_status()
    CACHE_PATH.parent.mkdir(parents=True, exist_ok=True)
    CACHE_PATH.write_text(resp.text)
    return resp.text


def fetch_games(season: int, include_postseason: bool = False, refresh: bool = False) -> list[Game]:
    text = _load_csv_text(refresh=refresh)
    reader = csv.DictReader(io.StringIO(text))

    games = []
    for row in reader:
        if int(row["season"]) != season:
            continue
        if row["game_type"] != "REG" and not include_postseason:
            continue
        games.append(
            Game(
                season=season,
                week=int(row["week"]),
                game_type=row["game_type"],
                date=row["gameday"],
                time=row["gametime"] or None,
                home_team=row["home_team"],
                away_team=row["away_team"],
                home_score=int(row["home_score"]) if row["home_score"] else None,
                away_score=int(row["away_score"]) if row["away_score"] else None,
                status="final",
            )
        )
    return games
