import argparse
import json
from pathlib import Path

from cfb_scraper.sources import espn

DATA_DIR = Path(__file__).resolve().parent.parent.parent / "data" / "cfb"

# ESPN seasontype for each game_type the scraper writes out.
_SEASONTYPE = {"REG": 2, "POST": 3}


def build_season_data(season: int, games: list[dict]) -> dict:
    teams = sorted({g["home_team"] for g in games} | {g["away_team"] for g in games})
    return {
        "season": season,
        "teams": teams,
        "games": games,
    }


def _completed_weeks(games: list[dict]) -> set[tuple[int, int]]:
    """Return (week, seasontype) pairs where every previously-recorded
    game is final - safe to skip re-fetching, since nothing about an
    already-final week can change.
    """
    by_week: dict[tuple[int, int], list[dict]] = {}
    for g in games:
        seasontype = _SEASONTYPE.get(g["game_type"])
        if seasontype is None:
            continue
        by_week.setdefault((g["week"], seasontype), []).append(g)
    return {
        key
        for key, week_games in by_week.items()
        if week_games and all(g["status"] == "final" and g["home_score"] is not None and g["away_score"] is not None for g in week_games)
    }


def _sort_key(g: dict) -> tuple:
    return (g["week"], g["date"], g["time"] or "")


def main():
    parser = argparse.ArgumentParser(description="Scrape FBS college football game results")
    parser.add_argument("season", type=int)
    parser.add_argument("--postseason", action="store_true", help="include postseason games")
    parser.add_argument(
        "--force",
        action="store_true",
        help="re-fetch every week, including ones already recorded as fully final",
    )
    parser.add_argument("--out", type=Path, default=None)
    args = parser.parse_args()

    out_path = args.out or DATA_DIR / f"{args.season}.json"

    old_games: list[dict] = []
    if out_path.exists() and not args.force:
        old_games = json.loads(out_path.read_text()).get("games", [])
    skip = _completed_weeks(old_games)

    new_games = [g.to_dict() for g in espn.fetch_games(args.season, include_postseason=args.postseason, skip=skip)]
    skipped_games = [g for g in old_games if (g["week"], _SEASONTYPE.get(g["game_type"])) in skip]
    games = sorted(new_games + skipped_games, key=_sort_key)

    data = build_season_data(args.season, games)
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(data, indent=2))

    summary = f"Wrote {len(games)} games ({len(data['teams'])} teams) to {out_path}"
    if skip:
        summary += f" ({len(skip)} already-final weeks skipped)"
    print(summary)


if __name__ == "__main__":
    main()
