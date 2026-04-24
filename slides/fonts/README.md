# Fonts

## Brand sans — shipped locally

- `SourceSans3-VariableFont_wght.ttf` — upright, weight axis 200–900
- `SourceSans3-Italic-VariableFont_wght.ttf` — italic, weight axis 200–900

Loaded by `colors_and_type.css` as `@font-face` declarations with `font-weight: 200 900` so any weight between 200 and 900 is available (e.g. `font-weight: 650`). Exposed as `var(--font-sans)` and as the `family-name 'Source Sans 3'`.

## Google Fonts (not supplied locally)

- **Caveat** — handwritten script accent (`var(--font-script)`)
- **JetBrains Mono** — monospace / vector display (`var(--font-mono)`)

Both are loaded from Google Fonts at runtime. If you want them offline too, download and add `@font-face` blocks the same way Source Sans 3 is set up:

- https://fonts.google.com/specimen/Caveat
- https://fonts.google.com/specimen/JetBrains+Mono
