# HTMX and JS Refactor Plan

This document captures the current conventions for HTMX and JavaScript usage in
the Adventure game UI.

## Goals

- Keep rendering server-driven by returning HTML fragments.
- Use HTMX swaps to update discrete regions (especially sidebar content).
- Limit JavaScript to behavior that cannot be expressed in HTML.

## HTMX patterns

- Prefer fragments and OOB swaps (`hx-swap-oob="true"`) when updating multiple
  regions.
- Avoid JSON responses that require client-side rendering.
- Use the existing sidebar partials (`templates/sidebar_left_oob.html`,
  `templates/sidebar_right_oob.html`) for OOB updates.

## JavaScript usage

- Keep JS minimal and focused on behaviors like dice animation.
- Scope HTMX event listeners to `#game` (e.g. check
  `evt.detail.target.id === "game"`).
- Avoid global side effects or extra timeouts.
- Do not use inline scripts; keep all JS in `static/js/` and load from layout
  templates.

## Testing

- If you change `static/js/app.js`, update `static/js/app.test.js` and run
  `make test-js` or `npm test`.
