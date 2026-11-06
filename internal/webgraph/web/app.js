const NEUTRAL_COLOR = "#999999";
const MAX_LEGEND_TEAMS = 6;

const statusEl = document.getElementById("status");
const leagueSelect = document.getElementById("league");
const conferenceSelect = document.getElementById("conference");
const yearSelect = document.getElementById("year");
const trimCheckbox = document.getElementById("trim");
const legendEl = document.getElementById("legend");
const weekControl = document.getElementById("week-control");
const weekSlider = document.getElementById("week-slider");
const weekLabel = document.getElementById("week-label");

let cy = null;
// Current highlight, or null: either a cycle (legend click) or a node
// (click on the graph) is isolated, never both at once.
let selection = null; // { type: "cycle", id } | { type: "node", team }
// Whether the week slider has been moved yet. Before the first move, the
// page shows the full, unfiltered season (today's default view, which for
// a completed season includes playoff weeks the slider itself can't
// reach) even though the slider control visually sits at its rightmost
// position.
let weekTouched = false;

// cyclesWithEdge returns the ids of every cycle in which winner beats loser
// consecutively - mirrors cycle.CyclesWithEdge on the Go side.
function cyclesWithEdge(cycles, winner, loser) {
  const ids = [];
  for (const c of cycles) {
    for (let i = 0; i < c.teams.length; i++) {
      if (c.teams[i] === winner && c.teams[(i + 1) % c.teams.length] === loser) {
        ids.push(c.id);
        break;
      }
    }
  }
  return ids;
}

// league returns which of the top-level dropdown groups d belongs to:
// "nfl", "cfb", or "cbb".
function league(d) {
  return d.league;
}

// hasConference reports whether leagueName is a conference-scoped league
// (cfb, cbb) as opposed to a single full-league dataset (nfl).
function hasConference(leagueName) {
  return leagueName === "cfb" || leagueName === "cbb";
}

// populateNav wires up the three dropdowns (league, conference, year) that
// together identify one dataset. Conference only applies to CFB/CBB, so
// it's hidden whenever the NFL league is selected.
function populateNav() {
  const current = window.DATASETS.find((d) => d.key === window.CURRENT_KEY);

  const leagues = [...new Set(window.DATASETS.map(league))].sort();
  for (const l of leagues) {
    const opt = document.createElement("option");
    opt.value = l;
    opt.textContent = l.toUpperCase();
    if (l === league(current)) opt.selected = true;
    leagueSelect.appendChild(opt);
  }

  populateConference(league(current), current.conference);
  populateYear(league(current), current.conference, current.season);

  leagueSelect.addEventListener("change", () => {
    populateConference(leagueSelect.value, undefined);
    populateYear(leagueSelect.value, conferenceSelect.value, undefined);
    navigate();
  });
  conferenceSelect.addEventListener("change", () => {
    populateYear(leagueSelect.value, conferenceSelect.value, undefined);
    navigate();
  });
  yearSelect.addEventListener("change", navigate);
}

function populateConference(selectedLeague, selectedConference) {
  conferenceSelect.innerHTML = "";
  conferenceSelect.classList.toggle("hidden", !hasConference(selectedLeague));
  if (!hasConference(selectedLeague)) return;

  const conferences = [...new Set(
    window.DATASETS.filter((d) => league(d) === selectedLeague).map((d) => d.conference)
  )].sort();
  for (const c of conferences) {
    const opt = document.createElement("option");
    opt.value = c;
    opt.textContent = c;
    if (c === selectedConference) opt.selected = true;
    conferenceSelect.appendChild(opt);
  }
}

function populateYear(selectedLeague, selectedConference, selectedYear) {
  yearSelect.innerHTML = "";
  const matches = window.DATASETS.filter(
    (d) => league(d) === selectedLeague && (!hasConference(selectedLeague) || d.conference === selectedConference)
  );
  const years = [...new Set(matches.map((d) => d.season))].sort((a, b) => b - a);
  for (const y of years) {
    const opt = document.createElement("option");
    opt.value = y;
    opt.textContent = y;
    if (y === selectedYear) opt.selected = true;
    yearSelect.appendChild(opt);
  }
}

// navigate is a real page navigation, not a fetch, so the browser's
// back/forward buttons and URL bar work normally when switching datasets.
function navigate() {
  const selectedLeague = leagueSelect.value;
  const selectedConference = hasConference(selectedLeague) ? conferenceSelect.value : undefined;
  const selectedYear = Number(yearSelect.value);
  const match = window.DATASETS.find(
    (d) =>
      league(d) === selectedLeague &&
      d.season === selectedYear &&
      (!hasConference(selectedLeague) || d.conference === selectedConference)
  );
  if (match) window.location.href = `${match.key}.html`;
}

function main() {
  populateNav();
  trimCheckbox.addEventListener("change", applyTrim);
  populateWeekSlider();
  if (!weekTouched) {
    render(window.GRAPH);
    updateStatus(window.GRAPH);
  }
}

function updateStatus(graph) {
  statusEl.textContent = `${graph.teams.length} teams, ${graph.cycles.length} distinct cycle${graph.cycles.length === 1 ? "" : "s"}`;
}

// populateWeekSlider wires up the NFL week slider: hidden entirely for
// non-NFL datasets or NFL datasets with no played weeks yet. A ?week=N
// query param on load is treated as an explicit, shareable selection (so
// it's applied immediately); absent that, the slider starts at its
// rightmost position without actually filtering anything, so the default
// view stays the full, unfiltered season.
function populateWeekSlider() {
  const current = window.DATASETS.find((d) => d.key === window.CURRENT_KEY);
  const weeks = window.GRAPH.weeks || [];
  if (league(current) !== "nfl" || weeks.length === 0) {
    weekControl.classList.toggle("hidden", true);
    return;
  }
  weekControl.classList.toggle("hidden", false);

  const maxWeek = weeks[weeks.length - 1].week;
  weekSlider.min = "1";
  weekSlider.max = String(maxWeek);

  const paramWeek = Number(new URLSearchParams(window.location.search).get("week"));
  const validParam = Number.isInteger(paramWeek) && paramWeek >= 1 && paramWeek <= maxWeek;

  weekTouched = validParam;
  const initialWeek = validParam ? paramWeek : maxWeek;
  weekSlider.value = String(initialWeek);
  updateWeekLabel(initialWeek);
  if (validParam) renderWeek(initialWeek);

  weekSlider.addEventListener("input", () => {
    weekTouched = true;
    const wk = Number(weekSlider.value);
    renderWeek(wk);
    const url = new URL(window.location.href);
    url.searchParams.set("week", String(wk));
    window.history.replaceState(null, "", url);
  });
}

function updateWeekLabel(wk) {
  weekLabel.textContent = weekTouched ? `through week ${wk}` : `week ${wk} (full season shown)`;
}

// renderWeek swaps in week wk's precomputed snapshot, sharing the full
// season's teams/teamColors (those don't vary by week).
function renderWeek(wk) {
  const entry = (window.GRAPH.weeks || []).find((w) => w.week === wk);
  if (!entry) return;
  updateWeekLabel(wk);
  const graph = { teams: window.GRAPH.teams, teamColors: window.GRAPH.teamColors, edges: entry.edges, cycles: entry.cycles };
  render(graph);
  updateStatus(graph);
}

function render(graph) {
  const nodeCycles = new Map(); // team -> [cycleId, ...]
  for (const team of graph.teams) nodeCycles.set(team, []);
  for (const c of graph.cycles) {
    for (const team of c.teams) nodeCycles.get(team).push(c.id);
  }

  // Order teams around the ring by the largest cycle's own sequence (it's
  // cycles[0] - DedupCycles sorts longest first), so that cycle draws as
  // a single clean loop instead of a tangled ball. Any team not on it
  // (no cycle at all, or only a smaller distinct one) is appended after,
  // still on the same ring.
  const order = new Map();
  if (graph.cycles.length > 0) {
    graph.cycles[0].teams.forEach((team, i) => order.set(team, i));
  }
  for (const team of graph.teams) {
    if (!order.has(team)) order.set(team, order.size);
  }

  // Cytoscape's built-in "circle" layout takes a `sort` option, but it's
  // broken as of 3.28 (it sorts a variable the final placement pass never
  // reads, so nodes land in their default order regardless) - so the ring
  // positions are computed by hand here and fed in via the "preset"
  // layout instead.
  const positions = {};
  const ringSize = order.size;
  for (const [team, idx] of order) {
    const theta = -Math.PI / 2 + (idx * 2 * Math.PI) / ringSize; // start at top, clockwise
    positions[team] = { x: 300 * Math.cos(theta), y: 300 * Math.sin(theta) };
  }

  const elements = [];
  for (const team of graph.teams) {
    const ids = nodeCycles.get(team);
    const colors = graph.teamColors[team] || { primary: NEUTRAL_COLOR, secondary: NEUTRAL_COLOR };
    elements.push({
      data: {
        id: team,
        cycleIds: ids,
        primary: colors.primary,
        secondary: colors.secondary,
        order: order.get(team),
        inCycle: ids.length > 0,
      },
    });
  }
  for (const e of graph.edges) {
    const ids = cyclesWithEdge(graph.cycles, e.winner, e.loser);
    if (ids.length === 0) {
      // Trimmable if both endpoints are already on some cycle: that
      // makes this edge visual noise once the cycle itself is drawn.
      const trimmable = nodeCycles.get(e.winner).length > 0 && nodeCycles.get(e.loser).length > 0;
      elements.push({
        data: { id: `${e.winner}->${e.loser}`, source: e.winner, target: e.loser, color: NEUTRAL_COLOR, cycleId: null, trimmable, neutral: true },
      });
      continue;
    }
    for (const id of ids) {
      elements.push({
        data: {
          id: `${e.winner}->${e.loser}#${id}`,
          source: e.winner,
          target: e.loser,
          color: graph.cycles[id].color,
          cycleId: id,
          neutral: false,
        },
      });
    }
  }

  if (cy) cy.destroy();
  cy = cytoscape({
    container: document.getElementById("cy"),
    elements,
    layout: {
      name: "preset",
      positions: (node) => positions[node.id()],
      fit: true,
      padding: 30,
    },
    style: [
      {
        selector: "node",
        style: {
          label: "data(id)",
          "background-color": "data(primary)",
          "border-width": 2,
          "border-color": "data(secondary)",
          color: "#fff",
          "text-halign": "center",
          "text-valign": "center",
          "font-size": 9,
          width: 28,
          height: 28,
          "z-index": 10,
        },
      },
      {
        selector: "edge",
        style: {
          width: 1.5,
          "line-color": "data(color)",
          "target-arrow-color": "data(color)",
          "target-arrow-shape": "triangle",
          "curve-style": "bezier",
          "control-point-step-size": 4,
        },
      },
      { selector: "node[!inCycle]", style: { opacity: 0.4 } },
      { selector: "edge[?neutral]", style: { opacity: 0.35, "z-index": 1 } },
      { selector: "edge[!neutral]", style: { opacity: 0.8, "z-index": 2 } },
      { selector: "edge.emphasized", style: { opacity: 0.9, "z-index": 5 } },
      { selector: ".dimmed", style: { opacity: 0.08 } },
      { selector: ".trimmed", style: { display: "none" } },
    ],
  });

  cy.on("tap", "node", (evt) => selectNode(evt.target.id()));
  cy.on("tap", (evt) => {
    if (evt.target === cy) clearSelection();
  });

  renderLegend(graph.cycles);
  applyTrim();
  applyHighlight();
}

function applyTrim() {
  if (!cy) return;
  cy.edges().forEach((e) => e.toggleClass("trimmed", trimCheckbox.checked && !!e.data("trimmable")));
}

function selectNode(team) {
  if (selection && selection.type === "node" && selection.team === team) {
    clearSelection();
    return;
  }
  selection = { type: "node", team };
  applyHighlight();
}

function isolateCycle(cycleId) {
  if (selection && selection.type === "cycle" && selection.id === cycleId) {
    clearSelection();
    return;
  }
  selection = { type: "cycle", id: cycleId };
  applyHighlight();
}

function clearSelection() {
  selection = null;
  applyHighlight();
}

// applyHighlight dims every element except those relevant to the current
// selection: a cycle's own members and edges, or - for a clicked node -
// every edge touching it (plus the node itself and whichever teams are on
// the other end of those edges), regardless of cycle membership.
function applyHighlight() {
  if (!cy) return;
  let keep = null;
  if (selection && selection.type === "cycle") {
    keep = cy.elements().filter((el) => {
      const ids = el.isNode() ? el.data("cycleIds") : [el.data("cycleId")];
      return ids.includes(selection.id);
    });
  } else if (selection && selection.type === "node") {
    keep = cy.getElementById(selection.team).closedNeighborhood();
  }
  cy.elements().forEach((el) => {
    const kept = !!keep && keep.contains(el);
    el.toggleClass("dimmed", !!keep && !kept);
    el.toggleClass("emphasized", kept && el.isEdge() && !!el.data("neutral"));
  });

  document.querySelectorAll(".legend-item").forEach((item) => {
    const isActiveCycle = selection && selection.type === "cycle" && Number(item.dataset.cycleId) === selection.id;
    item.classList.toggle("dimmed", !!selection && !isActiveCycle);
  });
}

function legendLabel(teams) {
  if (teams.length <= MAX_LEGEND_TEAMS) return teams.join(" -> ");
  const half = Math.floor(MAX_LEGEND_TEAMS / 2);
  const head = teams.slice(0, half).join(" -> ");
  const tail = teams.slice(teams.length - half).join(" -> ");
  return `${head} -> ... -> ${tail} (${teams.length} teams)`;
}

function renderLegend(cycles) {
  legendEl.innerHTML = "";
  for (const c of cycles) {
    const item = document.createElement("div");
    item.className = "legend-item";
    item.dataset.cycleId = c.id;
    item.innerHTML = `<span class="swatch" style="background:${c.color}"></span><span>${legendLabel(c.teams)}</span>`;
    item.addEventListener("click", () => isolateCycle(c.id));
    legendEl.appendChild(item);
  }
}

main();
