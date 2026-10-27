# garageai.eu and app.garageai.eu: who does what

Decided 2026-10-27 (Magnus): the portal stays a portal, not a landing page. The two sites have
different jobs, a few agreed touchpoints, and one rule for each touchpoint: one owner, the other
links.

| | **www.garageai.eu** (this repo, `www/`) | **app.garageai.eu** (the portal, Lovable, `garageai-portal`) |
|---|---|---|
| Job | **Explain and persuade.** Why GarageAI, how it works, what is safe, what is live and what is not | **Do.** Sign up, keys, credits, chat, connect a computer, My garages, Admin |
| Content | Narrative, Bridge, security, pricing explained, FAQ, roadmap, legal texts, landing pages per search intent (`/bridge`, `/api`, `/my-machine`) | Live data: the model catalogue, garages with grades and availability, prices, usage, earnings; the API reference |
| Changes when | The product or the narrative changes | Every heartbeat, price change and new garage |
| Built for | Search engines and AI answer engines: static HTML, titles per page, JSON-LD, sitemap, `llms.txt` | People who are signed in, or about to be: a React app |
| Audience language | English (the site is in English) | English and Swedish |

## Touchpoints, and who owns each

| Touchpoint | Owner | The other side |
|---|---|---|
| **Model catalogue** (models, prices, garages, grades) | Portal (`/models`, `/garages`) | The site links to it ("see the live models"); it never copies model lists or prices |
| **API reference** (base URL, auth, parameters, model IDs) | Portal (`/api`, Documentation) | The site's `/api` is the landing page for the search intent: why, a quickstart, what is not in place yet. It links to the portal's reference, and its examples use the same base URL and `garageai/auto` |
| **Narrative and positioning** (sharing first, an open innovation project, earnings as credits, payouts later) | Site | The portal's own short texts (page title, meta description, sign-up intro) use the site's wording, not their own |
| **Terms and privacy** | Site (`www/src/legal/`, `/terms`, `/privacy`) | The portal links to them and records the accepted version (`TERMS_VERSION`, now 2026-10-25) |
| **Security for garage owners** | Site (`/security`, `AUDIT.md` in this repo) | The portal links to it from Connect your GPU |
| **Brand** (mark, favicon, name) | Site (`www/public/brand/`, `docs/brand.md`) | The portal uses the same files |
| **Getting started** (the connect lines) | Portal (`/connect`, Windows and macOS/Linux tabs) | The site shows the same two lines and sends people to `/connect` |
| **Community** (Discord, GitHub Discussions) | Site links to it | The portal links to it from the footer or help |
| **Support** (account, payments) | Portal (the support chat to Magnus) | The site links there only for account questions |
| **Live status and figures** (garages online, tokens) | Portal or Ops, through a public endpoint | The site may show them only through that endpoint, never by copying numbers |

## Rules of thumb

- **If it changes on its own, it lives in the portal.** If it explains, it lives on the site.
- **Never two sources for the same fact.** When the site needs a fact the portal owns (a price, a model, a version), it links to it or reads it from a public endpoint at build time.
- **The site is the search surface.** The portal is a single-page app: its pages share one title and are empty without JavaScript, so AI answer engines see nothing there. Anything that should be found through search or quoted by an AI assistant needs a page on the site.
- **One wording.** When the narrative changes on the site, the portal's short texts follow in the same week.

## Open items (2026-10-27)

- Portal meta description still reads "earn money with your own GPU": change to the site's wording, for example "Share your GPU and your models, try everyone else's: open models from garages across Europe through one OpenAI-compatible API."
- Portal page titles: one per page (Models, Garages, Documentation) instead of the same title everywhere.
- **Pricing is not a menu item on either site** (decided 2026-10-27). What drives this audience is which models exist, whether they work and whether it is safe; price is a detail per model. Prices show per model in the portal's catalogue, and how pricing works is explained at `https://www.garageai.eu/#pricing`, linked from the catalogue. The section stays on the site, reachable by search and links, not from the nav.
- A public, signed-out endpoint with the catalogue (and later network figures), so the site can build a static, crawlable model list at build time without copying data.
