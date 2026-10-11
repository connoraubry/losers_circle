// Package cfbconf tracks which teams belong to each Power 4 college
// football conference (SEC, Big Ten, ACC, Big 12) in a given season.
//
// Conference realignment is infrequent and well-documented, so membership
// is hand-maintained here (keyed by ESPN team abbreviation, matching the
// data scraped by scraper/cfb_scraper) rather than sourced from an API.
package cfbconf

import (
	"fmt"
	"sort"
	"strings"

	"losers_circle/internal/nfldata"
)

// Canonical conference names.
const (
	SEC    = "SEC"
	BigTen = "Big Ten"
	ACC    = "ACC"
	Big12  = "Big 12"
)

// aliases maps normalized user-facing spellings to a canonical conference
// name.
var aliases = map[string]string{
	"sec":     SEC,
	"big ten": BigTen,
	"b1g":     BigTen,
	"big10":   BigTen,
	"acc":     ACC,
	"big 12":  Big12,
	"big12":   Big12,
}

// Canonicalize resolves a user-supplied conference name (case/spacing
// insensitive) to its canonical form, or "" if unrecognized.
func Canonicalize(name string) string {
	key := strings.ToLower(strings.Join(strings.Fields(name), " "))
	return aliases[key]
}

// membership[season][conference] holds ESPN team abbreviations.
var membership = map[int]map[string][]string{
	2023: {
		SEC:    {"ALA", "ARK", "AUB", "FLA", "UGA", "UK", "LSU", "MSST", "MIZ", "MISS", "SC", "TENN", "TA&M", "VAN"},
		BigTen: {"ILL", "IU", "IOWA", "MD", "MICH", "MSU", "MINN", "NEB", "NU", "OSU", "PSU", "PUR", "RUTG", "WIS"},
		ACC:    {"BC", "CLEM", "DUKE", "FSU", "GT", "LOU", "MIA", "NCSU", "UNC", "PITT", "SYR", "UVA", "VT", "WAKE"},
		Big12:  {"BAY", "BYU", "CIN", "HOU", "ISU", "KU", "KSU", "OU", "OKST", "TCU", "TEX", "TTU", "UCF", "WVU"},
	},
	2024: {
		SEC:    {"ALA", "ARK", "AUB", "FLA", "UGA", "UK", "LSU", "MSST", "MIZ", "MISS", "SC", "TENN", "TA&M", "VAN", "OU", "TEX"},
		BigTen: {"ILL", "IU", "IOWA", "MD", "MICH", "MSU", "MINN", "NEB", "NU", "OSU", "PSU", "PUR", "RUTG", "WIS", "UCLA", "USC", "ORE", "WASH"},
		ACC:    {"BC", "CLEM", "DUKE", "FSU", "GT", "LOU", "MIA", "NCSU", "UNC", "PITT", "SYR", "UVA", "VT", "WAKE", "CAL", "SMU", "STAN"},
		Big12:  {"BAY", "BYU", "CIN", "HOU", "ISU", "KU", "KSU", "OKST", "TCU", "TTU", "UCF", "WVU", "ARIZ", "ASU", "COLO", "UTAH"},
	},
}

func init() {
	// 2025 and 2026 carry forward the 2024 realignment; no further Power 4
	// membership changes are known at the time this table was written.
	membership[2025] = membership[2024]
	membership[2026] = membership[2024]
}

// Teams returns the sorted list of team abbreviations in conference for
// season, or nil if the season/conference combination isn't tracked.
func Teams(season int, conference string) []string {
	byConf, ok := membership[season]
	if !ok {
		return nil
	}
	teams, ok := byConf[conference]
	if !ok {
		return nil
	}
	out := make([]string, len(teams))
	copy(out, teams)
	sort.Strings(out)
	return out
}

// FilterSeason resolves conference (via Canonicalize) and returns a copy of
// s containing only that conference's teams and the games played between
// two of them.
func FilterSeason(s nfldata.Season, conference string) (nfldata.Season, error) {
	canon := Canonicalize(conference)
	if canon == "" {
		return nfldata.Season{}, fmt.Errorf("unknown conference %q (want SEC, Big Ten, ACC, or Big 12)", conference)
	}
	teams := Teams(s.SeasonYear, canon)
	if teams == nil {
		return nfldata.Season{}, fmt.Errorf("no %s membership data for season %d", canon, s.SeasonYear)
	}

	inConf := make(map[string]bool, len(teams))
	for _, t := range teams {
		inConf[t] = true
	}

	filtered := s
	filtered.Teams = teams
	filtered.Games = nil
	for _, g := range s.Games {
		if inConf[g.HomeTeam] && inConf[g.AwayTeam] {
			filtered.Games = append(filtered.Games, g)
		}
	}
	return filtered, nil
}
