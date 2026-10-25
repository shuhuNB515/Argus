# Report pages design QA

- Source visual truth: `/var/folders/qy/hg1xxyx5445fhc5bgm5l0rhc0000gn/T/codex-clipboard-d0b4dae8-133c-4439-abfa-13a802e79719.png`
- SRC implementation: `/Users/shu/Documents/src-agent-argus-harness-v2/report-src-final.png`
- CTF implementation: `/Users/shu/Documents/src-agent-argus-harness-v2/report-ctf-final.png`
- Browser: Codex in-app browser
- Viewport/CSS size: 1440 x 1000, device scale factor 1
- Source pixels: 2366 x 1591; implementation pixels: 1440 x 1000. Compared at matching desktop composition rather than raw pixel scale.
- State: light theme, empty real-data state.

## Full-view comparison

The implementation follows the reference screen's hierarchy: title area, five summary cards, one-line search and filters, and a bordered report table. Existing Argus navigation, typography, colors, radii and spacing tokens are preserved. SRC uses severity summaries; CTF uses challenge directions.

## Focused-region comparison

The dense control and table region was checked separately. Search, category/severity selection, time sorting, selection checkboxes, report viewing and CTF merge controls are represented with the existing component library. The empty table state stays inside the table frame and does not introduce sample data.

## Required fidelity surfaces

- Fonts and typography: existing Argus font stack, weights and hierarchy retained; no visible wrapping or truncation regression at the tested viewport.
- Spacing and layout rhythm: summary cards use equal tracks; controls and table align to the same content width; vertical spacing follows the reference.
- Colors and visual tokens: existing foreground, muted, border and semantic badge tokens retained.
- Image quality and asset fidelity: the reference contains no raster imagery; all visible icons use the project's existing Lucide icon set.
- Copy and content: report-specific labels replace vulnerability-list copy while keeping the reference information structure.

## Comparison history

- Initial pass: layout matched, but the browser console exposed a stale `nav.group.3` translation-key error from the sidebar search index.
- Fix: synchronized the sidebar search group mapping with the visible navigation groups.
- Final pass: both SRC and CTF pages render at 1440 x 1000 with no new console errors. Existing earlier console entries were timestamped before the fix.

## Primary interactions tested

- SRC and CTF routes load.
- Report-center return links are visible.
- Search and filter controls are enabled.
- SRC sorting switches between time order and severity order; equal severities fall back to newest first.
- CTF merge control is disabled when fewer than two reports are selected.
- Empty state uses zero real records.

No actionable P0, P1 or P2 differences remain. P3: exact card height and column widths may naturally vary with the host window and sidebar width.

final result: passed
