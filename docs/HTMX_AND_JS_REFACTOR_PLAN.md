# HTMX and JavaScript Refactor Plan

## Current state

### HTMX usage
- **Requests**: `/reroll`, `/begin`, `/play` are triggered by `hx-post` with `hx-target="#game"` and `hx-swap="innerHTML"`. Only the `#game` section is replaced; the layout (sidebars) is not re-rendered.
- **Response**: Handlers return a single HTML fragment (e.g. `game.html` or `start.html`) that is swapped into `#game`. No HTMX response headers (e.g. `HX-Push`) or out-of-band (OOB) swaps are used.
- **Sidebar sync**: Because the server only returns the main content, the left sidebar (stats, player dice) and right sidebar (enemies, enemy dice) are updated by **JavaScript** that runs after every swap: it reads hidden `data-*` elements inside `#game` (e.g. `.stats-update`, `.player-dice-update`, `.enemy-update`) and copies values into the sidebar DOM.

### JavaScript
- **Location**: ~170 lines of inline script in `templates/layout.html` (lines 10–179).
- **Responsibilities**:
  1. `updateSidebarStats()` – sync strength/luck/health from `#game .stats-update` to `#stat-*` in the left sidebar.
  2. `updateEnemySidebar()` – build list from `#game .enemy-update`, show/hide `.enemy-sidebar` and fill `.enemy-panel` name/strength/health.
  3. `setDieFace(el, face, animate)` – set a die’s `data-face` (optionally with a short “roll” animation).
  4. `updatePlayerDice()` – read `.player-dice-update` or `.stat-rolls` from `#game`, show last-roll or stat-roll dice and set faces.
  5. `updateEnemyDice()` – read `.enemy-dice-update`, show/hide enemy dice area and set faces.
  6. `initStatsUpdater()` – run all updaters on load and subscribe to `htmx:afterSwap` (with a 10 ms delay) to run them again after every swap; plus a 100 ms timeout “to catch late-loading content”.

### Layout
- **Single file**: `templates/layout.html` defines the full page: `define "layout.html"` with doctype, `<head>` (meta, HTMX script, CSS, inline JS), `<body>` (main, container, left aside, `#game`, right aside). No partials or shared fragments.

---

## Where we diverge from HTMX idioms

1. **Heavy JS for “sync”**  
   HTMX favours server-driven HTML. We use the server only for `#game` and drive sidebars entirely from JS by scraping the fragment. That’s a valid pattern but not the “return HTML, let HTMX swap it” style. A more idiomatic approach is **out-of-band (OOB) swaps**: the server returns the `#game` fragment **plus** extra elements with `hx-swap-oob="true"` (e.g. `id="sidebar-stats"`, `id="enemy-sidebar"`), and HTMX updates those regions from the same response. Then we need little or no JS for syncing.

2. **Global `htmx:afterSwap`**  
   We listen on `document.body` for **every** swap and then run all four updaters (with a timeout). We don’t scope to “only when `#game` was swapped,” so we do redundant work on any future OOB or other swaps. Prefer: listen for `htmx:afterSwap` and only run updaters when `evt.detail.target.id === 'game'` (or equivalent).

3. **Fragile “late content” timeout**  
   The 100 ms `setTimeout` is a code smell: it papers over timing/order issues. Once we have a single, well-defined “after game swap” path (and possibly OOB), we can remove it.

4. **Inline script**  
   Large inline script in the layout hurts maintainability, testability, and caching. HTMX docs and best practices recommend keeping JS in external files and using the event API.

5. **No tests**  
   The sidebar/dice logic is non-trivial and depends on DOM shape. It should be unit-tested with a fake DOM (e.g. jsdom) so refactors (and OOB migration) don’t regress.

---

## Goals

1. **Reassess HTMX usage**  
   Align with HTMX patterns: prefer server-returned HTML for all changing regions (e.g. OOB for sidebars), and minimal JS only where necessary.

2. **Move JavaScript out of the layout**  
   Put all current inline JS into one or more files under `static/` (e.g. `static/js/app.js`), load them from the layout, and ensure they run after HTMX is loaded and only hook once (e.g. on `DOMContentLoaded` + `htmx:afterSwap` scoped to `#game`).

3. **Add tests for the JavaScript**  
   Introduce a test runner and DOM environment (e.g. Node + Jest + jsdom, or Vitest + happy-dom), and add unit tests for:
   - `updateSidebarStats` (given a `#game` with `.stats-update`, sidebar elements get the right text).
   - `updateEnemySidebar` (given `.enemy-update` elements, right panels are shown/hidden and filled).
   - `setDieFace` (face 1–6, with and without animate).
   - `updatePlayerDice` (`.player-dice-update` vs `.stat-rolls`, correct sections and die faces).
   - `updateEnemyDice` (show/hide, die faces).
   - Init: after a synthetic `htmx:afterSwap` with target `#game`, updaters run (can assert on DOM state).

4. **Split the layout**  
   Break `layout.html` into smaller, named pieces so the main layout is structure + includes, and the big script block is gone (replaced by a link to `app.js`). Options:
   - **Option A (minimal)**: Extract only the script to `static/js/app.js`; keep one layout file but shrink it (no inline JS).
   - **Option B (partials)**: Add partials, e.g. `layout_head.html` (meta, link to CSS, script tags for HTMX and app.js), `sidebar_left.html`, `sidebar_right.html`, and a thin `layout.html` that defines the shell and includes these. Template syntax stays `{{template "layout_head.html" .}}` etc.; no new tech.

---

## Recommended implementation order

### Phase 1: Extract JS and scope HTMX listener (no OOB yet)
- Create `static/js/app.js` and move all current inline logic into it.
- Export or expose a single init function (e.g. `window.AdventureUI = { init: function() { ... } }`) that:
  - Runs the four updaters once.
  - Subscribes to `htmx:afterSwap` and runs updaters only when the swap target is `#game` (check `evt.detail.target.id === 'game'`).
  - Removes the 100 ms fallback timeout (or keeps it only as a temporary safety net with a comment).
- In `layout.html`, remove the inline `<script>...</script>` and add `<script src="/static/js/app.js"></script>` (after the HTMX script). Call the init (e.g. `AdventureUI.init()`) on `DOMContentLoaded` or at the end of `app.js` if DOM is already ready.
- Manually verify: full page load, reroll, begin, play choices, and battle flows still update sidebars and dice.

### Phase 2: JS test setup and tests
- Add a test runner and DOM env (e.g. `npm init -y`, Jest + jsdom, or Vitest + happy-dom) in a way that doesn’t require the Go app to run (tests run in Node).
- Create `static/js/app.test.js` (or `app.spec.js`) that:
  - Builds a minimal DOM (document with `#game`, sidebars, `.stats-update`, `.enemy-update`, dice containers, etc.).
  - Requires or imports the updater functions (you may need to refactor `app.js` to export individual functions for testing).
  - Tests each updater in isolation and one integration-style test that fires `htmx:afterSwap` and asserts sidebar/dice state.
- Add a short note in README (and optionally a Make target) for running JS tests.

### Phase 3: Split layout into partials
- Add `templates/layout_head.html`: doctype, `<html>`, `<head>`, meta, title, link to CSS, HTMX script, app.js script. It can take a minimal data map (e.g. page title) if you want.
- Add `templates/sidebar_left.html`: the left `<aside>` (character placeholder, stats, player-dice-area).
- Add `templates/sidebar_right.html`: the right `<aside>` (enemy panels, enemy-dice-area).
- Refactor `layout.html` to: `{{template "layout_head.html" .}}` (or no dot), `<body>`, `<main>`, `<div class="container">`, `{{template "sidebar_left.html" .}}`, `<section id="game">` with start/game template, `{{template "sidebar_right.html" .}}`, close container/main/body/html.
- Ensure `.State` / `.Start` are still passed where needed for `#stat-*` and any other template logic in sidebars.

### Phase 4 (optional): Move to OOB swaps
- Change handlers so that, for `/play` (and optionally `/begin`/`/reroll`), the response includes:
  - The current `#game` fragment.
  - An OOB fragment for the left sidebar (e.g. `<div id="sidebar-stats" hx-swap-oob="true">...</div>` with stats and optionally player dice).
  - An OOB fragment for the right sidebar (e.g. `<div id="enemy-sidebar-content" hx-swap-oob="true">...</div>` with enemy panels and enemy dice).
- Reduce or remove the JS that “syncs” from `#game` data elements to the sidebars; the server becomes the single source of truth for what the sidebars show.
- Keep only the dice **animation** (and any other purely visual behaviour) in JS, and add tests for that if it’s non-trivial.

---

## File and directory sketch after refactor

```
static/
  app.css
  js/
    app.js          # init, updaters, setDieFace, htmx:afterSwap listener
    app.test.js     # unit tests (Jest/Vitest + jsdom)

templates/
  layout.html       # thin shell: head partial, body, container, sidebar partials, #game, close
  layout_head.html  # <head> content: meta, CSS, HTMX, app.js
  sidebar_left.html
  sidebar_right.html
  game.html
  start.html
```

(If you prefer a single `app.js` at `static/app.js` and tests next to it, e.g. `static/app.test.js`, that’s fine; the important part is “no inline script, tests in repo.”)

---

## HTMX rules to follow going forward

1. **Server returns HTML**  
   Prefer returning fragments (and OOB fragments) that directly represent the new state of each region; avoid returning only “data” and painting the UI in JS.

2. **Minimal JS**  
   Use JS for: behaviour that can’t be expressed in HTML (e.g. dice roll animation), and narrow, event-driven glue (e.g. “on `htmx:afterSwap` for `#game`, run X”) until OOB can replace it.

3. **Events**  
   Use HTMX events (`htmx:afterSwap`, etc.) with a clear target (e.g. only when `#game` is swapped) instead of global side effects and timeouts.

4. **No inline script in layout**  
   Keep all script in external files so it can be cached, tested, and linted.

5. **Test the UI logic**  
   Any JS that interprets server output (e.g. data attributes) and updates the DOM should have unit tests so we can refactor (including toward OOB) safely.

---

## Summary

| Item                         | Action                                                                 |
|------------------------------|------------------------------------------------------------------------|
| HTMX usage                   | Reassess; plan OOB for sidebars so server drives all changing regions. |
| Inline JS in layout          | Move to `static/js/app.js` and load from layout.                      |
| HTMX listener                | Scope to `#game` swap only; remove or narrow 100 ms timeout.          |
| Layout file                  | Split into layout_head + sidebar_left + sidebar_right + thin layout.  |
| JS tests                     | Add test runner + jsdom; test each updater and init.                  |
| Optional later               | Implement OOB swaps and trim sync JS.                                 |

This plan keeps the current behaviour intact while making HTMX usage and JS maintainable and testable, then sets the stage for a more idiomatic, server-driven UI (OOB) if you choose to do it.
