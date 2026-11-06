// Package webgraph serves an interactive, browser-based view of a
// cycle.Graph. Cycle detection (Graph.LongestByTeam) is the expensive part
// of building that view, so Build precomputes it once per dataset and
// persists the result to disk, so a server restart doesn't redo the work
// unless the underlying season data actually changed.
package webgraph

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"losers_circle/internal/cbbconf"
	"losers_circle/internal/cfbconf"
	"losers_circle/internal/cycle"
	"losers_circle/internal/nfldata"
)

// conferences is the fixed set of Power 4 conferences tracked by both
// cfbconf and cbbconf (the two packages hand-maintain membership
// independently, but track the same four conference names).
var conferences = []string{cfbconf.SEC, cfbconf.BigTen, cfbconf.ACC, cfbconf.Big12}

// leagueOrder ranks each league for sorting the dataset list: NFL first,
// then CFB, then CBB.
var leagueOrder = map[string]int{"nfl": 0, "cfb": 1, "cbb": 2}

// Dataset identifies one graph the viewer can show: a full NFL season, or
// a single CFB or CBB conference's season.
type Dataset struct {
	Key        string `json:"key"`
	Label      string `json:"label"`
	Season     int    `json:"season"`
	League     string `json:"league"`
	Conference string `json:"conference,omitempty"`
}

// Entry is one dataset's precomputed graph and distinct cycles, ready to
// serve.
type Entry struct {
	Dataset Dataset
	Graph   *cycle.Graph
	Cycles  []cycle.DistinctCycle
	// Weeks holds one cumulative-through-week snapshot per regular-season
	// week that's been played so far, for the NFL week slider. It's nil
	// for non-NFL datasets and for NFL datasets with no played games yet.
	Weeks []WeekSnapshot
}

// WeekSnapshot is one week's cumulative (through that week) graph and
// distinct cycles, computed the same way as Entry's full-season Graph/
// Cycles but filtered to games through Week.
type WeekSnapshot struct {
	Week   int
	Graph  *cycle.Graph
	Cycles []cycle.DistinctCycle
}

// maxSliderWeek is the latest NFL regular-season week the week slider
// covers. Postseason weeks (19+) aren't included in this first version.
const maxSliderWeek = 18

// Cache holds every dataset's precomputed Entry, built once at startup.
type Cache struct {
	datasets []Dataset
	entries  map[string]Entry
}

// List returns every dataset in the cache, NFL seasons first then CFB
// conferences, newest season first within each.
func (c *Cache) List() []Dataset {
	return append([]Dataset(nil), c.datasets...)
}

// Get returns the dataset keyed by key, or false if there's no such
// dataset.
func (c *Cache) Get(key string) (Entry, bool) {
	e, ok := c.entries[key]
	return e, ok
}

// cacheFile is the on-disk shape of one dataset's persisted cycle
// computation: the expensive byTeam result, guarded by a hash of the
// source season JSON so a changed season (new games played) invalidates
// it automatically.
type cacheFile struct {
	SourceHash string              `json:"sourceHash"`
	ByTeam     map[string][]string `json:"byTeam"`
}

// Build loads every available dataset under dir (NFL seasons from
// dir/*.json, CFB conference-seasons from dir/cfb/*.json), computes or
// loads from cacheDir each one's longest-cycle-by-team result, and returns
// a ready-to-serve Cache. Datasets are built concurrently since
// LongestByTeam is CPU-bound and independent per dataset.
func Build(dir, cacheDir string) (*Cache, error) {
	jobs, err := discoverDatasets(dir)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	entries := make([]Entry, len(jobs))
	errs := make([]error, len(jobs))
	for i, job := range jobs {
		wg.Add(1)
		go func(i int, job datasetJob) {
			defer wg.Done()
			entries[i], errs[i] = job.build(cacheDir)
		}(i, job)
	}
	wg.Wait()

	c := &Cache{entries: map[string]Entry{}}
	for i, err := range errs {
		if err != nil {
			return nil, fmt.Errorf("dataset %s: %w", jobs[i].dataset.Key, err)
		}
		c.datasets = append(c.datasets, entries[i].Dataset)
		c.entries[entries[i].Dataset.Key] = entries[i]
	}
	sort.Slice(c.datasets, func(i, j int) bool {
		di, dj := c.datasets[i], c.datasets[j]
		if di.League != dj.League {
			return leagueOrder[di.League] < leagueOrder[dj.League]
		}
		if di.Season != dj.Season {
			return di.Season > dj.Season // newest season first
		}
		return di.Conference < dj.Conference
	})
	return c, nil
}

// datasetJob is one dataset's load+build work, resolved up front so Build
// can fan it out across goroutines.
type datasetJob struct {
	dataset Dataset
	season  nfldata.Season
}

func (job datasetJob) build(cacheDir string) (Entry, error) {
	raw, err := json.Marshal(job.season)
	if err != nil {
		return Entry{}, fmt.Errorf("hash source: %w", err)
	}
	sum := sha256.Sum256(raw)
	sourceHash := hex.EncodeToString(sum[:])

	cachePath := filepath.Join(cacheDir, job.dataset.Key+".json")
	byTeam, err := loadCachedByTeam(cachePath, sourceHash)
	if err != nil {
		return Entry{}, err
	}

	g := cycle.Build(job.season)
	if byTeam == nil {
		byTeam = g.LongestByTeam()
		if err := writeCachedByTeam(cachePath, sourceHash, byTeam); err != nil {
			return Entry{}, err
		}
	}

	var weeks []WeekSnapshot
	if job.dataset.League == "nfl" {
		weeks = buildWeekSnapshots(job.season)
	}

	return Entry{
		Dataset: job.dataset,
		Graph:   g,
		Cycles:  cycle.DedupCycles(byTeam),
		Weeks:   weeks,
	}, nil
}

// buildWeekSnapshots computes one cumulative-through-week snapshot per
// played regular-season week (capped at maxSliderWeek), for the NFL week
// slider. Each week's Graph (and so its edge list) is always rebuilt
// fresh, since games keep accumulating every week - but once some week's
// longest cycle already covers every team, that's provably the largest a
// cycle can ever get (it can't exceed the team count), so every later
// week reuses that same Cycles result instead of redoing the actual
// search, which is the expensive part (see package doc comment).
func buildWeekSnapshots(s nfldata.Season) []WeekSnapshot {
	latestPlayed := 0
	for _, gm := range s.Games {
		if gm.Played() && gm.Week <= maxSliderWeek && gm.Week > latestPlayed {
			latestPlayed = gm.Week
		}
	}
	if latestPlayed == 0 {
		return nil
	}

	weeks := make([]WeekSnapshot, 0, latestPlayed)
	var fullCycles []cycle.DistinctCycle
	for wk := 1; wk <= latestPlayed; wk++ {
		wg := cycle.Build(nfldata.FilterMaxWeek(s, wk))

		wc := fullCycles
		if wc == nil {
			wc = cycle.DedupCycles(wg.LongestByTeam())
			if len(wc) > 0 && len(wc[0].Teams) == len(s.Teams) {
				fullCycles = wc
			}
		}
		weeks = append(weeks, WeekSnapshot{Week: wk, Graph: wg, Cycles: wc})
	}
	return weeks
}

// loadCachedByTeam returns the cached byTeam result at path if it exists
// and matches sourceHash, or nil (with no error) if there's no usable
// cache yet.
func loadCachedByTeam(path, sourceHash string) (map[string][]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read cache %s: %w", path, err)
	}
	var cf cacheFile
	if err := json.Unmarshal(b, &cf); err != nil {
		return nil, fmt.Errorf("parse cache %s: %w", path, err)
	}
	if cf.SourceHash != sourceHash {
		return nil, nil
	}
	return cf.ByTeam, nil
}

func writeCachedByTeam(path, sourceHash string, byTeam map[string][]string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}
	b, err := json.Marshal(cacheFile{SourceHash: sourceHash, ByTeam: byTeam})
	if err != nil {
		return fmt.Errorf("marshal cache %s: %w", path, err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return fmt.Errorf("write cache %s: %w", path, err)
	}
	return nil
}

// discoverDatasets finds every NFL season under dir/*.json and every
// tracked CFB conference-season under dir/cfb/*.json, loading each
// season's data up front so Build's goroutines only need to compute (or
// load) cycles.
func discoverDatasets(dir string) ([]datasetJob, error) {
	var jobs []datasetJob

	nflFiles, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}
	for _, path := range nflFiles {
		season, err := nfldata.LoadFile(path)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, datasetJob{
			dataset: Dataset{
				Key:    fmt.Sprintf("nfl-%d", season.SeasonYear),
				Label:  fmt.Sprintf("%d NFL", season.SeasonYear),
				Season: season.SeasonYear,
				League: "nfl",
			},
			season: season,
		})
	}

	cfbFiles, err := filepath.Glob(filepath.Join(dir, "cfb", "*.json"))
	if err != nil {
		return nil, err
	}
	for _, path := range cfbFiles {
		full, err := nfldata.LoadFile(path)
		if err != nil {
			return nil, err
		}
		for _, conf := range conferences {
			if cfbconf.Teams(full.SeasonYear, conf) == nil {
				continue
			}
			filtered, err := cfbconf.FilterSeason(full, conf)
			if err != nil {
				return nil, fmt.Errorf("filter %s %d: %w", conf, full.SeasonYear, err)
			}
			jobs = append(jobs, datasetJob{
				dataset: Dataset{
					Key:        fmt.Sprintf("%s-%d", confSlug(conf), full.SeasonYear),
					Label:      fmt.Sprintf("%d %s", full.SeasonYear, conf),
					Season:     full.SeasonYear,
					League:     "cfb",
					Conference: conf,
				},
				season: filtered,
			})
		}
	}

	cbbFiles, err := filepath.Glob(filepath.Join(dir, "cbb", "*.json"))
	if err != nil {
		return nil, err
	}
	for _, path := range cbbFiles {
		full, err := nfldata.LoadFile(path)
		if err != nil {
			return nil, err
		}
		for _, conf := range conferences {
			if cbbconf.Teams(full.SeasonYear, conf) == nil {
				continue
			}
			filtered, err := cbbconf.FilterSeason(full, conf)
			if err != nil {
				return nil, fmt.Errorf("filter %s %d: %w", conf, full.SeasonYear, err)
			}
			jobs = append(jobs, datasetJob{
				dataset: Dataset{
					Key:        fmt.Sprintf("cbb-%s-%d", confSlug(conf), full.SeasonYear),
					Label:      fmt.Sprintf("%s %s (MBB)", cbbSeasonLabel(full.SeasonYear), conf),
					Season:     full.SeasonYear,
					League:     "cbb",
					Conference: conf,
				},
				season: filtered,
			})
		}
	}

	return jobs, nil
}

// confSlug turns a conference's display name into a URL/filename-safe key
// fragment, e.g. "Big Ten" -> "big-ten".
func confSlug(conference string) string {
	return strings.ToLower(strings.ReplaceAll(conference, " ", "-"))
}

// cbbSeasonLabel formats a CBB season's end year as the straddling
// "YY-YY" range it actually covers, e.g. 2027 -> "26-27". A bare end
// year on its own (unlike NFL/CFB's single-calendar-year seasons) reads
// like a single season that hasn't happened yet, when the season named
// "2027" is in fact the one that started in 2026.
func cbbSeasonLabel(endYear int) string {
	return fmt.Sprintf("%02d-%02d", (endYear-1)%100, endYear%100)
}
