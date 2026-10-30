const NEUTRAL_COLOR = "#999999";
const MAX_LEGEND_TEAMS = 6;

const statusEl = document.getElementById("status");
const datasetSelect = document.getElementById("dataset");
const trimCheckbox = document.getElementById("trim");
const legendEl = document.getElementById("legend");

let cy = null;
let highlightedCycle = null; // cycle id currently isolated, or null

async function main() {
  const datasets = await fetchJSON("data/datasets.json");
  for (const d of datasets) {
    const opt = document.createElement("option");
    opt.value = d.key;
    opt.textContent = d.label;
    datasetSelect.appendChild(opt);
  }
  datasetSelect.addEventListener("change", () => loadGraph(datasetSelect.value));
  trimCheckbox.addEventListener("change", applyTrim);
  if (datasets.length > 0) {
    datasetSelect.value = datasets[0].key;
    loadGraph(datasets[0].key);
  }
}

async function fetchJSON(url) {
  const resp = await fetch(url);
  if (!resp.ok) {
    throw new Error(`${url}: ${resp.status}`);
  }
  return resp.json();
}

async function loadGraph(key) {
  statusEl.textContent = "loading...";
  highlightedCycle = null;
  try {
    const graph = await fetchJSON(`data/${encodeURIComponent(key)}.json`);
    render(graph);
    statusEl.textContent = `${graph.teams.length} teams, ${graph.cycles.length} distinct cycle${graph.cycles.length === 1 ? "" : "s"}`;
  } catch (err) {
    statusEl.textContent = `error: ${err}`;
  }
}

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
    const color = ids.length > 0 ? graph.cycles[ids[0]].color : NEUTRAL_COLOR;
    elements.push({
      data: { id: team, cycleIds: ids, color, order: order.get(team) },
      classes: ids.length > 0 ? "in-cycle" : "",
    });
  }
  for (const e of graph.edges) {
    const ids = cyclesWithEdge(graph.cycles, e.winner, e.loser);
    if (ids.length === 0) {
      // Trimmable if both endpoints are already on some cycle: that
      // makes this edge visual noise once the cycle itself is drawn.
      const trimmable = nodeCycles.get(e.winner).length > 0 && nodeCycles.get(e.loser).length > 0;
      elements.push({
        data: { id: `${e.winner}->${e.loser}`, source: e.winner, target: e.loser, color: NEUTRAL_COLOR, cycleId: null, trimmable },
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
          "background-color": "data(color)",
          color: "#fff",
          "text-outline-width": 1,
          "text-outline-color": "data(color)",
          "font-size": 10,
          width: 24,
          height: 24,
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
          opacity: 0.8,
        },
      },
      { selector: ".dimmed", style: { opacity: 0.12 } },
      { selector: ".trimmed", style: { display: "none" } },
    ],
  });

  cy.on("tap", "node", (evt) => {
    const ids = evt.target.data("cycleIds");
    if (ids.length === 0) return;
    isolateCycle(ids[0]);
  });
  cy.on("tap", (evt) => {
    if (evt.target === cy) clearHighlight();
  });

  renderLegend(graph.cycles);
  applyTrim();
}

function applyTrim() {
  if (!cy) return;
  cy.edges().forEach((e) => e.toggleClass("trimmed", trimCheckbox.checked && !!e.data("trimmable")));
}

function isolateCycle(cycleId) {
  if (highlightedCycle === cycleId) {
    clearHighlight();
    return;
  }
  highlightedCycle = cycleId;
  cy.elements().forEach((el) => {
    const ids = el.isNode() ? el.data("cycleIds") : [el.data("cycleId")];
    const on = ids.includes(cycleId);
    el.toggleClass("dimmed", !on);
  });
  document.querySelectorAll(".legend-item").forEach((item) => {
    item.classList.toggle("dimmed", Number(item.dataset.cycleId) !== cycleId);
  });
}

function clearHighlight() {
  highlightedCycle = null;
  cy.elements().removeClass("dimmed");
  document.querySelectorAll(".legend-item").forEach((item) => item.classList.remove("dimmed"));
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
