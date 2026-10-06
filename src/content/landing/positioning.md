---
section: "positioning"
order: 3
title: "What Makes GarageAI Different"
subtitle: "Local GPUs and verified providers, one API, no crypto, no open ports"
description: "Decentralised GPU networks exist, and so do API aggregators. None of them, as far as we know, combines what GarageAI does. Here is how it compares, by category."
claim: "The first European marketplace where local GPUs and verified providers sell real-time inference in one OpenAI-compatible API, without crypto and without open ports."
columns:
  - "GarageAI"
  - "Crypto inference networks"
  - "GPU rental marketplaces"
  - "API aggregators"
rows:
  - feature: "Real-time inference from residential GPUs"
    values: ["yes", "yes", "partial", "no"]
    note: "Rental marketplaces hand you a whole machine by the hour; you build the inference yourself."
  - feature: "Named garages you can pick, plus a shared pool"
    values: ["yes", "no", "yes", "no"]
    note: "In token-based networks the node is anonymous: you never know where the model runs."
  - feature: "Local GPUs and datacenter providers in the same catalogue"
    values: ["yes", "no", "no", "no"]
    note: "Aggregators only resell cloud providers; GPU networks only have nodes."
  - feature: "Pay in euros by card, invoice for companies"
    values: ["yes", "no", "yes", "yes"]
    note: "Crypto networks pay out and charge in tokens or wallets."
  - feature: "EU gateway, self-hosted mesh, no US tunnel service in the path"
    values: ["yes", "no", "no", "partial"]
    note: "Where the request is processed is a first-class property, not a side effect."
  - feature: "Operators open no inbound ports"
    values: ["yes", "partial", "no", "n/a"]
    note: "The garage opens one outgoing WireGuard tunnel; only the gateway can reach the runtime port."
  - feature: "Onboarding from a web wizard in minutes"
    values: ["yes", "partial", "no", "n/a"]
    note: "Create a garage, run one command, pass the acceptance test, go live."
legend:
  yes: "Yes"
  partial: "Partly"
  no: "No"
  n/a: "Not applicable"
footnote: "Categories, not named products, because the field moves fast. If you know a service that already does all of this in Europe, tell us: powerup@garageai.eu."
---

Content for this section is defined in the frontmatter above.
