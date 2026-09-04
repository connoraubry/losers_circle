from dataclasses import dataclass
from typing import Optional


@dataclass
class Game:
    season: int
    week: int
    game_type: str  # REG, WC, DIV, CON, SB, etc.
    date: str  # ISO date, e.g. "2025-09-05"
    time: Optional[str]  # kickoff time, e.g. "20:20Z"
    home_team: str
    away_team: str
    home_score: Optional[int]
    away_score: Optional[int]
    status: str = "final"  # scheduled, in_progress, final

    @property
    def winner(self) -> Optional[str]:
        if self.home_score is None or self.away_score is None:
            return None
        if self.home_score == self.away_score:
            return None
        return self.home_team if self.home_score > self.away_score else self.away_team

    def to_dict(self) -> dict:
        return {
            "season": self.season,
            "week": self.week,
            "game_type": self.game_type,
            "date": self.date,
            "time": self.time,
            "home_team": self.home_team,
            "away_team": self.away_team,
            "home_score": self.home_score,
            "away_score": self.away_score,
            "status": self.status,
            "winner": self.winner,
        }
