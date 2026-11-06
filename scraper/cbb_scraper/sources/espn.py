import datetime as dt

import requests

from cbb_scraper.models import Game

SCOREBOARD_URL = "https://site.api.espn.com/apis/site/v2/sports/basketball/mens-college-basketball/scoreboard"

# Division I group id - excludes non-D1 teams from scoreboard results.
D1_GROUP = 50

# The scoreboard API has no week/seasontype params for basketball (unlike
# CFB) and doesn't accept date ranges, so games are fetched one date at a
# time. ESPN names a season by its end year (the 2024-25 season is
# "2025"), matching the `season` argument here.
#
# Conference tournaments run through mid-March and are classified by ESPN
# as regular-season games (not postseason), so a calendar cutoff - not
# ESPN's seasontype - is what separates "regular season" from
# "conference tournaments" for -postseason purposes. The NCAA tournament
# itself is never fetched: it's almost entirely cross-conference, so it
# would be dropped by conference filtering downstream anyway.
REG_SEASON_START = (11, 1)  # Nov 1 of season - 1
REG_SEASON_END = (3, 9)  # Mar 9 of season
POSTSEASON_END = (3, 17)  # Mar 17 of season - conference tournaments only


def _status(competition: dict) -> str:
    state = competition["status"]["type"]["state"]  # "pre", "in", "post"
    return {"pre": "scheduled", "in": "in_progress", "post": "final"}.get(state, "scheduled")


def _fetch_day(season: int, date: dt.date, week: int, game_type: str) -> list[Game]:
    resp = requests.get(
        SCOREBOARD_URL,
        params={"dates": date.strftime("%Y%m%d"), "groups": D1_GROUP},
        timeout=15,
    )
    resp.raise_for_status()
    data = resp.json()

    games = []
    for event in data.get("events", []):
        comp = event["competitions"][0]
        competitors = {c["homeAway"]: c for c in comp["competitors"]}
        home, away = competitors["home"], competitors["away"]
        status = _status(comp)

        def score(competitor):
            if status == "scheduled":
                return None
            raw = competitor.get("score")
            return int(raw) if raw is not None else None

        games.append(
            Game(
                season=season,
                week=week,
                game_type=game_type,
                date=event["date"][:10],
                time=event["date"][11:],
                home_team=home["team"]["abbreviation"],
                away_team=away["team"]["abbreviation"],
                home_score=score(home),
                away_score=score(away),
                status=status,
            )
        )
    return games


def _date_range(start: dt.date, end: dt.date) -> list[dt.date]:
    days = (end - start).days
    return [start + dt.timedelta(days=i) for i in range(days + 1)]


def fetch_games(
    season: int, include_postseason: bool = False, skip_dates: set[str] | None = None
) -> list[Game]:
    """Fetch every date in the season range not in skip_dates (ISO
    "YYYY-MM-DD" strings). skip_dates holds dates the caller already
    knows are fully final - see cli.py's _completed_dates - so a rerun
    doesn't re-request one day at a time for a ~130-day season when most
    of it can't have changed since the last run.
    """
    skip_dates = skip_dates or set()
    season_start = dt.date(season - 1, *REG_SEASON_START)
    reg_season_end = dt.date(season, *REG_SEASON_END)

    games = []
    for date in _date_range(season_start, reg_season_end):
        if date.isoformat() in skip_dates:
            continue
        week = (date - season_start).days // 7 + 1
        games.extend(_fetch_day(season, date, week, game_type="REG"))

    if include_postseason:
        postseason_end = dt.date(season, *POSTSEASON_END)
        for date in _date_range(reg_season_end + dt.timedelta(days=1), postseason_end):
            if date.isoformat() in skip_dates:
                continue
            week = (date - season_start).days // 7 + 1
            games.extend(_fetch_day(season, date, week, game_type="POST"))

    return games
