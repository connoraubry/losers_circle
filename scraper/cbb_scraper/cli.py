import argparse
import json
from pathlib import Path

from cbb_scraper.sources import espn

DATA_DIR = Path(__file__).resolve().parent.parent.parent / "data" / "cbb"


def build_season_data(season: int, games: list[dict]) -> dict:
    teams = sorted({g["home_team"] for g in games} | {g["away_team"] for g in games})
    return {
        "season": season,
        "teams": teams,
        "games": games,
    }


def _completed_dates(games: list[dict]) -> set[str]:
    """Return ISO date strings where every previously-recorded game that
    day is final - safe to skip re-fetching. Days with zero recorded
    games (no D1 games that date) aren't included here, since there's no
    record to tell "no games that day" apart from "never checked", so
    those still get re-requested every run; this is the only residual
    inefficiency, and it's a small minority of the ~130-day season.
    """
    by_date: dict[str, list[dict]] = {}
    for g in games:
        by_date.setdefault(g["date"], []).append(g)
    return {
        date
        for date, day_games in by_date.items()
        if day_games and all(g["status"] == "final" and g["home_score"] is not None and g["away_score"] is not None for g in day_games)
    }


def _sort_key(g: dict) -> tuple:
    return (g["date"], g["time"] or "")


def main():
    parser = argparse.ArgumentParser(description="Scrape D-I men's college basketball game results")
    parser.add_argument("season", type=int, help="season, named by its end year (e.g. 2025 for 2024-25)")
    parser.add_argument("--postseason", action="store_true", help="include conference tournament games")
    parser.add_argument(
        "--force",
        action="store_true",
        help="re-fetch every date, including ones already recorded as fully final",
    )
    parser.add_argument("--out", type=Path, default=None)
    args = parser.parse_args()

    out_path = args.out or DATA_DIR / f"{args.season}.json"

    old_games: list[dict] = []
    if out_path.exists() and not args.force:
        old_games = json.loads(out_path.read_text()).get("games", [])
    skip_dates = _completed_dates(old_games)

    new_games = [
        g.to_dict() for g in espn.fetch_games(args.season, include_postseason=args.postseason, skip_dates=skip_dates)
    ]
    skipped_games = [g for g in old_games if g["date"] in skip_dates]
    games = sorted(new_games + skipped_games, key=_sort_key)

    data = build_season_data(args.season, games)
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(data, indent=2))

    summary = f"Wrote {len(games)} games ({len(data['teams'])} teams) to {out_path}"
    if skip_dates:
        summary += f" ({len(skip_dates)} already-final days skipped)"
    print(summary)


if __name__ == "__main__":
    main()
