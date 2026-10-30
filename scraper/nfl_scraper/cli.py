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


def _game_key(game: dict) -> tuple:
    return (game["week"], game["game_type"], game["home_team"], game["away_team"])


def diff_games(old_games: list[dict], new_games: list[dict]) -> tuple[int, int]:
    """Return (added, updated) counts of new_games relative to old_games."""
    old_by_key = {_game_key(g): g for g in old_games}
    added = updated = 0
    for game in new_games:
        old = old_by_key.get(_game_key(game))
        if old is None:
            added += 1
        elif (old["home_score"], old["away_score"], old["status"]) != (
            game["home_score"],
            game["away_score"],
            game["status"],
        ):
            updated += 1
    return added, updated


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

    added = updated = None
    if out_path.exists():
        old_data = json.loads(out_path.read_text())
        added, updated = diff_games(old_data.get("games", []), data["games"])

    out_path.write_text(json.dumps(data, indent=2))

    summary = f"Wrote {len(games)} games ({len(data['teams'])} teams) to {out_path}"
    if added is not None:
        summary += f" ({added} added, {updated} updated)"
    print(summary)


if __name__ == "__main__":
    main()
