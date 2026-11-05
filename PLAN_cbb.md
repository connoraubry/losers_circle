# Plan: add P4 men's basketball

Add college men's basketball (SEC, Big Ten, ACC, Big 12) as a third
dataset type alongside NFL and CFB, reusing the existing sport-agnostic
cycle-detection engine. Decisions below were settled via /grill before
implementation started.

## Facts confirmed against the live ESPN API

- `site.api.espn.com/apis/site/v2/sports/basketball/mens-college-basketball/scoreboard`
  with `groups=50` returns full D-I slates, same pattern as the CFB
  scraper's `groups=80` (FBS). No sport-specific "major conference"
  group id exists — conference scoping happens client-side via the
  membership table, exactly like CFB already does.
- Team abbreviations match the CFB scraper's exactly for the same school
  (`UGA`, `AUB`, `ALA`, etc. all confirmed in a live response).
- The scoreboard API has **no `week` field** — CBB has no real concept
  of a week; games are fetched by date, not week/seasontype.
- ESPN labels a season by its **end** year: the 2024-25 season reports
  `season.year: 2025`, unlike CFB's single-calendar-year seasons.
- The API does not support date-range queries (`dates=20241104-20241110`
  did not return the expected multi-day set) — the scraper will fetch
  one day at a time across the season, same architecture as CFB's
  per-week fetch loop, just with many more requests (~150 days vs ~16
  weeks).

## Decisions

1. **Week field**: synthesize a calendar-week bucket (ISO week number
   relative to season start) rather than a running per-game index or
   leaving it unused. Keeps `-week`/`-upcoming`/`-sweep`/`-sweep-next`
   roughly meaningful for CBB datasets without a deeper rework of those
   flags' week-based grouping.
2. **Season file naming**: match ESPN's own end-year convention —
   `data/cbb/2025.json` for the 2024-25 season, label "2025 SEC (MBB)".
   Do not translate to CFB's start-year convention; keeps the scraper
   simple and matches what ESPN/most sites call the season.
3. **Conference scope**: intra-conference only, exactly reusing
   `cfbconf.FilterSeason`'s existing behavior pattern for the new
   `cbbconf` package. No combined cross-conference/NCAA-tournament
   dataset for now — smallest surface area, consistent with how CFB
   already drops cross-conference bowl games in its per-conference view.
4. **Dataset naming**: key `cbb-sec-2025`, label `"2025 SEC (MBB)"` —
   sport-prefixed key avoids collision with CFB's `sec-2025`/"2025 SEC".
5. **League dropdown value**: `"CBB"` — matches the `cbb-*` key prefix
   exactly (not `"MBB"`, which is used only in the per-dataset label
   suffix), so the key prefix and the dropdown value are the same
   string.
6. **CLI flag**: add `-sport` to `cmd/cycle`, defaulting to `cfb` when
   `-conference` is set (preserves today's `-conference SEC` meaning CFB
   unchanged); `-sport cbb -conference SEC` selects the new package/data
   dir.
7. **Postseason**: mirror CFB's opt-in `-postseason` flag rather than
   including conference-tournament games by default, for CLI-surface
   parity with the CFB scraper even though the filtering rationale
   differs slightly (NCAA tournament cross-conference games are dropped
   by the conference filter regardless of this flag; conference
   tournament games are intra-conference and would be included either
   way once fetched — this flag controls whether postseason date ranges
   are fetched at all).
8. **Backfill scope**: 2024, 2025, and the in-progress 2026 season.

## 1. Scraper: `scraper/cbb_scraper/`

Mirror `scraper/cfb_scraper/` structure (`models.py`, `cli.py`,
`sources/espn.py`, `__main__.py`).

- Fetch per-day across the regular season date range (~early Nov to
  early March), plus conference tournament dates under `-postseason`
  (decision 7).
- Compute `week` as a calendar-week bucket relative to season start
  (decision 1).
- Output the same JSON shape as CFB (`season`, `teams`, `games`) so
  `nfldata.LoadFile` reads it unmodified. Write to
  `data/cbb/<espn-end-year>.json` (decision 2).

## 2. Conference membership: `internal/cbbconf/`

Mirror `internal/cfbconf/cfbconf.go`'s structure but with its own
hand-maintained `membership[season][conference][]teamAbbrev` table —
basketball-only membership differs from football's (e.g. UConn joined
the Big 12 for basketball without a football program in the conference).
Cover 2024, 2025, 2026 (decision 8). Reuse `FilterSeason`'s
intra-conference-only filtering logic/shape (decision 3).

## 3. Team colors: `internal/webgraph/teamcolors.go`

Reuse `cfbTeamColors` directly — confirmed abbreviations match, so no
second hand-written color map needed for the schools already covered
there.

## 4. Backend wiring: `internal/webgraph/cache.go`, `cmd/cycle/main.go`

- Add an explicit `League` field to `Dataset` (not inferred from
  `Conference != ""` as today) with values `"nfl"`, `"cfb"`, `"cbb"`
  (decision 5 makes `"cbb"` the natural value to reuse here too).
- `discoverDatasets` adds a third glob (`dir/cbb/*.json`) and a second
  conference-filter loop using `cbbconf`.
- `cmd/cycle/main.go` adds `-sport` (decision 6), defaulting to `cfb`
  when `-conference` is set, switching between `cfbconf`/data dir and
  `cbbconf`/data dir.

## 5. Frontend: `internal/webgraph/web/app.js`, `nav.test.js`

- `league(d)` (`app.js:33`) becomes a real three-way switch reading
  `Dataset.League` instead of `d.conference ? "cfb" : "nfl"`.
- Nav dropdown logic (`populateConference`, `populateYear`) should
  generalize once league is a real field; `nav.test.js` needs new CBB
  fixtures and test cases added alongside the existing NFL/CFB ones, not
  just patched.

## 6. Data backfill

Scrape 2024, 2025, and 2026 (decision 8) once the scraper and conference
data are in place.

## Sequencing

1. Scraper (`scraper/cbb_scraper/`), verified against the live ESPN API
   shape already sampled above.
2. `cbbconf` membership data for 2024-2026.
3. Backend wiring (`Dataset.League`, `cache.go`, `cmd/cycle -sport`).
4. Frontend league field + `nav.test.js` updates.
5. Backfill 2024/2025/2026 data, smoke-test the viewer end to end.

Land as separate small commits per step above, on one feature branch,
merged to `main` once the viewer shows working CBB data end to end.
