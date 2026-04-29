---
version: alpha
name: Conductor
description: A calm, paper-textured task manager for households. Light gray chrome, hairline borders, a single forest-green accent for action and progress.
colors:
  primary: "#1c1a17"
  secondary: "#78746e"
  tertiary: "#3f6b5d"
  tertiary-hover: "#2f5a4d"
  tertiary-container: "#ecf2ef"
  neutral: "#fafafa"
  surface: "#ffffff"
  surface-sunken: "#f5f5f4"
  outline: "#e7e5e4"
  outline-hair: "#efede9"
  on-surface-muted: "#a8a29a"
  warning: "#8a6b20"
  warning-container: "#fbf3e4"
  warning-outline: "#ecddb8"
  danger: "#a13f3f"
  success: "#2f5a4d"
  success-container: "#eaf1ee"
  success-outline: "#cfe0d7"
typography:
  display:
    fontFamily: Inter
    fontSize: 28px
    fontWeight: 400
    lineHeight: 1.2
    letterSpacing: -0.4px
  h1:
    fontFamily: Inter
    fontSize: 26px
    fontWeight: 400
    lineHeight: 1.25
    letterSpacing: -0.3px
  h2:
    fontFamily: Inter
    fontSize: 18px
    fontWeight: 500
    lineHeight: 1.3
  title:
    fontFamily: Inter
    fontSize: 15px
    fontWeight: 500
    lineHeight: 1.4
  body-md:
    fontFamily: Inter
    fontSize: 14px
    fontWeight: 400
    lineHeight: 1.45
  body-sm:
    fontFamily: Inter
    fontSize: 13px
    fontWeight: 400
    lineHeight: 1.5
  meta:
    fontFamily: Inter
    fontSize: 12px
    fontWeight: 400
    lineHeight: 1.45
  caption:
    fontFamily: Inter
    fontSize: 11px
    fontWeight: 500
    lineHeight: 1.4
  label-caps:
    fontFamily: Inter
    fontSize: 12px
    fontWeight: 500
    lineHeight: 1.4
    letterSpacing: 0.5px
    textTransform: uppercase
  numeric:
    fontFamily: Inter
    fontSize: 11px
    fontWeight: 400
    fontFeature: "tnum"
rounded:
  xs: 3px
  sm: 4px
  md: 6px
  lg: 8px
  pill: 999px
spacing:
  xs: 4px
  sm: 6px
  md: 10px
  lg: 14px
  xl: 20px
  2xl: 28px
  3xl: 56px
components:
  app-shell:
    backgroundColor: "{colors.neutral}"
    textColor: "{colors.primary}"
    typography: "{typography.body-md}"
  sidebar:
    backgroundColor: "{colors.surface-sunken}"
    width: 200px
    padding: 16px 10px
  nav-item:
    textColor: "{colors.secondary}"
    typography: "{typography.body-sm}"
    rounded: "{rounded.sm}"
    padding: 6px 8px
  nav-item-active:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.primary}"
  button-primary:
    backgroundColor: "{colors.tertiary}"
    textColor: "#ffffff"
    rounded: "{rounded.md}"
    padding: 7px 12px
    typography: "{typography.body-sm}"
  button-primary-hover:
    backgroundColor: "{colors.tertiary-hover}"
  button-secondary:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.primary}"
    rounded: "{rounded.md}"
    padding: 7px 12px
  button-ghost:
    backgroundColor: "transparent"
    textColor: "{colors.secondary}"
  button-danger:
    backgroundColor: "{colors.danger}"
    textColor: "#ffffff"
    rounded: "{rounded.md}"
  card:
    backgroundColor: "{colors.surface}"
    rounded: "{rounded.lg}"
  card-header:
    backgroundColor: "{colors.surface-sunken}"
    padding: 14px 16px
  pill-neutral:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.secondary}"
    rounded: "{rounded.pill}"
    padding: 1px 8px
    typography: "{typography.caption}"
  pill-progress:
    backgroundColor: "{colors.warning-container}"
    textColor: "{colors.warning}"
    rounded: "{rounded.pill}"
  pill-done:
    backgroundColor: "{colors.success-container}"
    textColor: "{colors.success}"
    rounded: "{rounded.pill}"
  filter-chip:
    backgroundColor: "{colors.tertiary-container}"
    textColor: "{colors.tertiary-hover}"
    rounded: "{rounded.sm}"
    padding: 2px 8px
    typography: "{typography.caption}"
  input:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.primary}"
    rounded: "{rounded.sm}"
    padding: 7px 10px
    typography: "{typography.body-sm}"
  side-panel:
    backgroundColor: "{colors.surface}"
    width: 360px
  task-row:
    padding: 8px 16px 8px 42px
    typography: "{typography.body-sm}"
  task-row-selected:
    backgroundColor: "{colors.tertiary-container}"
  modal:
    backgroundColor: "{colors.surface}"
    rounded: "{rounded.lg}"
    padding: 22px
    width: 340px
  modal-scrim:
    backgroundColor: "rgba(28,26,23,0.4)"
  callout-warning:
    backgroundColor: "{colors.warning-container}"
    textColor: "{colors.warning}"
    rounded: "{rounded.sm}"
    padding: 10px 12px
---

## Overview

Conductor is a quiet, two-person household task manager. The interface should feel like a well-kept paper notebook: lots of warm white, hairline rules, restrained typography, and a single forest-green accent that does the work of every call-to-action and every "this is progressing" signal.

Visually it sits between **journal** and **utility software** — closer to a writing app than a productivity dashboard. Density is moderate: enough information per row to be glanceable for a household coordinating in real time, but never crowded. Color is used sparingly; structure is carried by hairlines and rhythm, not fills or shadows.

The design avoids enterprise tropes: no heavy headers, no avatar clusters, no badge swarms, no gradient accents, no novelty illustrations. When in doubt, remove a border and add whitespace.

## Colors

The palette is built around a warm-neutral surface stack, a single accent for action, and a small set of muted state colors. Saturation is intentionally low.

- **Primary `#1c1a17` — Ink.** The darkest value, used for headlines, body text on light backgrounds, and the dark scrim on modals. Never used as a fill behind text.
- **Secondary `#78746e` — Stone.** Metadata, captions, inactive labels, secondary icon strokes. Pairs with primary on neutral backgrounds.
- **Tertiary `#3f6b5d` — Forest.** The single accent. Used for primary buttons, progress bars (when work is happening), checked checkboxes, and links inside panels. Never decorative — every appearance signals action or progress.
- **Tertiary container `#ecf2ef`.** The wash behind active filter chips and selected task rows. The lightest possible echo of forest.
- **Neutral `#fafafa` — Page.** The app background. Slightly warm so it does not read as gray-blue.
- **Surface `#ffffff` — Card.** Project containers, panels, modals, inputs.
- **Surface sunken `#f5f5f4` — Tray.** The sidebar, project header rows, the nav rest state. One step below the page in elevation feel, even though it is technically lighter than `surface` in some contexts — the sunken feel comes from the hairline boundary.
- **Outline `#e7e5e4` — Edge.** Card borders, sidebar boundary, input borders. The strongest divider in the system, and still very quiet.
- **Outline hair `#efede9` — Rule.** Internal dividers between sections inside a card; column underlines. Should disappear at arm's length.
- **On-surface muted `#a8a29a` — Whisper.** Disabled icon strokes, count numerals beside nav items, "no data" copy.
- **Warning `#8a6b20` on `#fbf3e4`.** "In progress" pills, the new-category confirmation callout, the project-blocked notice. Mustard, never orange.
- **Success `#2f5a4d` on `#eaf1ee`.** "Done" pills only. A deeper, less saturated forest than the accent.
- **Danger `#a13f3f` — Brick.** Overdue dates, destructive button fills, delete actions.

People avatars are generated per-person at `oklch(0.78 0.055 hue)` for the fill and `oklch(0.28 0.04 hue)` for the initial — currently hue 14 (Maren, terracotta) and hue 210 (Jonas, slate). New household members should get hues from the same chroma/lightness band so the set stays cohesive.

## Typography

One family — **Inter** — at conservative weights (400 for body and headlines, 500 for titles, buttons, and caps labels). The system never goes heavier than 600. Display sizes use a mild negative tracking; everything else sits at default tracking.

- **Display 28 / 400.** Page titles ("All projects", "My tasks").
- **H1 26 / 400.** Single-project header.
- **H2 18 / 500.** Task title in the side panel.
- **Title 15 / 500.** Side-panel header, modal headline, mobile app bar.
- **Body 14 / 400.** Default text size, including nav labels at 13.
- **Body-sm 13 / 400.** Task row titles, button labels, form input text.
- **Meta 12 / 400.** Breadcrumbs, helper copy, secondary metadata.
- **Caption 11 / 500.** Pills, chips, due dates, count numerals.
- **Label-caps 12 / 500 / +0.5 tracking / uppercase.** Section headers ("To do", "In progress", "Done", "Overdue", "This week").
- **Numeric.** Any time digits stand alone (counts beside nav items, percentages, dates), apply `font-feature-settings: "tnum"` so columns line up.

Headlines use lowercase or sentence case; never title case. The wordmark is `conductor` lowercase.

## Layout

The app is a **two-pane shell**: a 200px sidebar on the left, a flexible content area on the right. Content uses generous outer padding (32px top, 56px sides on desktop) so the eye has room to land on the page title before scanning rows.

The base rhythm is **8px**, with most internal padding landing on multiples of 2 (so 6, 10, 14 are valid). Card-to-card gaps in the project list are 12px; section-to-section spacing is 28px.

The side panel for task detail is fixed at 360px, anchored to the right edge of the content area, and never covers the sidebar. Modals are centered on a `rgba(28,26,23,0.4)` scrim and capped at 340px wide for confirmations, larger only when the body content demands it.

## Elevation & Depth

Elevation is almost entirely conveyed by hairline borders and surface color, not shadows. Two exceptions:

- **Side panel:** `box-shadow: -8px 0 24px rgba(0,0,0,0.04)` — barely perceptible, just enough to lift the panel off the underlying list.
- **Modal:** `box-shadow: 0 12px 40px rgba(0,0,0,0.18)` — the only "real" shadow in the system. Reserved for the rare, attention-demanding case.
- **Mobile FAB:** `box-shadow: 0 8px 24px rgba(63,107,93,0.35)` — tinted to the accent.

Cards never get default shadows. Hover states change background fill, not elevation.

## Shapes

A small radius scale used consistently:

- **3–4px** — checkboxes, filter chips, inputs, nav items.
- **6px** — buttons, pills inside dense rows, project header inputs.
- **8px** — project cards, modals, side-panel-adjacent surfaces.
- **999px (pill)** — status and metadata pills only.

Iconography is **Feather-style**: 1.6px stroke, 24×24 viewBox, rounded caps and joins, currentColor stroke. Icons inside dense rows render at 11–14px; nav and button icons at 13–14px. Filled icons are reserved for the urgent priority flag.

## Components

Each component below resolves to a token entry above. Variants (hover, active, selected) are separate component entries with related names.

- **app-shell** — root flex container, sidebar + content.
- **sidebar** — sunken background, 200px wide, contains nav + household footer.
- **nav-item / nav-item-active** — icon + label + count row. Active state lifts to `surface` with a 1px outline; inactive sits flush in `surface-sunken`.
- **button-primary / button-primary-hover** — forest fill, white text. Used for "Create project", "Add task", "Save changes".
- **button-secondary** — white fill, outline border. The default button.
- **button-ghost** — transparent, muted text. Used for tertiary actions ("Edit", "Delete").
- **button-danger** — brick fill, white text. Used only for destructive confirmations.
- **card** — project containers, the toolbar above columns.
- **card-header** — sunken row inside a project card showing project name, counts, progress.
- **pill-neutral / pill-progress / pill-done** — tonal pills for category, status, counts. Always 11px caption type, always pill-shaped.
- **filter-chip** — active filter indicator. Subtle forest wash, with an `×` to remove.
- **input** — text inputs, selects, comboboxes, and date fields all share one visual.
- **side-panel** — 360px right-anchored drawer with header / scrollable body / footer action bar.
- **task-row / task-row-selected** — checkbox + priority flag + title + meta + assignee. Selected row gets `tertiary-container` wash.
- **modal / modal-scrim** — centered confirmation, ink scrim at 40% opacity.
- **callout-warning** — mustard inline notice for new-value confirmation, blocked deletes, and other inline warnings.

## Do's and Don'ts

**Do**
- Reach for the forest accent only when something is actionable or progressing. If you cannot answer "what action does this signify?", use stone or ink instead.
- Use hairline rules (`outline-hair`) to separate items inside a card; use `outline` only at card edges.
- Use tabular numerals for any column of numbers, dates, or percentages.
- Keep page titles lowercase and in the lightest available weight that still reads as a title (400 at display sizes).
- Let icons be quiet — 1.6px stroke, currentColor, never filled unless a fill carries meaning (urgent flag).

**Don't**
- Don't add gradients, glows, or colored shadows (the FAB shadow is the single permitted exception).
- Don't introduce a second accent color. If a new state needs to stand out, choose between mustard (warn) and brick (danger) — both already in the palette.
- Don't bold body copy to add hierarchy. Use size, color, or whitespace.
- Don't use emoji. Avatars carry person identity; icons carry action; nothing else needs glyphs.
- Don't add card shadows. If two surfaces need separation, use a hairline or a 1-step background shift.
- Don't crowd the sidebar with secondary nav. Three top-level items is the cap; everything else lives in the content area.
