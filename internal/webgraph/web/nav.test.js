// Tests for the league/conference/year dropdown logic in app.js. Runs
// under Node's built-in test runner (no npm dependencies): `node --test`.
//
// app.js is written for a real DOM, so each test builds a minimal select
// stub and runs app.js's source in a vm context against it, rather than
// a browser - this covers the pure dropdown wiring (what's selected,
// what's hidden, where a change navigates to), not rendering.
const { test } = require("node:test");
const assert = require("node:assert/strict");
const vm = require("node:vm");
const fs = require("node:fs");
const path = require("node:path");

const APP_JS = fs.readFileSync(path.join(__dirname, "app.js"), "utf8");

// A representative slice of real Dataset JSON (see webgraph.graphJSON):
// two NFL seasons, plus a couple of seasons each for two conferences.
const DATASETS = [
  { key: "nfl-2026", label: "2026 NFL", season: 2026 },
  { key: "nfl-2025", label: "2025 NFL", season: 2025 },
  { key: "acc-2026", label: "2026 ACC", season: 2026, conference: "ACC" },
  { key: "acc-2025", label: "2025 ACC", season: 2025, conference: "ACC" },
  { key: "sec-2026", label: "2026 SEC", season: 2026, conference: "SEC" },
  { key: "sec-2025", label: "2025 SEC", season: 2025, conference: "SEC" },
  { key: "sec-2024", label: "2024 SEC", season: 2024, conference: "SEC" },
];

function makeSelect() {
  const sel = {
    _options: [],
    _selectedIndex: -1,
    _listeners: [],
    classList: {
      _set: new Set(),
      toggle(cls, on) {
        if (on) this._set.add(cls);
        else this._set.delete(cls);
      },
      contains(cls) {
        return this._set.has(cls);
      },
    },
    appendChild(opt) {
      sel._options.push(opt);
      if (opt.selected || sel._selectedIndex === -1) sel._selectedIndex = sel._options.length - 1;
    },
    set innerHTML(_v) {
      sel._options = [];
      sel._selectedIndex = -1;
    },
    get options() {
      return sel._options;
    },
    get value() {
      return sel._selectedIndex >= 0 ? String(sel._options[sel._selectedIndex].value) : "";
    },
    set value(v) {
      const i = sel._options.findIndex((o) => String(o.value) === String(v));
      if (i >= 0) sel._selectedIndex = i;
    },
    addEventListener(evt, fn) {
      sel._listeners.push([evt, fn]);
    },
    fireChange() {
      for (const [evt, fn] of sel._listeners) if (evt === "change") fn();
    },
  };
  return sel;
}

// optionValues returns the select's option values in order, with "*"
// marking whichever one is currently selected.
function optionValues(sel) {
  return sel.options.map((o, i) => `${o.value}${i === sel._selectedIndex ? "*" : ""}`);
}

function stubElement() {
  return {
    set textContent(_v) {},
    set innerHTML(_v) {},
    appendChild() {},
    addEventListener() {},
    classList: { toggle() {}, contains() { return false; } },
  };
}

// loadApp evaluates app.js against a fresh fake window for the dataset
// identified by currentKey, returning the window (so tests can read
// window.location.href after a navigation) and the three dropdowns.
function loadApp(currentKey) {
  const elements = {
    league: makeSelect(),
    conference: makeSelect(),
    year: makeSelect(),
    trim: { addEventListener() {}, checked: false },
    status: stubElement(),
    legend: { innerHTML: "", appendChild() {} },
  };

  const emptyCollection = {
    forEach() {},
    filter() {
      return emptyCollection;
    },
    contains() {
      return false;
    },
  };

  const window = {
    CURRENT_KEY: currentKey,
    DATASETS,
    GRAPH: { teams: [], edges: [], cycles: [] },
    location: { href: `http://localhost/${currentKey}.html` },
    document: {
      getElementById(id) {
        return elements[id] ?? stubElement();
      },
      createElement(tag) {
        return tag === "option" ? { value: "", textContent: "", selected: false } : makeSelect();
      },
      querySelectorAll() {
        return [];
      },
    },
    cytoscape() {
      return {
        destroy() {},
        on() {},
        elements() {
          return emptyCollection;
        },
        edges() {
          return emptyCollection;
        },
        getElementById() {
          return { closedNeighborhood() { return emptyCollection; } };
        },
      };
    },
  };
  window.window = window;

  vm.runInContext(APP_JS, vm.createContext(window), { filename: "app.js" });
  return { window, elements };
}

test("NFL dataset hides the conference dropdown", () => {
  const { elements } = loadApp("nfl-2025");
  assert.equal(elements.conference.classList.contains("hidden"), true);
  assert.deepEqual(optionValues(elements.league), ["cfb", "nfl*"]);
  assert.deepEqual(optionValues(elements.year), ["2026", "2025*"]);
});

test("CFB dataset shows the conference dropdown, scoped to that league", () => {
  const { elements } = loadApp("sec-2024");
  assert.equal(elements.conference.classList.contains("hidden"), false);
  assert.deepEqual(optionValues(elements.conference), ["ACC", "SEC*"]);
  assert.deepEqual(optionValues(elements.year), ["2026", "2025", "2024*"]);
});

test("switching league to cfb reveals conference and defaults to a valid page", () => {
  const { window, elements } = loadApp("nfl-2025");
  elements.league.value = "cfb";
  elements.league.fireChange();

  assert.equal(elements.conference.classList.contains("hidden"), false);
  assert.deepEqual(optionValues(elements.conference), ["ACC*", "SEC"]);
  assert.equal(window.location.href, "acc-2026.html");
});

test("switching league back to nfl hides conference and navigates to an NFL page", () => {
  const { window, elements } = loadApp("sec-2024");
  elements.league.value = "nfl";
  elements.league.fireChange();

  assert.equal(elements.conference.classList.contains("hidden"), true);
  assert.equal(window.location.href, "nfl-2026.html");
});

test("switching conference re-scopes the year list and navigates to that conference's latest year", () => {
  const { window, elements } = loadApp("sec-2024");
  elements.conference.value = "ACC";
  elements.conference.fireChange();

  assert.deepEqual(optionValues(elements.year), ["2026*", "2025"]);
  assert.equal(window.location.href, "acc-2026.html");
});

test("switching year navigates within the same league/conference", () => {
  const { window, elements } = loadApp("sec-2024");
  elements.year.value = "2025";
  elements.year.fireChange();

  assert.equal(window.location.href, "sec-2025.html");
});
