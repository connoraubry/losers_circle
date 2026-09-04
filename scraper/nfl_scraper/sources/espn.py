import requests

from nfl_scraper.models import Game

SCOREBOARD_URL = "https://site.api.espn.com/apis/site/v2/sports/football/nfl/scoreboard"

# ESPN seasontype: 1=preseason, 2=regular season, 3=postseason
REG_SEASON_WEEKS = range(1, 19)
POSTSEASON_WEEKS = range(1, 6)  # WC, DIV, CON, PRO BOWL(skipped), SB


def _status(competition: dict) -> str:
    state = competition["status"]["type"]["state"]  # "pre", "in", "post"
    return {"pre": "scheduled", "in": "in_progress", "post": "final"}.get(state, "scheduled")


def _fetch_week(season: int, week: int, season_type: int) -> list[Game]:
    resp = requests.get(
        SCOREBOARD_URL,
        params={"dates": season, "seasontype": season_type, "week": week},
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
                game_type="REG" if season_type == 2 else "POST",
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


def fetch_games(season: int, include_postseason: bool = False) -> list[Game]:
    games = []
    for week in REG_SEASON_WEEKS:
        games.extend(_fetch_week(season, week, season_type=2))
    if include_postseason:
        for week in POSTSEASON_WEEKS:
            games.extend(_fetch_week(season, week, season_type=3))
    return games
