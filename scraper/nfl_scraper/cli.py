import argparse
import json
from pathlib import Path

from nfl_scraper.models import Game
from nfl_scraper.sources import espn, nflverse

DATA_DIR = Path(__file__).resolve().parent.parent.parent / "data"


def build_season_data(season: int, games: list[Game]) -> dict:
    teams = sorted({g.home_team for g in games} | {g.away_team for g in games})
    return {
        "season": season,
        "teams": teams,
        "games": [g.to_dict() for g in games],
    }


def main():
    parser = argparse.ArgumentParser(description="Scrape NFL season game results")
    parser.add_argument("season", type=int)
    parser.add_argument("--source", choices=["nflverse", "espn"], default="nflverse")
    parser.add_argument("--postseason", action="store_true", help="include postseason games")
    parser.add_argument("--refresh", action="store_true", help="bypass the nflverse CSV cache")
    parser.add_argument("--out", type=Path, default=None)
    args = parser.parse_args()

    if args.source == "nflverse":
        games = nflverse.fetch_games(args.season, include_postseason=args.postseason, refresh=args.refresh)
    else:
        games = espn.fetch_games(args.season, include_postseason=args.postseason)

    data = build_season_data(args.season, games)
    out_path = args.out or DATA_DIR / f"{args.season}.json"
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(data, indent=2))
    print(f"Wrote {len(games)} games ({len(data['teams'])} teams) to {out_path}")


if __name__ == "__main__":
    main()
