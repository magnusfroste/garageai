# GarageAI brand: logo and favicon

There is one mark, kept in one place. Every GarageAI surface uses the same files:
www.garageai.eu, app.garageai.eu and ops.garageai.eu. If the mark changes, change it
here and copy the new files to the other surfaces. Nobody draws a variant of their own.

## The mark

A roof and a dot on a rounded tile. The roof is the garage; the dot is the AI node.

| File | Use |
|---|---|
| [`public/brand/mark.svg`](../public/brand/mark.svg) | The mark: favicon, app header, anything square |
| [`public/favicon.svg`](../public/favicon.svg) | A copy of `mark.svg` at the path browsers look for |
| [`public/brand/apple-touch-icon.png`](../public/brand/apple-touch-icon.png) | 180×180, square corners (iOS rounds them itself) |
| [`public/brand/mark-512.png`](../public/brand/mark-512.png) | 512×512, for structured data, app manifests and social profiles |
| [`public/brand/logo-on-dark.svg`](../public/brand/logo-on-dark.svg) | Mark plus "GarageAI" in white, for dark backgrounds |
| [`public/brand/logo-on-light.svg`](../public/brand/logo-on-light.svg) | Mark plus "GarageAI" in near-black, for light backgrounds |

Published at `https://www.garageai.eu/brand/<file>`, so the portal and the Operations Center can
copy them from there.

## Rules

- **Name:** "GarageAI", written as one word with a capital G and AI. In running text "Garage AI" is
  also fine; never "GARAGE AI" or "Garageai".
- **Colours:** the tile is a gradient from `#00d4ff` (cyan) to `#00c37a` (green), top left to
  bottom right. The roof and the dot are `#06121a`.
- **Sizes:** the mark reads down to 16 px. Below 24 px use the mark alone, without the word.
- **In HTML,** put the mark next to the name as text (as the site header does) rather than using the
  logo SVGs, so the name uses the page's font.

## Where it is used

- **www.garageai.eu (this repo):** `public/favicon.svg`, the header (`HeaderReact.jsx`) and the
  Organization logo in `Layout.astro`.
- **app.garageai.eu (Lovable):** favicon and the header logo should be these files.
- **ops.garageai.eu (`infra/gateway/ops/index.html`):** the favicon and the logo in the sidebar
  should be this mark.
