# Web functional and visual verification

Browser UI changes require three distinct layers unless the test specification records why one is inapplicable.

## 1. Functional browser tests

Use the project's existing browser runner; prefer Playwright when it is already configured or its adoption was accepted. Cover critical journeys, real user-visible assertions, keyboard/focus behavior, and relevant loading, empty, error, and responsive states. Avoid tests that only click without asserting an outcome.

## 2. Automated visual regression

For stable changed states, use Playwright `expect(page).toHaveScreenshot()` or locator screenshots. Stabilize viewport, browser, device scale, color scheme, locale, timezone, fonts, animation, data, time, and network responses. Select browser projects and viewport matrix according to supported-product risk, not every permutation.

Treat baseline images as reviewed product artifacts. Update them only for an intentional accepted design change. Capture actual, expected, and diff paths plus trace/log evidence on failure.

## 3. Human/AI visual design review

Open captured screenshots and inspect layout hierarchy, spacing, typography, contrast, alignment, clipping/overflow, content density, responsive behavior, consistency with the design system, and loading/empty/error states. A pixel diff cannot decide whether a new design is good.

Record reviewer, revision, viewport/state set, findings, and approval result. Include accessibility assertions or ARIA snapshots where useful; they complement rather than replace pixels and visual judgment.

## Failure evidence

Retain screenshot/diff paths and hashes when supported. Capture Playwright trace for failed or flaky journeys, and diagnose nondeterminism before accepting a new baseline.
